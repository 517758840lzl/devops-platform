# ZZYD流水线

桌面演示项目，前后端分离：

- `backend/` — Go + Gin + GORM + SQLite
- `frontend/` — Vue 3 + Vite + Element Plus + Pinia

## 功能

- 登录鉴权（JWT）
- 工作台统计
- 项目 / 需求 / 任务 / Bug / 发布
- **迭代看板**（拖拽变更状态）
- **评论 & 操作日志**（任务/Bug/发布/项目）
- **发布审批流**（草稿 → 待审批 → 通过/驳回 → 发布）
- **成员权限**（owner/admin/developer/tester/viewer）

## 默认账号

| 账号 | 密码 | 说明 |
|------|------|------|
| admin | admin123 | 系统管理员，可审批发布 |
| dev | dev123 | 项目开发角色 |
| tester | test123 | 项目测试角色 |

## 启动

```bash
# 后端
cd backend
go mod tidy
go run ./cmd/server

# 前端（新终端）
cd frontend
npm install
npm run dev
```

- 后端：http://localhost:8080
- 前端：http://localhost:5173

> TODO(deploy)：上线后换成公网域名；「构建通知 → 平台访问地址」也要同步改，微信推送下载链接依赖该地址。

## 重置演示数据

删除 `backend/data/devops.db` 后重启后端，将重新 seed 示例数据。
