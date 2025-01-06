package types

import (
	"time"
)

// Project 表示一个项目及其包含的所有仓库信息
type Project struct {
	Name         string                 `json:"name"`         // 项目名称
	Repositories map[string]*Repository `json:"repositories"` // 项目下的仓库集合，按仓库名称映射
}

// Repository 表示一个仓库及其包含的 Artifact 信息
type Repository struct {
	Name     string      `json:"name"`      // 仓库名称
	Artifact []*Artifact `json:"artifacts"` // 仓库下的 Artifact 列表
}

// Artifact 表示一个 Artifact 及其相关信息
type Artifact struct {
	Tags []TagInfo `json:"tags"` // Artifact 下的标签信息
}

// TagInfo 表示单个标签的相关信息
type TagInfo struct {
	Name     string    `json:"name"`      // 标签名称
	PushTime time.Time `json:"push_time"` // 标签的推送时间
}

// SystemInfo 表示系统信息
type SystemInfo struct {
	Version string `json:"harbor_version"`
}

// NewProject 创建新的 Project
func NewProject(name string) *Project {
	return &Project{
		Name:         name,
		Repositories: make(map[string]*Repository),
	}
}

// AddRepository 为 Project 添加仓库
func (p *Project) AddRepository(repo *Repository) {
	if repo != nil {
		p.Repositories[repo.Name] = repo
	}
}

func (a *Artifact) AddTag(tag *TagInfo) {
	if a == nil || tag == nil {
		return
	}
	a.Tags = append(a.Tags, *tag)
}
