export function formatActivityAction(a) {
  if (!a) return '-'
  const map = {
    create: a.target_type === 'release' ? `创建发布 ${a.new_value || ''}`.trim() : `创建：${a.new_value || ''}`.trim(),
    status_change: `状态变更：${a.old_value || '-'} → ${a.new_value || '-'}`,
    comment: a.new_value ? `评论：${truncate(a.new_value, 80)}` : '添加了评论',
    submit_approval: '提交审批',
    approve: '审批通过',
    reject: '驳回发布',
    publish: '执行发布',
    add_member: `添加成员（${a.new_value || '-'}）`,
    update_member_role: `角色变更：${a.old_value || '-'} → ${a.new_value || '-'}`,
    remove_member: '移除成员',
    update_settings: a.new_value ? `更新项目配置（${a.new_value}）` : '更新项目配置',
    update_config_file: `更新配置文件 ${a.old_value || ''}${a.new_value ? ` / ${a.new_value}` : ''}`.trim(),
    upload_store_asset: `上传上架素材 ${a.old_value || ''} ${a.new_value || ''}`.trim(),
    delete_store_asset: `删除上架素材 ${a.old_value || ''} ${a.new_value || ''}`.trim(),
    trigger_build: `触发构建${a.old_value ? `（${a.old_value}）` : ''}`,
    build_success: `构建成功${a.new_value ? ` · ${String(a.new_value).slice(0, 8)}` : ''}`,
    build_failed: `构建失败${a.new_value ? `：${truncate(a.new_value, 60)}` : ''}`,
    delete_build_artifacts: `删除构建产物 ${a.old_value || ''}`.trim(),
    upload_attachment: a.new_value ? `上传附件：${a.new_value}` : '上传附件',
    update_progress: formatProgressUpdate(a),
    apply_progress_hints: a.new_value ? `根据缺口回退检查项（${a.new_value} 项）` : '根据缺口回退检查项',
  }
  return map[a.action] || a.action
}

function formatProgressUpdate(a) {
  const title = a.old_value ? String(a.old_value).split(' · ')[0] : '检查项'
  if (a.new_value) {
    return `更新进度：${title} → ${a.new_value}`
  }
  return `更新进度：${title}`
}

export function formatActivityTarget(a) {
  if (!a) return '-'
  if (a.target_label) return a.target_label
  const typeMap = {
    project: '项目',
    issue: 'Bug/任务',
    release: '发布',
    requirement: '需求',
  }
  const type = typeMap[a.target_type] || a.target_type || '-'
  if (a.target_type === 'project' && (a.project_name || a.project_code)) {
    if (a.project_name && a.project_code) return `${a.project_name}（${a.project_code}）`
    return a.project_name || a.project_code
  }
  return a.target_id ? `${type} #${a.target_id}` : type
}

function truncate(s, n) {
  const t = String(s || '')
  return t.length > n ? `${t.slice(0, n)}…` : t
}
