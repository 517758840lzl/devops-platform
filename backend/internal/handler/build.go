package handler

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devops-platform/internal/config"
	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"
	"devops-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BuildHandler struct {
	db  *gorm.DB
	cfg config.Config
}

func NewBuildHandler(db *gorm.DB, cfg config.Config) *BuildHandler {
	return &BuildHandler{db: db, cfg: cfg}
}

func (h *BuildHandler) List(c *gin.Context) {
	var items []model.BuildJob
	q := h.db.Order("created_at desc")
	if pid := c.Param("id"); pid != "" {
		q = q.Where("project_id = ?", pid)
	} else if pid := c.Query("project_id"); pid != "" {
		q = q.Where("project_id = ?", pid)
	}
	if rid := c.Query("release_id"); rid != "" {
		q = q.Where("release_id = ?", rid)
	}
	if err := q.Limit(50).Find(&items).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, items)
}

func (h *BuildHandler) Get(c *gin.Context) {
	var item model.BuildJob
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "build not found")
		return
	}
	response.OK(c, item)
}

type triggerBuildReq struct {
	ReleaseID  *uint  `json:"release_id"`
	Branch     string `json:"branch"`
	GitURL     string `json:"git_url"`
	CommitSHA  string `json:"commit_sha"`
}

func (h *BuildHandler) Trigger(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	var project model.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		response.Fail(c, 404, "project not found")
		return
	}
	userID, _ := c.Get("user_id")
	if _, err := service.GetProjectRole(h.db, projectID, userID.(uint)); err != nil {
		response.Fail(c, 403, "非项目成员，无法构建")
		return
	}

	var req triggerBuildReq
	_ = c.ShouldBindJSON(&req)

	repo, err := service.PickGitRepo(project.GitURL, project.GitBranch, project.GitRepos, req.GitURL, req.Branch)
	if err != nil {
		response.Fail(c, 400, "请先在项目设置中配置 Git 仓库地址")
		return
	}

	job := model.BuildJob{
		ProjectID:    projectID,
		ReleaseID:    req.ReleaseID,
		GitURL:       repo.GitURL,
		Branch:       repo.GitBranch,
		RepoName:     repo.Name,
		BuildWorkDir: repo.BuildWorkDir,
		CommitSHA:    strings.TrimSpace(req.CommitSHA),
		Status:       "pending",
		TriggeredBy:  userID.(uint),
	}
	err = h.db.Transaction(func(tx *gorm.DB) error {
		var maxNum uint
		if err := tx.Model(&model.BuildJob{}).
			Where("project_id = ?", projectID).
			Select("COALESCE(MAX(build_number), 0)").
			Scan(&maxNum).Error; err != nil {
			return err
		}
		job.BuildNumber = maxNum + 1
		return tx.Create(&job).Error
	})
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	service.LogActivity(h.db, "project", projectID, userID.(uint), "trigger_build", repo.GitBranch, repo.GitURL)
	if req.ReleaseID != nil {
		service.LogActivity(h.db, "release", *req.ReleaseID, userID.(uint), "trigger_build", repo.GitBranch, "")
	}

	go service.RunBuildJob(h.db, h.cfg, job.ID)

	response.OK(c, job)
}

func (h *BuildHandler) DownloadArtifacts(c *gin.Context) {
	var job model.BuildJob
	if err := h.db.First(&job, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "build not found")
		return
	}
	h.writeBuildArtifacts(c, job)
}

// DownloadByShareToken 微信推送免登录下载（公开链接，持有 token 即可）
func (h *BuildHandler) DownloadByShareToken(c *gin.Context) {
	token := strings.TrimSpace(c.Param("token"))
	if token == "" {
		response.Fail(c, 400, "invalid token")
		return
	}
	var job model.BuildJob
	if err := h.db.Where("share_token = ?", token).First(&job).Error; err != nil {
		response.Fail(c, 404, "build not found")
		return
	}
	h.writeBuildArtifacts(c, job)
}

func (h *BuildHandler) writeBuildArtifacts(c *gin.Context, job model.BuildJob) {
	if job.ArtifactsDeleted {
		response.Fail(c, 410, "产物已删除")
		return
	}
	if job.Status != "success" || job.ArtifactPath == "" {
		response.Fail(c, 404, "no artifacts")
		return
	}
	entries, err := service.ResolveBuildArtifacts(h.db, job)
	if err != nil || len(entries) == 0 {
		response.Fail(c, 404, "未找到 APK/AAB/IPA 构建产物，请重新构建")
		return
	}

	var project model.Project
	_ = h.db.First(&project, job.ProjectID)

	if len(entries) == 1 {
		entry := entries[0]
		filename := service.ArtifactDownloadFilename(project, job, entry.Name)
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		c.File(entry.Path)
		return
	}

	filename := service.ArtifactDownloadFilename(project, job, "artifacts.zip")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", "application/zip")
	zw := zip.NewWriter(c.Writer)
	defer zw.Close()

	for _, entry := range entries {
		data, err := os.ReadFile(entry.Path)
		if err != nil {
			continue
		}
		innerName := service.ArtifactDownloadFilename(project, job, entry.Name)
		w, err := zw.Create(innerName)
		if err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		if _, err := w.Write(data); err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
	}
}

func (h *BuildHandler) DeleteArtifacts(c *gin.Context) {
	var job model.BuildJob
	if err := h.db.First(&job, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "build not found")
		return
	}
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, job.ProjectID, userID.(uint))
	if err != nil || !service.CanWrite(role) {
		response.Fail(c, 403, "无删除权限")
		return
	}
	if job.Status != "success" {
		response.Fail(c, 400, "仅成功构建可删除产物")
		return
	}
	if job.ArtifactsDeleted {
		response.OK(c, job)
		return
	}

	if job.ArtifactPath != "" {
		_ = os.RemoveAll(job.ArtifactPath)
	}
	workspace := filepath.Join(h.cfg.WorkspaceDir, fmt.Sprintf("project_%d", job.ProjectID), fmt.Sprintf("build_%d", job.ID))
	_ = os.RemoveAll(workspace)

	job.ArtifactsDeleted = true
	job.ArtifactPath = ""
	if err := h.db.Save(&job).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	service.LogActivity(h.db, "project", job.ProjectID, userID.(uint), "delete_build_artifacts", fmt.Sprintf("build_%d", job.ID), "")
	response.OK(c, job)
}

func (h *BuildHandler) ListArtifactFiles(c *gin.Context) {
	var job model.BuildJob
	if err := h.db.First(&job, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "build not found")
		return
	}
	if job.ArtifactsDeleted || job.ArtifactPath == "" {
		response.OK(c, []gin.H{})
		return
	}
	entries, err := service.ResolveBuildArtifacts(h.db, job)
	if err != nil {
		response.OK(c, []gin.H{})
		return
	}
	files := make([]gin.H, 0, len(entries))
	var project model.Project
	_ = h.db.First(&project, job.ProjectID)
	for _, entry := range entries {
		files = append(files, gin.H{
			"name":          entry.Name,
			"download_name": service.ArtifactDownloadFilename(project, job, entry.Name),
			"path":          entry.Rel,
			"size":          entry.Size,
		})
	}
	response.OK(c, files)
}

func (h *BuildHandler) DownloadArtifactFile(c *gin.Context) {
	var job model.BuildJob
	if err := h.db.First(&job, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "build not found")
		return
	}
	if job.ArtifactsDeleted {
		response.Fail(c, 410, "产物已删除")
		return
	}
	filePath := strings.TrimSpace(c.Query("path"))
	if filePath == "" || strings.Contains(filePath, "..") {
		response.Fail(c, 400, "invalid path")
		return
	}
	diskPath := filepath.Join(job.ArtifactPath, filepath.FromSlash(filePath))
	if !strings.HasPrefix(diskPath, filepath.Clean(job.ArtifactPath)+string(os.PathSeparator)) && diskPath != filepath.Clean(job.ArtifactPath) {
		response.Fail(c, 400, "invalid path")
		return
	}
	info, err := os.Stat(diskPath)
	if err != nil || info.IsDir() {
		response.Fail(c, 404, "file not found")
		return
	}
	var project model.Project
	_ = h.db.First(&project, job.ProjectID)
	filename := service.ArtifactDownloadFilename(project, job, filepath.Base(diskPath))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.File(diskPath)
}
