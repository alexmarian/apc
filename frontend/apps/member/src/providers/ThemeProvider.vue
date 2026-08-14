<!-- src/providers/ThemeProvider.vue -->
<script setup lang="ts">
import { onUnmounted, ref } from 'vue'
import { darkTheme, lightTheme, NConfigProvider, NGlobalStyle, NMessageProvider } from 'naive-ui'

/*
 * The member portal is opened from a one-off invitation link by owners who have
 * no account and no stored preferences, so unlike apps/admin there is nothing to
 * persist a choice to and no chrome to hang a toggle off. Following the OS
 * setting is the behaviour a visitor expects from a page they opened from a link.
 */
const query = window.matchMedia('(prefers-color-scheme: dark)')
const prefersDark = ref(query.matches)

const onChange = (e: MediaQueryListEvent) => {
  prefersDark.value = e.matches
}
query.addEventListener('change', onChange)
onUnmounted(() => query.removeEventListener('change', onChange))
</script>

<template>
  <NConfigProvider :theme="prefersDark ? darkTheme : lightTheme">
    <!-- Paints the body from the active theme; without it the page keeps the
         browser default white behind themed components. -->
    <NGlobalStyle />
    <NMessageProvider>
      <slot></slot>
    </NMessageProvider>
  </NConfigProvider>
</template>
