package handler

import (
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

type ReleaseAssetHandler struct {
	db  *gorm.DB
	cfg config.Config
}

func NewReleaseAssetHandler(db *gorm.DB, cfg config.Config) *ReleaseAssetHandler {
	return &ReleaseAssetHandler{db: db, cfg: cfg}
}

var allowedScreenshotExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
}

var storeSlots = []string{
	"screenshot_1", "screenshot_2", "screenshot_3", "screenshot_4", "screenshot_5",
}

func (h *ReleaseAssetHandler) List(c *gin.Context) {
	var items []model.ReleaseAsset
	h.db.Where("release_id = ?", c.Param("id")).Order("slot asc").Find(&items)
	response.OK(c, items)
}

func (h *ReleaseAssetHandler) Upload(c *gin.Context) {
	releaseID := parseUint(c.Param("id"))
	slot := c.PostForm("slot")
	validSlot := false
	for _, s := range storeSlots {
		if s == slot {
			validSlot = true
			break
		}
	}
	if !validSlot {
		response.Fail(c, 400, "invalid slot, use screenshot_1..screenshot_5")
		return
	}

	var release model.Release
	if err := h.db.First(&release, releaseID).Error; err != nil {
		response.Fail(c, 404, "release not found")
		return
	}

	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, release.ProjectID, userID.(uint))
	if err != nil || !service.CanWrite(role) {
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

	dir := filepath.Join(h.cfg.UploadDir, "store", fmt.Sprintf("release_%d", releaseID))
	_ = os.MkdirAll(dir, 0o755)
	stored := fmt.Sprintf("%s_%d%s", slot, time.Now().UnixMilli(), ext)
	savePath := filepath.Join(dir, stored)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	url := "/uploads/store/release_" + fmt.Sprintf("%d", releaseID) + "/" + stored

	var asset model.ReleaseAsset
	if err := h.db.Where("release_id = ? AND slot = ?", releaseID, slot).First(&asset).Error; err == nil {
		if asset.URL != "" && asset.URL != url {
			oldPath := filepath.Join(h.cfg.UploadDir, strings.TrimPrefix(asset.URL, "/uploads/"))
			_ = os.Remove(oldPath)
		}
		asset.URL = url
		asset.OriginalName = file.Filename
		asset.UserID = userID.(uint)
		h.db.Save(&asset)
	} else {
		asset = model.ReleaseAsset{
			ReleaseID: releaseID, Slot: slot, URL: url,
			OriginalName: file.Filename, UserID: userID.(uint),
		}
		h.db.Create(&asset)
	}

	service.LogActivity(h.db, "release", releaseID, userID.(uint), "upload_store_asset", slot, file.Filename)
	response.OK(c, asset)
}

func (h *ReleaseAssetHandler) Delete(c *gin.Context) {
	var asset model.ReleaseAsset
	if err := h.db.First(&asset, c.Param("assetId")).Error; err != nil {
		response.Fail(c, 404, "not found")
		return
	}

	var release model.Release
	if err := h.db.First(&release, asset.ReleaseID).Error; err != nil {
		response.Fail(c, 404, "release not found")
		return
	}

	userID, _ := c.Get("user_id")
	role, err := service.GetProjectRole(h.db, release.ProjectID, userID.(uint))
	if err != nil || !service.CanDeleteStoreAsset(role) {
		response.Fail(c, 403, "仅项目 owner/admin 可删除上架素材")
		return
	}

	diskPath := filepath.Join(h.cfg.UploadDir, strings.TrimPrefix(asset.URL, "/uploads/"))
	_ = os.Remove(diskPath)
	h.db.Delete(&asset)
	service.LogActivity(h.db, "release", release.ID, userID.(uint), "delete_store_asset", asset.Slot, asset.OriginalName)
	response.OK(c, nil)
}
