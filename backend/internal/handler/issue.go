package handler

import (
	"strings"
	"time"

	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"
	"devops-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type IssueHandler struct {
	db *gorm.DB
}

func NewIssueHandler(db *gorm.DB) *IssueHandler {
	return &IssueHandler{db: db}
}

func validateBugSide(issueType, bugSide string) string {
	if issueType != "bug" {
		return ""
	}
	side := strings.TrimSpace(bugSide)
	if side != "frontend" && side != "backend" {
		return "Bug 必须选择端：frontend 或 backend"
	}
	return ""
}

func (h *IssueHandler) List(c *gin.Context) {
	var items []model.Issue
	q := h.db.Order("updated_at desc")
	if pid := c.Query("project_id"); pid != "" {
		q = q.Where("project_id = ?", pid)
	}
	if iid := c.Query("iteration_id"); iid != "" {
		q = q.Where("iteration_id = ?", iid)
	}
	if t := c.Query("type"); t != "" {
		q = q.Where("type = ?", t)
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	if assignee := c.Query("assignee_id"); assignee != "" {
		q = q.Where("assignee_id = ?", assignee)
	}
	if err := q.Find(&items).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, items)
}

func (h *IssueHandler) Get(c *gin.Context) {
	var item model.Issue
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "issue not found")
		return
	}
	response.OK(c, item)
}

func (h *IssueHandler) Create(c *gin.Context) {
	var item model.Issue
	if err := c.ShouldBindJSON(&item); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, item.ProjectID, userID.(uint))
	if err != nil || !service.CanWrite(role) {
		response.Fail(c, 403, "无写入权限")
		return
	}
	if item.ReporterID == 0 {
		item.ReporterID = userID.(uint)
	}
	if item.Status == "" {
		if item.Type == "bug" {
			item.Status = "open"
		} else {
			item.Status = "todo"
		}
	}
	if item.Priority == "" {
		item.Priority = "medium"
	}
	if msg := validateBugSide(item.Type, item.BugSide); msg != "" {
		response.Fail(c, 400, msg)
		return
	}
	if err := h.db.Create(&item).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	service.LogActivity(h.db, "issue", item.ID, userID.(uint), "create", "", item.Title)
	response.OK(c, item)
}

func (h *IssueHandler) Update(c *gin.Context) {
	var item model.Issue
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "issue not found")
		return
	}
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, item.ProjectID, userID.(uint))
	if err != nil || !service.CanWrite(role) {
		response.Fail(c, 403, "无写入权限")
		return
	}
	var req model.Issue
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}
	if msg := validateBugSide(item.Type, req.BugSide); msg != "" {
		response.Fail(c, 400, msg)
		return
	}
	oldStatus := item.Status
	item.Title = req.Title
	item.Description = req.Description
	item.Status = req.Status
	item.Priority = req.Priority
	item.Severity = req.Severity
	item.AssigneeID = req.AssigneeID
	item.IterationID = req.IterationID
	item.RequirementID = req.RequirementID
	item.ReleaseID = req.ReleaseID
	item.AppVersion = strings.TrimSpace(req.AppVersion)
	item.BugSide = req.BugSide
	item.DueDate = req.DueDate
	item.StepsToReproduce = req.StepsToReproduce
	item.ExpectedResult = req.ExpectedResult
	item.ActualResult = req.ActualResult
	item.Environment = req.Environment
	if (req.Status == "done" || req.Status == "closed" || req.Status == "resolved") && item.ClosedAt == nil {
		now := time.Now()
		item.ClosedAt = &now
	}
	if err := h.db.Save(&item).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	if oldStatus != item.Status {
		service.LogActivity(h.db, "issue", item.ID, userID.(uint), "status_change", oldStatus, item.Status)
	}
	response.OK(c, item)
}

type patchStatusReq struct {
	Status string `json:"status" binding:"required"`
}

func (h *IssueHandler) PatchStatus(c *gin.Context) {
	var item model.Issue
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "issue not found")
		return
	}
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, item.ProjectID, userID.(uint))
	if err != nil || !service.CanWrite(role) {
		response.Fail(c, 403, "无写入权限")
		return
	}
	var req patchStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}
	oldStatus := item.Status
	item.Status = req.Status
	if (req.Status == "done" || req.Status == "closed" || req.Status == "resolved") && item.ClosedAt == nil {
		now := time.Now()
		item.ClosedAt = &now
	}
	if err := h.db.Save(&item).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	service.LogActivity(h.db, "issue", item.ID, userID.(uint), "status_change", oldStatus, req.Status)
	response.OK(c, item)
}
