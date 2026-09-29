package handler

import (
	"io"
	"path/filepath"
	"strings"

	"devops-platform/internal/config"
	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"
	"devops-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ConfigFileHandler struct {
	db  *gorm.DB
	cfg config.Config
}

func NewConfigFileHandler(db *gorm.DB, cfg config.Config) *ConfigFileHandler {
	return &ConfigFileHandler{db: db, cfg: cfg}
}

func (h *ConfigFileHandler) List(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	category := c.Query("category")
	if category != "" && !h.allowConfigRead(c, projectID, category) {
		response.OK(c, []model.ProjectConfigFile{})
		return
	}
	q := h.db.Where("project_id = ?", projectID)
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var items []model.ProjectConfigFile
	if err := q.Order("category asc, locale asc").Find(&items).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	role, roleErr := service.GetProjectRole(h.db, projectID, userID.(uint))
	if roleErr != nil || !service.CanEditProjectConfig(role) {
		filtered := items[:0]
		for _, it := range items {
			if !isDevOnlyConfigCategory(it.Category) {
				filtered = append(filtered, it)
			}
		}
		items = filtered
	}
	response.OK(c, items)
}

func (h *ConfigFileHandler) GetContent(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	category := c.Param("category")
	if !h.allowConfigRead(c, projectID, category) {
		response.OK(c, gin.H{"exists": false, "content": ""})
		return
	}
	locale := c.Query("locale")

	var record model.ProjectConfigFile
	if err := h.db.Where("project_id = ? AND category = ? AND locale = ?", projectID, category, locale).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.OK(c, gin.H{"exists": false, "content": ""})
			return
		}
		response.Fail(c, 500, err.Error())
		return
	}
	if !service.ConfigFileOnDisk(h.cfg.UploadDir, &record) {
		response.OK(c, gin.H{"exists": false, "content": ""})
		return
	}
	content, err := service.ReadConfigFileContent(h.cfg.UploadDir, &record)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{
		"exists":        true,
		"content":       content,
		"original_name": record.OriginalName,
		"size":          record.Size,
		"item_count":    record.ItemCount,
		"updated_at":    record.UpdatedAt,
	})
}

func (h *ConfigFileHandler) Upload(c *gin.Context) {
	h.save(c, true)
}

func (h *ConfigFileHandler) Paste(c *gin.Context) {
	h.save(c, false)
}

type pasteConfigBody struct {
	Content      string `json:"content"`
	Locale       string `json:"locale"`
	OriginalName string `json:"original_name"`
}

func (h *ConfigFileHandler) save(c *gin.Context, fromUpload bool) {
	projectID := parseUint(c.Param("id"))
	category := c.Param("category")

	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, projectID, userID.(uint))
	if err != nil || !service.CanEditProjectConfig(role) {
		response.Fail(c, 403, "无编辑权限")
		return
	}
	if !service.ValidateConfigCategory(category) {
		response.Fail(c, 400, "invalid category")
		return
	}

	var content, locale, originalName string
	if fromUpload {
		locale = c.PostForm("locale")
		file, err := c.FormFile("file")
		if err != nil {
			response.Fail(c, 400, "file required")
			return
		}
		if file.Size > service.MaxConfigFileSize {
			response.Fail(c, 400, "file too large, max 2MB")
			return
		}
		f, err := file.Open()
		if err != nil {
			response.Fail(c, 400, err.Error())
			return
		}
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			response.Fail(c, 400, err.Error())
			return
		}
		content = string(data)
		originalName = file.Filename
	} else {
		var body pasteConfigBody
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Fail(c, 400, "invalid request")
			return
		}
		content = body.Content
		locale = body.Locale
		originalName = body.OriginalName
	}

	if originalName == "" {
		switch category {
		case "i18n_source":
			originalName = "app_strings.dart"
		case "i18n_locale":
			originalName = "app_strings_" + locale + ".dart"
		case "sms_template":
			originalName = "sms_filter_template.txt"
		default:
			originalName = "sms_filter_words.txt"
		}
	}
	ext := strings.ToLower(filepath.Ext(originalName))
	if category == "i18n_source" || category == "i18n_locale" {
		if ext != ".dart" {
			response.Fail(c, 400, "i18n 文件请使用 .dart 格式")
			return
		}
	}

	record, err := service.UpsertConfigFile(h.db, h.cfg.UploadDir, projectID, category, locale, originalName, content)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	service.LogActivity(h.db, "project", projectID, userID.(uint), "update_config_file", category, originalName)
	response.OK(c, record)
}

// allowConfigRead：非敏感配置可读；api_constants 仅 developer+（无权限时由调用方静默返回空）。
func isDevOnlyConfigCategory(category string) bool {
	return category == "api_constants"
}

func (h *ConfigFileHandler) allowConfigRead(c *gin.Context, projectID uint, category string) bool {
	if !isDevOnlyConfigCategory(category) {
		return true
	}
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, projectID, userID.(uint))
	if err != nil || !service.CanEditProjectConfig(role) {
		return false
	}
	return true
}

func (h *ConfigFileHandler) Preview(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	category := c.Param("category")
	if !h.allowConfigRead(c, projectID, category) {
		response.OK(c, gin.H{"items": gin.H{}})
		return
	}
	locale := c.Query("locale")

	var record model.ProjectConfigFile
	if err := h.db.Where("project_id = ? AND category = ? AND locale = ?", projectID, category, locale).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.OK(c, gin.H{"items": gin.H{}})
			return
		}
		response.Fail(c, 500, err.Error())
		return
	}
	if !service.ConfigFileOnDisk(h.cfg.UploadDir, &record) {
		response.OK(c, gin.H{"items": gin.H{}})
		return
	}
	items, err := service.PreviewConfigFile(h.cfg.UploadDir, &record)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"items": items, "item_count": record.ItemCount})
}

func (h *ConfigFileHandler) I18nDiff(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	locale := c.Query("locale")
	if locale == "" {
		response.Fail(c, 400, "locale required")
		return
	}
	result, err := service.DiffI18nLocale(h.db, h.cfg.UploadDir, projectID, locale)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *ConfigFileHandler) DefaultSmsTemplate(c *gin.Context) {
	response.OK(c, gin.H{"content": service.DefaultSmsTemplate()})
}
