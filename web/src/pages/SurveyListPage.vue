<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  NButton,
  NCard,
  NEmpty,
  NInput,
  NModal,
  NSpin,
  NSwitch,
  NTag,
  useDialog,
  useMessage,
} from 'naive-ui'
import { api, getToken, setToken } from '../api'

const router = useRouter()
const message = useMessage()
const dialog = useDialog()

const loading = ref(true)
const surveys = ref([])
const me = ref(null)
const keyword = ref('')
const showCreate = ref(false)
const creating = ref(false)
const form = reactive({ title: '', description: '', anonymous: true })

const statusMeta = {
  draft: { text: '草稿', type: 'default' },
  published: { text: '已发布', type: 'success' },
  closed: { text: '已关闭', type: 'warning' },
}

const filtered = computed(() => {
  const k = keyword.value.trim()
  if (!k) return surveys.value
  return surveys.value.filter((s) => s.title.includes(k) || (s.description || '').includes(k))
})

async function load() {
  loading.value = true
  try {
    const [list, user] = await Promise.all([api.surveys(), api.me()])
    surveys.value = list.list || []
    me.value = user
  } catch (e) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}
onMounted(load)

async function create() {
  if (!form.title.trim()) return message.warning('问卷标题不能为空')
  creating.value = true
  try {
    const sv = await api.createSurvey({
      title: form.title.trim(),
      description: form.description.trim(),
      anonymous: form.anonymous,
    })
    showCreate.value = false
    form.title = ''
    form.description = ''
    message.success('已创建，去添加题目吧')
    router.push(`/surveys/${sv.id}/edit`)
  } catch (e) {
    message.error(e.message)
  } finally {
    creating.value = false
  }
}

function goEdit(id) {
  router.push(`/surveys/${id}/edit`)
}
function goStats(id) {
  router.push(`/surveys/${id}/stats`)
}
function fillUrl(id) {
  return `${location.origin}/s/${id}`
}
async function copyLink(s) {
  if (s.status !== 'published') return message.warning('先发布问卷，再复制填写链接')
  try {
    await navigator.clipboard.writeText(fillUrl(s.id))
    message.success('填写链接已复制')
  } catch {
    message.info(fillUrl(s.id))
  }
}

async function togglePublish(s) {
  const next = s.status === 'published' ? 'closed' : 'published'
  const label = next === 'published' ? '发布' : '关闭'
  dialog.info({
    title: `${label}问卷`,
    content:
      next === 'published'
        ? `发布后「${s.title}」就可以通过链接填写了。`
        : `关闭后「${s.title}」将暂停收集答卷。`,
    positiveText: label,
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await api.setStatus(s.id, next)
        message.success(`已${label}`)
        load()
      } catch (e) {
        message.error(e.message)
      }
    },
  })
}

function remove(s) {
  dialog.warning({
    title: '删除问卷',
    content: `确定删除「${s.title}」吗？该问卷的答卷数据会一起删除，无法恢复。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await api.deleteSurvey(s.id)
        message.success('已删除')
        load()
      } catch (e) {
        message.error(e.message)
      }
    },
  })
}

function logout() {
  setToken('')
  router.replace('/login')
}

function fmt(iso) {
  if (!iso) return '-'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '-'
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h1 class="page-title">我的问卷</h1>
        <div class="muted">
          共 {{ surveys.length }} 份<template v-if="me"> · 你好，{{ me.nickname }}</template>
        </div>
      </div>
      <div class="row">
        <n-input v-model:value="keyword" placeholder="搜索标题" clearable style="width: 180px" />
        <n-button type="primary" @click="showCreate = true">新建问卷</n-button>
        <n-button quaternary @click="logout">退出</n-button>
      </div>
    </div>

    <n-spin :show="loading">
      <n-empty v-if="!loading && !filtered.length" description="还没有问卷，点右上角新建一份" style="padding: 60px 0" />

      <div class="list">
        <n-card v-for="s in filtered" :key="s.id" class="item" size="small">
          <div class="item-main">
            <div class="grow">
              <div class="row" style="gap: 8px; flex-wrap: wrap">
                <a class="item-title ellipsis" @click="goEdit(s.id)">{{ s.title }}</a>
                <n-tag size="small" :type="statusMeta[s.status]?.type || 'default'">
                  {{ statusMeta[s.status]?.text || s.status }}
                </n-tag>
                <n-tag v-if="s.anonymous" size="small" :bordered="false">匿名</n-tag>
              </div>
              <div class="muted ellipsis" style="margin-top: 4px">
                {{ s.description || '暂无描述' }}
              </div>
              <div class="muted" style="margin-top: 8px">
                {{ s.responseCount || 0 }} 份答卷 · 更新于 {{ fmt(s.updatedAt) }}
              </div>
            </div>

            <div class="item-actions">
              <n-button size="small" @click="goEdit(s.id)">编辑</n-button>
              <n-button size="small" @click="goStats(s.id)">统计</n-button>
              <n-button size="small" quaternary @click="copyLink(s)">复制链接</n-button>
              <n-button size="small" quaternary type="primary" @click="togglePublish(s)">
                {{ s.status === 'published' ? '关闭' : '发布' }}
              </n-button>
              <n-button size="small" quaternary type="error" @click="remove(s)">删除</n-button>
            </div>
          </div>
        </n-card>
      </div>
    </n-spin>

    <n-modal v-model:show="showCreate" preset="card" title="新建问卷" style="max-width: 520px">
      <div class="field">
        <div class="label">标题</div>
        <n-input v-model:value="form.title" placeholder="例如：数据结构课程反馈" maxlength="100" show-count />
      </div>
      <div class="field">
        <div class="label">描述</div>
        <n-input
          v-model:value="form.description"
          type="textarea"
          :rows="3"
          placeholder="简单说明这份问卷的用途（可不填）"
        />
      </div>
      <div class="field row" style="justify-content: space-between">
        <div>
          <div class="label">匿名填写</div>
          <div class="muted">开启后不记录填写人身份</div>
        </div>
        <n-switch v-model:value="form.anonymous" />
      </div>
      <template #footer>
        <div class="row" style="justify-content: flex-end">
          <n-button @click="showCreate = false">取消</n-button>
          <n-button type="primary" :loading="creating" @click="create">创建</n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.item {
  border-radius: 14px;
}
.item-main {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}
.item-title {
  font-weight: 600;
  font-size: 15px;
  cursor: pointer;
  color: var(--text);
}
.item-title:hover {
  color: var(--primary);
}
.item-actions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.field {
  margin-bottom: 16px;
}
.label {
  font-size: 13px;
  color: var(--muted);
  margin-bottom: 6px;
}
@media (max-width: 720px) {
  .item-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>
