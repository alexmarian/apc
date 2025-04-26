<!-- src/providers/ThemeProvider.vue -->
<script setup lang="ts">
import { ref, provide, computed, watchEffect, watch } from 'vue'
import {
  darkTheme,
  lightTheme
} from 'naive-ui'
import { usePreferences } from '@/stores/preferences.ts'

const preferences = usePreferences()
const props = defineProps({
  theme: {
    type: String
  }
})
if (!props.theme) {
  props.theme = preferences.getTheme()
}
const themes = {
  'darkTheme': darkTheme,
  'lightTheme': lightTheme
}
const currentTheme = computed(() => {
  if (themes[props.theme]) {
    return themes[props.theme]
  }
  return darkTheme
})
watch(currentTheme, () => {
  console.log(props.theme)
  if (themes[props.theme]) {
    preferences.setTheme(props.theme)
  }
})
</script>

<template>
  <NConfigProvider
    :theme="currentTheme"
  >
    <NLoadingBarProvider>
      <NDialogProvider>
        <NNotificationProvider>
          <NMessageProvider>
            <slot></slot>
          </NMessageProvider>
        </NNotificationProvider>
      </NDialogProvider>
    </NLoadingBarProvider>
  </NConfigProvider>
</template>

<style>
</style>
