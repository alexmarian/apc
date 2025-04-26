<!-- src/App.vue -->
<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { NSpace, NLayout, NLayoutHeader, NLayoutContent, NMenu, NButton, NDropdown } from 'naive-ui'
import { h, ref, computed } from 'vue'
import ThemeProvider from '@/providers/ThemeProvider.vue'
import LanguageSelector from '@/components/LanguageSelector.vue'
import UserProfileButton from '@/components/UserProfileButton.vue'
import { useI18n } from 'vue-i18n'
import { usePreferences } from '@/stores/preferences.ts'

const preferences = usePreferences()
const { t } = useI18n()

const isDark = ref(preferences.theme ? preferences.theme === 'darkTheme' : true)
const currentTheme = computed(() => {
  return isDark.value ? 'darkTheme' : 'lightTheme'
})
const menuOptions = [
  {
    label: () => h(RouterLink, { to: '/' }, { default: () => t('common.home', 'Home') }),
    key: 'home'
  },
  {
    label: () => h(RouterLink, { to: '/accounts' }, { default: () => t('accounts.title', 'Accounts') }),
    key: 'accounts'
  },
  {
    label: () => h(RouterLink, { to: '/expenses' }, { default: () => t('expenses.title', 'Expenses') }),
    key: 'expenses'
  },
  {
    label: () => h(RouterLink, { to: '/reports' }, { default: () => t('reports.title', 'Reports') }),
    key: 'reports'
  }
]

// Theme options for dropdown with explicit typing
const themeMenuOptions = [
  {
    label: 'Light Theme',
    key: 'light'
  },
  {
    label: 'Dark Theme',
    key: 'dark'
  },
  {
    label: 'Auto (System)',
    key: 'auto'
  }
]

</script>

<template>
  <ThemeProvider :theme="currentTheme">
    <NLayout class="main-layout">
      <NLayoutHeader bordered class="header">
        <div class="logo">
          <img alt="App logo" class="logo-img" src="@/assets/logo.svg" width="32" height="32" />
          <h1 class="app-title">APC</h1>
        </div>
        <NMenu mode="horizontal" :options="menuOptions" />
        <div class="header-right">
          <LanguageSelector />
          <NSwitch v-model:value="isDark" />
          <UserProfileButton />
        </div>
      </NLayoutHeader>
      <NLayoutContent class="content-margin">
        <RouterView />
      </NLayoutContent>
    </NLayout>
  </ThemeProvider>
</template>

<style scoped>
.main-layout {
  min-height: 100vh;
  width: 100%;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
}

.logo {
  display: flex;
  align-items: center;
}

.logo-img {
  margin-right: 8px;
}

.app-title {
  font-size: 1.25rem;
  font-weight: bold;
  margin: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}
.content-margin {
  margin: 10px; /* Adjust the margin value as needed */
}
</style>
