package types

import "time"

// Project 表示一个项目及其包含的所有仓库信息
type Project struct {
    Name         string                `json:"name"`
    Repositories map[string]*Repository `json:"repositories"` // 改用 map 便于查找
}

// Repository 表示一个仓库及其包含的标签信息
type Repository struct {
    Name string     `json:"name"`
    Tags []*TagInfo `json:"tags"`
}

// TagInfo 表示标签信息
type TagInfo struct {
    Name     string    `json:"name"`
    PushTime time.Time `json:"push_time"`
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

// AddTag 为 Repository 添加标签
func (r *Repository) AddTag(tag *TagInfo) {
    if tag != nil {
        r.Tags = append(r.Tags, tag)
    }
}