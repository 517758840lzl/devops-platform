package handler

import (
	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"
	"devops-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *ProjectHandler) GetSettings(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	userID, _ := c.Get("user_id")

	var setting model.ProjectSetting
	err := h.db.First(&setting, "project_id = ?", projectID).Error
	if err == gorm.ErrRecordNotFound {
		setting = defaultProjectSetting(projectID)
	} else if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	// 开发以下：仍可读非敏感配置，仅脱敏「应用配置」字段；不因非成员而 403
	role, roleErr := service.GetProjectRole(h.db, projectID, userID.(uint))
	if roleErr != nil || !service.CanEditProjectConfig(role) {
		redactDevOnlySettings(&setting)
	}
	response.OK(c, setting)
}

func (h *ProjectHandler) SaveSettings(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, projectID, userID.(uint))
	if err != nil || !service.CanEditProjectConfig(role) {
		response.Fail(c, 403, "无编辑权限")
		return
	}

	var req model.ProjectSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}
	req.ProjectID = projectID

	var existing model.ProjectSetting
	if err := h.db.First(&existing, "project_id = ?", projectID).Error; err == gorm.ErrRecordNotFound {
		if err := h.db.Create(&req).Error; err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
	} else if err != nil {
		response.Fail(c, 500, err.Error())
		return
	} else {
		if err := h.db.Save(&req).Error; err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
	}

	service.LogActivity(h.db, "project", projectID, userID.(uint), "update_settings", "", req.AppDisplayName)
	response.OK(c, req)
}

func defaultProjectSetting(projectID uint) model.ProjectSetting {
	return model.ProjectSetting{
		ProjectID:        projectID,
		ChannelIndex:     "0",
		SupportedLocales: `["zh","en"]`,
		DefaultLocale:    "zh",
		I18nRepoPath:     "lib/core/constants/app_strings.dart",
		SmsFilterWords:   `[]`,
		DisableEncBody:   "true",
		UrlReplaceRules:      `[]`,
		ExtraComplianceLinks: `[]`,
	}
}

// redactDevOnlySettings clears「应用配置」敏感字段，开发以下角色不可见。
func redactDevOnlySettings(s *model.ProjectSetting) {
	s.AppDisplayName = ""
	s.AppID = ""
	s.ChannelIndex = ""
	s.PackageAndroid = ""
	s.BundleIOS = ""
	s.AppleAppID = ""
	s.AppsflyerDevKey = ""
	s.RequestAesKey = ""
	s.ApiBaseURLProd = ""
	s.ApiBaseURLTest = ""
	s.ChannelCode = ""
	s.DisableEncBody = ""
}
