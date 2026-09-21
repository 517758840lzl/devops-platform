export function defaultArtifactPath(projectCode, projectId) {
  const code = projectCode || `project_${projectId}`
  return `data/artifacts/${code}`
}

export function resolveBuildCommand(profile, platform, customCommand) {
  if (profile === 'custom') {
    return customCommand || ''
  }
  const mode = profile === 'develop' ? 'debug' : 'release'
  const base = 'flutter pub get && '
  if (platform === 'ios') {
    return `${base}flutter build ios --${mode} --no-codesign`
  }
  return `${base}flutter build apk --${mode}`
}

export function inferProfileFromCommand(cmd) {
  if (!cmd) return 'develop'
  if (cmd.includes('--debug')) return 'develop'
  if (cmd.includes('--release')) return 'release'
  return 'custom'
}

export function inferPlatformFromCommand(cmd) {
  if (!cmd) return 'android'
  if (cmd.includes('build ios')) return 'ios'
  return 'android'
}
