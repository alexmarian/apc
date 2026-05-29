<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { computed } from 'vue'
import { usePreferences } from '@/stores/preferences.ts'

const { t } = useI18n()
const preferences = usePreferences()

// Get current theme to match the main app
const isDarkTheme = computed(() => {
  return preferences.theme === 'darkTheme' || preferences.theme === null
})
</script>

<template>
  <div class="auth-layout" :class="{ 'dark-theme': isDarkTheme, 'light-theme': !isDarkTheme }">
    <div class="auth-container">
      <div class="auth-logo">
        <img src="@/assets/logo.svg" alt="APC Logo" class="logo-img" />
        <h1 class="app-title">APC</h1>
      </div>
      <slot></slot>
    </div>
  </div>
</template>

<style scoped>
.auth-layout {
  min-height: 100vh;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  transition: background-color 0.3s ease, color 0.3s ease;
}

.dark-theme {
  background-color: #181818; /* Match darkTheme background */
  color: rgba(235, 235, 235, 0.64); /* Match dark theme text */
}

.light-theme {
  background-color: #f8f8f8; /* Match lightTheme background */
  color: #2c3e50; /* Match light theme text */
}

.auth-container {
  width: 100%;
  max-width: 400px;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transition: background-color 0.3s ease;
}

.dark-theme .auth-container {
  background-color: #222222; /* Slightly lighter than background for dark theme */
}

.light-theme .auth-container {
  background-color: #ffffff; /* White background for light theme */
}

.auth-logo {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 2rem;
}

.logo-img {
  width: 80px;
  height: 80px;
  margin-bottom: 1rem;
}

.app-title {
  font-size: 1.75rem;
  font-weight: bold;
  color: #41b883; /* Vue green color - consistent across themes */
  margin: 0;
}
</style>
