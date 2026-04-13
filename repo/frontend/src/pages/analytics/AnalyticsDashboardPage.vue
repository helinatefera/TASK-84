<template>
  <div class="dashboard-page">
    <h1>Analytics Dashboard</h1>
    <div v-if="!isReadonly" class="filter-bar">
      <div class="filter-group">
        <label>Item</label>
        <select v-model="filters.item_id" class="form-input">
          <option value="">All Items</option>
          <option v-for="it in itemOptions" :key="it.id" :value="it.id">{{ it.title }}</option>
        </select>
      </div>
      <div class="filter-group">
        <label>Start Date</label>
        <input v-model="filters.start_date" type="date" class="form-input" />
      </div>
      <div class="filter-group">
        <label>End Date</label>
        <input v-model="filters.end_date" type="date" class="form-input" />
      </div>
      <div class="filter-group">
        <label>Sentiment</label>
        <select v-model="filters.sentiment" class="form-input">
          <option value="">All</option>
          <option value="positive">Positive</option>
          <option value="neutral">Neutral</option>
          <option value="negative">Negative</option>
        </select>
      </div>
      <div class="filter-group">
        <label>Keywords</label>
        <input v-model="filters.keywords" type="text" placeholder="Search keywords..." class="form-input" />
      </div>
      <button @click="fetchDashboard" class="btn-primary">Apply Filters</button>
    </div>

    <div v-if="loading" class="loading">Loading analytics...</div>

    <div v-if="!loading && dashboardData" class="results">
      <h2>Aggregated Metrics</h2>
      <table class="data-table">
        <thead>
          <tr>
            <th>Period</th><th>Impressions</th><th>Clicks</th><th>Avg Dwell (s)</th><th>Favorites</th><th>Shares</th><th>Comments</th><th>Drill Down</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, idx) in dashboardData.data" :key="idx">
            <td>{{ row.period_start }}</td>
            <td>{{ row.impressions }}</td>
            <td>{{ row.clicks }}</td>
            <td>{{ row.avg_dwell_secs }}</td>
            <td>{{ row.favorites }}</td>
            <td>{{ row.shares }}</td>
            <td>{{ row.comments }}</td>
            <td><button @click="drillDown(row)" class="btn-sm">Sessions</button></td>
          </tr>
        </tbody>
      </table>
      <p v-if="!dashboardData.data?.length">No data for selected filters</p>
    </div>

    <div v-if="dashboardData?.data?.length" class="chart-section">
      <h2>Trends</h2>
      <v-chart :option="chartOption" style="height: 300px;" autoresize />
    </div>

    <div v-if="drillDownSessions.length" class="drill-down-section">
      <h3>Session Details <button @click="drillDownSessions = []; timelineEvents = []" class="btn-sm">Close</button></h3>
      <table class="data-table">
        <thead><tr><th>Session ID</th><th>Started</th><th>Last Active</th><th>Timeline</th></tr></thead>
        <tbody>
          <tr v-for="s in drillDownSessions" :key="s.session_id">
            <td>{{ s.session_id }}</td>
            <td>{{ s.started_at }}</td>
            <td>{{ s.last_active_at }}</td>
            <td><button @click="loadTimeline(s.session_id)" class="btn-sm">View</button></td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="timelineEvents.length" class="timeline-section">
      <h3>Event Timeline <button @click="timelineEvents = []" class="btn-sm">Close</button></h3>
      <ul class="timeline-list">
        <li v-for="evt in timelineEvents" :key="evt.id">
          <strong>{{ evt.event_type }}</strong> — {{ evt.server_ts }}
          <span v-if="evt.dwell_seconds"> (dwell: {{ evt.dwell_seconds }}s)</span>
        </li>
      </ul>
    </div>

    <div v-if="vizLoading" class="loading">Loading visualizations...</div>

    <div v-if="sentimentData.length" class="chart-section">
      <h2>Sentiment Heatmap</h2>
      <v-chart :option="sentimentChartOption" style="height: 280px;" autoresize />
    </div>

    <div class="viz-row">
      <div v-if="keywordsData.length" class="chart-section viz-half">
        <h2>Word Cloud</h2>
        <v-chart :option="wordCloudOption" style="height: 300px;" autoresize />
      </div>

      <div v-if="topicsData.length" class="chart-section viz-half">
        <h2>Topic Distribution</h2>
        <v-chart :option="topicsChartOption" style="height: 300px;" autoresize />
      </div>
    </div>

    <div v-if="cooccurrenceData.length" class="chart-section">
      <h2>Term Co-occurrence</h2>
      <v-chart :option="cooccurrenceChartOption" style="height: 400px;" autoresize />
    </div>

    <div v-if="isReadonly" class="readonly-banner">
      <p>You are viewing a shared dashboard snapshot. Filters and editing are disabled.</p>
    </div>

    <div v-if="!isReadonly" class="saved-views-section">
      <h2>Saved Views</h2>
      <div class="saved-views-controls">
        <input v-model="newViewName" type="text" placeholder="View name..." class="form-input" />
        <button @click="saveView" :disabled="!newViewName" class="btn-primary">Save Current View</button>
      </div>
      <div v-for="view in savedViews" :key="view.id" class="saved-view-item">
        <span>{{ view.name }}</span>
        <button @click="loadView(view)" class="btn-sm">Load</button>
        <button @click="shareView(view.id)" class="btn-sm">Share</button>
        <button @click="deleteView(view.id)" class="btn-sm btn-danger">Delete</button>
      </div>
      <p v-if="shareLink" class="share-link">Share link: <code>{{ shareLink }}</code></p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import apiClient from '@/api/client'
import VChart from 'vue-echarts'
import 'echarts'
import 'echarts-wordcloud'

const props = withDefaults(defineProps<{
  sharedToken?: string
  readonly?: boolean
}>(), { sharedToken: '', readonly: false })

const isReadonly = ref(props.readonly)
const itemOptions = ref<any[]>([])

const filters = reactive({ item_id: '', start_date: '', end_date: '', sentiment: '', keywords: '' })
const dashboardData = ref<any>(null)
const savedViews = ref<any[]>([])
const newViewName = ref('')
const shareLink = ref('')
const loading = ref(false)
const drillDownSessions = ref<any[]>([])
const timelineEvents = ref<any[]>([])
const sentimentData = ref<any[]>([])
const keywordsData = ref<any[]>([])
const topicsData = ref<any[]>([])
const cooccurrenceData = ref<any[]>([])
const vizLoading = ref(false)

const chartOption = computed(() => {
  if (!dashboardData.value?.data?.length) return {}
  const rows = dashboardData.value.data
  return {
    tooltip: { trigger: 'axis' },
    legend: { data: ['Impressions', 'Clicks', 'Favorites'] },
    xAxis: { type: 'category', data: rows.map((r: any) => r.period_start) },
    yAxis: { type: 'value' },
    series: [
      { name: 'Impressions', type: 'line', data: rows.map((r: any) => r.impressions) },
      { name: 'Clicks', type: 'bar', data: rows.map((r: any) => r.clicks) },
      { name: 'Favorites', type: 'line', data: rows.map((r: any) => r.favorites), lineStyle: { type: 'dashed' } },
    ],
  }
})

const sentimentChartOption = computed(() => {
  if (!sentimentData.value.length) return {}
  const labels = sentimentData.value.map((d: any) => d.sentiment_label)
  const counts = sentimentData.value.map((d: any) => d.count)
  const confidences = sentimentData.value.map((d: any) => +(d.avg_confidence * 100).toFixed(1))
  const colorMap: Record<string, string> = { positive: '#91cc75', neutral: '#fac858', negative: '#ee6666' }
  return {
    tooltip: { trigger: 'axis' },
    legend: { data: ['Count', 'Avg Confidence (%)'] },
    xAxis: { type: 'category', data: labels },
    yAxis: [
      { type: 'value', name: 'Count' },
      { type: 'value', name: 'Confidence %', max: 100 },
    ],
    series: [
      {
        name: 'Count',
        type: 'bar',
        data: counts.map((v: number, i: number) => ({
          value: v,
          itemStyle: { color: colorMap[labels[i]] || '#646cff' },
        })),
      },
      {
        name: 'Avg Confidence (%)',
        type: 'line',
        yAxisIndex: 1,
        data: confidences,
        lineStyle: { width: 2 },
        symbol: 'circle',
        symbolSize: 8,
      },
    ],
  }
})

const wordCloudOption = computed(() => {
  if (!keywordsData.value.length) return {}
  const maxW = Math.max(...keywordsData.value.map((k: any) => k.weight))
  return {
    series: [{
      type: 'wordCloud',
      sizeRange: [14, 60],
      rotationRange: [-45, 45],
      gridSize: 8,
      shape: 'circle',
      textStyle: { fontFamily: 'sans-serif', color: () => {
        const hue = Math.floor(Math.random() * 360)
        return `hsl(${hue}, 60%, 50%)`
      }},
      data: keywordsData.value.map((k: any) => ({
        name: k.keyword,
        value: Math.round((k.weight / maxW) * 100),
      })),
    }],
  }
})

const topicsChartOption = computed(() => {
  if (!topicsData.value.length) return {}
  return {
    tooltip: { trigger: 'item' },
    series: [{
      type: 'pie',
      radius: ['30%', '70%'],
      avoidLabelOverlap: true,
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      label: { show: true, formatter: '{b}: {c}' },
      data: topicsData.value.map((t: any) => ({ name: t.topic, value: t.count })),
    }],
  }
})

const cooccurrenceChartOption = computed(() => {
  if (!cooccurrenceData.value.length) return {}
  const nodeSet = new Set<string>()
  cooccurrenceData.value.forEach((d: any) => { nodeSet.add(d.term_a); nodeSet.add(d.term_b) })
  const nodes = Array.from(nodeSet).map(name => ({
    name,
    symbolSize: 20 + cooccurrenceData.value
      .filter((d: any) => d.term_a === name || d.term_b === name)
      .reduce((s: number, d: any) => s + d.frequency, 0) * 2,
  }))
  const links = cooccurrenceData.value.map((d: any) => ({
    source: d.term_a,
    target: d.term_b,
    value: d.frequency,
    lineStyle: { width: Math.max(1, Math.min(d.frequency, 8)) },
  }))
  return {
    tooltip: {},
    series: [{
      type: 'graph',
      layout: 'force',
      roam: true,
      label: { show: true, position: 'right' },
      force: { repulsion: 200, edgeLength: [80, 200] },
      data: nodes,
      links,
      lineStyle: { opacity: 0.6, curveness: 0.1 },
    }],
  }
})

async function fetchVisualizations() {
  vizLoading.value = true
  const params = { ...filters }
  const [kw, tp, co, se] = await Promise.allSettled([
    apiClient.get('/analytics/keywords', { params }),
    apiClient.get('/analytics/topics', { params }),
    apiClient.get('/analytics/cooccurrences', { params }),
    apiClient.get('/analytics/sentiment', { params }),
  ])
  keywordsData.value = kw.status === 'fulfilled' ? (kw.value.data.data || []) : []
  topicsData.value = tp.status === 'fulfilled' ? (tp.value.data.data || []) : []
  cooccurrenceData.value = co.status === 'fulfilled' ? (co.value.data.data || []) : []
  sentimentData.value = se.status === 'fulfilled' ? (se.value.data.data || []) : []
  vizLoading.value = false
}

async function drillDown(row: any) {
  try {
    const periodStart = typeof row.period_start === 'string'
      ? row.period_start.slice(0, 10)
      : new Date(row.period_start).toISOString().slice(0, 10)
    const { data } = await apiClient.get('/analytics/aggregate-sessions', {
      params: { item_id: row.item_id, period_start: periodStart },
    })
    drillDownSessions.value = data.data || []
  } catch { drillDownSessions.value = [] }
}

async function loadTimeline(sessionUUID: string) {
  try {
    const { data } = await apiClient.get(`/analytics/sessions/${sessionUUID}/timeline`)
    timelineEvents.value = data.events || data.data || []
  } catch { timelineEvents.value = [] }
}

async function fetchDashboard() {
  loading.value = true

  if (isReadonly.value && props.sharedToken) {
    // Shared view: fetch all data via the token-authorized endpoint
    try {
      const { data } = await apiClient.get(`/shared/${props.sharedToken}/data`)
      dashboardData.value = data.dashboard || null
      keywordsData.value = data.keywords?.data || []
      topicsData.value = data.topics?.data || []
      cooccurrenceData.value = data.cooccurrences?.data || []
      sentimentData.value = data.sentiment?.data || []
    } catch { dashboardData.value = null }
    loading.value = false
    return
  }

  try {
    const { data } = await apiClient.get('/analytics/dashboard', { params: filters })
    dashboardData.value = data
  } catch { dashboardData.value = null }
  loading.value = false
  fetchVisualizations()
}

async function loadSavedViews() {
  try {
    const { data } = await apiClient.get('/analytics/saved-views')
    savedViews.value = data.data || []
  } catch { savedViews.value = [] }
}

async function saveView() {
  try {
    await apiClient.post('/analytics/saved-views', { name: newViewName.value, filter_config: { ...filters } })
    newViewName.value = ''
    loadSavedViews()
  } catch (e: any) { alert(e.response?.data?.msg || 'Failed to save') }
}

async function loadView(view: any) {
  const config = typeof view.filter_config === 'string' ? JSON.parse(view.filter_config) : view.filter_config
  // Reset all filter fields first to avoid stale keys from prior views
  Object.assign(filters, { item_id: '', start_date: '', end_date: '', sentiment: '', keywords: '' }, config)
  fetchDashboard()
}

async function shareView(id: string) {
  try {
    const { data } = await apiClient.post(`/analytics/saved-views/${id}/share`)
    shareLink.value = `${window.location.origin}/analytics/shared/${data.token}`
  } catch (e: any) { alert(e.response?.data?.msg || 'Failed to share') }
}

async function deleteView(id: string) {
  try { await apiClient.delete(`/analytics/saved-views/${id}`); loadSavedViews() } catch {}
}

async function loadItems() {
  try {
    const { data } = await apiClient.get('/items', { params: { per_page: 100 } })
    itemOptions.value = data.data || data || []
  } catch { itemOptions.value = [] }
}

async function loadSharedView() {
  if (!props.sharedToken) return
  try {
    const { data } = await apiClient.get(`/shared/${props.sharedToken}`)
    isReadonly.value = true
    const config = typeof data.filter_config === 'string' ? JSON.parse(data.filter_config) : data.filter_config
    Object.assign(filters, { item_id: '', start_date: '', end_date: '', sentiment: '', keywords: '' }, config)
  } catch { /* shared link expired or invalid — load default dashboard */ }
}

onMounted(async () => {
  await loadItems()
  if (props.sharedToken) {
    await loadSharedView()
  }
  fetchDashboard()
  if (!isReadonly.value) {
    loadSavedViews()
  }
})
</script>

<style scoped>
.dashboard-page { max-width: 1200px; }
.filter-bar { display: flex; gap: 1rem; align-items: flex-end; flex-wrap: wrap; margin-bottom: 1.5rem; padding: 1rem; background: white; border-radius: 8px; border: 1px solid #e0e0e0; }
.filter-group { display: flex; flex-direction: column; gap: 0.3rem; }
.filter-group label { font-size: 0.85rem; font-weight: 600; }
.form-input { padding: 0.5rem; border: 1px solid #ddd; border-radius: 4px; }
.btn-primary { padding: 0.5rem 1rem; background: #646cff; color: white; border: none; border-radius: 4px; cursor: pointer; }
.btn-sm { padding: 0.3rem 0.6rem; border: 1px solid #ddd; border-radius: 4px; background: white; cursor: pointer; font-size: 0.85rem; }
.btn-danger { color: #e53e3e; border-color: #e53e3e; }
.data-table { width: 100%; border-collapse: collapse; background: white; border-radius: 8px; overflow: hidden; }
.data-table th, .data-table td { padding: 0.75rem; text-align: left; border-bottom: 1px solid #eee; }
.data-table th { background: #f7f7f7; font-weight: 600; font-size: 0.85rem; }
.loading { text-align: center; padding: 2rem; color: #888; }
.saved-views-section { margin-top: 2rem; }
.saved-views-controls { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
.saved-view-item { display: flex; align-items: center; gap: 0.5rem; padding: 0.5rem; border-bottom: 1px solid #eee; }
.saved-view-item span { flex: 1; }
.share-link { background: #f0f0f0; padding: 0.5rem; border-radius: 4px; margin-top: 0.5rem; }
.share-link code { word-break: break-all; }
.chart-section { margin: 1.5rem 0; background: white; padding: 1rem; border-radius: 8px; border: 1px solid #e0e0e0; }
.drill-down-section, .timeline-section { margin-top: 1.5rem; background: #fafafa; padding: 1rem; border-radius: 8px; border: 1px solid #e0e0e0; }
.timeline-list { list-style: none; padding: 0; }
.timeline-list li { padding: 0.4rem 0; border-bottom: 1px solid #eee; }
.viz-row { display: flex; gap: 1.5rem; flex-wrap: wrap; }
.viz-half { flex: 1; min-width: 300px; }
.readonly-banner { background: #fff3cd; border: 1px solid #ffc107; padding: 0.75rem 1rem; border-radius: 8px; margin-bottom: 1rem; }
.readonly-banner p { margin: 0; color: #856404; font-size: 0.9rem; }
</style>
