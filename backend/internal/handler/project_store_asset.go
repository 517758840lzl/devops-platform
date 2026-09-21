package handler

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devops-platform/internal/config"
	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"
	"devops-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProjectStoreAssetHandler struct {
	db  *gorm.DB
	cfg config.Config
}

func NewProjectStoreAssetHandler(db *gorm.DB, cfg config.Config) *ProjectStoreAssetHandler {
	return &ProjectStoreAssetHandler{db: db, cfg: cfg}
}

var iosStoreSlots = []string{
	"logo", "screenshot_1", "screenshot_2", "screenshot_3", "screenshot_4", "screenshot_5",
}

var androidStoreSlots = []string{
	"logo", "screenshot_1", "screenshot_2", "screenshot_3", "screenshot_4", "screenshot_5", "banner",
}

func (h *ProjectStoreAssetHandler) List(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	platform := c.Query("platform")
	q := h.db.Where("project_id = ?", projectID)
	if platform != "" {
		q = q.Where("platform = ?", platform)
	}
	var items []model.ProjectStoreAsset
	q.Order("platform asc, slot asc").Find(&items)
	response.OK(c, items)
}

func (h *ProjectStoreAssetHandler) Upload(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	platform := strings.ToLower(strings.TrimSpace(c.PostForm("platform")))
	slot := c.PostForm("slot")
	if platform != "ios" && platform != "android" {
		response.Fail(c, 400, "platform must be ios or android")
		return
	}
	if !isValidProjectStoreSlot(platform, slot) {
		response.Fail(c, 400, "invalid slot for platform")
		return
	}

	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, projectID, userID.(uint))
	if err != nil || !service.CanEditProjectConfig(role) {
		response.Fail(c, 403, "无上传权限")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, 400, "file required")
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedScreenshotExt[ext] {
		response.Fail(c, 400, "仅支持 jpg/png/webp")
		return
	}

	dir := filepath.Join(h.cfg.UploadDir, "store", fmt.Sprintf("project_%d", projectID), platform)
	_ = os.MkdirAll(dir, 0o755)
	stored := fmt.Sprintf("%s_%d%s", slot, time.Now().UnixMilli(), ext)
	savePath := filepath.Join(dir, stored)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	url := fmt.Sprintf("/uploads/store/project_%d/%s/%s", projectID, platform, stored)

	var asset model.ProjectStoreAsset
	if err := h.db.Where("project_id = ? AND platform = ? AND slot = ?", projectID, platform, slot).First(&asset).Error; err == nil {
		if asset.URL != "" && asset.URL != url {
			oldPath := filepath.Join(h.cfg.UploadDir, strings.TrimPrefix(asset.URL, "/uploads/"))
			_ = os.Remove(oldPath)
		}
		asset.URL = url
		asset.OriginalName = file.Filename
		asset.UserID = userID.(uint)
		h.db.Save(&asset)
	} else {
		asset = model.ProjectStoreAsset{
			ProjectID: projectID, Platform: platform, Slot: slot, URL: url,
			OriginalName: file.Filename, UserID: userID.(uint),
		}
		h.db.Create(&asset)
	}

	service.LogActivity(h.db, "project", projectID, userID.(uint), "upload_store_asset", platform+":"+slot, file.Filename)
	response.OK(c, asset)
}

func (h *ProjectStoreAssetHandler) DownloadZip(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	platform := strings.ToLower(strings.TrimSpace(c.Query("platform")))
	if platform == "" {
		platform = "all"
	}
	if platform != "all" && platform != "ios" && platform != "android" {
		response.Fail(c, 400, "platform must be all, ios or android")
		return
	}

	q := h.db.Where("project_id = ?", projectID)
	if platform != "all" {
		q = q.Where("platform = ?", platform)
	}
	var items []model.ProjectStoreAsset
	if err := q.Order("platform asc, slot asc").Find(&items).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	if len(items) == 0 {
		response.Fail(c, 404, "no store assets")
		return
	}

	filename := fmt.Sprintf("store-assets-project-%d-%s.zip", projectID, platform)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", "application/zip")
	zw := zip.NewWriter(c.Writer)
	defer zw.Close()

	for _, item := range items {
		diskPath := filepath.Join(h.cfg.UploadDir, strings.TrimPrefix(item.URL, "/uploads/"))
		data, err := os.ReadFile(diskPath)
		if err != nil {
			continue
		}
		ext := filepath.Ext(diskPath)
		entryName := item.Platform + "/" + item.Slot + ext
		if item.OriginalName != "" {
			entryName = item.Platform + "/" + item.OriginalName
		}
		w, err := zw.Create(entryName)
		if err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		if _, err := w.Write(data); err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
	}
}

func (h *ProjectStoreAssetHandler) DownloadOne(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	assetID := parseUint(c.Param("assetId"))
	var asset model.ProjectStoreAsset
	if err := h.db.Where("project_id = ? AND id = ?", projectID, assetID).First(&asset).Error; err != nil {
		response.Fail(c, 404, "not found")
		return
	}
	diskPath := filepath.Join(h.cfg.UploadDir, strings.TrimPrefix(asset.URL, "/uploads/"))
	name := asset.OriginalName
	if name == "" {
		name = asset.Platform + "_" + asset.Slot + filepath.Ext(diskPath)
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	c.File(diskPath)
}

func (h *ProjectStoreAssetHandler) Delete(c *gin.Context) {
	projectID := parseUint(c.Param("id"))
	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, projectID, userID.(uint))
	if err != nil || !service.CanDeleteStoreAsset(role) {
		response.Fail(c, 403, "仅项目 owner/admin 可删除上架素材")
		return
	}

	var asset model.ProjectStoreAsset
	if err := h.db.Where("project_id = ? AND id = ?", projectID, c.Param("assetId")).First(&asset).Error; err != nil {
		response.Fail(c, 404, "not found")
		return
	}
	diskPath := filepath.Join(h.cfg.UploadDir, strings.TrimPrefix(asset.URL, "/uploads/"))
	_ = os.Remove(diskPath)
	h.db.Delete(&asset)
	service.LogActivity(h.db, "project", projectID, userID.(uint), "delete_store_asset", asset.Platform+":"+asset.Slot, asset.OriginalName)
	response.OK(c, nil)
}

func isValidProjectStoreSlot(platform, slot string) bool {
	slots := iosStoreSlots
	if platform == "android" {
		slots = androidStoreSlots
	}
	for _, s := range slots {
		if s == slot {
			return true
		}
	}
	return false
}
