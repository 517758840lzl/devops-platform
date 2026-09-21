package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"devops-platform/internal/model"

	"gorm.io/gorm"
)

func NewShareToken() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// BuildShareDownloadURL returns a login-free download URL for WeChat.
// baseURL comes from ProjectSetting.NotifyPublicBaseURL（项目「构建通知 → 平台访问地址」）.
//
// TODO(deploy): 上线后把 NotifyPublicBaseURL 从局域网 IP 换成公网域名，例如 https://devops.example.com
// 推送里的「下载地址」= baseURL + /api/builds/share/{token}/download，域名不对手机就打不开。
func BuildShareDownloadURL(baseURL, shareToken string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	token := strings.TrimSpace(shareToken)
	if base == "" || token == "" {
		return ""
	}
	// TODO(deploy): 若生产环境前后端分离且无 /api 反代，这里可改为直连 API 域名
	return fmt.Sprintf("%s/api/builds/share/%s/download", base, token)
}

// NotifyBuildResult pushes build result to personal WeChat via Server酱 Turbo.
func NotifyBuildResult(db *gorm.DB, project model.Project, job model.BuildJob, success bool) {
	var setting model.ProjectSetting
	if err := db.First(&setting, "project_id = ?", project.ID).Error; err != nil {
		return
	}
	key := strings.TrimSpace(setting.ServerChanSendKey)
	if key == "" {
		return
	}

	status := "失败 ✗"
	if success {
		status = "成功 ✓"
	}
	title := fmt.Sprintf("[%s] 构建%s #%d", project.Name, status, job.BuildNumber)
	if job.BuildNumber == 0 {
		title = fmt.Sprintf("[%s] 构建%s", project.Name, status)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "- 项目：%s（%s）\n", project.Name, project.Code)
	if job.BuildNumber > 0 {
		fmt.Fprintf(&b, "- 构建号：#%d\n", job.BuildNumber)
	}
	if job.RepoName != "" {
		fmt.Fprintf(&b, "- 仓库：%s\n", job.RepoName)
	}
	if strings.TrimSpace(job.Branch) != "" {
		fmt.Fprintf(&b, "- 分支：%s\n", job.Branch)
	}
	if sha := strings.TrimSpace(job.CommitSHA); sha != "" {
		if len(sha) > 8 {
			sha = sha[:8]
		}
		fmt.Fprintf(&b, "- Commit：%s\n", sha)
	}
	fmt.Fprintf(&b, "- 结果：%s\n", status)

	base := strings.TrimRight(strings.TrimSpace(setting.NotifyPublicBaseURL), "/")
	if success {
		if link := BuildShareDownloadURL(base, job.ShareToken); link != "" {
			fmt.Fprintf(&b, "- 下载地址：%s\n", link)
			b.WriteString("- 手机点击链接即可下载安装包（无需登录）\n")
		} else if base != "" {
			fmt.Fprintf(&b, "- 平台：%s/builds\n", base)
			b.WriteString("- 请打开平台「打包构建」下载产物\n")
		} else {
			b.WriteString("- 请在「构建通知」填写平台访问地址后，推送将附带下载链接\n")
		}
	} else if base != "" {
		fmt.Fprintf(&b, "- 平台：%s/builds\n", base)
	}

	if err := SendServerChan(key, title, b.String()); err != nil {
		_ = db.Model(&job).Update("log", job.Log+"\n[微信推送失败] "+err.Error()).Error
	}
}

// SendServerChan posts to Server酱 Turbo API.
func SendServerChan(sendKey, title, desp string) error {
	sendKey = strings.TrimSpace(sendKey)
	if sendKey == "" {
		return fmt.Errorf("empty sendkey")
	}
	if title == "" {
		title = "通知"
	}
	endpoint := fmt.Sprintf("https://sctapi.ftqq.com/%s.send", sendKey)
	form := url.Values{}
	form.Set("title", title)
	form.Set("desp", desp)

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.PostForm(endpoint, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if strings.Contains(string(body), `"code":0`) || strings.Contains(string(body), `"code": 0`) {
		return nil
	}
	if strings.Contains(string(body), `"errno":0`) {
		return nil
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && !strings.Contains(string(body), `"code"`) {
		return nil
	}
	if strings.Contains(string(body), `"code"`) && !strings.Contains(string(body), `"code":0`) && !strings.Contains(string(body), `"code": 0`) {
		return fmt.Errorf("serverchan: %s", strings.TrimSpace(string(body)))
	}
	return nil
}
