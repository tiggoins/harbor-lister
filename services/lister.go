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
	if err := utils.CheckHarborVersion(h.client, h.config); err != nil {
		return fmt.Errorf("版本检查失败: %v", err)
	}

	projects, err := utils.FetchProjects(h.client, h.config)
	if err != nil {
		return fmt.Errorf("获取项目列表失败: %v", err)
	}

	// 创建带有最大并发数的 goroutine 池
	semaphore := make(chan struct{}, maxGoroutines) // 控制最大并发数
	var wg sync.WaitGroup

	for _, projectInfo := range projects {
		// 为每个项目启动一个 goroutine
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

				// 使用 FetchArtifacts 替换原来的 h.client.GetArtifacts
				artifacts, err := utils.FetchArtifacts(h.client, h.config, project.Name, repoName)
				if err != nil {
					fmt.Printf("警告: 获取仓库 '%s/%s' 的标签失败: %v\n",
						project.Name, repoName, err)
					continue
				}

				repository := &types.Repository{
					Name: repoName,
					Tags: make([]*types.TagInfo, 0),
				}

				// 处理标签
				for _, tag := range artifacts {
					if tag.Name != "" {
						tagInfo := &types.TagInfo{
							Name:     tag.Name,
							PushTime: tag.PushTime,
						}
						repository.AddTag(tagInfo)
					}
				}

				// 将仓库添加到项目
				if len(repository.Tags) > 0 {
					project.AddRepository(repository)
				}
			}

			// 如果项目包含仓库，则写入数据
			if len(project.Repositories) > 0 {
				if err := h.writer.WriteProject(project.Name, project); err != nil {
					fmt.Printf("写入项目 '%s' 数据失败: %v\n", project.Name, err)
					return
				}
				fmt.Printf("项目 '%s' 处理完成，包含 %d 个仓库\n",
					project.Name, len(project.Repositories))
			}

		}(projectInfo) // 传递项目信息给 goroutine

		// 控制最大并发数
		semaphore <- struct{}{}
	}

	// 等待所有 goroutine 完成
	wg.Wait()

	// 保存 Excel 文件
	if err := h.writer.Save(h.config.OutputFile); err != nil {
		return fmt.Errorf("保存Excel文件失败: %v", err)
	}

	return nil
}

// processProject 现在作为一个辅助方法来处理单个项目
func (h *HarborLister) processProject(projectName string) (*types.Project, error) {
	project := types.NewProject(projectName)

	repositories, err := utils.FetchRepositories(h.client, h.config, projectName)
	if err != nil {
		return nil, fmt.Errorf("获取仓库列表失败: %v", err)
	}

	for _, repo := range repositories {
		repoName := strings.TrimPrefix(repo.Name, projectName+"/")
		artifacts, err := utils.FetchArtifacts(h.client, h.config, projectName, repoName)
		if err != nil {
			continue
		}

		repository := &types.Repository{
			Name: repoName,
			Tags: make([]*types.TagInfo, 0),
		}

		for _, tag := range artifacts {
			if tag.Name != "" {
				tagInfo := &types.TagInfo{
					Name:     tag.Name,
					PushTime: tag.PushTime,
				}
				repository.AddTag(tagInfo)
			}
		}

		if len(repository.Tags) > 0 {
			project.AddRepository(repository)
		}
	}

	return project, nil
}
