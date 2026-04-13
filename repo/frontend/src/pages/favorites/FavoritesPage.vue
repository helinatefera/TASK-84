<template>
  <div class="favorites-page">
    <h1>Favorites & Wishlists</h1>
    <div class="tabs">
      <button :class="{ active: tab === 'favorites' }" @click="tab = 'favorites'">Favorites</button>
      <button :class="{ active: tab === 'wishlists' }" @click="tab = 'wishlists'">Wishlists</button>
    </div>

    <div v-if="tab === 'favorites'" class="tab-content">
      <div v-if="favLoading" class="loading">Loading...</div>
      <div v-for="fav in favorites" :key="fav.item_id" class="fav-item">
        <span>Item #{{ fav.item_id }}</span>
        <button @click="removeFavorite(fav.item_id)" class="btn-sm btn-danger">Remove</button>
      </div>
      <p v-if="!favLoading && !favorites.length" class="empty">No favorites yet</p>
    </div>

    <div v-if="tab === 'wishlists'" class="tab-content">
      <div class="create-row">
        <input v-model="newWishlistName" placeholder="New wishlist name..." class="form-input" />
        <button @click="createWishlist" :disabled="!newWishlistName" class="btn-primary">Create</button>
      </div>
      <div v-for="wl in wishlists" :key="wl.id" class="wishlist-item">
        <div class="wl-header">
          <strong>{{ wl.name }}</strong>
          <button @click="deleteWishlist(wl.id)" class="btn-sm btn-danger">Delete</button>
        </div>
      </div>
      <p v-if="!wishlists.length" class="empty">No wishlists yet</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import apiClient from '@/api/client'

const tab = ref('favorites')
const favorites = ref<any[]>([])
const wishlists = ref<any[]>([])
const favLoading = ref(false)
const newWishlistName = ref('')

async function fetchFavorites() {
  favLoading.value = true
  try { const { data } = await apiClient.get('/favorites'); favorites.value = data.data || data || [] } catch { favorites.value = [] }
  favLoading.value = false
}

async function removeFavorite(itemId: number) {
  try { await apiClient.delete(`/favorites/${itemId}`); fetchFavorites() } catch {}
}

async function fetchWishlists() {
  try { const { data } = await apiClient.get('/wishlists'); wishlists.value = data.data || data || [] } catch { wishlists.value = [] }
}

async function createWishlist() {
  try { await apiClient.post('/wishlists', { name: newWishlistName.value }); newWishlistName.value = ''; fetchWishlists() } catch {}
}

async function deleteWishlist(id: string) {
  try { await apiClient.delete(`/wishlists/${id}`); fetchWishlists() } catch {}
}

onMounted(() => { fetchFavorites(); fetchWishlists() })
</script>

<style scoped>
.favorites-page { max-width: 800px; }
.tabs { display: flex; border-bottom: 2px solid #e0e0e0; margin-bottom: 1rem; }
.tabs button { padding: 0.6rem 1.2rem; border: none; background: none; cursor: pointer; color: #888; border-bottom: 2px solid transparent; margin-bottom: -2px; }
.tabs button.active { color: #646cff; border-bottom-color: #646cff; }
.fav-item, .wishlist-item { display: flex; justify-content: space-between; align-items: center; padding: 0.75rem; border-bottom: 1px solid #eee; background: white; }
.wl-header { display: flex; justify-content: space-between; align-items: center; width: 100%; }
.create-row { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
.form-input { padding: 0.5rem; border: 1px solid #ddd; border-radius: 4px; flex: 1; }
.btn-primary { padding: 0.5rem 1rem; background: #646cff; color: white; border: none; border-radius: 4px; cursor: pointer; }
.btn-sm { padding: 0.3rem 0.6rem; border: 1px solid #ddd; border-radius: 4px; background: white; cursor: pointer; }
.btn-danger { color: #e53e3e; border-color: #e53e3e; }
.loading, .empty { text-align: center; padding: 2rem; color: #888; }
</style>
