package handler

import (
	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type IterationHandler struct {
	db *gorm.DB
}

func NewIterationHandler(db *gorm.DB) *IterationHandler {
	return &IterationHandler{db: db}
}

var kanbanColumns = []struct {
	Key   string
	Title string
}{
	{"todo", "待办"},
	{"in_progress", "进行中"},
	{"review", "待验收"},
	{"done", "已完成"},
}

func (h *IterationHandler) List(c *gin.Context) {
	var items []model.Iteration
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

func (h *IterationHandler) Get(c *gin.Context) {
	var item model.Iteration
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "iteration not found")
		return
	}
	response.OK(c, item)
}

func (h *IterationHandler) Kanban(c *gin.Context) {
	var iteration model.Iteration
	if err := h.db.First(&iteration, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "iteration not found")
		return
	}
	var issues []model.Issue
	h.db.Where("iteration_id = ? AND type = ?", iteration.ID, "task").Order("updated_at desc").Find(&issues)

	byStatus := map[string][]model.Issue{}
	for _, col := range kanbanColumns {
		byStatus[col.Key] = []model.Issue{}
	}
	for _, issue := range issues {
		if _, ok := byStatus[issue.Status]; ok {
			byStatus[issue.Status] = append(byStatus[issue.Status], issue)
		} else {
			byStatus["todo"] = append(byStatus["todo"], issue)
		}
	}

	columns := make([]gin.H, 0, len(kanbanColumns))
	for _, col := range kanbanColumns {
		columns = append(columns, gin.H{
			"key":    col.Key,
			"title":  col.Title,
			"issues": byStatus[col.Key],
		})
	}
	response.OK(c, gin.H{"iteration": iteration, "columns": columns})
}

func (h *IterationHandler) Create(c *gin.Context) {
	var item model.Iteration
	if err := c.ShouldBindJSON(&item); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}
	userID, _ := c.Get("user_id")
	if item.OwnerID == 0 {
		item.OwnerID = userID.(uint)
	}
	if item.Status == "" {
		item.Status = "planning"
	}
	if err := h.db.Create(&item).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, item)
}

func (h *IterationHandler) Update(c *gin.Context) {
	var item model.Iteration
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "iteration not found")
		return
	}
	var req model.Iteration
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}
	item.Name = req.Name
	item.Version = req.Version
	item.StartDate = req.StartDate
	item.TestDate = req.TestDate
	item.ReleaseDate = req.ReleaseDate
	item.Status = req.Status
	if err := h.db.Save(&item).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, item)
}
