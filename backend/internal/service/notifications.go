package service

import (
	"fmt"
	"sync"
	"time"

	"devops-platform/internal/model"

	"gorm.io/gorm"
)

const (
	AudienceMembers   = "members"
	AudienceWriters   = "writers"
	AudienceApprovers = "approvers"
	AudienceUser      = "user"
)

type NotificationVO struct {
	model.Notification
	Read bool `json:"read"`
}

type notifSub struct {
	ch        chan model.Notification
	userID    uint
	projectID uint
	role      string
}

type notificationHub struct {
	mu   sync.RWMutex
	subs map[chan model.Notification]*notifSub
}

// NotificationEvents pushes only to subscribers who may see that event.
var NotificationEvents = &notificationHub{subs: make(map[chan model.Notification]*notifSub)}

func (h *notificationHub) Subscribe(userID, projectID uint, role string) chan model.Notification {
	ch := make(chan model.Notification, 32)
	h.mu.Lock()
	h.subs[ch] = &notifSub{ch: ch, userID: userID, projectID: projectID, role: role}
	h.mu.Unlock()
	return ch
}

func (h *notificationHub) Unsubscribe(ch chan model.Notification) {
	h.mu.Lock()
	if _, ok := h.subs[ch]; ok {
		delete(h.subs, ch)
		close(ch)
	}
	h.mu.Unlock()
}

func (h *notificationHub) Publish(n model.Notification) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch, sub := range h.subs {
		if sub.projectID != 0 && sub.projectID != n.ProjectID {
			continue
		}
		if !NotificationVisibleTo(sub.userID, sub.role, n) {
			continue
		}
		select {
		case ch <- n:
		default:
		}
	}
}

func NotificationVisibleTo(userID uint, role string, n model.Notification) bool {
	if userID == 0 || role == "" {
		return false
	}
	switch n.Audience {
	case AudienceApprovers:
		return CanApproveRelease(role)
	case AudienceWriters:
		return CanWrite(role)
	case AudienceUser:
		return n.AudienceUserID != nil && *n.AudienceUserID == userID
	default:
		// members / 空：项目成员（含 viewer）
		return true
	}
}

func PublishNotification(db *gorm.DB, n model.Notification) {
	if n.Audience == "" {
		n.Audience = AudienceMembers
	}
	if n.EventKey != "" {
		var exist model.Notification
		if db.Where("event_key = ?", n.EventKey).First(&exist).Error == nil {
			return
		}
	}
	if err := db.Create(&n).Error; err != nil {
		return
	}
	NotificationEvents.Publish(n)
}

func PublishBuildNotification(db *gorm.DB, job model.BuildJob) {
	if job.Status != "success" && job.Status != "failed" {
		return
	}
	title := "构建失败"
	if job.Status == "success" {
		title = "构建成功"
	}
	num := job.BuildNumber
	if num == 0 {
		num = job.ID
	}
	msg := fmt.Sprintf("构建 #%d", num)
	if job.Branch != "" {
		msg += " · " + job.Branch
	}
	if sha := job.CommitSHA; len(sha) >= 8 {
		msg += " · " + sha[:8]
	}
	PublishNotification(db, model.Notification{
		ProjectID: job.ProjectID,
		Type:      "build",
		Title:     title,
		Message:   msg,
		Link:      "/builds",
		EventKey:  fmt.Sprintf("build:%d:%s", job.ID, job.Status),
		Audience:  AudienceMembers,
	})
}

func PublishBugAssigneeNotification(db *gorm.DB, projectID, bugID uint, title string, assigneeID, actorID uint) {
	if assigneeID == 0 || assigneeID == actorID {
		return
	}
	PublishNotification(db, model.Notification{
		ProjectID:      projectID,
		Type:           "bug",
		Title:          "指派给你的 Bug",
		Message:        fmt.Sprintf("#%d %s", bugID, title),
		Link:           "/bugs",
		EventKey:       fmt.Sprintf("bug:%d:assignee:%d:%d", bugID, assigneeID, time.Now().Unix()),
		Audience:       AudienceUser,
		AudienceUserID: &assigneeID,
	})
}

func PublishBugStatusNotification(db *gorm.DB, item model.Issue, actorID uint) {
	n := model.Notification{
		ProjectID: item.ProjectID,
		Type:      "bug",
		Title:     "Bug 状态变更",
		Message:   fmt.Sprintf("#%d %s → %s", item.ID, item.Title, item.Status),
		Link:      "/bugs",
		EventKey:  fmt.Sprintf("bug:%d:status:%s", item.ID, item.Status),
	}
	if item.AssigneeID != nil && *item.AssigneeID != 0 {
		if *item.AssigneeID == actorID {
			return
		}
		n.Audience = AudienceUser
		n.AudienceUserID = item.AssigneeID
	} else {
		n.Audience = AudienceWriters
	}
	PublishNotification(db, n)
}

func ListNotifications(db *gorm.DB, projectID, userID uint, role string) ([]NotificationVO, error) {
	var items []model.Notification
	if err := db.Where("project_id = ?", projectID).Order("id desc").Limit(80).Find(&items).Error; err != nil {
		return nil, err
	}
	readIDs := map[uint]bool{}
	ids := make([]uint, 0, len(items))
	for _, n := range items {
		if NotificationVisibleTo(userID, role, n) {
			ids = append(ids, n.ID)
		}
	}
	if len(ids) > 0 {
		var reads []model.NotificationRead
		_ = db.Where("user_id = ? AND notification_id IN ?", userID, ids).Find(&reads)
		for _, r := range reads {
			readIDs[r.NotificationID] = true
		}
	}
	out := make([]NotificationVO, 0, len(items))
	for _, n := range items {
		if !NotificationVisibleTo(userID, role, n) {
			continue
		}
		out = append(out, NotificationVO{Notification: n, Read: readIDs[n.ID]})
	}
	return out, nil
}

func MarkNotificationRead(db *gorm.DB, notificationID, userID uint) error {
	var n model.Notification
	if err := db.First(&n, notificationID).Error; err != nil {
		return err
	}
	role, err := GetProjectRole(db, n.ProjectID, userID)
	if err != nil || !NotificationVisibleTo(userID, role, n) {
		return ErrForbidden
	}
	row := model.NotificationRead{NotificationID: notificationID, UserID: userID}
	return db.Where("notification_id = ? AND user_id = ?", notificationID, userID).
		FirstOrCreate(&row).Error
}

func MarkProjectNotificationsRead(db *gorm.DB, projectID, userID uint, role string) error {
	var items []model.Notification
	if err := db.Where("project_id = ?", projectID).Find(&items).Error; err != nil {
		return err
	}
	for _, n := range items {
		if !NotificationVisibleTo(userID, role, n) {
			continue
		}
		row := model.NotificationRead{NotificationID: n.ID, UserID: userID}
		if err := db.Where("notification_id = ? AND user_id = ?", n.ID, userID).
			FirstOrCreate(&row).Error; err != nil {
			return err
		}
	}
	return nil
}
