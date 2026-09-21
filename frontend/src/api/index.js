import request from './request'

export const login = (data) => request.post('/auth/login', data)
export const getMe = () => request.get('/auth/me')
export const getDashboard = (params) => request.get('/dashboard/summary', { params })
export const listUsers = () => request.get('/users')

export const listGitBranches = (gitUrl) => request.get('/git/branches', { params: { url: gitUrl } })
export const listProjects = (params) => request.get('/projects', { params })
export const getProject = (id) => request.get(`/projects/${id}`)
export const createProject = (data) => request.post('/projects', data)
export const updateProject = (id, data) => request.put(`/projects/${id}`, data)
export const getProjectSettings = (projectId) => request.get(`/projects/${projectId}/settings`)
export const saveProjectSettings = (projectId, data) => request.put(`/projects/${projectId}/settings`, data)

export const listConfigFiles = (projectId, params) => request.get(`/projects/${projectId}/config-files`, { params })
export const getConfigFileContent = (projectId, category, locale = '') =>
  request.get(`/projects/${projectId}/config-files/${category}/content`, { params: locale ? { locale } : {} })
export const getConfigFilePreview = (projectId, category, locale = '') =>
  request.get(`/projects/${projectId}/config-files/${category}/preview`, { params: locale ? { locale } : {} })
export const getI18nDiff = (projectId, locale) =>
  request.get(`/projects/${projectId}/config-files/i18n/diff`, { params: { locale } })
export const getDefaultSmsTemplate = (projectId) =>
  request.get(`/projects/${projectId}/config-files/sms/default-template`)

export function uploadConfigFile(projectId, category, file, locale = '') {
  const form = new FormData()
  form.append('file', file)
  if (locale) form.append('locale', locale)
  const token = localStorage.getItem('token')
  return fetch(`/api/projects/${projectId}/config-files/${category}/upload`, {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: form,
  }).then(async (res) => {
    const body = await res.json()
    if (body.code !== 0) throw new Error(body.message || '上传失败')
    return body.data
  })
}

export function pasteConfigFile(projectId, category, data) {
  return request.post(`/projects/${projectId}/config-files/${category}/paste`, data)
}

export const listMembers = (projectId) => request.get(`/projects/${projectId}/members`)
export const addMember = (projectId, data) => request.post(`/projects/${projectId}/members`, data)
export const updateMember = (projectId, memberId, data) => request.put(`/projects/${projectId}/members/${memberId}`, data)
export const removeMember = (projectId, memberId) => request.delete(`/projects/${projectId}/members/${memberId}`)

export const listRequirements = (params) => request.get('/requirements', { params })
export const createRequirement = (data) => request.post('/requirements', data)
export const updateRequirement = (id, data) => request.put(`/requirements/${id}`, data)

export const listIterations = (params) => request.get('/iterations', { params })
export const getIterationKanban = (id) => request.get(`/iterations/${id}/kanban`)
export const createIteration = (data) => request.post('/iterations', data)

export const listIssues = (params) => request.get('/issues', { params })
export const getIssue = (id) => request.get(`/issues/${id}`)
export const createIssue = (data) => request.post('/issues', data)
export const updateIssue = (id, data) => request.put(`/issues/${id}`, data)
export const patchIssueStatus = (id, status) => request.patch(`/issues/${id}/status`, { status })

export const listReleases = (params) => request.get('/releases', { params })
export const getRelease = (id) => request.get(`/releases/${id}`)
export const createRelease = (data) => request.post('/releases', data)
export const updateRelease = (id, data) => request.put(`/releases/${id}`, data)
export const submitRelease = (id, data) => request.post(`/releases/${id}/submit`, data)
export const approveRelease = (id, data) => request.post(`/releases/${id}/approve`, data)
export const rejectRelease = (id, data) => request.post(`/releases/${id}/reject`, data)
export const publishRelease = (id, data) => request.post(`/releases/${id}/publish`, data)

export const listProjectStoreAssets = (projectId, platform) =>
  request.get(`/projects/${projectId}/store-assets`, { params: platform ? { platform } : {} })
export const deleteProjectStoreAsset = (projectId, assetId) => request.delete(`/projects/${projectId}/store-assets/${assetId}`)

export async function downloadProjectStoreAssetsZip(projectId, platform = 'all') {
  const token = localStorage.getItem('token')
  const res = await fetch(`/api/projects/${projectId}/store-assets/download?platform=${platform}`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.message || '下载失败')
  }
  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `store-assets-${platform}.zip`
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

export async function downloadProjectStoreAsset(projectId, assetId, filename) {
  const token = localStorage.getItem('token')
  const res = await fetch(`/api/projects/${projectId}/store-assets/${assetId}/download`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) throw new Error('下载失败')
  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename || 'asset'
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}
export function uploadProjectStoreAsset(projectId, platform, slot, file) {
  const form = new FormData()
  form.append('file', file)
  form.append('platform', platform)
  form.append('slot', slot)
  const token = localStorage.getItem('token')
  return fetch(`/api/projects/${projectId}/store-assets`, {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: form,
  }).then(async (res) => {
    const body = await res.json()
    if (body.code !== 0) throw new Error(body.message || '上传失败')
    return body.data
  })
}

export const listReleaseStoreAssets = (releaseId) => request.get(`/releases/${releaseId}/store-assets`)
export const deleteReleaseStoreAsset = (releaseId, assetId) => request.delete(`/releases/${releaseId}/store-assets/${assetId}`)
export function uploadReleaseStoreAsset(releaseId, slot, file) {
  const form = new FormData()
  form.append('file', file)
  form.append('slot', slot)
  const token = localStorage.getItem('token')
  return fetch(`/api/releases/${releaseId}/store-assets`, {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: form,
  }).then(async (res) => {
    const body = await res.json()
    if (body.code !== 0) throw new Error(body.message || '上传失败')
    return body.data
  })
}

export const listComments = (params) => request.get('/comments', { params })
export const createComment = (data) => request.post('/comments', data)
export const deleteComment = (id) => request.delete(`/comments/${id}`)

export const listActivities = (params) => request.get('/activities', { params })

export const getProjectProgress = (projectId) => request.get(`/projects/${projectId}/progress`)
export const patchChecklistItem = (projectId, itemId, data) =>
  request.patch(`/projects/${projectId}/checklist/${itemId}`, data)
export const applyProgressHints = (projectId) => request.post(`/projects/${projectId}/progress/apply-hints`)

export const listAttachments = (params) => request.get('/attachments', { params })
export const deleteAttachment = (id) => request.delete(`/attachments/${id}`)

export const browseDirs = (path = '') => request.get('/fs/browse', { params: path ? { path } : {} })

export const listBuilds = (params) => request.get('/builds', { params })
export const getBuild = (id) => request.get(`/builds/${id}`)
export const listBuildArtifacts = (buildId) => request.get(`/builds/${buildId}/artifacts`)
export const deleteBuildArtifacts = (buildId) => request.delete(`/builds/${buildId}/artifacts`)
export const triggerProjectBuild = (projectId, data) => request.post(`/projects/${projectId}/builds`, data)

function parseDownloadFilename(res, fallback) {
  const disposition = res.headers.get('Content-Disposition') || ''
  const match = disposition.match(/filename="?([^";]+)"?/)
  return match?.[1] || fallback
}

export async function downloadBuildArtifacts(buildId) {
  const token = localStorage.getItem('token')
  const res = await fetch(`/api/builds/${buildId}/download`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.message || '下载失败')
  }
  const filename = parseDownloadFilename(res, `build-${buildId}-artifacts.zip`)
  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
  return { filename, size: blob.size }
}

export async function downloadBuildArtifactFile(buildId, filePath, filename) {
  const token = localStorage.getItem('token')
  const res = await fetch(`/api/builds/${buildId}/download-file?path=${encodeURIComponent(filePath)}`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) throw new Error('下载失败')
  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename || filePath.split('/').pop() || 'artifact'
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

export function uploadAttachment(file, targetType = '', targetId = 0) {
  const form = new FormData()
  form.append('file', file)
  if (targetType) form.append('target_type', targetType)
  if (targetId) form.append('target_id', String(targetId))
  const token = localStorage.getItem('token')
  return fetch('/api/attachments/upload', {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: form,
  }).then(async (res) => {
    const body = await res.json()
    if (body.code !== 0) throw new Error(body.message || '上传失败')
    return body.data
  })
}
