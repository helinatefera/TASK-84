<template>
  <div class="item-detail" v-if="item">
    <div class="item-header">
      <h1>{{ item.title }}</h1>
      <span class="item-category">{{ item.category }}</span>
    </div>
    <p class="item-description">{{ item.description }}</p>

    <div class="tabs">
      <button :class="{ active: activeTab === 'reviews' }" @click="activeTab = 'reviews'">Reviews</button>
      <button :class="{ active: activeTab === 'qa' }" @click="activeTab = 'qa'; fetchQuestions()">Q&A</button>
    </div>

    <div v-if="activeTab === 'reviews'" class="tab-content">
      <router-link :to="{ name: 'writeReview', params: { itemId: item.id } }" class="write-review-btn">Write a Review</router-link>
      <div v-for="review in reviews" :key="review.id" class="review-card">
        <div class="review-rating">{{ '\u2605'.repeat(review.rating) }}{{ '\u2606'.repeat(5 - review.rating) }}</div>
        <p>{{ review.body }}</p>
        <small>{{ review.created_at }}</small>
      </div>
      <p v-if="reviews.length === 0" class="empty">No reviews yet</p>
    </div>

    <div v-if="activeTab === 'qa'" class="tab-content">
      <div class="ask-form">
        <h3>Ask a Question</h3>
        <textarea v-model="newQuestionBody" placeholder="Type your question..." class="form-input" rows="3"></textarea>
        <button @click="submitQuestion" :disabled="!newQuestionBody.trim()" class="btn-primary">Submit Question</button>
      </div>

      <div v-if="questionsLoading" class="loading">Loading questions...</div>

      <div v-for="q in questions" :key="q.id" class="question-card">
        <div class="question-body">
          <strong>Q:</strong> {{ q.body }}
        </div>
        <small class="question-meta">Asked {{ q.created_at }}</small>

        <div class="answers-section">
          <div v-for="a in q._answers" :key="a.id" class="answer-item">
            <strong>A:</strong> {{ a.body }}
            <small class="answer-meta">{{ a.created_at }}</small>
          </div>
          <p v-if="q._answersLoaded && q._answers.length === 0" class="empty-sm">No answers yet</p>
          <button v-if="!q._answersLoaded" @click="fetchAnswers(q)" class="btn-sm">Show Answers</button>
        </div>

        <div class="answer-form">
          <textarea v-model="q._newAnswer" placeholder="Write an answer..." class="form-input" rows="2"></textarea>
          <button @click="submitAnswer(q)" :disabled="!q._newAnswer?.trim()" class="btn-sm btn-answer">Post Answer</button>
        </div>
      </div>

      <p v-if="!questionsLoading && questions.length === 0" class="empty">No questions yet. Be the first to ask!</p>
    </div>
  </div>
  <div v-else class="loading">Loading...</div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { itemsApi } from '@/api/endpoints/items'
import { reviewsApi } from '@/api/endpoints/reviews'
import apiClient from '@/api/client'

const props = defineProps<{ id: string }>()
const item = ref<any>(null)
const reviews = ref<any[]>([])
const activeTab = ref('reviews')

const questions = ref<any[]>([])
const questionsLoading = ref(false)
const newQuestionBody = ref('')

async function fetchQuestions() {
  questionsLoading.value = true
  try {
    const { data } = await apiClient.get(`/items/${props.id}/questions`)
    questions.value = (data.data || data || []).map((q: any) => ({
      ...q,
      _answers: [],
      _answersLoaded: false,
      _newAnswer: '',
    }))
  } catch { questions.value = [] }
  questionsLoading.value = false
}

async function submitQuestion() {
  if (!newQuestionBody.value.trim()) return
  try {
    await apiClient.post(`/items/${props.id}/questions`, { body: newQuestionBody.value })
    newQuestionBody.value = ''
    fetchQuestions()
  } catch (e: any) { alert(e.response?.data?.msg || 'Failed to submit question') }
}

async function fetchAnswers(question: any) {
  try {
    const { data } = await apiClient.get(`/questions/${question.id}/answers`)
    question._answers = data.data || data || []
    question._answersLoaded = true
  } catch { question._answers = []; question._answersLoaded = true }
}

async function submitAnswer(question: any) {
  if (!question._newAnswer?.trim()) return
  try {
    await apiClient.post(`/questions/${question.id}/answers`, { body: question._newAnswer })
    question._newAnswer = ''
    fetchAnswers(question)
  } catch (e: any) { alert(e.response?.data?.msg || 'Failed to submit answer') }
}

const experimentVariant = ref<string | null>(null)

// Fetch experiment assignment and record exposure for any running experiment
// that the user may be part of. This drives A/B experiment participation.
async function checkExperimentAssignment() {
  try {
    // Try to get assignment for "item_detail_experiment" (convention: experiments
    // reference this endpoint). The backend returns the assigned variant or 404.
    const { data } = await apiClient.get('/experiments/assignment/item_detail')
    if (data?.variant?.name) {
      experimentVariant.value = data.variant.name
      // Record exposure so the experiment can measure results
      if (data.experiment_id) {
        await apiClient.post(`/experiments/${data.experiment_id}/expose`).catch(() => {})
      }
    }
  } catch {
    // No active experiment or not assigned — proceed normally
  }
}

onMounted(async () => {
  try {
    const { data } = await itemsApi.getById(props.id)
    item.value = data
  } catch { /* handle */ }
  try {
    const { data } = await reviewsApi.listByItem(props.id, {})
    reviews.value = data.data || data
  } catch { /* handle */ }
  // Fire-and-forget experiment check — doesn't block page rendering
  checkExperimentAssignment()
})
</script>

<style scoped>
.item-header { margin-bottom: 1rem; }
.item-header h1 { margin: 0 0 0.5rem; }
.item-category { background: #eef; padding: 0.2rem 0.6rem; border-radius: 4px; font-size: 0.85rem; color: #646cff; }
.item-description { color: #555; margin-bottom: 1.5rem; }
.tabs { display: flex; gap: 0; border-bottom: 2px solid #e0e0e0; margin-bottom: 1rem; }
.tabs button { padding: 0.6rem 1.2rem; border: none; background: none; cursor: pointer; font-size: 1rem; color: #888; border-bottom: 2px solid transparent; margin-bottom: -2px; }
.tabs button.active { color: #646cff; border-bottom-color: #646cff; }
.write-review-btn { display: inline-block; padding: 0.5rem 1rem; background: #646cff; color: white; text-decoration: none; border-radius: 4px; margin-bottom: 1rem; }
.review-card { background: white; padding: 1rem; border-radius: 6px; margin-bottom: 0.75rem; border: 1px solid #e0e0e0; }
.review-rating { color: #f5a623; font-size: 1.1rem; margin-bottom: 0.5rem; }
.ask-form { background: white; padding: 1rem; border-radius: 8px; border: 1px solid #e0e0e0; margin-bottom: 1.5rem; }
.ask-form h3 { margin: 0 0 0.5rem; }
.form-input { width: 100%; padding: 0.5rem; border: 1px solid #ddd; border-radius: 4px; resize: vertical; box-sizing: border-box; }
.btn-primary { padding: 0.5rem 1rem; background: #646cff; color: white; border: none; border-radius: 4px; cursor: pointer; margin-top: 0.5rem; }
.btn-primary:disabled { opacity: 0.5; }
.btn-sm { padding: 0.3rem 0.6rem; border: 1px solid #ddd; border-radius: 4px; background: white; cursor: pointer; font-size: 0.85rem; }
.btn-sm:disabled { opacity: 0.5; }
.btn-answer { background: #646cff; color: white; border-color: #646cff; margin-top: 0.5rem; }
.question-card { background: white; padding: 1rem; border-radius: 8px; border: 1px solid #e0e0e0; margin-bottom: 1rem; }
.question-body { font-size: 1rem; margin-bottom: 0.3rem; }
.question-meta { color: #999; display: block; margin-bottom: 0.75rem; }
.answers-section { margin-left: 1rem; padding-left: 1rem; border-left: 2px solid #e0e0e0; margin-bottom: 0.75rem; }
.answer-item { padding: 0.5rem 0; border-bottom: 1px solid #f0f0f0; }
.answer-item:last-child { border-bottom: none; }
.answer-meta { color: #999; display: block; margin-top: 0.2rem; }
.answer-form { margin-top: 0.5rem; }
.empty { color: #888; text-align: center; padding: 2rem; }
.empty-sm { color: #888; font-size: 0.85rem; padding: 0.5rem 0; }
.loading { text-align: center; padding: 3rem; color: #888; }
</style>
