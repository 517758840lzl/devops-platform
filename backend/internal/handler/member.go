package handler

import (
	"strconv"

	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"
	"devops-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MemberHandler struct {
	db *gorm.DB
}

func NewMemberHandler(db *gorm.DB) *MemberHandler {
	return &MemberHandler{db: db}
}

type memberVO struct {
	model.ProjectMember
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

func (h *MemberHandler) List(c *gin.Context) {
	projectID := c.Param("id")
	var members []model.ProjectMember
	if err := h.db.Where("project_id = ?", projectID).Find(&members).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	result := make([]memberVO, 0, len(members))
	for _, m := range members {
		var user model.User
		h.db.First(&user, m.UserID)
		result = append(result, memberVO{
			ProjectMember: m,
			Username:      user.Username,
			Name:          user.Name,
			Email:         user.Email,
		})
	}
	response.OK(c, result)
}

type addMemberReq struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

func (h *MemberHandler) Add(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, projectID, userID.(uint))
	if err != nil || !service.CanManageMembers(role) {
		response.Fail(c, 403, "无权限管理成员")
		return
	}
	var req addMemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}
	validRoles := map[string]bool{"owner": true, "super_admin": true, "admin": true, "developer": true, "tester": true, "viewer": true}
	if !validRoles[req.Role] {
		response.Fail(c, 400, "invalid role")
		return
	}
	var exists model.ProjectMember
	if err := h.db.Where("project_id = ? AND user_id = ?", projectID, req.UserID).First(&exists).Error; err == nil {
		response.Fail(c, 409, "成员已存在")
		return
	}
	m := model.ProjectMember{ProjectID: projectID, UserID: req.UserID, Role: req.Role}
	if err := h.db.Create(&m).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	service.LogActivity(h.db, "project", projectID, userID.(uint), "add_member", "", req.Role)
	var user model.User
	h.db.First(&user, req.UserID)
	response.OK(c, memberVO{ProjectMember: m, Username: user.Username, Name: user.Name, Email: user.Email})
}

type updateMemberReq struct {
	Role string `json:"role" binding:"required"`
}

func (h *MemberHandler) Update(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	memberID := c.Param("memberId")
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, projectID, userID.(uint))
	if err != nil || !service.CanManageMembers(role) {
		response.Fail(c, 403, "无权限管理成员")
		return
	}
	var m model.ProjectMember
	if err := h.db.Where("id = ? AND project_id = ?", memberID, projectID).First(&m).Error; err != nil {
		response.Fail(c, 404, "member not found")
		return
	}
	var req updateMemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}
	validRoles := map[string]bool{"owner": true, "super_admin": true, "admin": true, "developer": true, "tester": true, "viewer": true}
	if !validRoles[req.Role] {
		response.Fail(c, 400, "invalid role")
		return
	}
	if m.Role == "owner" && req.Role != "owner" {
		response.Fail(c, 400, "不能变更项目负责人角色")
		return
	}
	oldRole := m.Role
	m.Role = req.Role
	if err := h.db.Save(&m).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	service.LogActivity(h.db, "project", projectID, userID.(uint), "update_member_role", oldRole, req.Role)
	response.OK(c, m)
}

func (h *MemberHandler) Remove(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	memberID := c.Param("memberId")
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, projectID, userID.(uint))
	if err != nil || !service.CanManageMembers(role) {
		response.Fail(c, 403, "无权限管理成员")
		return
	}
	var m model.ProjectMember
	if err := h.db.Where("id = ? AND project_id = ?", memberID, projectID).First(&m).Error; err != nil {
		response.Fail(c, 404, "member not found")
		return
	}
	if m.Role == "owner" {
		response.Fail(c, 400, "不能移除项目负责人")
		return
	}
	h.db.Delete(&m)
	service.LogActivity(h.db, "project", projectID, userID.(uint), "remove_member", m.Role, "")
	response.OK(c, nil)
}

func parseUint(s string) uint {
	n, _ := strconv.ParseUint(s, 10, 64)
	return uint(n)
}
