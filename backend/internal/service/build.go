package service

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"devops-platform/internal/config"
	"devops-platform/internal/model"

	"gorm.io/gorm"
)

func RunBuildJob(db *gorm.DB, cfg config.Config, jobID uint) {
	var job model.BuildJob
	if err := db.First(&job, jobID).Error; err != nil {
		return
	}
	var project model.Project
	if err := db.First(&project, job.ProjectID).Error; err != nil {
		failJob(db, &job, "project not found")
		return
	}

	now := time.Now()
	job.Status = "running"
	job.StartedAt = &now
	job.Log = "开始构建...\n"
	db.Save(&job)
	PublishBuildEvent(job)

	var logBuf bytes.Buffer
	appendLog := func(s string) {
		logBuf.WriteString(s)
		if !strings.HasSuffix(s, "\n") {
			logBuf.WriteString("\n")
		}
		job.Log = logBuf.String()
		db.Model(&job).Update("log", job.Log)
	}

	workRoot := filepath.Join(cfg.WorkspaceDir, fmt.Sprintf("project_%d", job.ProjectID), fmt.Sprintf("build_%d", job.ID))
	_ = os.RemoveAll(workRoot)
	if err := os.MkdirAll(workRoot, 0o755); err != nil {
		failJob(db, &job, "create workspace failed: "+err.Error())
		return
	}

	targetCommit := strings.TrimSpace(job.CommitSHA)
	if targetCommit != "" {
		appendLog(fmt.Sprintf(">>> 按 Commit 构建: %s", targetCommit))
		appendLog(fmt.Sprintf(">>> git init + fetch %s", job.GitURL))
		if out, err := exec.Command("git", "init", workRoot).CombinedOutput(); err != nil {
			appendLog(string(out))
			failJob(db, &job, "git init failed: "+err.Error())
			return
		}
		if out, err := exec.Command("git", "-C", workRoot, "remote", "add", "origin", job.GitURL).CombinedOutput(); err != nil {
			appendLog(string(out))
			failJob(db, &job, "git remote add failed: "+err.Error())
			return
		}
		fetch := exec.Command("git", "-C", workRoot, "fetch", "--depth", "1", "origin", targetCommit)
		fetch.Env = os.Environ()
		out, err := fetch.CombinedOutput()
		appendLog(string(out))
		if err != nil {
			appendLog("浅层 fetch 失败，尝试完整 fetch…")
			fetch = exec.Command("git", "-C", workRoot, "fetch", "origin", targetCommit)
			fetch.Env = os.Environ()
			out, err = fetch.CombinedOutput()
			appendLog(string(out))
			if err != nil {
				failJob(db, &job, "git fetch commit failed: "+err.Error())
				return
			}
		}
		checkout := exec.Command("git", "-C", workRoot, "checkout", "--force", "FETCH_HEAD")
		out, err = checkout.CombinedOutput()
		appendLog(string(out))
		if err != nil {
			failJob(db, &job, "git checkout commit failed: "+err.Error())
			return
		}
	} else {
		appendLog(fmt.Sprintf(">>> git clone -b %s %s", job.Branch, job.GitURL))
		clone := exec.Command("git", "clone", "--depth", "1", "-b", job.Branch, job.GitURL, workRoot)
		clone.Env = os.Environ()
		out, err := clone.CombinedOutput()
		appendLog(string(out))
		if err != nil {
			failJob(db, &job, "git clone failed: "+err.Error())
			return
		}
	}

	revCmd := exec.Command("git", "-C", workRoot, "rev-parse", "HEAD")
	revOut, err := revCmd.Output()
	if err == nil {
		job.CommitSHA = strings.TrimSpace(string(revOut))
		db.Model(&job).Update("commit_sha", job.CommitSHA)
		appendLog("commit: " + job.CommitSHA)
	} else if targetCommit != "" {
		appendLog("warning: 无法解析 HEAD，保留请求的 commit: " + targetCommit)
	}

	buildDir := workRoot
	workDir := project.BuildWorkDir
	if job.BuildWorkDir != "" {
		workDir = job.BuildWorkDir
	}
	if workDir != "" {
		buildDir = filepath.Join(workRoot, workDir)
	}
	if job.RepoName != "" {
		appendLog(fmt.Sprintf(">>> 仓库: %s", job.RepoName))
	}

	cmdStr := ResolveBuildCommand(project)
	appendLog(fmt.Sprintf(">>> 构建模式: %s / %s", project.BuildProfile, project.BuildPlatform))
	appendLog(">>> " + cmdStr)

	build := exec.Command("sh", "-c", cmdStr)
	build.Dir = buildDir
	build.Env = os.Environ()
	out, err := build.CombinedOutput()
	appendLog(string(out))
	if err != nil {
		failJob(db, &job, "build failed: "+err.Error())
		return
	}

	outputDir := BuildArtifactDir(project, job.ID)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		failJob(db, &job, "create artifact dir failed: "+err.Error())
		return
	}
	copied := copyBuildArtifacts(buildDir, outputDir, project.BuildPlatform, appendLog)

	finished := time.Now()
	job.Status = "success"
	job.ArtifactPath = outputDir
	job.FinishedAt = &finished
	if job.ShareToken == "" {
		job.ShareToken = NewShareToken()
	}
	if copied > 0 {
		job.Log = logBuf.String() + fmt.Sprintf("\n已复制 %d 个产物到: %s\n构建成功 ✓", copied, outputDir)
	} else {
		job.Log = logBuf.String() + fmt.Sprintf("\n产物目录: %s（未找到标准 APK/IPA 路径，请检查构建输出）\n构建成功 ✓", outputDir)
	}
	db.Save(&job)
	PublishBuildEvent(job)

	if job.ReleaseID != nil {
		LogActivity(db, "release", *job.ReleaseID, job.TriggeredBy, "build_success", "", job.CommitSHA)
	}
	LogActivity(db, "project", job.ProjectID, job.TriggeredBy, "build_success", job.Branch, job.CommitSHA)
	NotifyBuildResult(db, project, job, true)
}

func copyBuildArtifacts(buildDir, outputDir, platform string, appendLog func(string)) int {
	candidates := artifactCandidates(buildDir, platform)
	count := 0
	for _, src := range candidates {
		if _, err := os.Stat(src); err != nil {
			continue
		}
		dst := filepath.Join(outputDir, filepath.Base(src))
		if err := copyPath(src, dst); err != nil {
			appendLog("copy failed: " + err.Error())
			continue
		}
		appendLog("产物: " + dst)
		count++
	}
	return count
}

func artifactCandidates(buildDir, platform string) []string {
	if platform == "ios" {
		return []string{
			filepath.Join(buildDir, "build/ios/iphoneos/Runner.app"),
			filepath.Join(buildDir, "build/ios/archive/Runner.xcarchive"),
		}
	}
	return []string{
		filepath.Join(buildDir, "build/app/outputs/flutter-apk/app-release.apk"),
		filepath.Join(buildDir, "build/app/outputs/flutter-apk/app-debug.apk"),
		filepath.Join(buildDir, "build/app/outputs/apk/release/app-release.apk"),
		filepath.Join(buildDir, "build/app/outputs/apk/debug/app-debug.apk"),
	}
}

var artifactExtensions = map[string]bool{
	".apk": true,
	".aab": true,
	".ipa": true,
}

type ArtifactFile struct {
	Name string
	Path string
	Rel  string
	Size int64
}

func IsArtifactFile(name string) bool {
	return artifactExtensions[strings.ToLower(filepath.Ext(name))]
}

func CollectArtifactFiles(artifactPath string) ([]ArtifactFile, error) {
	info, err := os.Stat(artifactPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if IsArtifactFile(info.Name()) {
			return []ArtifactFile{{
				Name: info.Name(),
				Path: artifactPath,
				Rel:  info.Name(),
				Size: info.Size(),
			}}, nil
		}
		return nil, fmt.Errorf("not an artifact file")
	}
	entries, err := os.ReadDir(artifactPath)
	if err != nil {
		return nil, err
	}
	var files []ArtifactFile
	for _, entry := range entries {
		if entry.IsDir() || !IsArtifactFile(entry.Name()) {
			continue
		}
		fi, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, ArtifactFile{
			Name: entry.Name(),
			Path: filepath.Join(artifactPath, entry.Name()),
			Rel:  entry.Name(),
			Size: fi.Size(),
		})
	}
	return files, nil
}

func ResolveBuildArtifacts(db *gorm.DB, job model.BuildJob) ([]ArtifactFile, error) {
	candidateDirs := []string{}
	if strings.TrimSpace(job.ArtifactPath) != "" {
		candidateDirs = append(candidateDirs, job.ArtifactPath)
	}
	var project model.Project
	if err := db.First(&project, job.ProjectID).Error; err == nil {
		candidateDirs = append(candidateDirs, BuildArtifactDir(project, job.ID))
	}
	workspaceRoot := filepath.Join("data", "workspaces", fmt.Sprintf("project_%d", job.ProjectID), fmt.Sprintf("build_%d", job.ID))
	candidateDirs = append(candidateDirs,
		filepath.Join(workspaceRoot, "build/app/outputs/flutter-apk"),
		filepath.Join(workspaceRoot, "build/app/outputs/apk/release"),
		filepath.Join(workspaceRoot, "build/app/outputs/apk/debug"),
	)

	seen := map[string]bool{}
	var merged []ArtifactFile
	for _, dir := range candidateDirs {
		files, err := CollectArtifactFiles(dir)
		if err != nil || len(files) == 0 {
			continue
		}
		for _, f := range files {
			if seen[f.Name] {
				continue
			}
			seen[f.Name] = true
			merged = append(merged, f)
		}
	}
	if len(merged) == 0 {
		return nil, fmt.Errorf("no artifacts")
	}
	return merged, nil
}

func copyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return exec.Command("cp", "-R", src, dst).Run()
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func failJob(db *gorm.DB, job *model.BuildJob, msg string) {
	finished := time.Now()
	job.Status = "failed"
	job.Log += "\n" + msg
	job.FinishedAt = &finished
	db.Save(job)
	PublishBuildEvent(*job)
	if job.ReleaseID != nil {
		LogActivity(db, "release", *job.ReleaseID, job.TriggeredBy, "build_failed", "", msg)
	}
	LogActivity(db, "project", job.ProjectID, job.TriggeredBy, "build_failed", job.Branch, msg)
	var project model.Project
	if db.First(&project, job.ProjectID).Error == nil {
		NotifyBuildResult(db, project, *job, false)
	}
}

func ValidateGitURL(url string) error {
	u := strings.TrimSpace(url)
	if u == "" {
		return fmt.Errorf("git url required")
	}
	if !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "git@") && !strings.HasPrefix(u, "http://") {
		return fmt.Errorf("unsupported git url scheme")
	}
	return nil
}
