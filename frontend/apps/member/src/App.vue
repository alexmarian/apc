<template>
  <NConfigProvider :theme-overrides="themeOverrides">
    <NMessageProvider>
      <div style="max-width: 720px; margin: 0 auto; padding: 24px 16px">
        <div style="display: flex; justify-content: flex-end; margin-bottom: 16px; gap: 8px">
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
    </NMessageProvider>
  </NConfigProvider>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { NConfigProvider, NMessageProvider, NButton } from 'naive-ui'

const { locale } = useI18n()

const themeOverrides = {}

const langs = [
  { value: 'ro', label: 'RO' },
  { value: 'ru', label: 'RU' },
]

function setLocale(lang: string) {
  locale.value = lang
  localStorage.setItem('userLocale', lang)
}
</script>
