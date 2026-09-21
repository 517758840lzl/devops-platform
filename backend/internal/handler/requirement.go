package handler

import (
	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RequirementHandler struct {
	db *gorm.DB
}

func NewRequirementHandler(db *gorm.DB) *RequirementHandler {
	return &RequirementHandler{db: db}
}

func (h *RequirementHandler) List(c *gin.Context) {
	var items []model.Requirement
	q := h.db.Order("updated_at desc")
	if pid := c.Query("project_id"); pid != "" {
		q = q.Where("project_id = ?", pid)
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Find(&items).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, items)
}

func (h *RequirementHandler) Get(c *gin.Context) {
	var item model.Requirement
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "requirement not found")
		return
	}
	response.OK(c, item)
}

func (h *RequirementHandler) Create(c *gin.Context) {
	var item model.Requirement
	if err := c.ShouldBindJSON(&item); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}
	userID, _ := c.Get("user_id")
	if item.ProposerID == 0 {
		item.ProposerID = userID.(uint)
	}
	if item.Status == "" {
		item.Status = "draft"
	}
	if item.Priority == "" {
		item.Priority = "medium"
	}
	if err := h.db.Create(&item).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, item)
}

func (h *RequirementHandler) Update(c *gin.Context) {
	var item model.Requirement
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "requirement not found")
		return
	}
	var req model.Requirement
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}
	item.Title = req.Title
	item.Description = req.Description
	item.Source = req.Source
	item.Priority = req.Priority
	item.Status = req.Status
	item.AssigneeID = req.AssigneeID
	if err := h.db.Save(&item).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, item)
}
