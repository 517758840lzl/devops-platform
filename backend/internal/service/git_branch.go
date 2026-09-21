package service

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

func contextWithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

func ListRemoteBranches(gitURL string) ([]string, error) {
	gitURL = strings.TrimSpace(gitURL)
	if gitURL == "" {
		return nil, fmt.Errorf("git url required")
	}

	ctx, cancel := contextWithTimeout(20 * time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--heads", gitURL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(out))
		lower := strings.ToLower(detail)
		switch {
		case strings.Contains(lower, "authentication failed"),
			strings.Contains(lower, "could not read username"),
			strings.Contains(lower, "terminal prompts disabled"),
			strings.Contains(lower, "permission denied"),
			strings.Contains(lower, "invalid credentials"),
			strings.Contains(lower, "repository not found"):
			return nil, fmt.Errorf("仓库需要登录权限（私有库）。可改用 SSH 地址，或 HTTPS 带 Token：https://<token>@github.com/owner/repo.git；也可直接手填分支名")
		case ctx.Err() != nil:
			return nil, fmt.Errorf("拉取分支超时，请检查网络")
		case detail != "":
			return nil, fmt.Errorf("拉取分支失败：%s", detail)
		default:
			return nil, fmt.Errorf("拉取分支失败，请检查仓库地址或网络")
		}
	}

	seen := map[string]struct{}{}
	var branches []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		ref := parts[1]
		const prefix = "refs/heads/"
		if !strings.HasPrefix(ref, prefix) {
			continue
		}
		name := strings.TrimPrefix(ref, prefix)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		branches = append(branches, name)
	}

	sort.Strings(branches)
	if len(branches) == 0 {
		return nil, fmt.Errorf("未找到远程分支")
	}
	return branches, nil
}
