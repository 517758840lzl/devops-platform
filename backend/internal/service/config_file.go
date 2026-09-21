package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devops-platform/internal/model"

	"gorm.io/gorm"
)

const MaxConfigFileSize = 2 << 20 // 2MB

var allowedConfigCategories = map[string]bool{
	"i18n_source":  true,
	"i18n_locale":  true,
	"api_constants": true,
	"sms_template": true,
	"sms_words":    true,
}

func ValidateConfigCategory(category string) bool {
	return allowedConfigCategories[category]
}

func ConfigFileDir(uploadDir string, projectID uint, category, locale string) string {
	base := filepath.Join(uploadDir, "config", fmt.Sprintf("project_%d", projectID), category)
	if locale != "" {
		return filepath.Join(base, locale)
	}
	return base
}

func CountConfigItems(category, content string) int {
	switch category {
	case "i18n_source", "i18n_locale", "api_constants":
		return len(ParseDartStringConsts(content))
	case "sms_template":
		return len(ParseLineList(content))
	case "sms_words":
		return len(ParseLineList(content))
	default:
		return 0
	}
}

func UpsertConfigFile(db *gorm.DB, uploadDir string, projectID uint, category, locale, originalName, content string) (*model.ProjectConfigFile, error) {
	if !ValidateConfigCategory(category) {
		return nil, fmt.Errorf("invalid category")
	}
	if len(content) > MaxConfigFileSize {
		return nil, fmt.Errorf("file too large, max 2MB")
	}
	if category == "i18n_locale" && locale == "" {
		return nil, fmt.Errorf("locale required")
	}
	if category != "i18n_locale" {
		locale = ""
	}

	dir := ConfigFileDir(uploadDir, projectID, category, locale)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	ext := filepath.Ext(originalName)
	if ext == "" {
		switch category {
		case "i18n_source", "i18n_locale", "api_constants":
			ext = ".dart"
		default:
			ext = ".txt"
		}
	}
	stored := "content" + ext
	savePath := filepath.Join(dir, stored)
	if err := os.WriteFile(savePath, []byte(content), 0o644); err != nil {
		return nil, err
	}

	itemCount := CountConfigItems(category, content)
	record := model.ProjectConfigFile{
		ProjectID:    projectID,
		Category:     category,
		Locale:       locale,
		FilePath:     savePath,
		OriginalName: originalName,
		Size:         int64(len(content)),
		ItemCount:    itemCount,
	}

	var existing model.ProjectConfigFile
	err := db.Where("project_id = ? AND category = ? AND locale = ?", projectID, category, locale).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		if err := db.Create(&record).Error; err != nil {
			return nil, err
		}
		return &record, nil
	}
	if err != nil {
		return nil, err
	}
	record.ID = existing.ID
	if err := db.Save(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func ReadConfigFileContent(record *model.ProjectConfigFile) (string, error) {
	data, err := os.ReadFile(record.FilePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type I18nDiffResult struct {
	TotalKeys   int      `json:"total_keys"`
	Translated  int      `json:"translated"`
	MissingKeys []string `json:"missing_keys"`
	ExtraKeys   []string `json:"extra_keys"`
}

func DiffI18nLocale(db *gorm.DB, uploadDir string, projectID uint, locale string) (*I18nDiffResult, error) {
	source, err := loadConfigKeys(db, uploadDir, projectID, "i18n_source", "")
	if err != nil {
		return nil, err
	}
	localeKeys, err := loadConfigKeys(db, uploadDir, projectID, "i18n_locale", locale)
	if err != nil {
		return nil, err
	}

	result := &I18nDiffResult{TotalKeys: len(source)}
	for k := range source {
		if _, ok := localeKeys[k]; ok {
			result.Translated++
		} else {
			result.MissingKeys = append(result.MissingKeys, k)
		}
	}
	for k := range localeKeys {
		if _, ok := source[k]; !ok {
			result.ExtraKeys = append(result.ExtraKeys, k)
		}
	}
	return result, nil
}

func loadConfigKeys(db *gorm.DB, uploadDir string, projectID uint, category, locale string) (map[string]string, error) {
	var record model.ProjectConfigFile
	if err := db.Where("project_id = ? AND category = ? AND locale = ?", projectID, category, locale).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return map[string]string{}, nil
		}
		return nil, err
	}
	content, err := ReadConfigFileContent(&record)
	if err != nil {
		return nil, err
	}
	return ParseDartStringConsts(content), nil
}

func PreviewConfigFile(record *model.ProjectConfigFile) (any, error) {
	content, err := ReadConfigFileContent(record)
	if err != nil {
		return nil, err
	}
	switch record.Category {
	case "i18n_source", "i18n_locale", "api_constants":
		return ParseDartStringConsts(content), nil
	case "sms_template", "sms_words":
		return ParseLineList(content), nil
	default:
		return nil, fmt.Errorf("unsupported category")
	}
}

func DefaultSmsTemplate() string {
	return strings.TrimSpace(`# 短信过滤词模板（开发维护）
# 格式说明：
# - 业务词库：每行一个词
# - 以 # 开头的行为注释，会被忽略
# - 空行忽略
#
# 示例：
# loan scam
# free money
`)
}
