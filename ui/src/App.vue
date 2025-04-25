<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { NSpace, NLayout, NLayoutHeader, NLayoutContent, NMenu, NButton, NDropdown } from 'naive-ui'
import { h, ref, inject } from 'vue'
import ThemeProvider from './providers/ThemeProvider.vue'

// Theme injection (TypeScript needs this defined)
const theme = inject('theme', {
  switchTheme: (theme: string) => {
  },
  isDark: ref(false),
  themeOptions: ['light', 'dark', 'auto']
})

// Menu options
const menuOptions = [
  {
    label: () => h(RouterLink, { to: '/' }, { default: () => 'Home' }),
    key: 'home'
  },
  {
    label: () => h(RouterLink, { to: '/about' }, { default: () => 'About' }),
    key: 'about'
  },
  {
    label: () => h(RouterLink, { to: '/accounts' }, { default: () => 'Accounts' }),
    key: 'accounts'
  }
]

// Theme options for dropdown
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
  theme.switchTheme(key)
}
</script>

<template>
  <ThemeProvider>
    <n-layout>
      <n-layout-header bordered>
        <div class="header-content">
          <div class="logo">
            <img alt="App logo" class="logo-img" src="@/assets/logo.svg" width="32" height="32" />
            <h1 class="app-title">APC Management</h1>
          </div>
          <n-space>
            <n-menu mode="horizontal" :options="menuOptions" />
            <n-dropdown
              trigger="click"
              :options="themeMenuOptions"
              @select="handleThemeChange"
            >
              <n-button>
                Theme
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
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}
</style>
