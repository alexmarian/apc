<script setup lang="ts">
import { NCard } from 'naive-ui'

defineProps<{
  /** Card title. Omit on success screens, where NResult supplies its own heading. */
  title?: string
}>()
</script>

<template>
  <div class="auth-layout">
    <div class="auth-container">
      <div class="auth-logo">
        <img src="/logo_128.png" alt="APC Logo" class="logo-img" width="128" height="128" />
      </div>
      <NCard :title="title">
        <slot></slot>
      </NCard>
    </div>
  </div>
</template>

<style scoped>
/* No background here: NGlobalStyle paints the body from the active naive theme. */
.auth-layout {
  min-height: 100vh;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

/* Horizontal padding is only a gutter for narrow viewports — the card supplies
 * the form's own inset, so the box stays roughly its previous 400px. */
.auth-container {
  width: 100%;
  max-width: 432px;
  padding: 2rem 1rem;
}

.auth-logo {
  display: flex;
  justify-content: center;
  margin-bottom: 2rem;
}

.logo-img {
  width: 128px;
  height: 128px;
}
</style>
