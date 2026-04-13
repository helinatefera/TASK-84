<template>
  <div class="experiments-page">
    <div class="page-header">
      <h1>Experiments</h1>
      <button @click="showCreate = !showCreate" class="btn-primary">{{ showCreate ? 'Cancel' : 'Create Experiment' }}</button>
    </div>

    <div v-if="showCreate" class="create-form">
      <div class="form-group"><label>Name</label><input v-model="form.name" class="form-input" /></div>
      <div class="form-group"><label>Description</label><textarea v-model="form.description" class="form-input"></textarea></div>
      <div class="form-group"><label>Min Sample Size</label><input v-model.number="form.min_sample_size" type="number" min="50" class="form-input" /></div>
      <h3>Variants (traffic % must sum to 100)</h3>
      <div v-for="(v, i) in form.variants" :key="i" class="variant-row">
        <input v-model="v.name" placeholder="Variant name" class="form-input" />
        <input v-model.number="v.traffic_pct" type="number" min="0" max="100" placeholder="%" class="form-input small" />
        <button @click="form.variants.splice(i, 1)" v-if="form.variants.length > 2" class="btn-sm btn-danger">X</button>
      </div>
      <button @click="form.variants.push({ name: '', traffic_pct: 0, config: '{}' })" class="btn-sm">+ Add Variant</button>
      <p v-if="createError" class="error-text">{{ createError }}</p>
      <button @click="createExperiment" class="btn-primary" style="margin-top:1rem">Create</button>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <table v-if="!loading && experiments.length" class="data-table">
      <thead><tr><th>Name</th><th>Status</th><th>Created</th><th>Actions</th></tr></thead>
      <tbody>
        <tr v-for="exp in experiments" :key="exp.id">
          <td><router-link :to="{ name: 'experimentDetail', params: { id: exp.id } }">{{ exp.name }}</router-link></td>
          <td><span class="badge" :class="'badge-' + exp.status">{{ exp.status }}</span></td>
          <td>{{ exp.created_at?.slice(0, 10) }}</td>
          <td><router-link :to="{ name: 'experimentDetail', params: { id: exp.id } }" class="btn-sm">View</router-link></td>
        </tr>
      </tbody>
    </table>
    <p v-if="!loading && !experiments.length" class="empty">No experiments yet</p>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import apiClient from '@/api/client'

const router = useRouter()
const experiments = ref<any[]>([])
const loading = ref(false)
const showCreate = ref(false)
const createError = ref('')
const form = reactive({
  name: '', description: '', min_sample_size: 100,
  variants: [{ name: 'control', traffic_pct: 50, config: '{}' }, { name: 'variant_a', traffic_pct: 50, config: '{}' }]
})

async function fetchExperiments() {
  loading.value = true
  try {
    const { data } = await apiClient.get('/experiments')
    experiments.value = data.data || data || []
  } catch { experiments.value = [] }
  loading.value = false
}

async function createExperiment() {
  createError.value = ''
  const sum = form.variants.reduce((s, v) => s + v.traffic_pct, 0)
  if (sum !== 100) { createError.value = `Traffic must sum to 100% (currently ${sum}%)`; return }
  try {
    const payload = {
      name: form.name, description: form.description, min_sample_size: form.min_sample_size,
      variants: form.variants.map(v => ({ name: v.name, traffic_pct: v.traffic_pct, config: JSON.parse(v.config || '{}') }))
    }
    await apiClient.post('/experiments', payload)
    showCreate.value = false
    fetchExperiments()
  } catch (e: any) { createError.value = e.response?.data?.msg || 'Failed to create' }
}

onMounted(fetchExperiments)
</script>

<style scoped>
.experiments-page { max-width: 900px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
.create-form { background: white; padding: 1.5rem; border-radius: 8px; border: 1px solid #e0e0e0; margin-bottom: 1.5rem; }
.form-group { margin-bottom: 1rem; display: flex; flex-direction: column; gap: 0.3rem; }
.form-group label { font-weight: 600; font-size: 0.9rem; }
.form-input { padding: 0.5rem; border: 1px solid #ddd; border-radius: 4px; }
.form-input.small { width: 80px; }
.variant-row { display: flex; gap: 0.5rem; margin-bottom: 0.5rem; align-items: center; }
.btn-primary { padding: 0.5rem 1rem; background: #646cff; color: white; border: none; border-radius: 4px; cursor: pointer; }
.btn-sm { padding: 0.3rem 0.6rem; border: 1px solid #ddd; border-radius: 4px; background: white; cursor: pointer; font-size: 0.85rem; text-decoration: none; color: #333; }
.btn-danger { color: #e53e3e; border-color: #e53e3e; }
.data-table { width: 100%; border-collapse: collapse; background: white; border-radius: 8px; overflow: hidden; }
.data-table th, .data-table td { padding: 0.75rem; text-align: left; border-bottom: 1px solid #eee; }
.data-table th { background: #f7f7f7; }
.data-table a { color: #646cff; text-decoration: none; }
.badge { padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.8rem; }
.badge-draft { background: #f3f4f6; color: #6b7280; }
.badge-running { background: #d1fae5; color: #065f46; }
.badge-paused { background: #fef3c7; color: #92400e; }
.badge-completed { background: #dbeafe; color: #1e40af; }
.badge-rolled_back { background: #fce7f3; color: #9d174d; }
.error-text { color: #e53e3e; }
.loading, .empty { text-align: center; padding: 2rem; color: #888; }
</style>
