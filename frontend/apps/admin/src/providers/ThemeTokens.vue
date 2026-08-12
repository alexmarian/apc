<!-- src/providers/ThemeTokens.vue -->
<script setup lang="ts">
import { computed, onUnmounted, watchEffect } from 'vue'
import { useThemeVars } from 'naive-ui'

// Naive UI only exposes its palette as `--n-*` variables scoped to its own
// component roots, so scoped component CSS has nothing to style against.
// This republishes the subset we use onto :root, keyed to the active theme.
// Must be rendered *inside* NConfigProvider for useThemeVars() to resolve the
// configured theme rather than the library default.
const vars = useThemeVars()

const tokens = computed<Record<string, string>>(() => {
  const v = vars.value
  return {
    '--body-color': v.bodyColor,
    '--card-color': v.cardColor,
    '--background-color': v.actionColor,
    '--background-alt-color': v.tableHeaderColor,
    '--hover-color': v.hoverColor,

    '--border-color': v.borderColor,
    '--divider-color': v.dividerColor,

    '--text-color': v.textColor2,
    '--text-color-1': v.textColor1,
    '--text-color-2': v.textColor2,
    '--text-color-3': v.textColor3,
    '--text-color-secondary': v.textColor3,
    '--text-color-disabled': v.textColorDisabled,

    '--input-color': v.inputColor,
    '--input-color-disabled': v.inputColorDisabled,

    '--primary-color': v.primaryColor,
    '--info-color': v.infoColor,
    '--success-color': v.successColor,
    '--warning-color': v.warningColor,
    '--error-color': v.errorColor,
  }
})

watchEffect(() => {
  const root = document.documentElement
  for (const [name, value] of Object.entries(tokens.value)) {
    root.style.setProperty(name, value)
  }
})

onUnmounted(() => {
  const root = document.documentElement
  for (const name of Object.keys(tokens.value)) {
    root.style.removeProperty(name)
  }
})
</script>

<template>
  <slot></slot>
</template>
