<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NButton,
  NCard,
  NDataTable,
  NEmpty,
  NModal,
  NPagination,
  NSpin,
  NTag,
  useMessage,
} from 'naive-ui'
import * as echarts from 'echarts'
import { api, request } from '../api'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const surveyId = Number(route.params.id)

const loading = ref(true)
const stats = ref(null)
const charts = {}
const exporting = ref(false)
const PALETTE = ['#3b82f6', '#60a5fa', '#818cf8', '#93c5fd', '#a5b4fc', '#c7d2fe']

const responses = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const responseLoading = ref(false)

const showDetail = ref(false)
const detail = ref(null)
const detailLoading = ref(false)

const typeText = { single: '单选', multiple: '多选', text: '填空', rating: '评分' }

const columns = [
  { title: '答卷 ID', key: 'id', width: 90 },
  { title: '提交时间', key: 'submittedAt', width: 180, render: (r) => fmtDate(r.submittedAt) },
  { title: '耗时', key: 'durationSec', width: 90, render: (r) => `${r.durationSec} 秒` },
  { title: '设备', key: 'device', width: 100 },
  { title: '', key: 'op', width: 80, render: () => '查看' },
]

function fmtDate(iso) {
  if (!iso) return '-'
  const d = new Date(iso)
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}

function chart(id) {
  const el = document.getElementById(id)
  if (!el) return null
  if (charts[id]) charts[id].dispose()
  charts[id] = echarts.init(el)
  return charts[id]
}

function renderCharts() {
  const s = stats.value
  if (!s) return

  const line = chart('chart-daily')
  if (line) {
    line.setOption({
      grid: { left: 40, right: 20, top: 30, bottom: 30 },
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: s.daily.map((d) => d.date.slice(5)) },
      yAxis: { type: 'value', minInterval: 1 },
      series: [
        {
          type: 'line',
          smooth: true,
          symbolSize: 6,
          data: s.daily.map((d) => d.count),
          areaStyle: { opacity: 0.15 },
          lineStyle: { width: 2.5 },
          itemStyle: { color: '#3b82f6' },
        },
      ],
    })
  }

  for (const q of s.questions) {
    if (q.type === 'single' || q.type === 'multiple') {
      const pie = chart(`chart-q-${q.id}`)
      if (pie) {
        pie.setOption({
          tooltip: { trigger: 'item', formatter: '{b}: {c} 票 ({d}%)' },
          color: PALETTE,
          legend: { bottom: 0, type: 'scroll', icon: 'circle', itemWidth: 8, itemHeight: 8 },
          series: [
            {
              type: 'pie',
              radius: ['46%', '72%'],
              center: ['50%', '44%'],
              avoidLabelOverlap: true,
              label: { formatter: (p) => (p.value ? `${p.name}\n${p.percent}%` : '') },
              labelLine: { length: 10, length2: 10 },
              data: (q.options || []).map((o) => ({ name: o.label, value: o.count })),
            },
          ],
        })
      }
    } else if (q.type === 'rating') {
      const bar = chart(`chart-q-${q.id}`)
      if (bar) {
        bar.setOption({
          grid: { left: 40, right: 20, top: 20, bottom: 30 },
          tooltip: {},
          xAxis: { type: 'category', data: (q.options || []).map((o) => `${o.label} 分`) },
          yAxis: { type: 'value', minInterval: 1 },
          series: [
            {
              type: 'bar',
              barWidth: '46%',
              itemStyle: { color: '#3b82f6', borderRadius: [6, 6, 0, 0] },
              data: (q.options || []).map((o) => o.count),
            },
          ],
        })
      }
    }
  }
}

async function load() {
  loading.value = true
  try {
    stats.value = await api.stats(surveyId)
    await nextTick()
    renderCharts()
    await loadResponses()
  } catch (e) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}

async function loadResponses() {
  responseLoading.value = true
  try {
    const data = await api.responses(surveyId, page.value, pageSize.value)
    responses.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    message.error(e.message)
  } finally {
    responseLoading.value = false
  }
}

function onPage(p) {
  page.value = p
  loadResponses()
}

async function openDetail(row) {
  showDetail.value = true
  detailLoading.value = true
  detail.value = null
  try {
    detail.value = await request(`/surveys/${surveyId}/responses/${row.id}`)
  } catch (e) {
    message.error(e.message)
    showDetail.value = false
  } finally {
    detailLoading.value = false
  }
}

const questionMap = computed(() => {
  const m = {}
  for (const q of stats.value?.questions || []) m[q.id] = q
  return m
})

function answerText(a) {
  const q = questionMap.value[a.questionId]
  if (!q) return a.text || '-'
  if (q.type === 'text') return a.text || '（空）'
  const labels = (q.options || []).filter((o) => (a.optionIds || []).includes(o.id)).map((o) => o.label)
  return labels.length ? labels.join('、') : '（未作答）'
}

async function exportCsv() {
  exporting.value = true
  try {
    const res = await api.exportCsv(surveyId)
    if (!res.ok) throw new Error('导出失败')
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${stats.value?.title || 'survey'}.csv`
    a.click()
    URL.revokeObjectURL(url)
    message.success('已导出 CSV')
  } catch (e) {
    message.error(e.message)
  } finally {
    exporting.value = false
  }
}

function resize() {
  Object.values(charts).forEach((c) => c.resize())
}

onMounted(() => {
  load()
  window.addEventListener('resize', resize)
})
onBeforeUnmount(() => {
  window.removeEventListener('resize', resize)
  Object.values(charts).forEach((c) => c.dispose())
})

watch(total, () => {})
</script>

<template>
  <div class="page">
    <div class="page-head">
      <div class="row" style="gap: 10px">
        <n-button quaternary @click="router.push('/surveys')">← 返回</n-button>
        <div>
          <h1 class="page-title">{{ stats?.title || '统计' }}</h1>
          <div class="muted">实时统计 · 数据随时间累积</div>
        </div>
      </div>
      <div class="row">
        <n-button quaternary @click="router.push(`/surveys/${surveyId}/edit`)">编辑问卷</n-button>
        <n-button type="primary" :loading="exporting" @click="exportCsv">导出 CSV</n-button>
      </div>
    </div>

    <n-spin :show="loading">
      <template v-if="stats">
        <div class="stat-cards">
          <n-card class="stat-card" size="small">
            <div class="stat-num">{{ stats.total }}</div>
            <div class="muted">累计答卷</div>
          </n-card>
          <n-card class="stat-card" size="small">
            <div class="stat-num">{{ stats.questions.length }}</div>
            <div class="muted">题目数量</div>
          </n-card>
          <n-card class="stat-card" size="small">
            <div class="stat-num">{{ stats.avgDurationSec }}<span class="unit">秒</span></div>
            <div class="muted">平均填写耗时</div>
          </n-card>
        </div>

        <n-card size="small" class="block" title="近 7 天答卷趋势">
          <div id="chart-daily" class="chart"></div>
        </n-card>

        <n-empty
          v-if="!stats.total"
          description="还没有人填写，发布问卷后把链接发出去吧"
          style="padding: 40px 0"
        />

        <template v-else>
          <div class="q-stats">
            <n-card
              v-for="(q, i) in stats.questions"
              :key="q.id"
              size="small"
              class="block"
            >
              <template #header>
                <div class="row" style="gap: 8px; flex-wrap: wrap">
                  <span class="q-no">{{ i + 1 }}</span>
                  <span class="q-title">{{ q.title }}</span>
                  <n-tag size="small" :bordered="false">{{ typeText[q.type] }}</n-tag>
                  <span class="muted" style="font-size: 12px">{{ q.answered }} 人作答</span>
                </div>
              </template>

              <div v-if="q.type === 'single' || q.type === 'multiple'">
                <div :id="`chart-q-${q.id}`" class="chart"></div>
                <div class="legend">
                  <div v-for="o in q.options" :key="o.id" class="legend-row">
                    <span class="legend-label ellipsis">{{ o.label }}</span>
                    <span class="muted">{{ o.count }} 票 · {{ o.percent }}%</span>
                  </div>
                </div>
              </div>

              <div v-else-if="q.type === 'rating'">
                <div :id="`chart-q-${q.id}`" class="chart"></div>
              </div>

              <div v-else>
                <n-empty
                  v-if="!q.words || !q.words.length"
                  description="暂无文字内容"
                  style="padding: 16px 0"
                />
                <div v-else class="words">
                  <div v-for="w in q.words" :key="w.word" class="word">
                    <span class="word-text ellipsis">{{ w.word }}</span>
                    <span class="word-count">{{ w.count }}</span>
                  </div>
                </div>
              </div>
            </n-card>
          </div>

          <n-card size="small" class="block" title="答卷记录">
            <n-data-table
              :columns="columns"
              :data="responses"
              :loading="responseLoading"
              :bordered="false"
              :row-props="(row) => ({ style: 'cursor:pointer', onClick: () => openDetail(row) })"
              size="small"
            />
            <div style="display: flex; justify-content: flex-end; margin-top: 12px">
              <n-pagination
                :page="page"
                :page-size="pageSize"
                :item-count="total"
                @update:page="onPage"
              />
            </div>
          </n-card>
        </template>
      </template>
    </n-spin>

    <n-modal v-model:show="showDetail" preset="card" title="答卷详情" style="max-width: 560px">
      <n-spin :show="detailLoading">
        <div v-if="detail">
          <div class="muted" style="margin-bottom: 12px">
            #{{ detail.id }} · {{ fmtDate(detail.submittedAt) }} · 耗时
            {{ detail.durationSec }} 秒 · {{ detail.device }}
          </div>
          <div v-for="a in detail.answers" :key="a.questionId" class="ans">
            <div class="ans-q">{{ questionMap[a.questionId]?.title || `题目 ${a.questionId}` }}</div>
            <div class="ans-a">{{ answerText(a) }}</div>
          </div>
        </div>
      </n-spin>
    </n-modal>
  </div>
</template>

<style scoped>
.stat-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}
.stat-card {
  border-radius: 14px;
}
.stat-num {
  font-size: 26px;
  font-weight: 700;
  color: var(--text);
}
.unit {
  font-size: 13px;
  font-weight: 400;
  color: var(--muted);
  margin-left: 2px;
}
.block {
  border-radius: 14px;
  margin-bottom: 16px;
}
.chart {
  width: 100%;
  height: 260px;
}
.q-stats {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}
.legend {
  border-top: 1px dashed var(--border);
  padding-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 150px;
  overflow: auto;
}
.legend-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
}
.legend-label {
  max-width: 65%;
}
.words {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.word {
  display: flex;
  align-items: center;
  gap: 6px;
  background: #f1f5f9;
  border-radius: 999px;
  padding: 4px 10px;
  font-size: 13px;
}
.word-count {
  color: #4f46e5;
  font-weight: 600;
}
.q-no {
  width: 20px;
  height: 20px;
  border-radius: 6px;
  background: #eef2ff;
  color: #4f46e5;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.q-title {
  font-size: 14px;
  font-weight: 600;
}
.ans {
  margin-bottom: 12px;
}
.ans-q {
  font-size: 13px;
  color: var(--muted);
}
.ans-a {
  margin-top: 2px;
  font-size: 14px;
}
@media (max-width: 720px) {
  .stat-cards {
    grid-template-columns: 1fr;
  }
  .q-stats {
    grid-template-columns: 1fr;
  }
}
</style>
