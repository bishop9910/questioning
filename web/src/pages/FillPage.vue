<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  NButton,
  NCard,
  NInput,
  NRadio,
  NRadioGroup,
  NRate,
  NSpin,
  NCheckbox,
  NCheckboxGroup,
  useMessage,
} from 'naive-ui'
import { api, deviceOf } from '../api'

const route = useRoute()
const message = useMessage()
const surveyId = Number(route.params.id)

const loading = ref(true)
const survey = ref(null)
const loadError = ref('')

const answers = reactive({}) // questionId -> { optionIds: [], text: '' }
const path = ref([]) // 走过的题目下标
const cursor = ref(0)
const submitted = ref(false)
const submitting = ref(false)
const startedAt = Date.now()

const current = computed(() => {
  if (!survey.value) return null
  return survey.value.questions[cursor.value] || null
})

const progress = computed(() => {
  if (!survey.value || !survey.value.questions.length) return 0
  return Math.round(((path.value.length + (submitted.value ? 1 : 0)) / survey.value.questions.length) * 100)
})

function ensure(id) {
  if (!answers[id]) answers[id] = { optionIds: [], text: '' }
  return answers[id]
}

async function load() {
  loading.value = true
  try {
    survey.value = await api.publicSurvey(surveyId)
    for (const q of survey.value.questions || []) ensure(q.id)
    path.value = survey.value.questions.length ? [0] : []
    cursor.value = 0
  } catch (e) {
    loadError.value = e.message
  } finally {
    loading.value = false
  }
}
onMounted(load)

const isChoice = (q) => q && (q.type === 'single' || q.type === 'multiple')

function answered(q) {
  const a = answers[q.id]
  if (!a) return false
  if (q.type === 'text') return !!a.text.trim()
  if (q.type === 'rating') return a.optionIds.length > 0
  return a.optionIds.length > 0
}

function singleValue(q) {
  const a = answers[q.id]
  if (!a || !a.optionIds.length) return null
  return a.optionIds[0]
}
function setSingle(q, v) {
  ensure(q.id).optionIds = v === null ? [] : [v]
}

function setMulti(q, vals) {
  let list = Array.isArray(vals) ? vals.slice() : []
  if (q.maxSelect > 0 && list.length > q.maxSelect) {
    message.warning(`这道题最多只能选 ${q.maxSelect} 项`)
    list = list.slice(0, q.maxSelect)
  }
  ensure(q.id).optionIds = list
}

function isPicked(q, o) {
  const a = answers[q.id]
  return !!a && a.optionIds.includes(o.id)
}

function toggleMulti(q, o) {
  if (multiDisabled(q, o)) return
  const cur = answers[q.id]?.optionIds || []
  setMulti(q, isPicked(q, o) ? cur.filter((x) => x !== o.id) : [...cur, o.id])
}

// 已经选满上限时，未选中的选项禁用
function multiDisabled(q, o) {
  const a = answers[q.id]
  const picked = a ? a.optionIds.includes(o.id) : false
  if (picked) return false
  return q.maxSelect > 0 && (a ? a.optionIds.length : 0) >= q.maxSelect
}

function ratingValue(q) {
  const a = answers[q.id]
  if (!a || !a.optionIds.length) return 0
  const opt = (q.options || []).find((o) => o.id === a.optionIds[0])
  return opt ? Number(opt.label) || 0 : 0
}
function setRating(q, n) {
  const opt = (q.options || []).find((o) => Number(o.label) === n)
  ensure(q.id).optionIds = opt ? [opt.id] : []
}

function nextIndex(q) {
  let next = 0
  const a = answers[q.id]
  if (isChoice(q) && a) {
    for (const o of q.options || []) {
      if (a.optionIds.includes(o.id) && o.nextQuestionId) next = o.nextQuestionId
    }
  }
  if (next) {
    const i = survey.value.questions.findIndex((x) => x.id === next)
    if (i >= 0) return i
  }
  return cursor.value + 1
}

async function next() {
  const q = current.value
  if (!q) return
  if (q.required && !answered(q)) return message.warning('这道题是必答题，填完再继续')
  if (q.type === 'multiple') {
    const n = (answers[q.id]?.optionIds || []).length
    if (q.minSelect > 0 && n > 0 && n < q.minSelect) {
      return message.warning(`这道题至少要选 ${q.minSelect} 项`)
    }
    if (q.maxSelect > 0 && n > q.maxSelect) {
      return message.warning(`这道题最多只能选 ${q.maxSelect} 项`)
    }
  }
  const to = nextIndex(q)
  if (to >= survey.value.questions.length) return submit()
  path.value.push(to)
  cursor.value = to
}

function back() {
  if (path.value.length <= 1) return
  path.value.pop()
  cursor.value = path.value[path.value.length - 1]
}

async function submit() {
  submitting.value = true
  try {
    const payload = path.value
      .map((i) => survey.value.questions[i])
      .filter((q) => answered(q) || !q.required)
      .map((q) => {
        const a = answers[q.id]
        return {
          questionId: q.id,
          optionIds: q.type === 'text' ? [] : a.optionIds,
          text: q.type === 'text' ? a.text.trim() : '',
        }
      })
    await api.submit(surveyId, {
      answers: payload,
      durationSec: Math.round((Date.now() - startedAt) / 1000),
      device: deviceOf(),
    })
    submitted.value = true
  } catch (e) {
    message.error(e.message)
  } finally {
    submitting.value = false
  }
}

function restart() {
  for (const k of Object.keys(answers)) delete answers[k]
  for (const q of survey.value.questions || []) ensure(q.id)
  path.value = survey.value.questions.length ? [0] : []
  cursor.value = 0
  submitted.value = false
}
</script>

<template>
  <div class="fill-wrap">
    <div v-if="loading" class="fill-card loading-box">
      <n-spin size="large" />
    </div>

    <template v-else>
      <div v-if="loadError" class="fill-card">
        <n-card class="block-card">
          <div class="done-icon">!</div>
          <h2 class="done-title">{{ loadError }}</h2>
          <p class="muted">这份问卷可能还没发布，或者已经被关闭了。</p>
        </n-card>
      </div>

      <div v-else-if="submitted" class="fill-card">
        <n-card class="block-card">
          <div class="done-icon ok">✓</div>
          <h2 class="done-title">提交成功</h2>
          <p class="muted">感谢你的填写，答案已经记录下来啦。</p>
          <n-button type="primary" ghost style="margin-top: 8px" @click="restart">再填一份</n-button>
        </n-card>
      </div>

      <div v-else-if="survey" class="fill-card">
        <div class="survey-head">
          <h1 class="survey-title">{{ survey.title }}</h1>
          <p v-if="survey.description" class="survey-desc">{{ survey.description }}</p>
          <div class="progress">
            <div class="progress-bar" :style="{ width: progress + '%' }"></div>
          </div>
        </div>

        <n-card v-if="current" class="block-card">
          <div class="q-label">
            第 {{ cursor + 1 }} 题
            <span v-if="current.required" class="required">*</span>
          </div>
          <div class="q-title">{{ current.title }}</div>
          <div v-if="current.description" class="q-desc">{{ current.description }}</div>

          <div class="opts">
            <template v-if="current.type === 'single'">
              <div class="opt-list">
                <label
                  v-for="o in current.options"
                  :key="o.id"
                  class="opt"
                  :class="{ picked: singleValue(current) === o.id }"
                  @click.prevent="setSingle(current, o.id)"
                >
                  <n-radio :checked="singleValue(current) === o.id" :value="o.id" />
                  <span>{{ o.label }}</span>
                </label>
              </div>
            </template>

            <template v-else-if="current.type === 'multiple'">
              <div class="opt-list">
                <label
                  v-for="o in current.options"
                  :key="o.id"
                  class="opt"
                  :class="{
                    picked: isPicked(current, o),
                    muted: multiDisabled(current, o),
                  }"
                  @click.prevent="toggleMulti(current, o)"
                >
                  <n-checkbox
                    :checked="isPicked(current, o)"
                    :disabled="multiDisabled(current, o)"
                  />
                  <span>{{ o.label }}</span>
                </label>
              </div>
              <div class="multi-hint">
                <span>已选 {{ (answers[current.id]?.optionIds || []).length }} 项</span>
                <span v-if="current.minSelect > 0"> · 至少 {{ current.minSelect }} 项</span>
                <span v-if="current.maxSelect > 0"> · 最多 {{ current.maxSelect }} 项</span>
                <span v-if="!current.minSelect && !current.maxSelect"> · 可多选</span>
              </div>
            </template>

            <template v-else-if="current.type === 'rating'">
              <div class="rating-row">
                <n-rate
                  :value="ratingValue(current)"
                  @update:value="(v) => setRating(current, v)"
                />
                <span class="muted">{{ ratingValue(current) ? ratingValue(current) + ' 分' : '请打分' }}</span>
              </div>
            </template>

            <template v-else>
              <n-input
                v-model:value="answers[current.id].text"
                type="textarea"
                :rows="4"
                :maxlength="500"
                show-count
                placeholder="请输入你的回答"
              />
            </template>
          </div>

          <div class="fill-actions">
            <n-button v-if="path.length > 1" quaternary @click="back">上一题</n-button>
            <n-button type="primary" size="large" :loading="submitting" @click="next">
              {{ cursor + 1 >= survey.questions.length ? '提交' : '下一题' }}
            </n-button>
          </div>
        </n-card>
      </div>
    </template>
  </div>
</template>

<style scoped>
.fill-wrap {
  min-height: 100%;
  padding: 24px 16px 48px;
  display: block;
}
.fill-card {
  margin: 0 auto;
  max-width: 640px;
}
.loading-box {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 240px;
}
.survey-head {
  margin-bottom: 16px;
}
.survey-title {
  font-size: 22px;
  margin: 0 0 8px;
}
.survey-desc {
  color: var(--muted);
  font-size: 14px;
  margin: 0 0 12px;
  line-height: 1.6;
}
.progress {
  height: 6px;
  background: #e5e7eb;
  border-radius: 999px;
  overflow: hidden;
}
.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #6366f1);
  transition: width 0.25s;
}
.block-card {
  border-radius: 16px;
  box-shadow: 0 10px 30px rgba(31, 39, 51, 0.06);
}
.q-label {
  font-size: 12px;
  color: var(--muted);
  margin-bottom: 6px;
}
.required {
  color: #ef4444;
  margin-left: 2px;
}
.q-title {
  font-size: 17px;
  font-weight: 600;
  line-height: 1.5;
}
.q-desc {
  color: var(--muted);
  font-size: 13px;
  margin-top: 6px;
}
.opts {
  margin-top: 16px;
}
.opts :deep(.n-radio-group),
.opts :deep(.n-checkbox-group) {
  display: block;
  width: 100%;
}
.opt-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}
.opt {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border: 1px solid var(--border);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.15s;
  font-size: 14px;
  box-sizing: border-box;
  width: 100%;
}
.opt:hover {
  border-color: #c7d2fe;
  background: #f8faff;
}
.opt.picked {
  border-color: var(--primary);
  background: #eff6ff;
}
.opt.muted {
  opacity: 0.45;
  cursor: not-allowed;
}
.opt.muted:hover {
  border-color: var(--border);
  background: transparent;
}
.multi-hint {
  margin-top: 8px;
  font-size: 12px;
  color: var(--muted);
}
.rating-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.fill-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 24px;
}
.done-icon {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: #fef3c7;
  color: #b45309;
  font-size: 28px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
}
.done-icon.ok {
  background: #dcfce7;
  color: #15803d;
}
.done-title {
  margin: 0 0 6px;
  font-size: 20px;
}
@media (max-width: 720px) {
  .fill-wrap {
    padding: 16px 12px 32px;
  }
  .survey-title {
    font-size: 19px;
  }
  .opt {
    padding: 12px;
  }
}
</style>
