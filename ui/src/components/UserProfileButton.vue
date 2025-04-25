<script setup lang="ts">
import { computed } from 'vue'
import { NDropdown, NButton, NAvatar } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const authStore = useAuthStore()
const router = useRouter()

// Get username from auth store
const username = computed(() => {
  return authStore.user || t('auth.anonymous', 'Anonymous')
})

// User menu options
const userMenuOptions = [
  {
    label: t('auth.logout', 'Logout'),
    key: 'logout',
    props: {
      onClick: () => handleLogout()
    }
  }
]

// Handle logout
const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <div class="user-profile-button">
    <NDropdown
      v-if="useAuthStore.isAuthenticated"
      trigger="click"
      :options="userMenuOptions"
      placement="bottom-end"
    >
      <NButton text>
        <div class="user-button-content">
          <NAvatar round size="small" color="#3366ff">
            {{ username.charAt(0).toUpperCase() }}
          </NAvatar>
          <span class="username">{{ username }}</span>
        </div>
      </NButton>
    </NDropdown>

    <NButton v-else @click="router.push('/login')" type="primary">
      {{ t('auth.login', 'Login') }}
    </NButton>
  </div>
</template>

<style scoped>
.user-profile-button {
  display: flex;
  align-items: center;
}

.user-button-content {
  display: flex;
  align-items: center;
  gap: 8px;
}

.username {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
