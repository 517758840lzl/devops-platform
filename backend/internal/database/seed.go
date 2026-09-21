package database

import (
	"devops-platform/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	devHash, _ := bcrypt.GenerateFromPassword([]byte("dev123"), bcrypt.DefaultCost)
	testHash, _ := bcrypt.GenerateFromPassword([]byte("test123"), bcrypt.DefaultCost)

	admin := model.User{Username: "admin", PasswordHash: string(hash), Name: "管理员", Email: "admin@example.com", Role: "admin", Status: "active"}
	dev := model.User{Username: "dev", PasswordHash: string(devHash), Name: "开发-张三", Email: "dev@example.com", Role: "user", Status: "active"}
	tester := model.User{Username: "tester", PasswordHash: string(testHash), Name: "测试-李四", Email: "tester@example.com", Role: "user", Status: "active"}
	for _, u := range []*model.User{&admin, &dev, &tester} {
		if err := db.Create(u).Error; err != nil {
			return err
		}
	}

	project := model.Project{
		Name: "Prime Cedi Loan", Code: "PCL",
		Description: "示例项目，用于演示研发交付流水线",
		GitURL:       "https://github.com/octocat/Hello-World.git",
		GitBranch:    "master",
		BuildCommand: "cat README && echo 'build ok'",
		BuildWorkDir: "",
		OwnerID: admin.ID, Status: "active",
	}
	if err := db.Create(&project).Error; err != nil {
		return err
	}

	members := []model.ProjectMember{
		{ProjectID: project.ID, UserID: admin.ID, Role: "owner"},
		{ProjectID: project.ID, UserID: dev.ID, Role: "developer"},
		{ProjectID: project.ID, UserID: tester.ID, Role: "tester"},
	}
	if err := db.Create(&members).Error; err != nil {
		return err
	}

	_ = db.Create(&model.ProjectSetting{
		ProjectID: project.ID, AppDisplayName: "Prime Cedi Loan",
		AppID: "PrimeCreditLoan", ChannelIndex: "0",
		PackageAndroid: "com.bluebird.pcl", BundleIOS: "com.bluebird.pcl",
		AppleAppID: "6808185084", AppsflyerDevKey: "PFfRT77vnCVpKaZuU3Pghg",
		// TODO(deploy): 以下为演示域名，按真实项目替换
		ApiBaseURLProd: "https://www.bluebirdfintech.com/",
		ChannelCode: "GP",
		PrivacyURLProd: "https://example.com/privacy", PrivacyURLTest: "https://test.example.com/privacy",
		TermsURLProd: "https://example.com/terms", TermsURLTest: "https://test.example.com/terms",
		SupportedLocales: `["en","fr"]`, DefaultLocale: "en", I18nRepoPath: "assets/i18n",
		SmsFilterWords: `["loan scam","free money"]`, SmsSignName: "PrimeCedi", SmsTemplateID: "TPL001",
		UrlReplaceRules: `[{"env":"prod","from":"test.example.com","to":"example.com"}]`,
		// TODO(deploy): NotifyPublicBaseURL 上线改为公网域名；开发期可填局域网前端地址
		NotifyPublicBaseURL: "http://192.168.1.10:5173",
	}).Error

	req1 := model.Requirement{
		ProjectID: project.ID, Title: "用户登录与权限管理", Description: "支持账号密码登录、JWT 鉴权、角色权限",
		Source: "产品", Priority: "high", Status: "approved", ProposerID: admin.ID, AssigneeID: &dev.ID,
	}
	req2 := model.Requirement{
		ProjectID: project.ID, Title: "贷款申请流程优化", Description: "简化 KYC 步骤，提升转化率",
		Source: "业务", Priority: "medium", Status: "in_progress", ProposerID: admin.ID,
	}
	if err := db.Create(&req1).Error; err != nil {
		return err
	}
	if err := db.Create(&req2).Error; err != nil {
		return err
	}

	iter := model.Iteration{
		ProjectID: project.ID, Name: "v1.2 迭代", Version: "1.2.0",
		Status: "developing", OwnerID: admin.ID,
	}
	if err := db.Create(&iter).Error; err != nil {
		return err
	}

	iterID := iter.ID
	req1ID := req1.ID
	devID := dev.ID
	issues := []model.Issue{
		{ProjectID: project.ID, IterationID: &iterID, RequirementID: &req1ID, Type: "task", Title: "实现 JWT 登录接口", Status: "done", Priority: "high", ReporterID: admin.ID, AssigneeID: &devID},
		{ProjectID: project.ID, IterationID: &iterID, Type: "task", Title: "前端登录页联调", Status: "in_progress", Priority: "medium", ReporterID: admin.ID, AssigneeID: &devID},
		{ProjectID: project.ID, IterationID: &iterID, Type: "task", Title: "迭代看板拖拽", Status: "todo", Priority: "medium", ReporterID: admin.ID, AssigneeID: &devID},
		{ProjectID: project.ID, IterationID: &iterID, Type: "task", Title: "发布审批流联调", Status: "review", Priority: "high", ReporterID: admin.ID, AssigneeID: &devID},
		{ProjectID: project.ID, Type: "bug", Title: "iOS 14 闪退：启动页白屏", Status: "open", Priority: "high", Severity: "critical", ReporterID: tester.ID, Environment: "iOS 14", StepsToReproduce: "冷启动 App", ActualResult: "白屏 3 秒后闪退"},
		{ProjectID: project.ID, Type: "bug", Title: "还款页金额显示错误", Status: "in_progress", Priority: "medium", Severity: "major", ReporterID: tester.ID, Environment: "Android"},
	}
	for i := range issues {
		if err := db.Create(&issues[i]).Error; err != nil {
			return err
		}
	}

	pubID := admin.ID
	release := model.Release{
		ProjectID: project.ID, IterationID: &iterID, Version: "1.2.0", Title: "v1.2.0 发布",
		Status: "pending_approval", ReleaseNotes: "- 登录优化\n- 迭代看板\n- 发布审批流", PublisherID: &pubID,
	}
	if err := db.Create(&release).Error; err != nil {
		return err
	}

	_ = db.Create(&model.ReleaseApproval{
		ReleaseID: release.ID, UserID: admin.ID, Action: "submit", Comment: "请审批 v1.2.0 发布",
	}).Error
	_ = db.Create(&model.Comment{
		TargetType: "release", TargetID: release.ID, UserID: admin.ID, Content: "本次包含登录优化和看板功能，请尽快审批",
	}).Error
	if len(issues) > 1 {
		_ = db.Create(&model.ActivityLog{
			TargetType: "issue", TargetID: issues[1].ID, UserID: dev.ID, Action: "status_change", OldValue: "todo", NewValue: "in_progress",
		}).Error
	}

	return nil
}
