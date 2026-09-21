package service

import (
	"encoding/json"
	"strings"
)

type GitRepoEntry struct {
	Name         string `json:"name"`
	GitURL       string `json:"git_url"`
	GitBranch    string `json:"git_branch"`
	BuildWorkDir string `json:"build_work_dir"`
}

func ParseGitRepos(raw string) []GitRepoEntry {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil
	}
	var items []GitRepoEntry
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	out := make([]GitRepoEntry, 0, len(items))
	for _, item := range items {
		item.GitURL = strings.TrimSpace(item.GitURL)
		if item.GitURL == "" {
			continue
		}
		if item.GitBranch == "" {
			item.GitBranch = "main"
		}
		out = append(out, item)
	}
	return out
}

type ResolvedGitRepo struct {
	Name         string
	GitURL       string
	GitBranch    string
	BuildWorkDir string
	IsPrimary    bool
}

func ResolveProjectGitRepos(gitURL, gitBranch, gitReposJSON string) []ResolvedGitRepo {
	repos := make([]ResolvedGitRepo, 0, 1+len(ParseGitRepos(gitReposJSON)))
	primaryURL := strings.TrimSpace(gitURL)
	if primaryURL != "" {
		branch := strings.TrimSpace(gitBranch)
		if branch == "" {
			branch = "main"
		}
		repos = append(repos, ResolvedGitRepo{
			Name:      "主仓库",
			GitURL:    primaryURL,
			GitBranch: branch,
			IsPrimary: true,
		})
	}
	for _, item := range ParseGitRepos(gitReposJSON) {
		if primaryURL != "" && item.GitURL == primaryURL {
			continue
		}
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = item.GitURL
		}
		repos = append(repos, ResolvedGitRepo{
			Name:         name,
			GitURL:       item.GitURL,
			GitBranch:    item.GitBranch,
			BuildWorkDir: item.BuildWorkDir,
		})
	}
	return repos
}

func PickGitRepo(gitURL, gitBranch, gitReposJSON, requestedURL, requestedBranch string) (ResolvedGitRepo, error) {
	repos := ResolveProjectGitRepos(gitURL, gitBranch, gitReposJSON)
	if len(repos) == 0 {
		return ResolvedGitRepo{}, ValidateGitURL("")
	}
	requestedURL = strings.TrimSpace(requestedURL)
	if requestedURL == "" {
		picked := repos[0]
		if strings.TrimSpace(requestedBranch) != "" {
			picked.GitBranch = requestedBranch
		}
		return picked, nil
	}
	for _, repo := range repos {
		if repo.GitURL == requestedURL {
			picked := repo
			if strings.TrimSpace(requestedBranch) != "" {
				picked.GitBranch = requestedBranch
			}
			return picked, nil
		}
	}
	return ResolvedGitRepo{}, ValidateGitURL(requestedURL)
}
