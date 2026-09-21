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
	UserName    string `json:"user_name"`
	TargetLabel string `json:"target_label"`
	ProjectName string `json:"project_name"`
	ProjectCode string `json:"project_code"`
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
	projectCache := map[uint]model.Project{}
	issueCache := map[uint]model.Issue{}
	releaseCache := map[uint]model.Release{}

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

		vo := activityVO{ActivityLog: log, UserName: name}
		pid := log.ProjectID
		if pid == 0 {
			pid = service.ResolveActivityProjectID(h.db, log.TargetType, log.TargetID)
		}
		if pid > 0 {
			proj, ok := projectCache[pid]
			if !ok {
				h.db.Select("id, code, name").First(&proj, pid)
				projectCache[pid] = proj
			}
			vo.ProjectName = proj.Name
			vo.ProjectCode = proj.Code
		}
		vo.TargetLabel = h.resolveTargetLabel(log, projectCache, issueCache, releaseCache)
		result = append(result, vo)
	}
	response.OK(c, result)
}

func (h *ActivityHandler) resolveTargetLabel(
	log model.ActivityLog,
	projectCache map[uint]model.Project,
	issueCache map[uint]model.Issue,
	releaseCache map[uint]model.Release,
) string {
	switch log.TargetType {
	case "project":
		proj, ok := projectCache[log.TargetID]
		if !ok {
			h.db.Select("id, code, name").First(&proj, log.TargetID)
			projectCache[log.TargetID] = proj
		}
		if proj.Name != "" && proj.Code != "" {
			return proj.Name + "（" + proj.Code + "）"
		}
		if proj.Name != "" {
			return proj.Name
		}
		if proj.Code != "" {
			return proj.Code
		}
		return "项目 #" + strconv.FormatUint(uint64(log.TargetID), 10)
	case "issue":
		issue, ok := issueCache[log.TargetID]
		if !ok {
			h.db.Select("id, title, type, project_id").First(&issue, log.TargetID)
			issueCache[log.TargetID] = issue
		}
		prefix := "任务"
		if issue.Type == "bug" || issue.Type == "" {
			prefix = "Bug"
		}
		var title string
		if issue.Title != "" {
			title = prefix + "：" + issue.Title
		} else {
			title = prefix + " #" + strconv.FormatUint(uint64(log.TargetID), 10)
		}
		return withProjectPrefix(h, log, projectCache, title)
	case "release":
		rel, ok := releaseCache[log.TargetID]
		if !ok {
			h.db.Select("id, version, title, project_id").First(&rel, log.TargetID)
			releaseCache[log.TargetID] = rel
		}
		var title string
		if rel.Version != "" {
			if rel.Title != "" {
				title = "发布 " + rel.Version + "（" + rel.Title + "）"
			} else {
				title = "发布 " + rel.Version
			}
		} else {
			title = "发布 #" + strconv.FormatUint(uint64(log.TargetID), 10)
		}
		return withProjectPrefix(h, log, projectCache, title)
	case "requirement":
		var req model.Requirement
		h.db.Select("id, title, project_id").First(&req, log.TargetID)
		var title string
		if req.Title != "" {
			title = "需求：" + req.Title
		} else {
			title = "需求 #" + strconv.FormatUint(uint64(log.TargetID), 10)
		}
		return withProjectPrefix(h, log, projectCache, title)
	default:
		if log.TargetID > 0 {
			return withProjectPrefix(h, log, projectCache, log.TargetType+" #"+strconv.FormatUint(uint64(log.TargetID), 10))
		}
		return log.TargetType
	}
}

func withProjectPrefix(
	h *ActivityHandler,
	log model.ActivityLog,
	projectCache map[uint]model.Project,
	label string,
) string {
	pid := log.ProjectID
	if pid == 0 {
		pid = service.ResolveActivityProjectID(h.db, log.TargetType, log.TargetID)
	}
	if pid == 0 {
		return label
	}
	proj, ok := projectCache[pid]
	if !ok {
		h.db.Select("id, code, name").First(&proj, pid)
		projectCache[pid] = proj
	}
	name := proj.Name
	if name == "" {
		name = proj.Code
	}
	if name == "" {
		return label
	}
	if proj.Code != "" && proj.Name != "" {
		return proj.Name + "（" + proj.Code + "）· " + label
	}
	return name + " · " + label
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
