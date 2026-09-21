package database

import (
	"os"
	"path/filepath"

	"devops-platform/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Connect(dbPath string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	return gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.ProjectSetting{},
		&model.ProjectConfigFile{},
		&model.ProjectStoreAsset{},
		&model.ProjectMember{},
		&model.ReleaseAsset{},
		&model.Requirement{},
		&model.Iteration{},
		&model.Issue{},
		&model.Release{},
		&model.ReleaseApproval{},
		&model.ReleaseItem{},
		&model.Comment{},
		&model.Attachment{},
		&model.BuildJob{},
		&model.ActivityLog{},
		&model.ProjectMilestone{},
		&model.ProjectChecklistItem{},
	); err != nil {
		return err
	}
	if err := backfillBuildNumbers(db); err != nil {
		return err
	}
	return backfillActivityProjectIDs(db)
}

// 历史构建按项目内创建时间补全独立编号
func backfillBuildNumbers(db *gorm.DB) error {
	var jobs []model.BuildJob
	if err := db.Where("build_number = 0 OR build_number IS NULL").Order("project_id asc, id asc").Find(&jobs).Error; err != nil {
		return err
	}
	if len(jobs) == 0 {
		return nil
	}
	nextByProject := map[uint]uint{}
	var maxes []struct {
		ProjectID uint
		MaxNum    uint
	}
	_ = db.Model(&model.BuildJob{}).
		Select("project_id, COALESCE(MAX(build_number), 0) as max_num").
		Where("build_number > 0").
		Group("project_id").
		Scan(&maxes)
	for _, m := range maxes {
		nextByProject[m.ProjectID] = m.MaxNum
	}
	for _, job := range jobs {
		nextByProject[job.ProjectID]++
		if err := db.Model(&job).Update("build_number", nextByProject[job.ProjectID]).Error; err != nil {
			return err
		}
	}
	return nil
}

func backfillActivityProjectIDs(db *gorm.DB) error {
	var logs []model.ActivityLog
	if err := db.Where("project_id = 0 OR project_id IS NULL").Find(&logs).Error; err != nil {
		return err
	}
	for _, log := range logs {
		var projectID uint
		switch log.TargetType {
		case "project":
			projectID = log.TargetID
		case "issue":
			var issue model.Issue
			if db.Select("project_id").First(&issue, log.TargetID).Error == nil {
				projectID = issue.ProjectID
			}
		case "release":
			var release model.Release
			if db.Select("project_id").First(&release, log.TargetID).Error == nil {
				projectID = release.ProjectID
			}
		}
		if projectID == 0 {
			continue
		}
		if err := db.Model(&log).Update("project_id", projectID).Error; err != nil {
			return err
		}
	}
	return nil
}
