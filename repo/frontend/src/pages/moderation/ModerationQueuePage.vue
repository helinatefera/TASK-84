<template>
  <div class="moderation-page">
    <h1>Moderation Queue</h1>
    <div class="filter-bar">
      <select v-model="statusFilter" class="form-input" @change="fetchQueue">
        <option value="">All Statuses</option>
        <option value="pending">Pending</option>
        <option value="in_review">In Review</option>
        <option value="resolved">Resolved</option>
        <option value="dismissed">Dismissed</option>
      </select>
      <select v-model="typeFilter" class="form-input" @change="fetchQueue">
        <option value="">All Types</option>
        <option value="review">Review</option>
        <option value="question">Question</option>
        <option value="answer">Answer</option>
      </select>
    </div>

    <div v-if="loading" class="loading">Loading...</div>

    <div v-for="report in reports" :key="report.id" class="report-card">
      <div class="report-header">
        <span class="badge" :class="'badge-' + report.status">{{ report.status }}</span>
        <span class="badge badge-type">{{ report.target_type }}</span>
        <span class="priority">Priority: {{ report.priority }}</span>
        <span class="category">{{ report.category }}</span>
      </div>
      <p v-if="report.description" class="report-desc">{{ report.description }}</p>
      <small class="report-meta">Report #{{ report.id }} | Target ID: {{ report.target_id }} | {{ report.created_at }}</small>

      <div v-if="report.status === 'pending' || report.status === 'in_review'" class="report-actions">
        <textarea v-model="report._notes" placeholder="Resolution notes (required)..." class="form-input notes-input"></textarea>
        <div class="action-buttons">
          <button @click="resolveReport(report, 'resolved')" :disabled="!report._notes" class="btn-primary">Resolve</button>
          <button @click="resolveReport(report, 'dismissed')" :disabled="!report._notes" class="btn-secondary">Dismiss</button>
        </div>
      </div>
    </div>
    <p v-if="!loading && reports.length === 0" class="empty">No reports in queue</p>

    <h2 style="margin-top:2rem">Appeals</h2>
    <div class="filter-bar">
      <select v-model="appealStatusFilter" class="form-input" @change="fetchAppeals">
        <option value="">All Statuses</option>
        <option value="pending">Pending</option>
        <option value="needs_edit">Needs Edit</option>
        <option value="approved">Approved</option>
        <option value="rejected">Rejected</option>
      </select>
    </div>

    <div v-for="appeal in appeals" :key="appeal.id" class="report-card">
      <div class="report-header">
        <span class="badge" :class="'badge-' + appeal.status">{{ appeal.status }}</span>
      </div>
      <p class="report-desc">{{ appeal.body }}</p>
      <small class="report-meta">Appeal #{{ appeal.id }} | Created {{ appeal.created_at }}</small>

      <div v-if="appeal.status === 'pending'" class="report-actions">
        <textarea v-model="appeal._note" placeholder="Moderator note (required)..." class="form-input notes-input"></textarea>
        <div class="action-buttons">
          <button @click="handleAppeal(appeal, 'approved')" :disabled="!appeal._note" class="btn-primary">Approve</button>
          <button @click="handleAppeal(appeal, 'rejected')" :disabled="!appeal._note" class="btn-secondary">Reject</button>
          <button @click="handleAppeal(appeal, 'needs_edit')" :disabled="!appeal._note" class="btn-warning">Needs Edit</button>
        </div>
      </div>
    </div>
    <p v-if="!loading && appeals.length === 0" class="empty">No appeals</p>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import apiClient from '@/api/client'

const reports = ref<any[]>([])
const statusFilter = ref('')
const typeFilter = ref('')
const loading = ref(false)
const appeals = ref<any[]>([])
const appealStatusFilter = ref('')

async function fetchQueue() {
  loading.value = true
  try {
    const params: any = {}
    if (statusFilter.value) params.status = statusFilter.value
    if (typeFilter.value) params.target_type = typeFilter.value
    const { data } = await apiClient.get('/moderation/queue', { params })
    reports.value = (data.data || data || []).map((r: any) => ({ ...r, _notes: '' }))
  } catch { reports.value = [] }
  loading.value = false
}

async function resolveReport(report: any, status: string) {
  if (!report._notes) return
  try {
    await apiClient.put(`/moderation/reports/${report.id}`, {
      status,
      resolution_note: report._notes,
    })
    fetchQueue()
  } catch (e: any) { alert(e.response?.data?.msg || 'Failed') }
}

async function fetchAppeals() {
  try {
    const params: any = {}
    if (appealStatusFilter.value) params.status = appealStatusFilter.value
    const { data } = await apiClient.get('/moderation/appeals', { params })
    appeals.value = (data.data || []).map((a: any) => ({ ...a, _note: '' }))
  } catch { appeals.value = [] }
}

async function handleAppeal(appeal: any, status: string) {
  if (!appeal._note) return
  try {
    await apiClient.put(`/moderation/appeals/${appeal.id}`, { status, note: appeal._note })
    fetchAppeals()
  } catch (e: any) { alert(e.response?.data?.msg || 'Failed') }
}

onMounted(() => { fetchQueue(); fetchAppeals() })
</script>

<style scoped>
.moderation-page { max-width: 900px; }
.filter-bar { display: flex; gap: 1rem; margin-bottom: 1.5rem; }
.form-input { padding: 0.5rem; border: 1px solid #ddd; border-radius: 4px; }
.report-card { background: white; border: 1px solid #e0e0e0; border-radius: 8px; padding: 1rem; margin-bottom: 1rem; }
.report-header { display: flex; gap: 0.5rem; align-items: center; margin-bottom: 0.5rem; }
.badge { padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.8rem; font-weight: 600; }
.badge-pending { background: #fef3c7; color: #92400e; }
.badge-in_review { background: #dbeafe; color: #1e40af; }
.badge-resolved { background: #d1fae5; color: #065f46; }
.badge-dismissed { background: #f3f4f6; color: #6b7280; }
.badge-type { background: #ede9fe; color: #5b21b6; }
.priority { font-size: 0.85rem; color: #666; }
.category { font-size: 0.85rem; color: #646cff; font-weight: 600; }
.report-desc { color: #555; margin: 0.5rem 0; }
.report-meta { color: #999; }
.report-actions { margin-top: 1rem; }
.notes-input { width: 100%; min-height: 60px; margin-bottom: 0.5rem; resize: vertical; }
.action-buttons { display: flex; gap: 0.5rem; }
.btn-primary { padding: 0.5rem 1rem; background: #646cff; color: white; border: none; border-radius: 4px; cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; }
.btn-secondary { padding: 0.5rem 1rem; background: #6b7280; color: white; border: none; border-radius: 4px; cursor: pointer; }
.btn-secondary:disabled { opacity: 0.5; }
.btn-warning { padding: 0.5rem 1rem; background: #f59e0b; color: white; border: none; border-radius: 4px; cursor: pointer; }
.btn-warning:disabled { opacity: 0.5; }
.badge-needs_edit { background: #fef3c7; color: #92400e; }
.badge-approved { background: #d1fae5; color: #065f46; }
.badge-rejected { background: #fce7f3; color: #9d174d; }
.loading, .empty { text-align: center; padding: 2rem; color: #888; }
</style>
