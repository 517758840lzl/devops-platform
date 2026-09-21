package handler

import (
	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"
	"devops-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	db *gorm.DB
}

func NewProjectHandler(db *gorm.DB) *ProjectHandler {
	return &ProjectHandler{db: db}
}

func (h *ProjectHandler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	var user model.User
	if err := h.db.First(&user, uid).Error; err != nil {
		response.Fail(c, 401, "unauthorized")
		return
	}

	q := h.db.Model(&model.Project{}).Order("updated_at desc")
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	// 平台 admin 看全部；其余只看自己是成员的项目
	if user.Role != "admin" {
		q = q.Where(
			"id IN (?)",
			h.db.Model(&model.ProjectMember{}).Select("project_id").Where("user_id = ?", uid),
		)
	}

	var items []model.Project
	if err := q.Find(&items).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	type projectVO struct {
		model.Project
		MyRole string `json:"my_role"`
	}
	result := make([]projectVO, 0, len(items))
	for _, p := range items {
		role := ""
		if user.Role == "admin" {
			role = "owner"
		} else {
			var m model.ProjectMember
			if h.db.Where("project_id = ? AND user_id = ?", p.ID, uid).First(&m).Error == nil {
				role = m.Role
			}
		}
		result = append(result, projectVO{Project: p, MyRole: role})
	}
	response.OK(c, result)
}

func (h *ProjectHandler) Get(c *gin.Context) {
	var item model.Project
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "project not found")
		return
	}
	response.OK(c, item)
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var item model.Project
	if err := c.ShouldBindJSON(&item); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}
	userID, _ := c.Get("user_id")
	if item.OwnerID == 0 {
		item.OwnerID = userID.(uint)
	}
	if item.Status == "" {
		item.Status = "active"
	}
	if item.GitBranch == "" {
		item.GitBranch = "main"
	}
	if item.BuildProfile == "" {
		item.BuildProfile = "develop"
	}
	if item.BuildPlatform == "" {
		item.BuildPlatform = "android"
	}
	if err := h.db.Create(&item).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	// 创建者自动加入项目成员（否则非平台 admin 在列表中看不到自己的项目）
	_ = h.db.Where("project_id = ? AND user_id = ?", item.ID, item.OwnerID).FirstOrCreate(&model.ProjectMember{
		ProjectID: item.ID,
		UserID:    item.OwnerID,
		Role:      "owner",
	}).Error
	response.OK(c, item)
}

func (h *ProjectHandler) Update(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, projectID, userID.(uint))
	if err != nil || !service.CanEditProjectConfig(role) {
		response.Fail(c, 403, "仅开发者及以上可修改项目设置")
		return
	}

	var item model.Project
	if err := h.db.First(&item, projectID).Error; err != nil {
		response.Fail(c, 404, "project not found")
		return
	}
	var req model.Project
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}
	item.Name = req.Name
	item.Description = req.Description
	if req.Status != "" {
		item.Status = req.Status
	} else if item.Status == "" {
		item.Status = "active"
	}
	item.GitURL = req.GitURL
	if req.GitBranch == "" {
		req.GitBranch = "main"
	}
	item.GitBranch = req.GitBranch
	item.GitRepos = req.GitRepos
	item.BuildProfile = req.BuildProfile
	item.BuildPlatform = req.BuildPlatform
	item.BuildCommand = req.BuildCommand
	item.BuildWorkDir = req.BuildWorkDir
	item.ArtifactOutputDir = service.SanitizeArtifactOutputDir(req.ArtifactOutputDir)
	if item.BuildProfile == "" {
		item.BuildProfile = "develop"
	}
	if item.BuildPlatform == "" {
		item.BuildPlatform = "android"
	}
	if err := h.db.Save(&item).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, item)
}
