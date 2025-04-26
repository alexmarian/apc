<!-- src/providers/ThemeProvider.vue -->
<script setup lang="ts">
import { ref, provide, computed, watchEffect, watch } from 'vue'
import {
  darkTheme,
  lightTheme,
} from 'naive-ui'


const props = defineProps({
  theme: {
    type: String,
    default: 'darkTheme'
  }
})
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
</script>

<template>
  <n-config-provider
    :theme="currentTheme"
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
</style>
