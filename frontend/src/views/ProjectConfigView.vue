<template>
  <div>
    <div class="page-header">
      <h2>
        项目配置
        <el-tag v-if="projectStore.current" size="small" type="info">{{ projectStore.current.name }}</el-tag>
      </h2>
      <el-tag v-if="isDirty && tab !== 'store'" type="warning" size="small">有未保存的修改</el-tag>
      <el-button v-if="tab !== 'store' && canEditProjectConfig" type="primary" :loading="saving" @click="save">保存配置</el-button>
    </div>

    <el-tabs v-model="tab" lazy>
      <el-tab-pane v-if="canEditProjectConfig" label="应用配置" name="identity">
        <el-form :model="form" label-width="140px" class="tab-form wide" :disabled="!canEditProjectConfig">
          <el-form-item label="应用显示名">
            <el-input v-model="form.app_display_name" placeholder="AppStrings.appTitle" />
          </el-form-item>
          <el-form-item label="App ID">
            <el-input v-model="form.app_id" placeholder="environment_config → acqChannel / 请求头 BridgeKey" />
            <div class="hint">如 Easy Money：<code>PrimeCreditLoan</code></div>
          </el-form-item>
          <el-form-item label="渠道序号">
            <el-input v-model="form.channel_index" placeholder="acqChannelIndex / 请求头 Position" />
            <div class="hint">通常为 <code>0</code>，对应 lib/core/config/environment_config.dart</div>
          </el-form-item>
          <el-form-item label="Android 包名">
            <el-input v-model="form.package_android" placeholder="android/app/build.gradle.kts → applicationId" />
          </el-form-item>
          <el-form-item label="iOS Bundle ID">
            <el-input v-model="form.bundle_ios" placeholder="ios Runner → PRODUCT_BUNDLE_IDENTIFIER" />
          </el-form-item>
          <el-form-item label="Apple App ID">
            <el-input v-model="form.apple_app_id" placeholder="App Store Connect 数字 ID / AppsFlyer iOS" />
            <div class="hint">app_constants.dart → afAppleAppId，如 <code>6808185084</code></div>
          </el-form-item>
          <el-form-item label="AppsFlyer Dev Key">
            <el-input v-model="form.appsflyer_dev_key" placeholder="app_constants.dart → afDevKey" />
          </el-form-item>
          <el-form-item label="API 请求秘钥">
            <el-input
              v-model="form.request_aes_key"
              type="password"
              show-password
              placeholder="request_security_config.dart → requestAesKey"
            />
            <div class="hint">请求体 AES 加密密钥，IV 取 UTF-8 前 16 字节；接口路径映射请到「接口混淆」Tab</div>
          </el-form-item>
          <el-form-item label="API 域名">
            <el-input v-model="form.api_base_url_prod" placeholder="environment_config → baseUrl（生产）" />
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <el-tab-pane label="需求与设计" name="product-design">
        <el-form :model="form" label-width="140px" class="tab-form wide" :disabled="!canEditProjectConfig">
          <el-divider content-position="left">需求文档</el-divider>
          <el-form-item label="PRD / 需求文档">
            <div class="link-row">
              <el-input
                v-model="form.requirement_doc_url"
                placeholder="飞书文档 / Notion / Google Docs 链接..."
              />
              <el-button v-if="form.requirement_doc_url" link type="primary" @click="openUrl(form.requirement_doc_url)">打开</el-button>
            </div>
            <div class="hint">产品需求说明、版本规划、交互说明等</div>
          </el-form-item>

          <el-divider content-position="left">Figma</el-divider>
          <el-form-item label="主设计稿">
            <div class="link-row">
              <el-input
                v-model="form.figma_design_url"
                placeholder="https://www.figma.com/design/xxxxx/..."
              />
              <el-button v-if="form.figma_design_url" link type="primary" @click="openUrl(form.figma_design_url)">打开</el-button>
            </div>
            <div v-if="figmaDesignFileKey" class="hint">
              File Key：<code>{{ figmaDesignFileKey }}</code>（可复制给 Cursor 做 design-to-code）
            </div>
          </el-form-item>
          <el-form-item label="FigJam 白板">
            <div class="link-row">
              <el-input
                v-model="form.figma_figjam_url"
                placeholder="https://www.figma.com/board/xxxxx/..."
              />
              <el-button v-if="form.figma_figjam_url" link type="primary" @click="openUrl(form.figma_figjam_url)">打开</el-button>
            </div>
            <div class="hint">流程图、评审记录（可选）</div>
          </el-form-item>

          <el-form-item label="备注">
            <el-input
              v-model="form.figma_notes"
              type="textarea"
              :rows="4"
              placeholder="如：v1.2 需求见 PRD 第 3 章；主设计稿 Page「Home」为首页；设计 @xxx"
            />
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <el-tab-pane label="协议链接" name="compliance">
        <el-form :model="form" label-width="140px" class="tab-form wide" :disabled="!canEditProjectConfig">
          <el-form-item label="官网">
            <div class="link-row">
              <el-input v-model="form.official_website_url_prod" placeholder="https://www.example.com/" />
              <el-button
                v-if="form.official_website_url_prod"
                link
                type="primary"
                @click="openUrl(form.official_website_url_prod)"
              >打开</el-button>
            </div>
            <!-- TODO(deploy): 示例域名，按项目替换为真实官网 -->
            <div class="hint">品牌官网 / 营销落地页，如 Easy Money：<code>https://www.bluebirdfintech.com/</code></div>
          </el-form-item>
          <el-form-item label="隐私协议">
            <div class="link-row">
              <el-input v-model="form.privacy_url_prod" placeholder="https://..." />
              <el-button
                v-if="form.privacy_url_prod"
                link
                type="primary"
                @click="openUrl(form.privacy_url_prod)"
              >打开</el-button>
            </div>
          </el-form-item>
          <el-form-item label="用户协议">
            <div class="link-row">
              <el-input v-model="form.terms_url_prod" placeholder="https://..." />
              <el-button
                v-if="form.terms_url_prod"
                link
                type="primary"
                @click="openUrl(form.terms_url_prod)"
              >打开</el-button>
            </div>
          </el-form-item>
          <el-form-item label="贷款合同">
            <div class="link-row">
              <el-input v-model="form.loan_contract_url_prod" placeholder="privacy_policy_config → loanAgreementUrl" />
              <el-button
                v-if="form.loan_contract_url_prod"
                link
                type="primary"
                @click="openUrl(form.loan_contract_url_prod)"
              >打开</el-button>
            </div>
            <div class="hint">借款合同页，如 <code>.../customer-contract/contract.html</code></div>
          </el-form-item>

          <div v-for="(link, idx) in extraLinks" :key="idx" class="link-block">
            <el-form-item :label="link.name || `链接 ${idx + 1}`">
              <div class="link-fields">
                <el-input v-model="link.name" placeholder="链接名称，如：帮助中心" class="link-name" />
                <div class="link-row">
                  <el-input v-model="link.url_prod" placeholder="正式 URL" />
                  <el-button
                    v-if="link.url_prod"
                    link
                    type="primary"
                    @click="openUrl(link.url_prod)"
                  >打开</el-button>
                  <el-button link type="danger" @click="extraLinks.splice(idx, 1)">删除</el-button>
                </div>
              </div>
            </el-form-item>
          </div>
          <el-form-item label=" ">
            <el-button v-if="canEditProjectConfig" size="small" @click="extraLinks.push({ name: '', url_prod: '', url_test: '' })">添加链接</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <el-tab-pane label="多语言" name="i18n">
        <el-form :model="form" label-width="140px" class="tab-form wide" :disabled="!canEditProjectConfig">
          <el-form-item label="默认语言">
            <el-select v-model="form.default_locale">
              <el-option v-for="l in localeOptions" :key="l" :label="localeLabel(l)" :value="l" />
            </el-select>
          </el-form-item>
          <el-form-item label="支持语言">
            <el-checkbox-group v-model="locales">
              <el-checkbox v-for="l in allLocales" :key="l.code" :value="l.code">{{ l.label }}</el-checkbox>
            </el-checkbox-group>
          </el-form-item>

          <el-divider content-position="left">开发定义 Key（源文件）</el-divider>
          <el-form-item label="app_strings.dart">
            <div class="editor-stack">
              <ConfigFileEditor
                v-if="projectStore.currentIdOrDefault"
                :key="`src-${projectStore.currentIdOrDefault}`"
                :project-id="projectStore.currentIdOrDefault"
                category="i18n_source"
                accept=".dart"
                placeholder="粘贴 lib/core/constants/app_strings.dart 完整内容..."
                hint="开发上传/粘贴 Key 定义文件。翻译按这里的 static const String key 进行，单文件建议不超过 2MB。"
                :rows="14"
                :readonly="!canEditProjectConfig"
                @saved="onSourceSaved"
              />
            </div>
          </el-form-item>

          <el-divider content-position="left">各语言翻译</el-divider>
          <el-form-item
            v-for="loc in translationLocales"
            :key="loc"
            :label="`${localeLabel(loc)} 翻译`"
          >
            <div class="editor-stack">
              <el-tag
                v-if="i18nDiff[loc]"
                size="small"
                class="diff-tag"
                :type="i18nDiff[loc].missing_keys?.length ? 'warning' : 'success'"
              >
                {{ i18nDiff[loc].translated || 0 }}/{{ i18nDiff[loc].total_keys || 0 }} keys
              </el-tag>
              <ConfigFileEditor
                :project-id="projectStore.currentIdOrDefault"
                category="i18n_locale"
                :locale="loc"
                accept=".dart"
                :placeholder="`粘贴 ${localeLabel(loc)} 的 app_strings 翻译文件...`"
                hint="业务/翻译同学上传同结构 dart 文件，Key 须与源文件一致。"
                :rows="10"
                :readonly="!canEditProjectConfig"
                @saved="() => refreshDiff(loc)"
              />
              <div v-if="i18nDiff[loc]?.missing_keys?.length" class="missing">
                缺失 {{ i18nDiff[loc].missing_keys.length }} 个 key，如：{{ i18nDiff[loc].missing_keys.slice(0, 5).join(', ') }}
              </div>
            </div>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <el-tab-pane v-if="canEditProjectConfig" label="接口混淆" name="api-map">
        <el-form label-width="140px" class="tab-form wide" :disabled="!canEditProjectConfig">
          <el-form-item label="api_constants.dart">
            <div class="editor-stack">
              <ConfigFileEditor
                v-if="projectStore.currentIdOrDefault"
                :key="`api-map-${projectStore.currentIdOrDefault}`"
                :project-id="projectStore.currentIdOrDefault"
                category="api_constants"
                accept=".dart"
                placeholder="static const String login = '/primecl/system/auth/enter';"
                hint="每行一个 static const String 接口标识 = '混淆后路径'，与代码中 ApiConstants 保持一致。"
                :rows="14"
                :readonly="!canEditProjectConfig"
                @saved="refreshApiMapPreview"
              />
            </div>
          </el-form-item>
          <el-form-item v-if="apiMapRows.length" label="映射预览">
            <el-table :data="apiMapRows" size="small" stripe max-height="420" class="api-map-table">
              <el-table-column prop="key" label="接口标识" min-width="200" />
              <el-table-column prop="path" label="混淆后路径" min-width="300" show-overflow-tooltip />
            </el-table>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <el-tab-pane label="短信词库" name="sms">
        <el-form :model="form" label-width="140px" class="tab-form wide" :disabled="!canEditProjectConfig">
          <el-form-item label="短信过滤词">
            <div class="editor-stack">
              <ConfigFileEditor
                v-if="projectStore.currentIdOrDefault"
                :key="`sms-words-${projectStore.currentIdOrDefault}`"
                :project-id="projectStore.currentIdOrDefault"
                category="sms_words"
                accept=".txt,.csv"
                placeholder="每行一个词，可直接从 Excel 粘贴&#10;loan scam&#10;free money&#10;guaranteed approval"
                hint="业务直接粘贴词库，一行一个词。# 开头为注释。内容存文件，数据库只记条数和大小。"
                :rows="16"
                :readonly="!canEditProjectConfig"
              />
            </div>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <el-tab-pane v-if="canEditProjectConfig" label="构建通知" name="notify">
        <el-form :model="form" label-width="140px" class="tab-form wide" :disabled="!canEditProjectConfig">
          <el-alert type="info" show-icon :closable="false" style="margin-bottom:16px">
            使用 <a href="https://sct.ftqq.com" target="_blank" rel="noopener">Server酱 Turbo</a>
            推送到个人微信。构建<strong>成功</strong>后会推送免登录下载链接；「发送测试推送」仅用于验证通道。
          </el-alert>
          <el-form-item label="SendKey">
            <el-input
              v-model="form.serverchan_send_key"
              type="password"
              show-password
              clearable
              placeholder="SCTxxxxxxxx"
            />
            <div class="hint">免费额度有限；密钥勿泄露。留空则关闭推送。</div>
          </el-form-item>
          <el-form-item label="平台访问地址">
            <!-- TODO(deploy): placeholder 里的局域网 IP 仅开发用；上线改为公网域名，构建推送下载地址会跟着变 -->
            <el-input
              v-model="form.notify_public_base_url"
              clearable
              placeholder="http://192.168.x.x:5173 或 https://devops.example.com"
            />
            <div class="hint">
              必填才能带下载链接。当前填手机能访问的前端地址；
              <!-- TODO(deploy): 上线后改成公网域名，推送「下载地址」= 本字段 + /api/builds/share/.../download -->
              上线后换成公网域名，微信里点开才能下包。
            </div>
          </el-form-item>
          <el-form-item label=" ">
            <el-button :loading="testingNotify" @click="onTestNotify">发送测试推送</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <el-tab-pane label="上架素材" name="store">
        <el-form label-width="140px" class="tab-form wide">
          <el-form-item label=" ">
            <el-button
              type="primary"
              plain
              size="small"
              :disabled="!iosAssetCount && !androidAssetCount"
              :loading="downloadingAll"
              @click="downloadStoreZip('all')"
            >
              一键下载全部素材
            </el-button>
            <span class="hint inline-hint">协同上传的文件可在此打包下载，按 android/、ios/ 分目录</span>
          </el-form-item>
          <el-form-item label="Android 进度">
            <span class="asset-count">{{ androidAssetCount }}/7（1 Logo + 5 截图 + 1 Banner）</span>
            <el-button
              size="small"
              :disabled="!androidAssetCount"
              :loading="downloadingPlatform === 'android'"
              @click="downloadStoreZip('android')"
            >
              一键下载 Android
            </el-button>
          </el-form-item>
          <el-form-item label="Android 素材">
            <StoreAssetsPanel
              v-if="projectStore.currentIdOrDefault"
              :key="`store-android-${projectStore.currentIdOrDefault}`"
              :project-id="projectStore.currentIdOrDefault"
              platform="android"
              :can-upload="canEditProjectConfig"
              :can-delete="canDeleteStoreAsset"
              @change="(n) => (androidAssetCount = n)"
            />
          </el-form-item>

          <el-divider content-position="left">iOS</el-divider>
          <el-form-item label="iOS 进度">
            <span class="asset-count">{{ iosAssetCount }}/6（1 Logo + 5 截图）</span>
            <el-button
              size="small"
              :disabled="!iosAssetCount"
              :loading="downloadingPlatform === 'ios'"
              @click="downloadStoreZip('ios')"
            >
              一键下载 iOS
            </el-button>
          </el-form-item>
          <el-form-item label="iOS 素材">
            <StoreAssetsPanel
              v-if="projectStore.currentIdOrDefault"
              :key="`store-ios-${projectStore.currentIdOrDefault}`"
              :project-id="projectStore.currentIdOrDefault"
              platform="ios"
              :can-upload="canEditProjectConfig"
              :can-delete="canDeleteStoreAsset"
              @change="(n) => (iosAssetCount = n)"
            />
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { onBeforeRouteLeave, useRoute } from 'vue-router'
import {
  downloadProjectStoreAssetsZip,
  getConfigFilePreview,
  getI18nDiff,
  getProjectSettings,
  listProjectStoreAssets,
  saveProjectSettings,
  testProjectNotify,
} from '../api'
import ConfigFileEditor from '../components/ConfigFileEditor.vue'
import StoreAssetsPanel from '../components/StoreAssetsPanel.vue'
import { useProjectPermission } from '../composables/useProjectPermission'
import { useProjectStore } from '../stores/project'
import { ElMessage, ElMessageBox } from 'element-plus'

const route = useRoute()
const projectStore = useProjectStore()
const { canEditProjectConfig, canDeleteStoreAsset, loaded, refresh: refreshPermission } = useProjectPermission()
const allTabs = ['identity', 'product-design', 'compliance', 'i18n', 'api-map', 'sms', 'notify', 'store']
const developerOnlyTabs = ['identity', 'api-map', 'notify']
const defaultPublicTab = 'product-design'

function resolveTab(raw) {
  const name = raw === 'figma' ? 'product-design' : raw
  if (!allTabs.includes(name)) {
    return canEditProjectConfig.value ? 'identity' : defaultPublicTab
  }
  if (developerOnlyTabs.includes(name) && !canEditProjectConfig.value) {
    return defaultPublicTab
  }
  return name
}

const initialTab = route.query.tab === 'figma' ? 'product-design' : route.query.tab
const tab = ref(allTabs.includes(initialTab) ? initialTab : 'identity')

function ensureAllowedTab() {
  tab.value = resolveTab(tab.value)
}
const apiMapRows = ref([])
const iosAssetCount = ref(0)
const androidAssetCount = ref(0)
const downloadingAll = ref(false)
const downloadingPlatform = ref('')
const saving = ref(false)
const testingNotify = ref(false)

async function downloadStoreZip(platform) {
  if (!projectStore.currentIdOrDefault) return
  if (platform === 'all') downloadingAll.value = true
  else downloadingPlatform.value = platform
  try {
    await downloadProjectStoreAssetsZip(projectStore.currentIdOrDefault, platform)
    ElMessage.success('下载已开始')
  } catch (err) {
    ElMessage.error(err.message || '下载失败')
  } finally {
    downloadingAll.value = false
    downloadingPlatform.value = ''
  }
}

const preservedFields = reactive({
  url_replace_rules: '[]',
  sms_filter_words: '[]',
})
const savedSnapshot = ref('')

function buildSnapshot() {
  return JSON.stringify({
    form: { ...form },
    extraLinks: extraLinks.value,
    locales: locales.value,
  })
}

const isDirty = computed(() => savedSnapshot.value !== '' && buildSnapshot() !== savedSnapshot.value)

async function confirmLeaveIfDirty() {
  if (!isDirty.value) return true
  try {
    await ElMessageBox.confirm('当前有未保存的文字配置修改，确定离开吗？', '未保存的修改', {
      type: 'warning',
      confirmButtonText: '离开',
      cancelButtonText: '留在此页',
    })
    return true
  } catch {
    return false
  }
}

function onBeforeUnload(e) {
  if (isDirty.value) {
    e.preventDefault()
    e.returnValue = ''
  }
}

async function refreshAssetCount() {
  if (!projectStore.currentIdOrDefault) {
    iosAssetCount.value = 0
    androidAssetCount.value = 0
    return
  }
  const [iosList, androidList] = await Promise.all([
    listProjectStoreAssets(projectStore.currentIdOrDefault, 'ios'),
    listProjectStoreAssets(projectStore.currentIdOrDefault, 'android'),
  ])
  iosAssetCount.value = iosList.length
  androidAssetCount.value = androidList.length
}
const form = reactive({
  app_display_name: '', app_id: '', channel_index: '0',
  package_android: '', bundle_ios: '', apple_app_id: '', appsflyer_dev_key: '',
  request_aes_key: '', disable_enc_body: 'true',
  api_base_url_prod: '', api_base_url_test: '',
  official_website_url_prod: '', official_website_url_test: '',
  privacy_url_prod: '', privacy_url_test: '', terms_url_prod: '', terms_url_test: '',
  loan_contract_url_prod: '', loan_contract_url_test: '',
  requirement_doc_url: '', figma_design_url: '', figma_figjam_url: '', figma_notes: '',
  default_locale: 'zh',
  serverchan_send_key: '',
  notify_public_base_url: '',
})

const figmaDesignFileKey = computed(() => parseFigmaFileKey(form.figma_design_url))

function parseFigmaFileKey(url) {
  if (!url) return ''
  const m = url.match(/figma\.com\/(?:design|file|board|make)\/([a-zA-Z0-9]+)/)
  return m?.[1] || ''
}

function openUrl(url) {
  const raw = String(url || '').trim()
  if (!raw) return
  const href = /^https?:\/\//i.test(raw) ? raw : `https://${raw}`
  window.open(href, '_blank', 'noopener')
}

const extraLinks = ref([])
const locales = ref(['zh', 'en'])
const i18nDiff = ref({})
const allLocales = [
  { code: 'zh', label: '中文' },
  { code: 'en', label: 'English' },
  { code: 'fr', label: 'Français' },
  { code: 'es', label: 'Español' },
  { code: 'pt', label: 'Português' },
  { code: 'sw', label: 'Kiswahili' },
]

const localeOptions = computed(() => locales.value.length ? locales.value : ['zh'])
const translationLocales = computed(() => locales.value.filter((l) => l !== form.default_locale))

function localeLabel(code) {
  return allLocales.find((l) => l.code === code)?.label || code
}

function parseJSON(raw, fallback) {
  try { return JSON.parse(raw || '') } catch { return fallback }
}

async function load() {
  if (isDirty.value) {
    const ok = await confirmLeaveIfDirty()
    if (!ok) return
  }
  const data = await getProjectSettings(projectStore.currentIdOrDefault)
  preservedFields.url_replace_rules = data.url_replace_rules || '[]'
  preservedFields.sms_filter_words = data.sms_filter_words || '[]'
  Object.assign(form, {
    app_display_name: data.app_display_name || '',
    app_id: data.app_id || '',
    channel_index: data.channel_index || data.channel_code || '0',
    package_android: data.package_android || '',
    bundle_ios: data.bundle_ios || '',
    apple_app_id: data.apple_app_id || '',
    appsflyer_dev_key: data.appsflyer_dev_key || '',
    request_aes_key: data.request_aes_key || '',
    disable_enc_body: data.disable_enc_body || 'true',
    api_base_url_prod: data.api_base_url_prod || '',
    api_base_url_test: data.api_base_url_test || '',
    official_website_url_prod: data.official_website_url_prod || '',
    official_website_url_test: data.official_website_url_test || '',
    privacy_url_prod: data.privacy_url_prod || '',
    privacy_url_test: data.privacy_url_test || '',
    terms_url_prod: data.terms_url_prod || '',
    terms_url_test: data.terms_url_test || '',
    loan_contract_url_prod: data.loan_contract_url_prod || '',
    loan_contract_url_test: data.loan_contract_url_test || '',
    requirement_doc_url: data.requirement_doc_url || '',
    figma_design_url: data.figma_design_url || '',
    figma_figjam_url: data.figma_figjam_url || '',
    figma_notes: data.figma_notes || '',
    default_locale: data.default_locale || 'zh',
    serverchan_send_key: data.serverchan_send_key || '',
    notify_public_base_url: data.notify_public_base_url || '',
  })
  extraLinks.value = parseJSON(data.extra_compliance_links, [])
  locales.value = parseJSON(data.supported_locales, ['zh', 'en'])
  savedSnapshot.value = buildSnapshot()
  await refreshAllDiff()
  if (canEditProjectConfig.value) {
    await refreshApiMapPreview()
  } else {
    apiMapRows.value = []
  }
}

async function refreshDiff(locale) {
  try {
    i18nDiff.value[locale] = await getI18nDiff(projectStore.currentIdOrDefault, locale)
  } catch {
    i18nDiff.value[locale] = null
  }
}

async function refreshAllDiff() {
  await Promise.all(translationLocales.value.map((loc) => refreshDiff(loc)))
}

async function onSourceSaved() {
  await refreshAllDiff()
}

async function refreshApiMapPreview() {
  if (!canEditProjectConfig.value || !projectStore.currentIdOrDefault) {
    apiMapRows.value = []
    return
  }
  try {
    const data = await getConfigFilePreview(projectStore.currentIdOrDefault, 'api_constants')
    const items = data.items || {}
    apiMapRows.value = Object.entries(items)
      .map(([key, path]) => ({ key, path }))
      .sort((a, b) => a.key.localeCompare(b.key))
  } catch {
    apiMapRows.value = []
  }
}

async function onTestNotify() {
  if (!projectStore.currentIdOrDefault) return
  testingNotify.value = true
  try {
    await testProjectNotify(projectStore.currentIdOrDefault, {
      serverchan_send_key: form.serverchan_send_key,
    })
    ElMessage.success('已发送，请查看微信')
  } catch (e) {
    ElMessage.error(e.message || '推送失败')
  } finally {
    testingNotify.value = false
  }
}

async function save() {
  if (!canEditProjectConfig.value) {
    ElMessage.warning('当前角色无项目配置编辑权限')
    return
  }
  saving.value = true
  try {
    await saveProjectSettings(projectStore.currentIdOrDefault, {
      ...form,
      url_replace_rules: preservedFields.url_replace_rules,
      extra_compliance_links: JSON.stringify(extraLinks.value),
      supported_locales: JSON.stringify(locales.value),
      sms_filter_words: preservedFields.sms_filter_words,
    })
    savedSnapshot.value = buildSnapshot()
    ElMessage.success('配置已保存')
  } finally {
    saving.value = false
  }
}

onBeforeRouteLeave(async (_to, _from, next) => {
  if (await confirmLeaveIfDirty()) next()
  else next(false)
})

onMounted(async () => {
  await refreshPermission()
  ensureAllowedTab()
  window.addEventListener('beforeunload', onBeforeUnload)
  await load()
  await refreshAssetCount()
})
onBeforeUnmount(() => {
  window.removeEventListener('beforeunload', onBeforeUnload)
})
watch(() => projectStore.currentId, async () => {
  await refreshPermission()
  ensureAllowedTab()
  await load()
  await refreshAssetCount()
})
watch(translationLocales, refreshAllDiff)
watch(tab, (v) => { if (v === 'api-map' && canEditProjectConfig.value) refreshApiMapPreview() })
watch(canEditProjectConfig, () => ensureAllowedTab())
watch(loaded, () => ensureAllowedTab())
watch(() => route.query.tab, (v) => {
  tab.value = resolveTab(v)
})
</script>

<style scoped>
.tab-form { max-width: 640px; margin-top: 8px; }
.tab-form.wide { max-width: 960px; }
.tab-form.wide :deep(.el-form-item__content) {
  flex: 1;
  min-width: 0;
}
.link-block { margin-bottom: 4px; }
.link-fields { display: flex; flex-direction: column; gap: 8px; width: 100%; }
.link-name { max-width: 280px; }
.link-row { display: flex; align-items: center; gap: 8px; width: 100%; }
.link-row .el-input { flex: 1; }
.editor-stack {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.diff-tag { align-self: flex-start; }
.missing { font-size: 12px; color: #e6a23c; line-height: 1.5; }
.asset-count { font-size: 13px; color: #606266; margin-right: 12px; }
.inline-hint { margin-left: 12px; }
.hint { font-size: 12px; color: #909399; margin-top: 6px; line-height: 1.5; }
.hint code { background: #f4f4f5; padding: 1px 4px; border-radius: 3px; }
.api-map-table { width: 100%; }
</style>
