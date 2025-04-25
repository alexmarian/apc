<!-- src/components/LanguageSelector.vue -->
<script setup lang="ts">
import { ref } from 'vue'
import { NDropdown, NButton } from 'naive-ui'
import { useI18n } from 'vue-i18n'

// Get the i18n composable
const i18n = useI18n()

// Available languages
const languages = [
  { label: 'English', key: 'en' },
  { label: 'Română', key: 'ro' }
]

// Current language
const currentLanguage = ref(languages.find(lang => lang.key === i18n.locale.value) || languages[0])

// Handle language change
const handleLanguageChange = (key: string) => {
  // Change the locale
  i18n.locale.value = key
  // Update the current language
  currentLanguage.value = languages.find(lang => lang.key === key) || languages[0]
}
</script>

<template>
  <div class="language-selector">
    <n-dropdown
      trigger="click"
      :options="languages"
      @select="handleLanguageChange"
    >
      <n-button>
        {{ currentLanguage.label }}
      </n-button>
    </n-dropdown>
  </div>
</template>

<style scoped>
.language-selector {
  display: inline-block;
}
</style>
