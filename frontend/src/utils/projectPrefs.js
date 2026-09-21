const KEY = 'devops_project_build_prefs'

function loadAll() {
  try {
    return JSON.parse(localStorage.getItem(KEY) || '{}')
  } catch {
    return {}
  }
}

export function getProjectPrefs(projectId) {
  if (!projectId) return {}
  return loadAll()[projectId] || {}
}

export function saveProjectPrefs(projectId, prefs) {
  if (!projectId) return
  const all = loadAll()
  all[projectId] = { ...(all[projectId] || {}), ...prefs }
  localStorage.setItem(KEY, JSON.stringify(all))
}

export const defaultBuildPrefs = {
  git_branch: 'main',
  build_profile: 'develop',
  build_platform: 'android',
}
