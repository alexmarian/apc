<!-- src/App.vue -->
<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { NIcon, NLayout, NLayoutHeader, NLayoutContent, NMenu, NButton, NDropdown } from 'naive-ui'
import { h, ref, computed } from 'vue'
import ThemeProvider from '@/providers/ThemeProvider.vue'
import LanguageSelector from '@/components/LanguageSelector.vue'
import UserProfileButton from '@/components/UserProfileButton.vue'
import { useI18n } from 'vue-i18n'
import { usePreferences } from '@/stores/preferences.ts'
import {
  AttachMoneyRound,
  AccountBalanceRound,
  HomeRound,
  BedroomParentRound
} from '@vicons/material'

const preferences = usePreferences()
const { t } = useI18n()

const isDark = ref(preferences.theme ? preferences.theme === 'darkTheme' : true)
const currentTheme = computed(() => {
  return isDark.value ? 'darkTheme' : 'lightTheme'
})

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

const menuOptions = [
  {
    label: () => h(RouterLink, { to: '/' }, { default: () => t('common.home', 'Home') }),
    key: 'home',
    icon: renderIcon(HomeRound)
  },
  {
    label: () => h(RouterLink, { to: '/accounts' }, { default: () => t('accounts.title', 'Accounts') }),
    key: 'accounts',
    icon: renderIcon(AccountBalanceRound)
  },
  {
    label: 'Expenses',
    key: 'expenses-group',
    icon: renderIcon(AttachMoneyRound),
    children: [
      {
        label: () => h(RouterLink, { to: '/expenses' }, { default: () => t('expenses.title', 'Management') }),
        key: 'expenses-management'
      },
      {
        label: () => h(RouterLink, { to: '/reports' }, { default: () => t('reports.title', 'Reports') }),
        key: 'expenses-reports'
      },
      {
        label: () => h(RouterLink, { to: '/expenses/distribution' }, { default: () => t('distribution.title', 'Distribution') }),
        key: 'expense-distribution'

      }
    ]
  },
  {
    label: 'Units',
    key: 'units-group',
    icon: renderIcon(BedroomParentRound),
    children: [
      {
        label: () => h(RouterLink, { to: '/units' }, { default: () => t('units.title', 'Management') }),
        key: 'units-management'
      },
      {
        label: () => h(RouterLink, { to: '/owners/report' }, { default: () => t('owners.report', 'Owners Report') }),
        key: 'owners-report'
      }
    ]
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
