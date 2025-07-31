<script setup lang="ts">
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { NIcon, NLayout, NLayoutHeader, NLayoutContent, NMenu, NButton, NDropdown, NSwitch } from 'naive-ui'
import { h, ref, computed, watch, nextTick } from 'vue'
import type { Component } from 'vue';
import ThemeProvider from '@/providers/ThemeProvider.vue'
import LanguageSelector from '@/components/LanguageSelector.vue'
import UserProfileButton from '@/components/UserProfileButton.vue'
import { useI18n } from 'vue-i18n'
import { usePreferences } from '@/stores/preferences.ts'
import { useAuthStore } from '@/stores/auth'
import {
  AttachMoneyRound,
  AccountBalanceRound,
  HomeRound,
  BedroomParentRound,
  PeopleRound
} from '@vicons/material'

const preferences = usePreferences()
const { t } = useI18n()
const route = useRoute()
const authStore = useAuthStore()

// Use visibility class to prevent flickering
const appReady = ref(false)
nextTick(() => {
  appReady.value = true
})

// Check if current route is an auth page (login, register, etc.)
const isAuthPage = computed(() => route.meta.isAuthPage)

const isDark = ref(preferences.theme ? preferences.theme === 'darkTheme' : true)
const currentTheme = computed(() => {
  return isDark.value ? 'darkTheme' : 'lightTheme'
})

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

// Using computed to ensure the menu options update when the language changes
const menuOptions = computed(() => [
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
    label: () => h(RouterLink, { to: '/gatherings' }, { default: () => t('gatherings.title', 'Gatherings') }),
    key: 'gatherings',
    icon: renderIcon(PeopleRound)
  },
  {
    label: () => t('expenses.title', 'Expenses'),
    key: 'expenses-group',
    icon: renderIcon(AttachMoneyRound),
    children: [
      {
        label: () => h(RouterLink, { to: '/expenses' }, { default: () => t('expenses.management', 'Management') }),
        key: 'expenses-management'
      },
      {
        label: () => h(RouterLink, { to: '/reports' }, { default: () => t('reports.title', 'Reports') }),
        key: 'expenses-reports'
      },
      {
        label: () => h(RouterLink, { to: '/expenses/distribution' }, { default: () => t('expenses.distribution', 'Distribution') }),
        key: 'expense-distribution'
      }
    ]
  },
  {
    label: () => t('units.title', 'Units'),
    key: 'units-group',
    icon: renderIcon(BedroomParentRound),
    children: [
      {
        label: () => h(RouterLink, { to: '/units' }, { default: () => t('units.management', 'Management') }),
        key: 'units-management'
      },
      {
        label: () => h(RouterLink, { to: '/owners/report' }, { default: () => t('owners.report', 'Owners Report') }),
        key: 'owners-report'
      },
      {
        label: () => h(RouterLink, { to: '/owners/voting' }, { default: () => t('owners.votingReport', 'Voting Report') }),
        key: 'owners-voting-report'
      }
    ]
  }
])

// Theme options as computed property to make them translatable
const themeMenuOptions = computed(() => [
  {
    label: t('theme.light', 'Light Theme'),
    key: 'light'
  },
  {
    label: t('theme.dark', 'Dark Theme'),
    key: 'dark'
  },
  {
    label: t('theme.auto', 'Auto (System)'),
    key: 'auto'
  }
])

// Watch for theme changes
watch(isDark, () => {
  const newTheme = isDark.value ? 'darkTheme' : 'lightTheme'
  preferences.setTheme(newTheme)
})
</script>

<template>
  <ThemeProvider :theme="currentTheme">
    <!-- Add a class when app is ready to prevent flickering -->
    <div :class="{ 'app-visible': appReady, 'app-hidden': !appReady }">
      <!-- Render the standard layout only for non-auth pages -->
      <template v-if="!isAuthPage">
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
      </template>

      <!-- For auth pages, render just the RouterView without the layout -->
      <template v-else>
        <RouterView />
      </template>
    </div>
  </ThemeProvider>
</template>

<style scoped>
/* Hidden class to prevent flickering on initial load */
.app-hidden {
  visibility: hidden;
  opacity: 0;
}

.app-visible {
  visibility: visible;
  opacity: 1;
  transition: opacity 0.2s ease-in-out;
}

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
