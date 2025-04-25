<!-- src/providers/ThemeProvider.vue -->
<script setup lang="ts">
import { ref, provide, computed, watchEffect } from 'vue'
import {
  darkTheme,
  lightTheme,
  useOsTheme
} from 'naive-ui'
import { ThemeKey, themeOptions, type ThemeOption } from '@/utils/theme'

// Current theme preference - load from localStorage if available
const storedTheme = localStorage.getItem('theme') as ThemeOption | null
const themePreference = ref<ThemeOption>(
  themeOptions.includes(storedTheme as ThemeOption) ? storedTheme as ThemeOption : 'auto'
)

// OS theme detection
const osThemeRef = useOsTheme()

// Computed actual theme
const currentTheme = computed(() => {
  if (themePreference.value === 'auto') {
    return osThemeRef.value === 'dark' ? darkTheme : lightTheme
  }
  return themePreference.value === 'dark' ? darkTheme : lightTheme
})

// Is dark mode active?
const isDark = computed(() => {
  if (themePreference.value === 'auto') {
    return osThemeRef.value === 'dark'
  }
  return themePreference.value === 'dark'
})

// Custom theme overrides
const themeOverrides = {
  common: {
    primaryColor: '#3366ff',
    primaryColorHover: '#5c85ff',
    primaryColorPressed: '#254edb',
    infoColor: '#2080f0',
    successColor: '#18a058',
    warningColor: '#f0a020',
    errorColor: '#d03050'
  },
  Button: {
    textColor: isDark.value ? '#ffffff' : undefined
  }
}

// Switch theme function
const switchTheme = (theme: ThemeOption) => {
  console.log('Switching theme to:', theme)
  themePreference.value = theme
  localStorage.setItem('theme', theme)
}

// Apply a CSS class to the body for global styling
watchEffect(() => {
  if (isDark.value) {
    document.body.classList.add('dark-theme')
    document.body.classList.remove('light-theme')
  } else {
    document.body.classList.add('light-theme')
    document.body.classList.remove('dark-theme')
  }
})

// Provide theme-related values to child components
provide(ThemeKey, {
  current: currentTheme,
  overrides: themeOverrides,
  isDark,
  switchTheme,
  themeOptions
})
</script>

<template>
  <n-config-provider
    :theme="currentTheme"
    :theme-overrides="themeOverrides"
  >
    <n-loading-bar-provider>
      <n-dialog-provider>
        <n-notification-provider>
          <n-message-provider>
            <slot></slot>
          </n-message-provider>
        </n-notification-provider>
      </n-dialog-provider>
    </n-loading-bar-provider>
  </n-config-provider>
</template>

<style>
/* Global theme variables */
:root {
  --primary-color: #3366ff;
  --background-color: #ffffff;
  --text-color: #333333;
  --border-color: #e5e5e5;
}

body.dark-theme {
  --background-color: #121212;
  --text-color: #e0e0e0;
  --border-color: #333333;
  background-color: var(--background-color);
  color: var(--text-color);
}

body.light-theme {
  --background-color: #ffffff;
  --text-color: #333333;
  --border-color: #e5e5e5;
  background-color: var(--background-color);
  color: var(--text-color);
}
</style>
