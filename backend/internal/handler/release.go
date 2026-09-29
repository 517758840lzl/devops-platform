package handler

import (
	"fmt"
	"strings"
	"time"

	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"
	"devops-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReleaseHandler struct {
	db *gorm.DB
}

func NewReleaseHandler(db *gorm.DB) *ReleaseHandler {
	return &ReleaseHandler{db: db}
}

type releaseDetailVO struct {
	model.Release
	Approvals []approvalVO `json:"approvals"`
}

type approvalVO struct {
	model.ReleaseApproval
	UserName string `json:"user_name"`
}

func (h *ReleaseHandler) List(c *gin.Context) {
	var items []model.Release
	q := h.db.Order("updated_at desc")
	if pid := c.Query("project_id"); pid != "" {
		q = q.Where("project_id = ?", pid)
	}
	if err := q.Find(&items).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, items)
}

func (h *ReleaseHandler) Get(c *gin.Context) {
	var item model.Release
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "release not found")
		return
	}
	var approvals []model.ReleaseApproval
	h.db.Where("release_id = ?", item.ID).Order("created_at asc").Find(&approvals)
	approvalList := make([]approvalVO, 0, len(approvals))
	for _, a := range approvals {
		var user model.User
		h.db.First(&user, a.UserID)
		approvalList = append(approvalList, approvalVO{ReleaseApproval: a, UserName: user.Name})
	}
	response.OK(c, releaseDetailVO{Release: item, Approvals: approvalList})
}

func (h *ReleaseHandler) Create(c *gin.Context) {
	var item model.Release
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
	if item.PublisherID == nil {
		uid := userID.(uint)
		item.PublisherID = &uid
	}
	if item.Status == "" {
		item.Status = "draft"
	}
	if err := h.db.Create(&item).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	service.LogActivity(h.db, "release", item.ID, userID.(uint), "create", "", item.Version)
	response.OK(c, item)
}

func (h *ReleaseHandler) Update(c *gin.Context) {
	var item model.Release
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "release not found")
		return
	}
	if item.Status != "draft" && item.Status != "rejected" {
		response.Fail(c, 400, "仅草稿或已驳回状态可编辑")
		return
	}
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, item.ProjectID, userID.(uint))
	if err != nil || !service.CanWrite(role) {
		response.Fail(c, 403, "无写入权限")
		return
	}
	var req model.Release
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}
	item.Title = req.Title
	item.Version = req.Version
	item.ReleaseNotes = req.ReleaseNotes
	item.RollbackPlan = req.RollbackPlan
	item.PlannedAt = req.PlannedAt
	item.IterationID = req.IterationID
	if err := h.db.Save(&item).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, item)
}

type approvalActionReq struct {
	Comment string `json:"comment"`
	Reason  string `json:"reason"`
}

func (h *ReleaseHandler) recordApproval(releaseID, userID uint, action, comment string) {
	_ = h.db.Create(&model.ReleaseApproval{
		ReleaseID: releaseID,
		UserID:    userID,
		Action:    action,
		Comment:   comment,
	}).Error
}

func (h *ReleaseHandler) Submit(c *gin.Context) {
	var item model.Release
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "release not found")
		return
	}
	if item.Status != "draft" && item.Status != "rejected" {
		response.Fail(c, 400, "当前状态不可提交审批")
		return
	}
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, item.ProjectID, userID.(uint))
	if err != nil || !service.CanWrite(role) {
		response.Fail(c, 403, "无写入权限")
		return
	}
	var req approvalActionReq
	_ = c.ShouldBindJSON(&req)
	now := time.Now()
	item.Status = "pending_approval"
	item.SubmittedAt = &now
	item.RejectReason = ""
	if err := h.db.Save(&item).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	h.recordApproval(item.ID, userID.(uint), "submit", req.Comment)
	service.LogActivity(h.db, "release", item.ID, userID.(uint), "submit_approval", "draft", "pending_approval")
	service.PublishNotification(h.db, model.Notification{
		ProjectID: item.ProjectID,
		Type:      "release",
		Title:     "发布待审批",
		Message:   strings.TrimSpace(item.Version + " " + item.Title),
		Link:      "/releases",
		EventKey:  fmt.Sprintf("release:%d:pending", item.ID),
		Audience:  service.AudienceApprovers,
	})
	response.OK(c, item)
}

func (h *ReleaseHandler) Approve(c *gin.Context) {
	var item model.Release
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "release not found")
		return
	}
	if item.Status != "pending_approval" {
		response.Fail(c, 400, "当前状态不可审批")
		return
	}
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, item.ProjectID, userID.(uint))
	if err != nil || !service.CanApproveRelease(role) {
		response.Fail(c, 403, "无审批权限")
		return
	}
	var req approvalActionReq
	_ = c.ShouldBindJSON(&req)
	now := time.Now()
	uid := userID.(uint)
	item.Status = "approved"
	item.ApproverID = &uid
	item.ApprovedAt = &now
	if err := h.db.Save(&item).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	h.recordApproval(item.ID, uid, "approve", req.Comment)
	service.LogActivity(h.db, "release", item.ID, uid, "approve", "pending_approval", "approved")
	response.OK(c, item)
}

func (h *ReleaseHandler) Reject(c *gin.Context) {
	var item model.Release
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "release not found")
		return
	}
	if item.Status != "pending_approval" {
		response.Fail(c, 400, "当前状态不可驳回")
		return
	}
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, item.ProjectID, userID.(uint))
	if err != nil || !service.CanApproveRelease(role) {
		response.Fail(c, 403, "无审批权限")
		return
	}
	var req approvalActionReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Reason == "" {
		response.Fail(c, 400, "请填写驳回原因")
		return
	}
	uid := userID.(uint)
	item.Status = "rejected"
	item.RejectReason = req.Reason
	item.ApproverID = &uid
	if err := h.db.Save(&item).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	h.recordApproval(item.ID, uid, "reject", req.Reason)
	service.LogActivity(h.db, "release", item.ID, uid, "reject", "pending_approval", "rejected")
	response.OK(c, item)
}

func (h *ReleaseHandler) Publish(c *gin.Context) {
	var item model.Release
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "release not found")
		return
	}
	if item.Status != "approved" {
		response.Fail(c, 400, "需先审批通过才能发布")
		return
	}
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, item.ProjectID, userID.(uint))
	if err != nil || !service.CanApproveRelease(role) {
		response.Fail(c, 403, "无发布权限")
		return
	}
	var req approvalActionReq
	_ = c.ShouldBindJSON(&req)
	now := time.Now()
	item.Status = "published"
	item.PublishedAt = &now
	if err := h.db.Save(&item).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	h.recordApproval(item.ID, userID.(uint), "publish", req.Comment)
	service.LogActivity(h.db, "release", item.ID, userID.(uint), "publish", "approved", "published")
	response.OK(c, item)
}

func (h *ReleaseHandler) Approvals(c *gin.Context) {
	var approvals []model.ReleaseApproval
	h.db.Where("release_id = ?", c.Param("id")).Order("created_at asc").Find(&approvals)
	result := make([]approvalVO, 0, len(approvals))
	for _, a := range approvals {
		var user model.User
		h.db.First(&user, a.UserID)
		result = append(result, approvalVO{ReleaseApproval: a, UserName: user.Name})
	}
	response.OK(c, result)
}
