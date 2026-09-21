package handler

import (
	"fmt"
	"strings"

	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"
	"devops-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProgressHandler struct {
	db *gorm.DB
}

func NewProgressHandler(db *gorm.DB) *ProgressHandler {
	return &ProgressHandler{db: db}
}

func (h *ProgressHandler) Get(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, projectID, userID.(uint))
	if err != nil {
		// 无成员权限：静默返回空进度，避免前端 toast；列表侧已过滤不可见项目
		response.OK(c, &service.ProgressSummary{
			Milestones: []service.MilestoneProgressVO{},
			Hints:      []service.ProgressHint{},
			CanEdit:    false,
		})
		return
	}

	summary, err := service.LoadProjectProgress(h.db, projectID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	summary.CanEdit = service.CanEditProjectConfig(role)
	response.OK(c, summary)
}

type patchChecklistReq struct {
	Status        *string `json:"status"`
	AssigneeID    *uint   `json:"assignee_id"`
	Note          *string `json:"note"`
	ClearAssignee bool    `json:"clear_assignee"`
}

func (h *ProgressHandler) PatchChecklist(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	itemID := parseUint(c.Param("itemId"))
	userID, _ := c.Get("user_id")

	role, err := service.GetProjectRole(h.db, projectID, userID.(uint))
	if err != nil || !service.CanEditProjectConfig(role) {
		response.Fail(c, 403, "仅 developer 及以上可改进度")
		return
	}

	var item model.ProjectChecklistItem
	if err := h.db.First(&item, itemID).Error; err != nil {
		response.Fail(c, 404, "检查项不存在")
		return
	}
	var ms model.ProjectMilestone
	if err := h.db.First(&ms, item.MilestoneID).Error; err != nil || ms.ProjectID != projectID {
		response.Fail(c, 404, "检查项不属于该项目")
		return
	}

	var req patchChecklistReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}

	updates := map[string]interface{}{}
	oldParts := []string{}
	newParts := []string{}

	if req.Status != nil {
		st := strings.TrimSpace(*req.Status)
		if st != "todo" && st != "doing" && st != "done" && st != "blocked" {
			response.Fail(c, 400, "status 须为 todo/doing/done/blocked")
			return
		}
		if item.Status != st {
			oldParts = append(oldParts, "status="+item.Status)
			newParts = append(newParts, "status="+st)
			updates["status"] = st
		}
	}
	if req.ClearAssignee {
		if item.AssigneeID != nil {
			oldParts = append(oldParts, fmt.Sprintf("assignee=%d", *item.AssigneeID))
			newParts = append(newParts, "assignee=")
			updates["assignee_id"] = nil
		}
	} else if req.AssigneeID != nil {
		aid := *req.AssigneeID
		if aid == 0 {
			if item.AssigneeID != nil {
				oldParts = append(oldParts, fmt.Sprintf("assignee=%d", *item.AssigneeID))
				newParts = append(newParts, "assignee=")
				updates["assignee_id"] = nil
			}
		} else {
			var prev uint
			if item.AssigneeID != nil {
				prev = *item.AssigneeID
			}
			if prev != aid {
				oldParts = append(oldParts, fmt.Sprintf("assignee=%d", prev))
				newParts = append(newParts, fmt.Sprintf("assignee=%d", aid))
				updates["assignee_id"] = aid
			}
		}
	}
	if req.Note != nil {
		note := strings.TrimSpace(*req.Note)
		if len(note) > 512 {
			response.Fail(c, 400, "备注过长")
			return
		}
		if item.Note != note {
			oldParts = append(oldParts, "note")
			newParts = append(newParts, "note")
			updates["note"] = note
		}
	}

	if len(updates) == 0 {
		_ = h.db.Preload("Assignee").First(&item, item.ID)
		response.OK(c, item)
		return
	}

	if err := h.db.Model(&item).Updates(updates).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	oldVal := item.Title
	if len(oldParts) > 0 {
		oldVal = item.Title + " · " + strings.Join(oldParts, ",")
	}
	newVal := strings.Join(newParts, ",")
	service.LogActivity(h.db, "project", projectID, userID.(uint), "update_progress", oldVal, newVal)

	_ = h.db.Preload("Assignee").First(&item, item.ID)
	response.OK(c, item)
}

func (h *ProgressHandler) ApplyHints(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, projectID, userID.(uint))
	if err != nil || !service.CanEditProjectConfig(role) {
		response.Fail(c, 403, "仅 developer 及以上可改进度")
		return
	}
	n, err := service.ApplyProgressHints(h.db, projectID, userID.(uint))
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	if n > 0 {
		service.LogActivity(h.db, "project", projectID, userID.(uint), "apply_progress_hints", "", fmt.Sprintf("%d", n))
	}
	summary, err := service.LoadProjectProgress(h.db, projectID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	summary.CanEdit = true
	response.OK(c, gin.H{"updated": n, "progress": summary})
}
