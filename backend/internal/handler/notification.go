package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"devops-platform/internal/pkg/response"
	"devops-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	db *gorm.DB
}

func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{db: db}
}

func (h *NotificationHandler) List(c *gin.Context) {
	projectID := parseUint(c.Query("project_id"))
	userID, _ := c.Get("user_id")
	if projectID == 0 {
		response.Fail(c, 400, "project_id required")
		return
	}
	role, err := service.GetProjectRole(h.db, projectID, userID.(uint))
	if err != nil {
		response.OK(c, []service.NotificationVO{})
		return
	}
	items, err := service.ListNotifications(h.db, projectID, userID.(uint), role)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, items)
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID, _ := c.Get("user_id")
	id := parseUint(c.Param("id"))
	if err := service.MarkNotificationRead(h.db, id, userID.(uint)); err != nil {
		if err == service.ErrForbidden || err == service.ErrNotMember {
			response.Fail(c, 403, "无权限")
			return
		}
		response.Fail(c, 404, "notification not found")
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	projectID := parseUint(c.Query("project_id"))
	userID, _ := c.Get("user_id")
	if projectID == 0 {
		response.Fail(c, 400, "project_id required")
		return
	}
	role, err := service.GetProjectRole(h.db, projectID, userID.(uint))
	if err != nil {
		response.Fail(c, 403, "无项目权限")
		return
	}
	if err := service.MarkProjectNotificationsRead(h.db, projectID, userID.(uint), role); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (h *NotificationHandler) Events(c *gin.Context) {
	userID, _ := c.Get("user_id")
	projectID := parseUint(strings.TrimSpace(c.Query("project_id")))
	if projectID == 0 {
		response.Fail(c, 400, "project_id required")
		return
	}
	role, err := service.GetProjectRole(h.db, projectID, userID.(uint))
	if err != nil {
		response.Fail(c, 403, "无项目权限")
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.Fail(c, 500, "streaming unsupported")
		return
	}
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	flusher.Flush()

	ch := service.NotificationEvents.Subscribe(userID.(uint), projectID, role)
	defer service.NotificationEvents.Unsubscribe(ch)

	// 心跳保活，不是拉业务数据
	ping := time.NewTicker(25 * time.Second)
	defer ping.Stop()
	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ping.C:
			_, _ = fmt.Fprintf(c.Writer, ": ping\n\n")
			flusher.Flush()
		case n, open := <-ch:
			if !open {
				return
			}
			vo := service.NotificationVO{Notification: n, Read: false}
			payload, err := json.Marshal(vo)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintf(c.Writer, "event: notification\ndata: %s\n\n", payload)
			flusher.Flush()
		}
	}
}
