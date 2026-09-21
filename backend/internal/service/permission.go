package service

import (
	"errors"

	"devops-platform/internal/model"

	"gorm.io/gorm"
)

var ErrForbidden = errors.New("forbidden")
var ErrNotMember = errors.New("not a project member")

func GetProjectRole(db *gorm.DB, projectID, userID uint) (string, error) {
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return "", err
	}
	if user.Role == "admin" {
		return "owner", nil
	}
	var member model.ProjectMember
	if err := db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error; err != nil {
		return "", ErrNotMember
	}
	return member.Role, nil
}

func RequireProjectRole(db *gorm.DB, projectID, userID uint, allowed ...string) error {
	role, err := GetProjectRole(db, projectID, userID)
	if err != nil {
		return err
	}
	for _, r := range allowed {
		if role == r {
			return nil
		}
	}
	return ErrForbidden
}

func CanManageMembers(role string) bool {
	return role == "owner" || role == "super_admin" || role == "admin"
}

func CanDeleteStoreAsset(role string) bool {
	return role == "owner" || role == "super_admin" || role == "admin"
}

func CanApproveRelease(role string) bool {
	return role == "owner" || role == "super_admin" || role == "admin"
}

func CanWrite(role string) bool {
	return role != "viewer"
}

// CanEditProjectConfig: owner/admin/developer may edit project settings & config files; tester/viewer read-only.
func CanEditProjectConfig(role string) bool {
	return role == "owner" || role == "super_admin" || role == "admin" || role == "developer"
}
