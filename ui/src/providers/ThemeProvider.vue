<script setup lang="ts">
import { ref, provide, computed } from 'vue'
import {
  darkTheme,
  lightTheme,
  useOsTheme,
  createDiscreteThemeOverrides,
  GlobalThemeOverrides
} from 'naive-ui'

// Theme options
const themeOptions = ['light', 'dark', 'auto'] as const
type ThemeOption = typeof themeOptions[number]

// Current theme preference
const themePreference = ref<ThemeOption>(
  localStorage.getItem('theme') as ThemeOption || 'auto'
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
const themeOverrides: GlobalThemeOverrides = {
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
    textColor: '#FFF'
  }
}

// Switch theme function
const switchTheme = (theme: ThemeOption) => {
  themePreference.value = theme
  localStorage.setItem('theme', theme)
}

// Provide theme-related values to child components
provide('theme', {
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
