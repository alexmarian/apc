<template>
  <NConfigProvider :theme-overrides="themeOverrides">
    <NMessageProvider>
      <div style="max-width: 720px; margin: 0 auto; padding: 24px 16px">
        <RouterView />
      </div>
    </NMessageProvider>
  </NConfigProvider>
</template>

<script setup lang="ts">
import { NConfigProvider, NMessageProvider } from 'naive-ui'

const themeOverrides = {}
</script>
