package handler

import (
	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"
	"devops-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommentHandler struct {
	db *gorm.DB
}

func NewCommentHandler(db *gorm.DB) *CommentHandler {
	return &CommentHandler{db: db}
}

type commentVO struct {
	model.Comment
	UserName string `json:"user_name"`
}

func (h *CommentHandler) List(c *gin.Context) {
	targetType := c.Query("target_type")
	targetID := c.Query("target_id")
	if targetType == "" || targetID == "" {
		response.Fail(c, 400, "target_type and target_id required")
		return
	}
	var comments []model.Comment
	if err := h.db.Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("created_at asc").Find(&comments).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	result := make([]commentVO, 0, len(comments))
	for _, cm := range comments {
		var user model.User
		h.db.First(&user, cm.UserID)
		result = append(result, commentVO{Comment: cm, UserName: user.Name})
	}
	response.OK(c, result)
}

type createCommentReq struct {
	TargetType string `json:"target_type" binding:"required"`
	TargetID   uint   `json:"target_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

func (h *CommentHandler) Create(c *gin.Context) {
	var req createCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}
	userID, _ := c.Get("user_id")
	cm := model.Comment{
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		UserID:     userID.(uint),
		Content:    req.Content,
	}
	if err := h.db.Create(&cm).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	service.LogActivity(h.db, req.TargetType, req.TargetID, userID.(uint), "comment", "", req.Content)
	var user model.User
	h.db.First(&user, userID)
	response.OK(c, commentVO{Comment: cm, UserName: user.Name})
}

func (h *CommentHandler) Delete(c *gin.Context) {
	var cm model.Comment
	if err := h.db.First(&cm, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "comment not found")
		return
	}
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	if cm.UserID != userID.(uint) && role.(string) != "admin" {
		response.Fail(c, 403, "forbidden")
		return
	}
	h.db.Delete(&cm)
	response.OK(c, nil)
}
