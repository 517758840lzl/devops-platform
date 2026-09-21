package service

import (
	"devops-platform/internal/model"

	"gorm.io/gorm"
)

func resolveActivityProjectID(db *gorm.DB, targetType string, targetID uint) uint {
	switch targetType {
	case "project":
		return targetID
	case "issue":
		var issue model.Issue
		if err := db.Select("project_id").First(&issue, targetID).Error; err == nil {
			return issue.ProjectID
		}
	case "release":
		var release model.Release
		if err := db.Select("project_id").First(&release, targetID).Error; err == nil {
			return release.ProjectID
		}
	case "requirement":
		var req model.Requirement
		if err := db.Select("project_id").First(&req, targetID).Error; err == nil {
			return req.ProjectID
		}
	}
	return 0
}

func LogActivity(db *gorm.DB, targetType string, targetID, userID uint, action, oldValue, newValue string) {
	_ = db.Create(&model.ActivityLog{
		ProjectID:  resolveActivityProjectID(db, targetType, targetID),
		TargetType: targetType,
		TargetID:   targetID,
		UserID:     userID,
		Action:     action,
		OldValue:   oldValue,
		NewValue:   newValue,
	}).Error
}
