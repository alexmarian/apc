<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { NForm, NFormItem, NInput, NButton, NAlert, NSpace } from 'naive-ui'
import { useAuthStore } from '../stores/auth'
import { useI18n } from 'vue-i18n'
import type { FormRules } from 'naive-ui'
import AuthLayout from '@/layouts/AuthLayout.vue'
import { usePreferences } from '@/stores/preferences.ts'

// Get auth store
const auth = useAuthStore()
const router = useRouter()
const { t } = useI18n()
const preferences = usePreferences()

// Check if we're using dark theme
const isDarkTheme = computed(() => {
  return preferences.theme === 'darkTheme' || preferences.theme === null
})

// Form data
const formData = reactive({
  login: '',
  password: '',
  totp: ''
})

// Form validation rules
const rules: FormRules = {
  login: [
    {
      required: true,
      message: t('validation.required', {
        field: t('auth.username', 'Username')
      }),
      trigger: 'blur'
    }
  ],
  password: [
    {
      required: true,
      message: t('validation.required', {
        field: t('auth.password', 'Password')
      }),
      trigger: 'blur'
    }
  ],
  totp: [
    {
      required: true,
      message: t('validation.required', {
        field: t('auth.totp', 'TOTP code')
      }),
      trigger: 'blur'
    }, {
      type: 'string',
      len: 6,
      message: t('auth.totpLength', 'TOTP code must be 6 digits'),
      trigger: 'blur'
    }
  ]
}

// State
const formRef = ref(null)
const rememberMe = ref(false)

// Handle login
const handleLogin = async (e: MouseEvent) => {
  e.preventDefault()

  if (!formRef.value) return

  try {
    // @ts-ignore - Naive UI types issue with form ref
    await formRef.value.validate()

    const success = await auth.login(formData)

    if (success) {
      router.push('/')
    }
  } catch (err) {
    console.error('Validation failed:', err)
  }
}
</script>

<template>
  <AuthLayout>
    <NAlert v-if="auth.error" type="error" style="margin-bottom: 24px;">
      {{ auth.error }}
    </NAlert>

    <NForm
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-placement="top"
    >
      <h2 class="form-title" :class="{ 'light': !isDarkTheme }">{{ t('auth.login', 'Login') }}</h2>

      <NFormItem :label="t('auth.username', 'Username')" path="login">
        <NInput
          v-model:value="formData.login"
          :placeholder="t('auth.enterUsername', 'Enter your username')"
          autofocus
        />
      </NFormItem>

      <NFormItem :label="t('auth.password', 'Password')" path="password">
        <NInput
          v-model:value="formData.password"
          type="password"
          :placeholder="t('auth.enterPassword', 'Enter your password')"
          show-password-on="click"
        />
      </NFormItem>

      <NFormItem :label="t('auth.totp', 'TOTP Code')" path="totp">
        <NInput
          v-model:value="formData.totp"
          :placeholder="t('auth.enterTotp', 'Enter your 6-digit code')"
        />
      </NFormItem>

      <div style="margin-top: 24px;">
        <NSpace vertical align="center">
          <NButton
            type="primary"
            block
            @click="handleLogin"
            :loading="auth.loading"
          >
            {{ t('auth.login', 'Login') }}
          </NButton>
        </NSpace>
      </div>
    </NForm>
  </AuthLayout>
</template>

<style scoped>
.form-title {
  color: #e0e0e0;
  text-align: center;
  margin-bottom: 24px;
  font-size: 1.5rem;
}

.form-title.light {
  color: #2c3e50;
}

:deep(.n-button) {
  font-weight: bold;
  height: 40px;
}
</style>
