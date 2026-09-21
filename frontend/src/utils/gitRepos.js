export function parseGitRepos(raw) {
  if (!raw) return []
  try {
    const list = JSON.parse(raw)
    return Array.isArray(list) ? list.filter((r) => r?.git_url?.trim()) : []
  } catch {
    return []
  }
}

export function listProjectGitRepos(project) {
  if (!project) return []
  const repos = []
  if (project.git_url?.trim()) {
    repos.push({
      key: project.git_url,
      name: '主仓库',
      git_url: project.git_url,
      git_branch: project.git_branch || 'main',
      build_work_dir: project.build_work_dir || '',
      isPrimary: true,
    })
  }
  parseGitRepos(project.git_repos).forEach((r, idx) => {
    if (repos.some((x) => x.git_url === r.git_url)) return
    repos.push({
      key: r.git_url || `extra-${idx}`,
      name: r.name || r.git_url,
      git_url: r.git_url,
      git_branch: r.git_branch || 'main',
      build_work_dir: r.build_work_dir || '',
      isPrimary: false,
    })
  })
  return repos
}
