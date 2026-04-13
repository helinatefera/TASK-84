<template>
  <div class="catalog-page">
    <div class="catalog-header">
      <h1>Catalog</h1>
      <input v-model="searchQuery" type="text" placeholder="Search items..." class="search-input" @input="debouncedFetch" />
    </div>
    <div v-if="loading" class="loading">Loading...</div>
    <div v-else class="item-grid">
      <div v-for="item in items" :key="item.id" class="item-card" @click="goToItem(item.id)">
        <h3>{{ item.title }}</h3>
        <p class="item-category">{{ item.category || 'Uncategorized' }}</p>
        <p class="item-desc">{{ item.description?.slice(0, 100) }}{{ (item.description?.length ?? 0) > 100 ? '...' : '' }}</p>
      </div>
    </div>
    <div v-if="!loading && items.length === 0" class="empty-state">No items found</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { itemsApi } from '@/api/endpoints/items'

const router = useRouter()
const items = ref<any[]>([])
const searchQuery = ref('')
const loading = ref(false)
let debounceTimer: ReturnType<typeof setTimeout>

async function fetchItems() {
  loading.value = true
  try {
    const { data } = await itemsApi.list({ search: searchQuery.value || undefined })
    items.value = data.data || data
  } catch {
    items.value = []
  } finally {
    loading.value = false
  }
}

function debouncedFetch() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(fetchItems, 300)
}

function goToItem(id: string) {
  router.push({ name: 'itemDetail', params: { id } })
}

onMounted(fetchItems)
</script>

<style scoped>
.catalog-page { max-width: 1200px; }
.catalog-header { display: flex; align-items: center; gap: 1rem; margin-bottom: 1.5rem; }
.catalog-header h1 { margin: 0; }
.search-input { flex: 1; padding: 0.6rem; border: 1px solid #ddd; border-radius: 4px; font-size: 1rem; }
.item-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 1rem; }
.item-card { background: white; border-radius: 8px; padding: 1.2rem; cursor: pointer; transition: box-shadow 0.2s; border: 1px solid #e0e0e0; }
.item-card:hover { box-shadow: 0 2px 12px rgba(0,0,0,0.1); }
.item-card h3 { margin: 0 0 0.5rem; }
.item-category { color: #646cff; font-size: 0.85rem; margin: 0 0 0.5rem; }
.item-desc { color: #666; font-size: 0.9rem; margin: 0; }
.loading { text-align: center; padding: 2rem; color: #888; }
.empty-state { text-align: center; padding: 3rem; color: #888; }
</style>
