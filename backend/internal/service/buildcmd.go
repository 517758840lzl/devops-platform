package service

import (
	"fmt"
	"path/filepath"
	"strings"

	"devops-platform/internal/model"
)

func DefaultArtifactDir(project model.Project) string {
	code := strings.TrimSpace(project.Code)
	if code == "" {
		code = fmt.Sprintf("project_%d", project.ID)
	}
	return filepath.Join("data", "artifacts", code)
}

func ResolveBuildCommand(p model.Project) string {
	if p.BuildProfile == "custom" {
		cmd := strings.TrimSpace(p.BuildCommand)
		if cmd != "" {
			return cmd
		}
	}
	platform := p.BuildPlatform
	if platform == "" {
		platform = "android"
	}
	mode := "release"
	if p.BuildProfile == "develop" {
		mode = "debug"
	}
	base := "flutter pub get && "
	switch platform {
	case "ios":
		return base + fmt.Sprintf("flutter build ios --%s --no-codesign", mode)
	default:
		return base + fmt.Sprintf("flutter build apk --%s", mode)
	}
}

func SanitizeArtifactOutputDir(dir string) string {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return ""
	}
	clean := filepath.Clean(dir)
	if filepath.IsAbs(clean) {
		return ""
	}
	prefix := filepath.Join("data", "artifacts")
	if clean != prefix && !strings.HasPrefix(clean, prefix+string(filepath.Separator)) {
		return ""
	}
	return clean
}

func ResolveArtifactBaseDir(p model.Project) string {
	dir := SanitizeArtifactOutputDir(p.ArtifactOutputDir)
	if dir == "" {
		return DefaultArtifactDir(p)
	}
	return dir
}

func BuildArtifactDir(p model.Project, job model.BuildJob) string {
	return filepath.Join(ResolveArtifactBaseDir(p), JobDirName(job))
}

// JobDirName 与网页「#构建号」一致，避免磁盘 build_12 对应页面 #8 的歧义。
func JobDirName(job model.BuildJob) string {
	n := job.BuildNumber
	if n == 0 {
		n = job.ID
	}
	return fmt.Sprintf("build_%d", n)
}

// JobWorkspaceDir 源码/构建工作区路径：workspaces/project_{id}/build_{构建号}
func JobWorkspaceDir(workspaceRoot string, job model.BuildJob) string {
	return filepath.Join(workspaceRoot, fmt.Sprintf("project_%d", job.ProjectID), JobDirName(job))
}

func InferBuildMode(p model.Project, originalName string) string {
	name := strings.ToLower(originalName)
	if strings.Contains(name, "debug") {
		return "debug"
	}
	if strings.Contains(name, "release") {
		return "release"
	}
	cmd := strings.ToLower(p.BuildCommand)
	if strings.Contains(cmd, "--debug") {
		return "debug"
	}
	if strings.Contains(cmd, "--release") {
		return "release"
	}
	if p.BuildProfile == "develop" {
		return "debug"
	}
	if p.BuildProfile == "release" {
		return "release"
	}
	return "release"
}

func SanitizeDownloadName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	var b strings.Builder
	lastUnderscore := false
	for _, r := range name {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastUnderscore = false
		case r >= 0x4e00 && r <= 0x9fff:
			b.WriteRune(r)
			lastUnderscore = false
		default:
			if !lastUnderscore {
				b.WriteByte('_')
				lastUnderscore = true
			}
		}
	}
	return strings.Trim(b.String(), "_")
}

// ArtifactDownloadFilename: {项目名}_{debug|release}_{yyyyMMdd_HHmmss}.ext
func ArtifactDownloadFilename(p model.Project, job model.BuildJob, originalName string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	if ext == "" {
		ext = ".apk"
	}
	base := SanitizeDownloadName(p.Name)
	if base == "" {
		base = SanitizeDownloadName(p.Code)
	}
	if base == "" {
		base = fmt.Sprintf("project_%d", p.ID)
	}
	mode := InferBuildMode(p, originalName)
	stamp := job.CreatedAt
	if job.FinishedAt != nil {
		stamp = *job.FinishedAt
	}
	return fmt.Sprintf("%s_%s_%s%s", base, mode, stamp.Format("20060102_150405"), ext)
}
