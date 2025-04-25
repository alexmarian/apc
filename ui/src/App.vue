<!-- src/App.vue -->
<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { NSpace, NLayout, NLayoutHeader, NLayoutContent, NMenu, NButton, NDropdown } from 'naive-ui'
import { h, inject } from 'vue'
import ThemeProvider from '@/providers/ThemeProvider.vue'
import LanguageSelector from '@/components/LanguageSelector.vue'
import { ThemeKey, defaultThemeProvider } from '@/utils/theme'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
// Theme injection using the Symbol key for more reliable injection
const theme = inject(ThemeKey, defaultThemeProvider)

// Menu options
const menuOptions = [
  {
    label: () => h(RouterLink, { to: '/' }, { default: () => t('common.home', 'Home') }),
    key: 'home'
  },
  {
    label: () => h(RouterLink, { to: '/accounts' }, { default: () => t('accounts.title', 'Accounts') }),
    key: 'accounts'
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

// Handle theme change
const handleThemeChange = (key: string) => {
  console.log('Theme change triggered', key)
  // Check if key is valid before switching
  if (['light', 'dark', 'auto'].includes(key)) {
    theme.switchTheme(key as any)
  } else {
    console.error('Invalid theme key:', key)
  }
}
</script>

<template>
  <ThemeProvider>
    <n-layout class="main-layout">
      <n-layout-header bordered class="header">
        <div class="header-content">
          <div class="logo">
            <img alt="App logo" class="logo-img" src="@/assets/logo.svg" width="32" height="32" />
            <h1 class="app-title">APC Management</h1>
          </div>
          <n-space>
            <n-menu mode="horizontal" :options="menuOptions" />
            <!-- Temporarily comment out until fully set up -->
            <LanguageSelector />
            <n-dropdown
              trigger="click"
              :options="themeMenuOptions"
              @select="handleThemeChange"
            >
              <n-button>
                Theme: {{ theme.isDark ? 'Dark' : 'Light' }}
              </n-button>
            </n-dropdown>
          </n-space>
        </div>
      </n-layout-header>

      <n-layout-content>
        <div class="content-container">
          <RouterView />
        </div>
      </n-layout-content>
    </n-layout>
  </ThemeProvider>
</template>

<style scoped>
.main-layout {
  min-height: 100vh;
}

.header {
  position: sticky;
  top: 0;
  z-index: 999;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 2rem;
  height: 64px;
}

.logo {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.app-title {
  font-size: 1.5rem;
  font-weight: 500;
  margin: 0;
}

.content-container {
  width: 100%;
  margin: 0 auto;
  padding: 2rem;
}
</style>
