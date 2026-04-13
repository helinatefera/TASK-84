<template>
  <div class="auth-layout">
    <div class="auth-card">
      <h1 class="auth-title">Local Insights Portal</h1>
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.auth-layout {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
}
.auth-card {
  background: white;
  border-radius: 8px;
  padding: 2rem;
  width: 100%;
  max-width: 420px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}
.auth-title {
  text-align: center;
  margin-bottom: 1.5rem;
  font-size: 1.5rem;
  color: #333;
}
</style>
