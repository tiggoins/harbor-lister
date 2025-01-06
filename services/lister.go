package services

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tiggoins/harbor-lister/config"
	"github.com/tiggoins/harbor-lister/types"
	"github.com/tiggoins/harbor-lister/utils"
)

const (
	maxGoroutines = 50 // 最大并发数
)

type HarborLister struct {
	client *http.Client
	config *config.Config
	writer *ExcelWriter
}

func NewHarborLister(config *config.Config) *HarborLister {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureSSL,
		},
	}

	return &HarborLister{
		client: &http.Client{
			Transport: tr,
			Timeout:   60 * time.Second,
		},
		config: config,
		writer: NewExcelWriter(),
	}
}

func (h *HarborLister) List() error {
	// 检查 Harbor 版本
	if err := utils.CheckHarborVersion(h.client, h.config); err != nil {
		return fmt.Errorf("版本检查失败: %v", err)
	}

	// 获取项目列表
	projects, err := utils.FetchProjects(h.client, h.config)
	if err != nil {
		return fmt.Errorf("获取项目列表失败: %v", err)
	}

	// 创建带有最大并发数的 goroutine 池
	semaphore := make(chan struct{}, maxGoroutines)
	var wg sync.WaitGroup

	for _, projectInfo := range projects {
		wg.Add(1)
		go func(projectInfo types.Project) {
			defer wg.Done()

			fmt.Printf("正在处理项目: %s\n", projectInfo.Name)
			project := types.NewProject(projectInfo.Name)

			// 获取仓库列表
			repositories, err := utils.FetchRepositories(h.client, h.config, project.Name)
			if err != nil {
				fmt.Printf("警告: 获取项目 '%s' 的仓库列表失败: %v\n", project.Name, err)
				return
			}

			// 遍历每个仓库
			for _, repo := range repositories {
				repoName := strings.TrimPrefix(repo.Name, project.Name+"/")

				// 获取 Artifact 列表
				artifacts, err := utils.FetchArtifacts(h.client, h.config, project.Name, repoName)
				if err != nil {
					fmt.Printf("警告: 获取仓库 '%s/%s' 的 Artifact 列表失败: %v\n",
						project.Name, repoName, err)
					continue
				}

				// 构建 Repository
				repository := &types.Repository{
					Name:     repoName,
					Artifact: make([]*types.Artifact, 0),
				}

				// 处理每个 Artifact
				for _, artifact := range artifacts {
					newArtifact := &types.Artifact{
						Tags: make([]types.TagInfo, 0),
					}

					// 添加标签到 Artifact
					for _, tag := range artifact.Tags {
						newArtifact.Tags = append(newArtifact.Tags, types.TagInfo{
							Name:     tag.Name,
							PushTime: tag.PushTime,
						})
					}

					// 如果 Artifact 包含标签，则添加到 Repository
					if len(newArtifact.Tags) > 0 {
						repository.Artifact = append(repository.Artifact, newArtifact)
					}
				}

				// 如果 Repository 包含 Artifact，则添加到 Project
				if len(repository.Artifact) > 0 {
					project.AddRepository(repository)
				}
			}

			// 如果项目包含仓库，则写入数据
			if len(project.Repositories) > 0 {
				if err := h.writer.WriteProject(project); err != nil {
					fmt.Printf("写入项目 [%s] 数据失败: [%v]\n", project.Name, err)
					return
				}
				fmt.Printf("项目 [%s] 处理完成，包含 [%d] 个仓库\n",
					project.Name, len(project.Repositories))
			} else {
				fmt.Printf("\033[31m项目 [%s] 处理完成，包含 0 个仓库\033[0m\n",
					project.Name)
			}

			<-semaphore // 释放一个 goroutine 配额
		}(projectInfo)

		// 控制最大并发数
		semaphore <- struct{}{}
	}

	// 等待所有 goroutine 完成
	wg.Wait()

	// 保存 Excel 文件
	if err := h.writer.Save(h.config.OutputFile); err != nil {
		return fmt.Errorf("保存 Excel 文件失败: %v", err)
	}

	return nil
}
