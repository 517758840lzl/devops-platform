package handler

import (
	"strconv"
	"strings"
	"time"

	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"
	"devops-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ActivityHandler struct {
	db *gorm.DB
}

func NewActivityHandler(db *gorm.DB) *ActivityHandler {
	return &ActivityHandler{db: db}
}

type activityVO struct {
	model.ActivityLog
	UserName string `json:"user_name"`
}

func (h *ActivityHandler) List(c *gin.Context) {
	targetType := c.Query("target_type")
	targetID := c.Query("target_id")
	projectID := c.Query("project_id")
	userID := c.Query("user_id")
	action := strings.TrimSpace(c.Query("action"))
	actionGroup := strings.TrimSpace(c.Query("action_group"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	from := strings.TrimSpace(c.Query("from"))
	to := strings.TrimSpace(c.Query("to"))

	q := h.db.Model(&model.ActivityLog{}).Order("created_at desc")

	switch {
	case targetType != "" && targetID != "":
		q = q.Where("target_type = ? AND target_id = ?", targetType, targetID)
	case projectID != "":
		pid, _ := strconv.ParseUint(projectID, 10, 64)
		if pid == 0 {
			response.Fail(c, 400, "invalid project_id")
			return
		}
		uid, _ := c.Get("user_id")
		role, err := service.GetProjectRole(h.db, uint(pid), uid.(uint))
		if err != nil || !service.CanManageMembers(role) {
			response.Fail(c, 403, "仅管理员及以上可查看项目操作日志")
			return
		}
		scopeType := strings.TrimSpace(c.Query("scope_type"))
		var issueIDs []uint
		h.db.Model(&model.Issue{}).Where("project_id = ?", pid).Pluck("id", &issueIDs)
		var releaseIDs []uint
		h.db.Model(&model.Release{}).Where("project_id = ?", pid).Pluck("id", &releaseIDs)

		parts := []string{}
		args := []interface{}{}
		addProject := scopeType == "" || scopeType == "project"
		addIssue := scopeType == "" || scopeType == "issue"
		addRelease := scopeType == "" || scopeType == "release"

		if addProject {
			parts = append(parts, "project_id = ?", "(target_type = ? AND target_id = ?)")
			args = append(args, uint(pid), "project", uint(pid))
		}
		if addIssue && len(issueIDs) > 0 {
			parts = append(parts, "(target_type = ? AND target_id IN ?)")
			args = append(args, "issue", issueIDs)
		} else if addIssue && scopeType == "issue" {
			q = q.Where("1 = 0")
		}
		if addRelease && len(releaseIDs) > 0 {
			parts = append(parts, "(target_type = ? AND target_id IN ?)")
			args = append(args, "release", releaseIDs)
		} else if addRelease && scopeType == "release" {
			q = q.Where("1 = 0")
		}
		if len(parts) > 0 {
			q = q.Where(strings.Join(parts, " OR "), args...)
		}
		if scopeType == "project" {
			q = q.Where("target_type = ?", "project")
		}
	default:
		response.Fail(c, 400, "project_id 或 target_type+target_id 必填")
		return
	}

	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if actions := actionGroupActions(actionGroup); len(actions) > 0 {
		q = q.Where("action IN ?", actions)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("action LIKE ? OR old_value LIKE ? OR new_value LIKE ?", like, like, like)
	}
	if t := parseDayStart(from); t != nil {
		q = q.Where("created_at >= ?", *t)
	}
	if t := parseDayEnd(to); t != nil {
		q = q.Where("created_at <= ?", *t)
	}

	limit := 300
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}

	var logs []model.ActivityLog
	if err := q.Limit(limit).Find(&logs).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	userCache := map[uint]string{}
	result := make([]activityVO, 0, len(logs))
	for _, log := range logs {
		name, ok := userCache[log.UserID]
		if !ok {
			var user model.User
			h.db.Select("id, name, username").First(&user, log.UserID)
			name = user.Name
			if name == "" {
				name = user.Username
			}
			userCache[log.UserID] = name
		}
		result = append(result, activityVO{ActivityLog: log, UserName: name})
	}
	response.OK(c, result)
}

func actionGroupActions(group string) []string {
	switch group {
	case "build":
		return []string{"trigger_build", "build_success", "build_failed", "delete_build_artifacts"}
	case "config":
		return []string{"update_settings", "update_config_file"}
	case "member":
		return []string{"add_member", "update_member_role", "remove_member"}
	case "release":
		return []string{"create", "submit_approval", "approve", "reject", "publish"}
	case "bug":
		return []string{"create", "status_change", "comment", "upload_attachment"}
	case "asset":
		return []string{"upload_store_asset", "delete_store_asset"}
	case "progress":
		return []string{"update_progress", "apply_progress_hints"}
	default:
		return nil
	}
}

func parseDayStart(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return nil
	}
	return &t
}

func parseDayEnd(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return nil
	}
	end := t.Add(24*time.Hour - time.Nanosecond)
	return &end
}
