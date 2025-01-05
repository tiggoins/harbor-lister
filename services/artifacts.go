package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tiggoins/harbor-lister/types"
)

func fetchArtifacts(client *http.Client, projectName, repoName string) ([]types.Artifact, error) {
	var allArtifacts []types.Artifact
	page := 1
	pageSize := 10 // 每页获取的工件数量

	for {
		encodedRepoName := url.QueryEscape(repoName)
		path := fmt.Sprintf("/projects/%s/repositories/%s/artifacts?page=%d&page_size=%d",
			projectName, encodedRepoName, page, pageSize)
		var artifacts []types.Artifact
		resp, err := client.Get(path)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&artifacts); err != nil {
			return nil, err
		}

		allArtifacts = append(allArtifacts, artifacts...)

		if len(artifacts) < pageSize {
			break
		}
		page++
	}

	return allArtifacts, nil
}
