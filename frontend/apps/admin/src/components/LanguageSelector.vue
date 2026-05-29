<!-- src/components/LanguageSelector.vue -->
<script setup lang="ts">
import { ref, computed } from 'vue'
import { NDropdown, NButton } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { usePreferences } from '@/stores/preferences.ts'
// Get the i18n composable
const i18n = useI18n()
const preferences = usePreferences()

// Available languages
const languages = [
  { label: 'English', key: 'en' },
  { label: 'Română', key: 'ro' }
]

const selectedLocale = ref(preferences.locale || 'en')
i18n.locale.value = selectedLocale.value
const currentLanguage = computed(() => {
  return languages.find(language => language.key === selectedLocale.value) || languages[0]
})
// Handle language change
const handleLanguageChange = (key: string) => {
  // Change the locale
  i18n.locale.value = key
  // Save the selected language in preferences
  preferences.setLocale(key)
  selectedLocale.value= key
}
</script>

<template>
  <div class="language-selector">
    <n-dropdown
      v-model:value="selectedLocale"
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
