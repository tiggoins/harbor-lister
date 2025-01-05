package services

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tiggoins/harbor-lister/config"
	"github.com/tiggoins/harbor-lister/types"
	"github.com/tiggoins/harbor-lister/utils"
)

const (
	TagsKey = "tags"
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

	for _, project := range projects {
		fmt.Printf("正在处理项目: %s\n", project.Name)

		projectMap := make(types.ProjectMap)

		repositories, err := utils.FetchRepositories(h.client, h.config, project.Name)
		if err != nil {
			fmt.Printf("警告: 获取项目 '%s' 的仓库列表失败: %v\n", project.Name, err)
			continue
		}

		for _, repo := range repositories {
			repoName := strings.TrimPrefix(repo.Name, project.Name+"/")
			artifacts, err := fetchArtifacts(h.client, project.Name, repoName)
			if err != nil {
				fmt.Printf("警告: 获取仓库 '%s/%s' 的标签失败: %v\n",
					project.Name, repoName, err)
				continue
			}

			var tagInfos []types.TagInfo
			for _, artifact := range artifacts {
				for _, tag := range artifact.Tags {
					if tag.Name != "" {
						tagInfos = append(tagInfos, types.TagInfo{
							Name:     tag.Name,
							PushTime: artifact.PushTime,
						})
					}
				}
			}

			if len(tagInfos) > 0 {
				projectMap[repoName] = map[string][]types.TagInfo{
					TagsKey: tagInfos,
				}
			}
		}

		if len(projectMap) > 0 {
			if err := h.writer.WriteProject(project.Name, projectMap); err != nil {
				return fmt.Errorf("写入项目 '%s' 数据失败: %v", project.Name, err)
			}
			fmt.Printf("项目 '%s' 处理完成，包含 %d 个仓库\n",
				project.Name, len(projectMap))
		}

		projectMap = nil
	}

	if err := h.writer.Save(h.config.OutputFile); err != nil {
		return fmt.Errorf("保存Excel文件失败: %v", err)
	}

	return nil
}
