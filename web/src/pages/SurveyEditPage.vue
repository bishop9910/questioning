<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NButton,
  NCard,
  NEmpty,
  NInput,
  NSelect,
  NSwitch,
  NTag,
  useDialog,
  useMessage,
} from 'naive-ui'
import { api } from '../api'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()

const surveyId = Number(route.params.id)
const survey = ref(null)
const questions = ref([])
const selectedKey = ref(null)
const loading = ref(true)
const saving = ref(false)
const history = ref([])
const future = ref([])

const titleForm = reactive({ title: '', description: '', anonymous: false })

let uid = 1
const nkey = (p) => `${p}${uid++}`

const typeOptions = [
  { label: '单选题', value: 'single' },
  { label: '多选题', value: 'multiple' },
  { label: '填空题', value: 'text' },
  { label: '评分题（1-5）', value: 'rating' },
]
const typeText = { single: '单选', multiple: '多选', text: '填空', rating: '评分' }
const statusText = { draft: '草稿', published: '已发布', closed: '已关闭' }

function fromServer(list) {
  return (list || []).map((q) => ({
    key: nkey('q'),
    id: q.id || 0,
    type: q.type,
    title: q.title,
    description: q.description || '',
    required: !!q.required,
    options: (q.options || []).map((o) => ({
      key: nkey('o'),
      id: o.id || 0,
      label: o.label,
      nextQuestionId: o.nextQuestionId || 0,
    })),
  }))
}

function applySurvey(sv) {
  survey.value = sv
  titleForm.title = sv.title
  titleForm.description = sv.description || ''
  titleForm.anonymous = !!sv.anonymous
  const prev = selectedKey.value
  questions.value = fromServer(sv.questions)
  selectedKey.value = questions.value.length
    ? prev && questions.value.some((q) => q.key === prev)
      ? prev
      : questions.value[0].key
    : null
}

async function load() {
  loading.value = true
  try {
    applySurvey(await api.survey(surveyId))
  } catch (e) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}
onMounted(() => {
  load()
  window.addEventListener('keydown', onKey)
})
onBeforeUnmount(() => window.removeEventListener('keydown', onKey))

function onKey(e) {
  if (!(e.ctrlKey || e.metaKey) || e.key.toLowerCase() !== 'z') return
  const tag = (e.target && e.target.tagName ? e.target.tagName : '').toLowerCase()
  if (tag === 'input' || tag === 'textarea') return
  e.preventDefault()
  if (e.shiftKey) redo()
  else undo()
}

const selected = computed(() => questions.value.find((q) => q.key === selectedKey.value) || null)
const hasChoice = (q) => q && (q.type === 'single' || q.type === 'multiple')
const canUndo = computed(() => history.value.length > 0)
const canRedo = computed(() => future.value.length > 0)

function snapshot() {
  history.value.push(JSON.stringify(questions.value))
  if (history.value.length > 60) history.value.shift()
  future.value = []
}

function restore(json) {
  questions.value = JSON.parse(json).map((q) => ({
    ...q,
    options: (q.options || []).map((o) => ({ ...o})),
  }))
  if (!questions.value.some((q) => q.key === selectedKey.value)) {
    selectedKey.value = questions.value.length ? questions.value[0].key : null
  }
}

function undo() {
  if (!canUndo.value) return
  future.value.push(JSON.stringify(questions.value))
  restore(history.value.pop())
}

function redo() {
  if (!canRedo.value) return
  history.value.push(JSON.stringify(questions.value))
  restore(future.value.pop())
}

function addQuestion(type = 'single') {
  snapshot()
  const q = {
    key: nkey('q'),
    id: 0,
    type,
    title: '',
    description: '',
    required: true,
    options:
      type === 'single' || type === 'multiple'
        ? [makeOption('选项 1'), makeOption('选项 2')]
        : [],
  }
  questions.value.push(q)
  selectedKey.value = q.key
}

function makeOption(label) {
  return { key: nkey('o'), id: 0, label, nextQuestionId: 0 }
}

function removeQuestion(q) {
  snapshot()
  questions.value = questions.value.filter((x) => x.key !== q.key)
  if (selectedKey.value === q.key) {
    selectedKey.value = questions.value.length ? questions.value[0].key : null
  }
}

function move(index, delta) {
  const to = index + delta
  if (to < 0 || to >= questions.value.length) return
  snapshot()
  const arr = questions.value.slice()
  const [item] = arr.splice(index, 1)
  arr.splice(to, 0, item)
  questions.value = arr
}

function changeType(q) {
  if (hasChoice(q) && q.options.length < 2) {
    q.options = [makeOption('选项 1'), makeOption('选项 2')]
  }
  if (!hasChoice(q)) q.options = []
}

function addOption(q) {
  snapshot()
  q.options.push(makeOption(`选项 ${q.options.length + 1}`))
}

function removeOption(q, o) {
  snapshot()
  q.options = q.options.filter((x) => x.key !== o.key)
}

/* 拖拽重排 */
const dragIndex = ref(-1)
const dropIndex = ref(-1)
function onDragStart(i) {
  dragIndex.value = i
}
function onDragOver(i, e) {
  e.preventDefault()
  dropIndex.value = i
}
function onDrop(i) {
  const from = dragIndex.value
  dragIndex.value = -1
  dropIndex.value = -1
  if (from < 0 || from === i) return
  snapshot()
  const arr = questions.value.slice()
  const [item] = arr.splice(from, 1)
  arr.splice(i, 0, item)
  questions.value = arr
}

/* 跳转下拉：值 0 不跳转；已有题用真实 id；新题用 -(index+1) */
const jumpOptions = computed(() =>
  questions.value.map((q, i) => ({
    label: `第 ${i + 1} 题 · ${q.title || '未命名'}`,
    value: q.id > 0 ? q.id : -(i + 1),
  })),
)

async function saveInfo() {
  if (!titleForm.title.trim()) return message.warning('问卷标题不能为空')
  try {
    await api.updateSurvey(surveyId, {
      title: titleForm.title.trim(),
      description: titleForm.description.trim(),
      anonymous: titleForm.anonymous,
    })
    message.success('问卷信息已保存')
    await load()
  } catch (e) {
    message.error(e.message)
  }
}

async function saveQuestions() {
  for (let i = 0; i < questions.value.length; i++) {
    const q = questions.value[i]
    if (!q.title.trim()) return message.warning(`第 ${i + 1} 题还没写题干`)
    if (hasChoice(q)) {
      if (q.options.length < 2) return message.warning(`第 ${i + 1} 题至少要 2 个选项`)
      if (q.options.some((o) => !o.label.trim())) return message.warning(`第 ${i + 1} 题有选项内容为空`)
    }
  }
  saving.value = true
  try {
    const payload = questions.value.map((q) => ({
      id: q.id || 0,
      type: q.type,
      title: q.title.trim(),
      description: (q.description || '').trim(),
      required: !!q.required,
      options: (q.options || []).map((o) => ({
        label: o.label.trim(),
        nextQuestionId: o.nextQuestionId || 0,
      })),
    }))
    const sv = await api.saveQuestions(surveyId, payload)
    applySurvey(sv)
    history.value = []
    future.value = []
    message.success('题目已保存')
  } catch (e) {
    message.error(e.message)
  } finally {
    saving.value = false
  }
}

async function publish() {
  if (questions.value.length === 0) return message.warning('先添加题目再发布')
  try {
    await saveQuestions()
    await api.setStatus(surveyId, 'published')
    message.success('已发布，可以复制链接分享了')
    await load()
  } catch (e) {
    message.error(e.message)
  }
}

async function copyLink() {
  const url = `${location.origin}/s/${surveyId}`
  try {
    await navigator.clipboard.writeText(url)
    message.success('填写链接已复制')
  } catch {
    message.info(url)
  }
}

function preview() {
  window.open(`/s/${surveyId}`, '_blank')
}

function back() {
  router.push('/surveys')
}

function removeSurvey() {
  dialog.warning({
    title: '删除问卷',
    content: '删除后问卷和答卷都会消失，确定吗？',
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await api.deleteSurvey(surveyId)
        message.success('已删除')
        router.replace('/surveys')
      } catch (e) {
        message.error(e.message)
      }
    },
  })
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <div class="row" style="gap: 10px">
        <n-button quaternary @click="back">← 返回</n-button>
        <div>
          <h1 class="page-title">{{ titleForm.title || '未命名问卷' }}</h1>
          <div class="muted">
            <n-tag size="small" :bordered="false">{{ statusText[survey?.status] || '草稿' }}</n-tag>
            <span style="margin-left: 8px">{{ questions.length }} 道题</span>
          </div>
        </div>
      </div>
      <div class="row">
        <n-button :disabled="!canUndo" quaternary @click="undo">撤销</n-button>
        <n-button :disabled="!canRedo" quaternary @click="redo">重做</n-button>
        <n-button quaternary @click="preview">预览</n-button>
        <n-button quaternary @click="copyLink">复制链接</n-button>
        <n-button type="primary" ghost :loading="saving" @click="saveQuestions">保存题目</n-button>
        <n-button type="primary" @click="publish">发布</n-button>
      </div>
    </div>

    <div class="edit-grid">
      <div>
        <n-card size="small" title="问卷信息" class="soft">
          <div class="field">
            <n-input v-model:value="titleForm.title" placeholder="问卷标题" maxlength="100" />
          </div>
          <div class="field">
            <n-input
              v-model:value="titleForm.description"
              type="textarea"
              :rows="2"
              placeholder="问卷说明（可不填）"
            />
          </div>
          <div class="row" style="justify-content: space-between">
            <span class="muted">匿名填写</span>
            <div class="row">
              <n-switch v-model:value="titleForm.anonymous" size="small" />
              <n-button size="small" @click="saveInfo">保存</n-button>
            </div>
          </div>
        </n-card>

        <div class="row" style="margin: 16px 0 10px; justify-content: space-between">
          <div style="font-weight: 600">题目列表</div>
          <div class="muted" style="font-size: 12px">拖拽卡片可排序</div>
        </div>

        <n-empty v-if="!questions.length" description="还没有题目，点下面按钮添加" style="padding: 32px 0" />

        <div class="q-list">
          <div
            v-for="(q, i) in questions"
            :key="q.key"
            class="q-card"
            :class="{ active: q.key === selectedKey, dragging: dropIndex === i && dragIndex !== -1 }"
            draggable="true"
            @dragstart="onDragStart(i)"
            @dragover="onDragOver(i, $event)"
            @drop="onDrop(i)"
            @click="selectedKey = q.key"
          >
            <div class="q-index">{{ i + 1 }}</div>
            <div class="grow">
              <div class="q-title ellipsis">{{ q.title || '未命名题目' }}</div>
              <div class="muted" style="font-size: 12px; margin-top: 2px">
                {{ typeText[q.type] }}<template v-if="q.required"> · 必答</template>
                <template v-if="hasChoice(q)"> · {{ q.options.length }} 个选项</template>
              </div>
            </div>
            <div class="q-ops" @click.stop>
              <n-button text size="tiny" @click="move(i, -1)">↑</n-button>
              <n-button text size="tiny" @click="move(i, 1)">↓</n-button>
              <n-button text size="tiny" type="error" @click="removeQuestion(q)">删除</n-button>
            </div>
          </div>
        </div>

        <div class="row" style="margin-top: 12px">
          <n-button dashed style="flex: 1" @click="addQuestion('single')">+ 单选题</n-button>
          <n-button dashed style="flex: 1" @click="addQuestion('multiple')">+ 多选题</n-button>
          <n-button dashed style="flex: 1" @click="addQuestion('text')">+ 填空题</n-button>
          <n-button dashed style="flex: 1" @click="addQuestion('rating')">+ 评分题</n-button>
        </div>
      </div>

      <div>
        <n-card size="small" class="soft">
          <template #header>
            <div class="row" style="justify-content: space-between">
              <span>题目设置</span>
              <n-button text type="error" size="small" @click="removeSurvey">删除问卷</n-button>
            </div>
          </template>

          <n-empty v-if="!selected" description="在左侧选一道题" style="padding: 40px 0" />

          <div v-else>
            <div class="field">
              <div class="label">题型</div>
              <n-select
                v-model:value="selected.type"
                :options="typeOptions"
                @update:value="changeType(selected)"
              />
            </div>
            <div class="field">
              <div class="label">题干</div>
              <n-input
                v-model:value="selected.title"
                type="textarea"
                :rows="2"
                maxlength="200"
                show-count
                placeholder="请输入题目内容"
              />
            </div>
            <div class="field">
              <div class="label">补充说明</div>
              <n-input v-model:value="selected.description" placeholder="可不填" />
            </div>
            <div class="field row" style="justify-content: space-between">
              <span class="label" style="margin: 0">必答题</span>
              <n-switch v-model:value="selected.required" size="small" />
            </div>

            <template v-if="hasChoice(selected)">
              <div class="label" style="margin-bottom: 8px">选项与跳转</div>
              <div class="opt-list">
                <div v-for="(o, oi) in selected.options" :key="o.key" class="opt-row">
                  <div class="opt-idx">{{ oi + 1 }}</div>
                  <n-input v-model:value="o.label" placeholder="选项内容" size="small" />
                  <n-select
                    v-model:value="o.nextQuestionId"
                    :options="[{ label: '不跳转', value: 0 }, ...jumpOptions]"
                    size="small"
                    style="width: 168px"
                  />
                  <n-button text type="error" size="tiny" @click="removeOption(selected, o)">✕</n-button>
                </div>
              </div>
              <n-button dashed size="small" block style="margin-top: 8px" @click="addOption(selected)">
                + 添加选项
              </n-button>
              <div class="muted" style="font-size: 12px; margin-top: 8px">
                选择某个选项后跳转到指定题目，未选择则按顺序进入下一题。
              </div>
            </template>

            <div v-else-if="selected.type === 'rating'" class="muted" style="margin-top: 8px">
              评分题固定为 1-5 分，无需设置选项。
            </div>
            <div v-else class="muted" style="margin-top: 8px">填空题由填写者自由输入文字。</div>
          </div>
        </n-card>
      </div>
    </div>
  </div>
</template>

<style scoped>
.edit-grid {
  display: grid;
  grid-template-columns: minmax(320px, 1fr) minmax(360px, 460px);
  gap: 20px;
  align-items: start;
}
.q-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.q-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: #fff;
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.q-card:hover {
  border-color: #c7d2fe;
}
.q-card.active {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.12);
}
.q-card.dragging {
  border-style: dashed;
  border-color: var(--primary);
}
.q-index {
  width: 22px;
  height: 22px;
  border-radius: 7px;
  background: #eef2ff;
  color: #4f46e5;
  font-size: 12px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: none;
}
.q-title {
  font-size: 14px;
  font-weight: 500;
}
.q-ops {
  display: flex;
  gap: 4px;
  opacity: 0.75;
}
.opt-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.opt-row {
  display: flex;
  align-items: center;
  gap: 6px;
}
.opt-idx {
  width: 16px;
  font-size: 12px;
  color: var(--muted);
  text-align: center;
  flex: none;
}
.field {
  margin-bottom: 14px;
}
.field:last-child {
  margin-bottom: 0;
}
.label {
  font-size: 13px;
  color: var(--muted);
  margin-bottom: 6px;
}
@media (max-width: 900px) {
  .edit-grid {
    grid-template-columns: 1fr;
  }
  .opt-row {
    flex-wrap: wrap;
  }
  .opt-row :deep(.n-select) {
    width: 100% !important;
  }
}
</style>
