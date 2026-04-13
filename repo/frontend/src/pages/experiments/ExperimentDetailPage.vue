<template>
  <div class="experiment-detail" v-if="experiment">
    <h1>{{ experiment.name }}</h1>
    <p class="desc">{{ experiment.description }}</p>
    <span class="badge" :class="'badge-' + experiment.status">{{ experiment.status }}</span>

    <div class="actions" v-if="experiment.status !== 'completed' && experiment.status !== 'rolled_back'">
      <button v-if="experiment.status === 'draft'" @click="action('start')" class="btn-primary">Start</button>
      <button v-if="experiment.status === 'running'" @click="action('pause')" class="btn-warning">Pause</button>
      <button v-if="experiment.status === 'running' || experiment.status === 'paused'" @click="action('complete')" class="btn-primary">Complete</button>
      <button v-if="experiment.status === 'running' || experiment.status === 'paused'" @click="action('rollback')" class="btn-danger">Rollback</button>
    </div>

    <h2>Variants</h2>
    <table class="data-table">
      <thead><tr><th>Name</th><th>Traffic %</th><th v-if="canEditTraffic">Adjust</th></tr></thead>
      <tbody>
        <tr v-for="v in experiment.variants" :key="v.name">
          <td>{{ v.name }}</td>
          <td>{{ v.traffic_pct }}%</td>
          <td v-if="canEditTraffic">
            <input type="number" min="0" max="100" step="1" v-model.number="trafficEdits[v.name]" class="traffic-input" />
          </td>
        </tr>
      </tbody>
    </table>
    <div v-if="canEditTraffic" class="traffic-controls">
      <span class="traffic-sum" :class="{ invalid: trafficSum !== 100 }">Sum: {{ trafficSum }}%</span>
      <button @click="saveTraffic" :disabled="trafficSum !== 100" class="btn-primary">Save Traffic</button>
    </div>

    <h2>Results</h2>
    <div v-if="results">
      <p><strong>Confidence State:</strong> <span class="badge" :class="'badge-' + results.confidence_state">{{ results.confidence_state }}</span></p>
      <table class="data-table" v-if="results.variants">
        <thead><tr><th>Variant</th><th>Sample Size</th><th>Exposures</th></tr></thead>
        <tbody><tr v-for="v in results.variants" :key="v.name"><td>{{ v.name }}</td><td>{{ v.sample_size }}</td><td>{{ v.exposures }}</td></tr></tbody>
      </table>
    </div>
    <p v-else class="empty">No results yet</p>
  </div>
  <div v-else class="loading">Loading...</div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import apiClient from '@/api/client'

const props = defineProps<{ id: string }>()
const experiment = ref<any>(null)
const results = ref<any>(null)
const trafficEdits = ref<Record<string, number>>({})

const canEditTraffic = computed(() => {
  const s = experiment.value?.status
  return s === 'running' || s === 'paused'
})

const trafficSum = computed(() =>
  Object.values(trafficEdits.value).reduce((sum: number, v) => sum + (v as number), 0)
)

watch(experiment, (exp) => {
  if (exp?.variants) {
    const edits: Record<string, number> = {}
    for (const v of exp.variants) { edits[v.name] = v.traffic_pct }
    trafficEdits.value = edits
  }
})

async function fetchExperiment() {
  try {
    const { data } = await apiClient.get(`/experiments/${props.id}`)
    experiment.value = data
  } catch {}
}

async function fetchResults() {
  try {
    const { data } = await apiClient.get(`/experiments/${props.id}/results`)
    results.value = data
  } catch {}
}

async function saveTraffic() {
  const variants = Object.entries(trafficEdits.value).map(([name, pct]) => ({ name, traffic_pct: pct }))
  try {
    await apiClient.put(`/experiments/${props.id}/traffic`, { variants })
    fetchExperiment()
  } catch (e: any) { alert(e.response?.data?.msg || 'Failed to update traffic') }
}

async function action(act: string) {
  try {
    await apiClient.post(`/experiments/${props.id}/${act}`)
    fetchExperiment()
    fetchResults()
  } catch (e: any) { alert(e.response?.data?.msg || 'Action failed') }
}

onMounted(() => { fetchExperiment(); fetchResults() })
</script>

<style scoped>
.experiment-detail { max-width: 800px; }
.desc { color: #666; margin-bottom: 1rem; }
.actions { display: flex; gap: 0.5rem; margin: 1rem 0; }
.btn-primary { padding: 0.5rem 1rem; background: #646cff; color: white; border: none; border-radius: 4px; cursor: pointer; }
.btn-warning { padding: 0.5rem 1rem; background: #f59e0b; color: white; border: none; border-radius: 4px; cursor: pointer; }
.btn-danger { padding: 0.5rem 1rem; background: #e53e3e; color: white; border: none; border-radius: 4px; cursor: pointer; }
.data-table { width: 100%; border-collapse: collapse; background: white; margin: 1rem 0; }
.data-table th, .data-table td { padding: 0.75rem; text-align: left; border-bottom: 1px solid #eee; }
.data-table th { background: #f7f7f7; }
.badge { padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.8rem; display: inline-block; }
.badge-draft { background: #f3f4f6; } .badge-running { background: #d1fae5; } .badge-paused { background: #fef3c7; }
.badge-completed { background: #dbeafe; } .badge-rolled_back { background: #fce7f3; }
.badge-insufficient_data { background: #f3f4f6; } .badge-monitoring { background: #fef3c7; }
.badge-recommend_keep { background: #d1fae5; } .badge-recommend_rollback { background: #fce7f3; }
.loading, .empty { text-align: center; padding: 2rem; color: #888; }
.traffic-input { width: 70px; padding: 0.3rem; border: 1px solid #ddd; border-radius: 4px; text-align: center; }
.traffic-controls { display: flex; align-items: center; gap: 1rem; margin: 0.5rem 0 1rem; }
.traffic-sum { font-size: 0.9rem; font-weight: 600; }
.traffic-sum.invalid { color: #e53e3e; }
</style>
