<!-- src/providers/ThemeProvider.vue -->
<script setup lang="ts">
import { computed, watch } from 'vue'
import {
  darkTheme,
  lightTheme,
  NGlobalStyle
} from 'naive-ui'
import { usePreferences } from '@/stores/preferences.ts'
import ThemeTokens from '@/providers/ThemeTokens.vue'

const preferences = usePreferences()
const props = defineProps<{
  theme?: string
}>()

const themes: Record<string, typeof darkTheme> = {
  darkTheme: darkTheme,
  lightTheme: lightTheme,
}

const currentTheme = computed(() => {
  if (props.theme && themes[props.theme]) {
    return themes[props.theme]
  }else if (preferences.theme && themes[preferences.theme]) {
    return themes[preferences.theme]
  }
  return darkTheme
})

watch(currentTheme, () => {
  console.log(props.theme)
  if (props.theme && themes[props.theme]) {
    preferences.setTheme(props.theme)
  }
})
</script>

<template>
  <NConfigProvider
    :theme="currentTheme"
  >
    <NGlobalStyle />
    <ThemeTokens>
      <NLoadingBarProvider>
        <NDialogProvider>
          <NNotificationProvider>
            <NMessageProvider>
              <slot></slot>
            </NMessageProvider>
          </NNotificationProvider>
        </NDialogProvider>
      </NLoadingBarProvider>
    </ThemeTokens>
  </NConfigProvider>
</template>
