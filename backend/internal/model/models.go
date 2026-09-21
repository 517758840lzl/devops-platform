package model

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64" json:"username"`
	PasswordHash string    `gorm:"size:255" json:"-"`
	Name         string    `gorm:"size:64" json:"name"`
	Email        string    `gorm:"size:128" json:"email"`
	Role         string    `gorm:"size:32" json:"role"`
	Status       string    `gorm:"size:32" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Project struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:128" json:"name"`
	Code         string    `gorm:"size:32;uniqueIndex" json:"code"`
	Description  string    `gorm:"type:text" json:"description"`
	GitURL            string `gorm:"size:512" json:"git_url"`
	GitBranch         string `gorm:"size:64" json:"git_branch"`
	GitRepos          string `gorm:"type:text" json:"git_repos"`
	BuildProfile      string `gorm:"size:16" json:"build_profile"`
	BuildPlatform     string `gorm:"size:16" json:"build_platform"`
	BuildCommand      string `gorm:"type:text" json:"build_command"`
	BuildWorkDir      string `gorm:"size:128" json:"build_work_dir"`
	ArtifactOutputDir string `gorm:"size:512" json:"artifact_output_dir"`
	OwnerID      uint      `json:"owner_id"`
	Status       string    `gorm:"size:32" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type BuildJob struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	ProjectID    uint       `gorm:"index" json:"project_id"`
	BuildNumber  uint       `gorm:"index" json:"build_number"`
	ReleaseID    *uint      `json:"release_id"`
	GitURL        string `gorm:"size:512" json:"git_url"`
	Branch        string `gorm:"size:64" json:"branch"`
	RepoName      string `gorm:"size:128" json:"repo_name"`
	BuildWorkDir  string `gorm:"size:128" json:"build_work_dir"`
	CommitSHA    string     `gorm:"size:64" json:"commit_sha"`
	Status       string     `gorm:"size:32;index" json:"status"`
	Log          string     `gorm:"type:text" json:"log"`
	ArtifactPath     string     `gorm:"size:512" json:"artifact_path"`
	ArtifactsDeleted bool       `json:"artifacts_deleted"`
	TriggeredBy      uint       `json:"triggered_by"`
	StartedAt    *time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ProjectSetting struct {
	ProjectID        uint      `gorm:"primaryKey" json:"project_id"`
	AppDisplayName   string    `gorm:"size:128" json:"app_display_name"`
	AppID            string    `gorm:"size:128" json:"app_id"`
	ChannelIndex     string    `gorm:"size:16" json:"channel_index"`
	PackageAndroid   string    `gorm:"size:128" json:"package_android"`
	BundleIOS        string    `gorm:"size:128" json:"bundle_ios"`
	AppleAppID       string    `gorm:"size:32" json:"apple_app_id"`
	AppsflyerDevKey  string    `gorm:"size:128" json:"appsflyer_dev_key"`
	RequestAesKey    string    `gorm:"size:128" json:"request_aes_key"`
	DisableEncBody   string    `gorm:"size:8" json:"disable_enc_body"`
	ApiBaseURLProd   string    `gorm:"size:512" json:"api_base_url_prod"`
	ApiBaseURLTest   string    `gorm:"size:512" json:"api_base_url_test"`
	ChannelCode      string    `gorm:"size:64" json:"channel_code"`
	OfficialWebsiteURLProd string `gorm:"size:512" json:"official_website_url_prod"`
	OfficialWebsiteURLTest string `gorm:"size:512" json:"official_website_url_test"`
	PrivacyURLProd         string `gorm:"size:512" json:"privacy_url_prod"`
	PrivacyURLTest         string `gorm:"size:512" json:"privacy_url_test"`
	TermsURLProd         string `gorm:"size:512" json:"terms_url_prod"`
	TermsURLTest         string `gorm:"size:512" json:"terms_url_test"`
	LoanContractURLProd  string `gorm:"size:512" json:"loan_contract_url_prod"`
	LoanContractURLTest  string `gorm:"size:512" json:"loan_contract_url_test"`
	ExtraComplianceLinks string `gorm:"type:text" json:"extra_compliance_links"`
	UrlReplaceRules      string `gorm:"type:text" json:"url_replace_rules"`
	SupportedLocales string    `gorm:"type:text" json:"supported_locales"`
	I18nRepoPath     string    `gorm:"size:255" json:"i18n_repo_path"`
	DefaultLocale    string    `gorm:"size:16" json:"default_locale"`
	SmsFilterWords   string    `gorm:"type:text" json:"sms_filter_words"`
	SmsTemplateID    string    `gorm:"size:64" json:"sms_template_id"`
	SmsSignName      string    `gorm:"size:64" json:"sms_sign_name"`
	RequirementDocURL string   `gorm:"size:512" json:"requirement_doc_url"`
	FigmaDesignURL    string   `gorm:"size:512" json:"figma_design_url"`
	FigmaFigJamURL    string   `gorm:"size:512" json:"figma_figjam_url"`
	FigmaNotes        string   `gorm:"type:text" json:"figma_notes"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ProjectConfigFile struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ProjectID    uint      `gorm:"index:idx_proj_cfg_cat_locale,unique" json:"project_id"`
	Category     string    `gorm:"size:32;index:idx_proj_cfg_cat_locale,unique" json:"category"`
	Locale       string    `gorm:"size:16;index:idx_proj_cfg_cat_locale,unique" json:"locale"`
	FilePath     string    `gorm:"size:512" json:"file_path"`
	OriginalName string    `gorm:"size:255" json:"original_name"`
	Size         int64     `json:"size"`
	ItemCount    int       `json:"item_count"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ProjectStoreAsset struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ProjectID    uint      `gorm:"index:idx_proj_store_plat_slot,unique" json:"project_id"`
	Platform     string    `gorm:"size:16;index:idx_proj_store_plat_slot,unique" json:"platform"`
	Slot         string    `gorm:"size:32;index:idx_proj_store_plat_slot,unique" json:"slot"`
	URL          string    `gorm:"size:512" json:"url"`
	OriginalName string    `gorm:"size:255" json:"original_name"`
	UserID       uint      `json:"user_id"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ReleaseAsset struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ReleaseID    uint      `gorm:"index" json:"release_id"`
	Slot         string    `gorm:"size:32;index" json:"slot"`
	URL          string    `gorm:"size:512" json:"url"`
	OriginalName string    `gorm:"size:255" json:"original_name"`
	UserID       uint      `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type ProjectMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `json:"project_id"`
	UserID    uint      `json:"user_id"`
	Role      string    `gorm:"size:32" json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Requirement struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProjectID   uint      `json:"project_id"`
	Title       string    `gorm:"size:255" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Source      string    `gorm:"size:64" json:"source"`
	Priority    string    `gorm:"size:32" json:"priority"`
	Status      string    `gorm:"size:32" json:"status"`
	ProposerID  uint      `json:"proposer_id"`
	AssigneeID  *uint     `json:"assignee_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Iteration struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	ProjectID   uint       `json:"project_id"`
	Name        string     `gorm:"size:128" json:"name"`
	Version     string     `gorm:"size:32" json:"version"`
	StartDate   *time.Time `json:"start_date"`
	TestDate    *time.Time `json:"test_date"`
	ReleaseDate *time.Time `json:"release_date"`
	Status      string     `gorm:"size:32" json:"status"`
	OwnerID     uint       `json:"owner_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Issue struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	ProjectID          uint       `json:"project_id"`
	IterationID        *uint      `json:"iteration_id"`
	RequirementID      *uint      `json:"requirement_id"`
	Type               string     `gorm:"size:16;index" json:"type"`
	Title              string     `gorm:"size:255" json:"title"`
	Description        string     `gorm:"type:text" json:"description"`
	Status             string     `gorm:"size:32;index" json:"status"`
	Priority           string     `gorm:"size:32" json:"priority"`
	Severity           string     `gorm:"size:32" json:"severity"`
	AssigneeID         *uint      `json:"assignee_id"`
	ReleaseID          *uint      `json:"release_id"`
	AppVersion         string     `gorm:"size:64" json:"app_version"`
	BugSide            string     `gorm:"size:16" json:"bug_side"`
	ReporterID         uint       `json:"reporter_id"`
	DueDate            *time.Time `json:"due_date"`
	StepsToReproduce   string     `gorm:"type:text" json:"steps_to_reproduce"`
	ExpectedResult     string     `gorm:"type:text" json:"expected_result"`
	ActualResult       string     `gorm:"type:text" json:"actual_result"`
	Environment        string     `gorm:"size:32" json:"environment"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	ClosedAt           *time.Time `json:"closed_at"`
}

type Release struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	ProjectID    uint       `json:"project_id"`
	IterationID  *uint      `json:"iteration_id"`
	Version      string     `gorm:"size:32" json:"version"`
	Title        string     `gorm:"size:255" json:"title"`
	Status       string     `gorm:"size:32" json:"status"`
	ReleaseNotes string     `gorm:"type:text" json:"release_notes"`
	RollbackPlan string     `gorm:"type:text" json:"rollback_plan"`
	PublisherID  *uint      `json:"publisher_id"`
	ApproverID   *uint      `json:"approver_id"`
	RejectReason string     `gorm:"type:text" json:"reject_reason"`
	PlannedAt    *time.Time `json:"planned_at"`
	SubmittedAt  *time.Time `json:"submitted_at"`
	ApprovedAt   *time.Time `json:"approved_at"`
	PublishedAt  *time.Time `json:"published_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ReleaseApproval struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ReleaseID uint      `json:"release_id"`
	UserID    uint      `json:"user_id"`
	Action    string    `gorm:"size:32" json:"action"`
	Comment   string    `gorm:"type:text" json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type ReleaseItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ReleaseID uint      `json:"release_id"`
	ItemType  string    `gorm:"size:32" json:"item_type"`
	ItemID    uint      `json:"item_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TargetType string    `gorm:"size:32;index" json:"target_type"`
	TargetID   uint      `gorm:"index" json:"target_id"`
	UserID     uint      `json:"user_id"`
	Content    string    `gorm:"type:text" json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type Attachment struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TargetType   string    `gorm:"size:32;index" json:"target_type"`
	TargetID     uint      `gorm:"index" json:"target_id"`
	UserID       uint      `json:"user_id"`
	FileName     string    `gorm:"size:255" json:"file_name"`
	OriginalName string    `gorm:"size:255" json:"original_name"`
	MimeType     string    `gorm:"size:128" json:"mime_type"`
	MediaType    string    `gorm:"size:16" json:"media_type"`
	Size         int64     `json:"size"`
	URL          string    `gorm:"size:512" json:"url"`
	CreatedAt    time.Time `json:"created_at"`
}

type ActivityLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ProjectID  uint      `gorm:"index" json:"project_id"`
	TargetType string    `gorm:"size:32;index" json:"target_type"`
	TargetID   uint      `gorm:"index" json:"target_id"`
	UserID     uint      `json:"user_id"`
	Action     string    `gorm:"size:64" json:"action"`
	OldValue   string    `gorm:"size:255" json:"old_value"`
	NewValue   string    `gorm:"size:255" json:"new_value"`
	CreatedAt  time.Time `json:"created_at"`
}

// ProjectMilestone is a delivery stage in the project progress pipeline.
type ProjectMilestone struct {
	ID        uint                   `gorm:"primaryKey" json:"id"`
	ProjectID uint                   `gorm:"index;uniqueIndex:idx_project_milestone_key" json:"project_id"`
	Key       string                 `gorm:"size:64;uniqueIndex:idx_project_milestone_key" json:"key"`
	Title     string                 `gorm:"size:128" json:"title"`
	Sort      int                    `json:"sort"`
	Status    string                 `gorm:"size:16;default:todo" json:"status"` // todo | doing | done | blocked
	Weight    int                    `gorm:"default:1" json:"weight"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Items     []ProjectChecklistItem `gorm:"foreignKey:MilestoneID" json:"items,omitempty"`
}

// ProjectChecklistItem is a concrete task under a milestone.
type ProjectChecklistItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	MilestoneID uint      `gorm:"index;uniqueIndex:idx_milestone_item_key" json:"milestone_id"`
	Key         string    `gorm:"size:64;uniqueIndex:idx_milestone_item_key" json:"key"`
	Title       string    `gorm:"size:255" json:"title"`
	Status      string    `gorm:"size:16;default:todo" json:"status"` // todo | doing | done | blocked
	AssigneeID  *uint     `json:"assignee_id"`
	Note        string    `gorm:"size:512" json:"note"`
	Weight      int       `gorm:"default:1" json:"weight"`
	Sort        int       `json:"sort"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Assignee    *User     `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
}
