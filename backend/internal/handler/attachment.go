package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devops-platform/internal/config"
	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"
	"devops-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttachmentHandler struct {
	db  *gorm.DB
	cfg config.Config
}

func NewAttachmentHandler(db *gorm.DB, cfg config.Config) *AttachmentHandler {
	return &AttachmentHandler{db: db, cfg: cfg}
}

var allowedImageExt = map[string]string{
	".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png",
	".gif": "image/gif", ".webp": "image/webp",
}

var allowedVideoExt = map[string]string{
	".mp4": "video/mp4", ".mov": "video/quicktime", ".webm": "video/webm",
}

const maxImageSize = 10 << 20  // 10MB
const maxVideoSize = 100 << 20 // 100MB

func (h *AttachmentHandler) List(c *gin.Context) {
	targetType := c.Query("target_type")
	targetID := c.Query("target_id")
	if targetType == "" || targetID == "" {
		response.Fail(c, 400, "target_type and target_id required")
		return
	}
	var items []model.Attachment
	if err := h.db.Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("created_at asc").Find(&items).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, items)
}

func (h *AttachmentHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, 400, "file required")
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	mimeType, mediaType := "", ""
	if m, ok := allowedImageExt[ext]; ok {
		mimeType, mediaType = m, "image"
		if file.Size > maxImageSize {
			response.Fail(c, 400, "图片不能超过 10MB")
			return
		}
	} else if m, ok := allowedVideoExt[ext]; ok {
		mimeType, mediaType = m, "video"
		if file.Size > maxVideoSize {
			response.Fail(c, 400, "视频不能超过 100MB")
			return
		}
	} else {
		response.Fail(c, 400, "仅支持 jpg/png/gif/webp/mp4/mov/webm")
		return
	}

	targetType := c.PostForm("target_type")
	targetID := parseUint(c.PostForm("target_id"))
	if targetType != "" && targetID > 0 {
		userID, _ := c.Get("user_id")
		var issue model.Issue
		if targetType == "issue" {
			if err := h.db.First(&issue, targetID).Error; err != nil {
				response.Fail(c, 404, "issue not found")
				return
			}
			role, err := service.GetProjectRole(h.db, issue.ProjectID, userID.(uint))
			if err != nil || !service.CanWrite(role) {
				response.Fail(c, 403, "无上传权限")
				return
			}
		}
	}

	if err := os.MkdirAll(h.cfg.UploadDir, 0o755); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	storedName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	savePath := filepath.Join(h.cfg.UploadDir, storedName)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	att := model.Attachment{
		TargetType:   targetType,
		TargetID:     targetID,
		UserID:       userID.(uint),
		FileName:     storedName,
		OriginalName: file.Filename,
		MimeType:     mimeType,
		MediaType:    mediaType,
		Size:         file.Size,
		URL:          "/uploads/" + storedName,
	}
	if err := h.db.Create(&att).Error; err != nil {
		os.Remove(savePath)
		response.Fail(c, 500, err.Error())
		return
	}
	if targetType != "" && targetID > 0 {
		service.LogActivity(h.db, targetType, targetID, userID.(uint), "upload_attachment", "", file.Filename)
	}
	response.OK(c, att)
}

func (h *AttachmentHandler) Delete(c *gin.Context) {
	var att model.Attachment
	if err := h.db.First(&att, c.Param("id")).Error; err != nil {
		response.Fail(c, 404, "attachment not found")
		return
	}
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	if att.UserID != userID.(uint) && role.(string) != "admin" {
		response.Fail(c, 403, "forbidden")
		return
	}
	_ = os.Remove(filepath.Join(h.cfg.UploadDir, att.FileName))
	h.db.Delete(&att)
	response.OK(c, nil)
}
