package types

import "time"

type TagInfo struct {
	Name     string
	PushTime time.Time
}

// ProjectMap 表示项目、仓库和标签信息的层级结构
type ProjectMap map[string]map[string][]TagInfo

type Project struct {
	Name string `json:"name"`
}

type Repository struct {
	Name string `json:"name"`
}

type Artifact struct {
	Tags []struct {
		Name string `json:"name"`
	} `json:"tags"`
	PullTime time.Time `json:"pull_time"`
	PushTime time.Time `json:"push_time"`
}

type SystemInfo struct {
	Version string `json:"harbor_version"`
}

