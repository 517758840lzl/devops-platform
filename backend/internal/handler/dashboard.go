package handler

import (
	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"
	"devops-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

func (h *DashboardHandler) Summary(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	projectIDs, err := accessibleProjectIDs(h.db, uid)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	// 可选：再收窄到当前选中项目（须在可访问范围内）
	if pid := parseUint(c.Query("project_id")); pid > 0 {
		allowed := false
		for _, id := range projectIDs {
			if id == pid {
				allowed = true
				break
			}
		}
		if allowed {
			projectIDs = []uint{pid}
		} else {
			projectIDs = nil
		}
	}

	stats := gin.H{
		"projects":     0,
		"requirements": 0,
		"tasks":        0,
		"bugs":         0,
		"releases":     0,
		"my_tasks":     0,
		"open_bugs":    0,
	}
	recentIssues := []model.Issue{}

	if len(projectIDs) == 0 {
		response.OK(c, gin.H{"stats": stats, "recent_issues": recentIssues})
		return
	}

	var projectCount, reqCount, taskCount, bugCount, releaseCount int64
	var myTasks, openBugs int64

	h.db.Model(&model.Project{}).Where("id IN ? AND status = ?", projectIDs, "active").Count(&projectCount)
	h.db.Model(&model.Requirement{}).Where("project_id IN ?", projectIDs).Count(&reqCount)
	h.db.Model(&model.Issue{}).Where("project_id IN ? AND type = ?", projectIDs, "task").Count(&taskCount)
	h.db.Model(&model.Issue{}).Where("project_id IN ? AND type = ?", projectIDs, "bug").Count(&bugCount)
	h.db.Model(&model.Release{}).Where("project_id IN ?", projectIDs).Count(&releaseCount)

	h.db.Model(&model.Issue{}).
		Where("project_id IN ? AND type = ? AND assignee_id = ? AND status NOT IN ?",
			projectIDs, "task", uid, []string{"done", "closed"}).
		Count(&myTasks)
	h.db.Model(&model.Issue{}).
		Where("project_id IN ? AND type = ? AND status NOT IN ?",
			projectIDs, "bug", []string{"resolved", "closed"}).
		Count(&openBugs)

	h.db.Where("project_id IN ?", projectIDs).Order("updated_at desc").Limit(10).Find(&recentIssues)

	response.OK(c, gin.H{
		"stats": gin.H{
			"projects":     projectCount,
			"requirements": reqCount,
			"tasks":        taskCount,
			"bugs":         bugCount,
			"releases":     releaseCount,
			"my_tasks":     myTasks,
			"open_bugs":    openBugs,
		},
		"recent_issues": recentIssues,
	})
}

func accessibleProjectIDs(db *gorm.DB, userID uint) ([]uint, error) {
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	var ids []uint
	if user.Role == "admin" {
		if err := db.Model(&model.Project{}).Pluck("id", &ids).Error; err != nil {
			return nil, err
		}
		return ids, nil
	}
	if err := db.Model(&model.ProjectMember{}).Where("user_id = ?", userID).Pluck("project_id", &ids).Error; err != nil {
		return nil, err
	}
	_ = service.ErrNotMember // keep import used if needed — actually remove unused
	return ids, nil
}
