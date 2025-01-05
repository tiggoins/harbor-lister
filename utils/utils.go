package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tiggoins/harbor-lister/config"
	"github.com/tiggoins/harbor-lister/types"
)

func fetch[T any](client *http.Client, config *config.Config, path string) (T, error) {
	var result T
	url := fmt.Sprintf("%s%s", config.HarborURL, path)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result, fmt.Errorf("创建请求失败: %w", err)
	}
	req.SetBasicAuth(config.Username, config.Password)

	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("服务器返回非预期状态码: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("解析响应数据失败: %w", err)
	}

	return result, nil
}

func FetchProjects(client *http.Client, config *config.Config) ([]types.Project, error) {
	return fetch[[]types.Project](client, config, "/projects")
}

func FetchRepositories(client *http.Client, config *config.Config, projectName string) ([]types.Repository, error) {
	path := fmt.Sprintf("/projects/%s/repositories", projectName)
	return fetch[[]types.Repository](client, config, path)
}

func FetchArtifacts(client *http.Client, config *config.Config, projectName, repoName string) ([]types.TagInfo, error) {
    var allTags []types.TagInfo
    pageSize := 10
    page := 1
    
    for {
        path := fmt.Sprintf("/projects/%s/repositories/%s/artifacts?page=%d&page_size=%d", 
            projectName, repoName, page, pageSize)
        tags, err := fetch[[]types.TagInfo](client, config, path)
        if err != nil {
            return nil, fmt.Errorf("fetching artifacts page %d: %w", page, err)
        }
        
        if len(tags) == 0 {
            break
        }
        
        allTags = append(allTags, tags...)
        if len(tags) < pageSize {
            break
        }
        page++
    }
    
    return allTags, nil
}

func CheckHarborVersion(client *http.Client, config *config.Config) error {
	sysInfo, err := fetch[types.SystemInfo](client, config, "/systeminfo")
	if err != nil {
		return fmt.Errorf("获取系统信息失败: %w", err)
	}

	version := strings.Split(sysInfo.Version, ".")
	if len(version) > 0 {
		if version[0] < "2" {
			return fmt.Errorf("Harbor 版本必须是 2.0 或以上版本，当前版本: %s", sysInfo.Version)
		}
	} else {
		return fmt.Errorf("无法解析 Harbor 版本: %s", sysInfo.Version)
	}

	return nil
}

func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.Local
	}
	return t.In(location).Format("2006-01-02 15:04:05")
}
