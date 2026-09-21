package service

import (
	"fmt"
	"strings"

	"devops-platform/internal/model"

	"gorm.io/gorm"
)

type checklistDef struct {
	Key    string
	Title  string
	Weight int
}

type milestoneDef struct {
	Key    string
	Title  string
	Weight int
	Items  []checklistDef
}

// LoanAppProgressTemplate is the fixed delivery pipeline for lending apps (v1).
var LoanAppProgressTemplate = []milestoneDef{
	{
		Key: "req", Title: "需求设计", Weight: 1,
		Items: []checklistDef{
			{Key: "req_doc", Title: "需求文档已确认", Weight: 1},
			{Key: "req_figma", Title: "Figma 设计稿已对齐", Weight: 1},
			{Key: "req_scope", Title: "范围与验收标准已确认", Weight: 1},
		},
	},
	{
		Key: "ui", Title: "页面UI", Weight: 1,
		Items: []checklistDef{
			{Key: "ui_login", Title: "登录/注册页 UI", Weight: 1},
			{Key: "ui_home", Title: "首页/额度页 UI", Weight: 1},
			{Key: "ui_loan", Title: "借款流程页 UI", Weight: 1},
			{Key: "ui_repay", Title: "还款流程页 UI", Weight: 1},
			{Key: "ui_profile", Title: "个人中心/设置页 UI", Weight: 1},
		},
	},
	{
		Key: "api", Title: "接口对接", Weight: 1,
		Items: []checklistDef{
			{Key: "api_base", Title: "API 域名/环境配置完成", Weight: 1},
			{Key: "api_auth", Title: "鉴权/加密通道联调", Weight: 1},
			{Key: "api_error", Title: "错误码与异常态处理", Weight: 1},
		},
	},
	{
		Key: "login", Title: "登录注册", Weight: 1,
		Items: []checklistDef{
			{Key: "login_sms", Title: "短信登录/注册", Weight: 1},
			{Key: "login_ocr", Title: "实名 OCR 对接", Weight: 1},
			{Key: "login_kyc", Title: "KYC/活体认证", Weight: 1},
			{Key: "login_session", Title: "会话与登出", Weight: 1},
		},
	},
	{
		Key: "loan", Title: "借款流程", Weight: 2,
		Items: []checklistDef{
			{Key: "loan_quota", Title: "额度展示与刷新", Weight: 1},
			{Key: "loan_apply", Title: "借款申请提交", Weight: 2},
			{Key: "loan_bank", Title: "绑卡/收款账户", Weight: 1},
			{Key: "loan_contract", Title: "借款合同确认", Weight: 1},
			{Key: "loan_disburse", Title: "放款结果页", Weight: 1},
			{Key: "loan_history", Title: "借款记录列表", Weight: 1},
		},
	},
	{
		Key: "repay", Title: "还款流程", Weight: 2,
		Items: []checklistDef{
			{Key: "repay_plan", Title: "还款计划展示", Weight: 1},
			{Key: "repay_pay", Title: "主动还款", Weight: 2},
			{Key: "repay_early", Title: "提前结清", Weight: 1},
			{Key: "repay_result", Title: "还款结果与账单", Weight: 1},
		},
	},
	{
		Key: "qa", Title: "测试验收", Weight: 1,
		Items: []checklistDef{
			{Key: "qa_smoke", Title: "冒烟用例通过", Weight: 1},
			{Key: "qa_regression", Title: "回归用例通过", Weight: 1},
			{Key: "qa_bugs_closed", Title: "阻塞/严重 Bug 已关闭", Weight: 2},
			{Key: "qa_signoff", Title: "测试签字验收", Weight: 1},
		},
	},
	{
		Key: "store", Title: "上架素材", Weight: 1,
		Items: []checklistDef{
			{Key: "store_android", Title: "Android 上架素材齐全", Weight: 1},
			// iOS 与 Android 不同步，暂不纳入进度统计
			{Key: "store_copy", Title: "商店文案/关键词就绪", Weight: 1},
		},
	},
	{
		Key: "compliance", Title: "合规链接", Weight: 1,
		Items: []checklistDef{
			{Key: "comp_privacy", Title: "隐私协议正式 URL", Weight: 1},
			{Key: "comp_terms", Title: "用户协议正式 URL", Weight: 1},
			{Key: "comp_website", Title: "官网正式 URL", Weight: 1},
			// 借款合同暂不纳入进度统计
		},
	},
	{
		Key: "release", Title: "发版上线", Weight: 1,
		Items: []checklistDef{
			{Key: "rel_build", Title: "可用 Release 安装包", Weight: 1},
			{Key: "rel_notes", Title: "发版说明就绪", Weight: 1},
			{Key: "rel_publish", Title: "商店提交/上线", Weight: 1},
		},
	},
}

// excludedFromProgress: 暂不计入进度与缺口（如 iOS 素材、借款合同）。
func excludedFromProgress(itemKey string) bool {
	return itemKey == "store_ios" || itemKey == "comp_loan"
}

// EnsureProgressSeeded creates the loan-app checklist template if the project has none.
func EnsureProgressSeeded(db *gorm.DB, projectID uint) error {
	var count int64
	if err := db.Model(&model.ProjectMilestone{}).Where("project_id = ?", projectID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		for i, ms := range LoanAppProgressTemplate {
			m := model.ProjectMilestone{
				ProjectID: projectID,
				Key:       ms.Key,
				Title:     ms.Title,
				Sort:      i + 1,
				Status:    "todo",
				Weight:    ms.Weight,
			}
			if err := tx.Create(&m).Error; err != nil {
				return err
			}
			for j, it := range ms.Items {
				w := it.Weight
				if w <= 0 {
					w = 1
				}
				item := model.ProjectChecklistItem{
					MilestoneID: m.ID,
					Key:         it.Key,
					Title:       it.Title,
					Status:      "todo",
					Weight:      w,
					Sort:        j + 1,
				}
				if err := tx.Create(&item).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

type ProgressHint struct {
	Key      string `json:"key"`
	Category string `json:"category"` // compliance | store | bug | build
	Level    string `json:"level"`    // info | warning | critical
	Message  string `json:"message"`
	ItemKey  string `json:"item_key,omitempty"` // related checklist key for apply-hints
}

type MilestoneProgressVO struct {
	model.ProjectMilestone
	DoneCount  int     `json:"done_count"`
	TotalCount int     `json:"total_count"`
	Percent    float64 `json:"percent"`
}

type ProgressSummary struct {
	Percent           float64               `json:"percent"`
	CurrentStageKey   string                `json:"current_stage_key"`
	CurrentStageTitle string                `json:"current_stage_title"`
	BlockedCount      int                   `json:"blocked_count"`
	DoneWeight        int                   `json:"done_weight"`
	TotalWeight       int                   `json:"total_weight"`
	UpdatedAt         *string               `json:"updated_at,omitempty"`
	UpdatedByName     string                `json:"updated_by_name,omitempty"`
	Milestones        []MilestoneProgressVO `json:"milestones"`
	Hints             []ProgressHint        `json:"hints"`
	CanEdit           bool                  `json:"can_edit"`
}

func LoadProjectProgress(db *gorm.DB, projectID uint) (*ProgressSummary, error) {
	if err := EnsureProgressSeeded(db, projectID); err != nil {
		return nil, err
	}

	var milestones []model.ProjectMilestone
	if err := db.Where("project_id = ?", projectID).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort asc").Preload("Assignee")
		}).
		Order("sort asc").
		Find(&milestones).Error; err != nil {
		return nil, err
	}

	summary := &ProgressSummary{
		Milestones: make([]MilestoneProgressVO, 0, len(milestones)),
		Hints:      CollectProgressHints(db, projectID),
	}

	doneWeight, totalWeight, blocked := 0, 0, 0
	currentKey, currentTitle := "", ""
	var latestUpdate *string
	openCriticalBugs := countOpenSevereBugs(db, projectID)

	for i := range milestones {
		ms := milestones[i]
		visible := make([]model.ProjectChecklistItem, 0, len(ms.Items))
		for _, it := range ms.Items {
			if excludedFromProgress(it.Key) {
				continue
			}
			visible = append(visible, it)
		}
		ms.Items = visible

		done, total, msDoneW, msTotalW := 0, len(ms.Items), 0, 0
		msBlocked := false
		allDone := total > 0
		for _, it := range ms.Items {
			w := it.Weight
			if w <= 0 {
				w = 1
			}
			msTotalW += w
			totalWeight += w
			switch it.Status {
			case "done":
				done++
				msDoneW += w
				doneWeight += w
			case "blocked":
				blocked++
				msBlocked = true
				allDone = false
			default:
				allDone = false
			}
		}
		if total == 0 {
			allDone = ms.Status == "done"
		}

		status := ms.Status
		if msBlocked {
			status = "blocked"
		} else if allDone {
			status = "done"
		} else if done > 0 {
			status = "doing"
		} else {
			status = "todo"
		}
		// Persist derived status lightly (best-effort)
		if status != ms.Status {
			_ = db.Model(&ms).Update("status", status).Error
			ms.Status = status
		}

		pct := 0.0
		if msTotalW > 0 {
			pct = float64(msDoneW) / float64(msTotalW) * 100
		}
		summary.Milestones = append(summary.Milestones, MilestoneProgressVO{
			ProjectMilestone: ms,
			DoneCount:        done,
			TotalCount:       total,
			Percent:          pct,
		})

		if currentKey == "" && !allDone {
			currentKey = ms.Key
			currentTitle = ms.Title
			// If QA stage and open severe bugs, stay / mark blocked feel
			if ms.Key == "qa" && openCriticalBugs > 0 {
				currentTitle = ms.Title + "（有未关严重 Bug）"
			}
		}

		for _, it := range ms.Items {
			ts := it.UpdatedAt.Format("2006-01-02 15:04:05")
			if latestUpdate == nil || ts > *latestUpdate {
				latestUpdate = &ts
			}
		}
	}

	if currentKey == "" && len(milestones) > 0 {
		last := milestones[len(milestones)-1]
		currentKey = last.Key
		currentTitle = last.Title
	}

	summary.Percent = 0
	if totalWeight > 0 {
		summary.Percent = float64(doneWeight) / float64(totalWeight) * 100
	}
	summary.CurrentStageKey = currentKey
	summary.CurrentStageTitle = currentTitle
	summary.BlockedCount = blocked
	summary.DoneWeight = doneWeight
	summary.TotalWeight = totalWeight
	summary.UpdatedAt = latestUpdate

	var lastAct model.ActivityLog
	if err := db.Where("project_id = ? AND action = ?", projectID, "update_progress").
		Order("created_at desc").First(&lastAct).Error; err == nil {
		var user model.User
		if db.Select("name, username").First(&user, lastAct.UserID).Error == nil {
			summary.UpdatedByName = user.Name
			if summary.UpdatedByName == "" {
				summary.UpdatedByName = user.Username
			}
		}
		ts := lastAct.CreatedAt.Format("2006-01-02 15:04:05")
		summary.UpdatedAt = &ts
	}

	return summary, nil
}

func CollectProgressHints(db *gorm.DB, projectID uint) []ProgressHint {
	hints := []ProgressHint{}

	var setting model.ProjectSetting
	if err := db.Where("project_id = ?", projectID).First(&setting).Error; err == nil {
		if strings.TrimSpace(setting.PrivacyURLProd) == "" {
			hints = append(hints, ProgressHint{
				Key: "privacy_empty", Category: "compliance", Level: "warning",
				Message: "隐私协议正式 URL 未填", ItemKey: "comp_privacy",
			})
		}
		if strings.TrimSpace(setting.TermsURLProd) == "" {
			hints = append(hints, ProgressHint{
				Key: "terms_empty", Category: "compliance", Level: "warning",
				Message: "用户协议正式 URL 未填", ItemKey: "comp_terms",
			})
		}
		if strings.TrimSpace(setting.OfficialWebsiteURLProd) == "" {
			hints = append(hints, ProgressHint{
				Key: "website_empty", Category: "compliance", Level: "warning",
				Message: "官网正式 URL 未填", ItemKey: "comp_website",
			})
		}
	} else {
		hints = append(hints, ProgressHint{
			Key: "settings_missing", Category: "compliance", Level: "warning",
			Message: "项目配置尚未初始化（隐私/用户协议/官网等）",
		})
	}

	androidFilled, androidTotal := countStoreSlots(db, projectID, "android", []string{
		"logo", "screenshot_1", "screenshot_2", "screenshot_3", "screenshot_4", "screenshot_5", "banner",
	})
	if androidFilled < androidTotal {
		hints = append(hints, ProgressHint{
			Key: "android_assets", Category: "store", Level: "warning",
			Message: fmt.Sprintf("Android 上架素材 %d/%d", androidFilled, androidTotal),
			ItemKey: "store_android",
		})
	}
	var openBugs []model.Issue
	_ = db.Where("project_id = ? AND type = ? AND status IN ?", projectID, "bug", []string{"open", "in_progress"}).
		Find(&openBugs)
	if len(openBugs) > 0 {
		critical, major := 0, 0
		for _, b := range openBugs {
			sev := strings.ToLower(b.Severity)
			pri := strings.ToLower(b.Priority)
			if sev == "critical" || sev == "blocker" || pri == "p0" || pri == "critical" {
				critical++
			} else if sev == "major" || sev == "high" || pri == "p1" || pri == "high" {
				major++
			}
		}
		msg := fmt.Sprintf("未关闭 Bug %d 个", len(openBugs))
		level := "info"
		if critical > 0 {
			msg = fmt.Sprintf("未关闭 Bug %d 个（含 %d 个 Critical）", len(openBugs), critical)
			level = "critical"
		} else if major > 0 {
			msg = fmt.Sprintf("未关闭 Bug %d 个（含 %d 个 Major）", len(openBugs), major)
			level = "warning"
		}
		hints = append(hints, ProgressHint{
			Key: "open_bugs", Category: "bug", Level: level,
			Message: msg, ItemKey: "qa_bugs_closed",
		})
	}

	var successBuild int64
	_ = db.Model(&model.BuildJob{}).
		Where("project_id = ? AND status = ? AND artifacts_deleted = ?", projectID, "success", false).
		Count(&successBuild)
	if successBuild == 0 {
		hints = append(hints, ProgressHint{
			Key: "no_release_build", Category: "build", Level: "warning",
			Message: "尚无可用安装包", ItemKey: "rel_build",
		})
	}

	return hints
}

func countStoreSlots(db *gorm.DB, projectID uint, platform string, slots []string) (filled, total int) {
	total = len(slots)
	var assets []model.ProjectStoreAsset
	_ = db.Where("project_id = ? AND platform = ?", projectID, platform).Find(&assets)
	seen := map[string]bool{}
	for _, a := range assets {
		seen[a.Slot] = true
	}
	for _, s := range slots {
		if seen[s] {
			filled++
		}
	}
	return
}

func countOpenSevereBugs(db *gorm.DB, projectID uint) int {
	var bugs []model.Issue
	_ = db.Where("project_id = ? AND type = ? AND status IN ?", projectID, "bug", []string{"open", "in_progress"}).
		Find(&bugs)
	n := 0
	for _, b := range bugs {
		sev := strings.ToLower(b.Severity)
		pri := strings.ToLower(b.Priority)
		if sev == "critical" || sev == "blocker" || sev == "major" ||
			pri == "p0" || pri == "p1" || pri == "critical" || pri == "high" {
			n++
		}
	}
	return n
}

// ApplyProgressHints marks checklist items related to active hints as todo (not done).
func ApplyProgressHints(db *gorm.DB, projectID, userID uint) (int, error) {
	if err := EnsureProgressSeeded(db, projectID); err != nil {
		return 0, err
	}
	hints := CollectProgressHints(db, projectID)
	keys := map[string]bool{}
	for _, h := range hints {
		if h.ItemKey != "" {
			keys[h.ItemKey] = true
		}
	}
	if len(keys) == 0 {
		return 0, nil
	}
	var milestones []model.ProjectMilestone
	if err := db.Where("project_id = ?", projectID).Preload("Items").Find(&milestones).Error; err != nil {
		return 0, err
	}
	updated := 0
	for _, ms := range milestones {
		for _, it := range ms.Items {
			if excludedFromProgress(it.Key) || !keys[it.Key] {
				continue
			}
			if it.Status == "done" {
				old := it.Status
				if err := db.Model(&it).Update("status", "todo").Error; err != nil {
					return updated, err
				}
				LogActivity(db, "project", projectID, userID, "update_progress",
					fmt.Sprintf("%s:%s", it.Title, old), "todo")
				updated++
			}
		}
	}
	return updated, nil
}
