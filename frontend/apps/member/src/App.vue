<template>
  <ThemeProvider>
    <div class="member-shell">
      <div class="member-shell__toolbar">
        <NButton
          v-for="lang in langs"
          :key="lang.value"
          :type="locale === lang.value ? 'primary' : 'default'"
          size="small"
          @click="setLocale(lang.value)"
        >
          {{ lang.label }}
        </NButton>
      </div>
      <RouterView />
    </div>
  </ThemeProvider>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { NButton } from 'naive-ui'
import ThemeProvider from '@/providers/ThemeProvider.vue'

const { locale } = useI18n()

const langs = [
  { value: 'ro', label: 'RO' },
  { value: 'ru', label: 'RU' },
]

function setLocale(lang: string) {
  locale.value = lang
  localStorage.setItem('userLocale', lang)
  document.documentElement.lang = lang
}
</script>

<style scoped>
.member-shell {
  max-width: 720px;
  margin: 0 auto;
  padding: 24px 16px;
}

.member-shell__toolbar {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-bottom: 16px;
}
</style>
