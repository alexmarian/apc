<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  NForm,
  NFormItem,
  NInput,
  NButton,
  NAlert,
  NSpace,
  NCheckbox,
  NResult
} from 'naive-ui'
import AuthLayout from '@/layouts/AuthLayout.vue'
import { useI18n } from 'vue-i18n'
import { authApi } from '@/services/api.ts'
import { usePreferences } from '@/stores/preferences.ts'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const preferences = usePreferences()

// Check if we're using dark theme
const isDarkTheme = computed(() => {
  return preferences.theme === 'darkTheme' || preferences.theme === null
})

// Form data
const formData = reactive({
  token: '',
  newPassword: '',
  confirmPassword: '',
  resetTOTPSecret: false
})

// State
const loading = ref(false)
const error = ref('')
const success = ref(false)
const qrCode = ref('')
const formRef = ref(null)

// Get token from route
onMounted(() => {
  const token = route.query.token as string
  if (token) {
    formData.token = token
  }
})

// Handle password reset
const handleResetPassword = async (e: MouseEvent) => {
  e.preventDefault()

  if (!formRef.value) return

  try {
    // @ts-ignore - Naive UI types issue with form ref
    await formRef.value.validate()

    if (formData.newPassword !== formData.confirmPassword) {
      error.value = t('validation.passwordMatch', 'Passwords do not match')
      return
    }

    loading.value = true
    error.value = ''

    const response = await authApi.resetPassword({
      token: formData.token,
      new_password: formData.newPassword,
      reset_totp_secret: formData.resetTOTPSecret
    })

    success.value = true

    // If TOTP was reset, store the QR code
    if (formData.resetTOTPSecret && response.data.qrCode) {
      qrCode.value = response.data.qrCode
    }
  } catch (err: any) {
    error.value = err.response?.data?.msg || 'An error occurred. Please try again.'
    console.error('Password reset failed:', err)
  } finally {
    loading.value = false
  }
}

// Navigate to login page
const goToLogin = () => {
  router.push('/login')
}
</script>

<template>
  <AuthLayout>
    <!-- Reset Success -->
    <div v-if="success">
      <NResult
        status="success"
        :title="t('auth.passwordResetComplete', 'Password Reset Complete')"
        :description="t('auth.passwordResetSuccess', 'Your password has been reset successfully.')"
        class="reset-result"
      >
        <template #icon v-if="qrCode">
          <div class="qr-code-container">
            <h3 class="qr-code-title" :class="{ 'light': !isDarkTheme }">
              {{ t('auth.scanQRDescription', 'Please scan this QR code with your authenticator app')
              }}</h3>
            <img :src="'data:image/png;base64,' + qrCode" alt="QR Code for TOTP" class="qr-code" />
          </div>
        </template>
        <template #footer>
          <NButton type="primary" @click="goToLogin">
            {{ t('auth.proceedToLogin', 'Proceed to Login') }}
          </NButton>
        </template>
      </NResult>
    </div>

    <!-- Reset Form -->
    <div v-else>
      <NAlert v-if="error" type="error" style="margin-bottom: 16px;">
        {{ error }}
      </NAlert>

      <NForm
        ref="formRef"
        :model="formData"
        label-placement="top"
      >
        <h2 class="form-title" :class="{ 'light': !isDarkTheme }">{{ t('auth.resetPassword', 'Reset Password') }}</h2>

        <NFormItem :label="t('auth.newPassword', 'New Password')" path="newPassword">
          <NInput
            v-model:value="formData.newPassword"
            type="password"
            :placeholder="t('auth.enterNewPassword', 'Enter new password')"
            show-password-on="click"
          />
        </NFormItem>

        <NFormItem :label="t('auth.confirmPassword', 'Confirm Password')" path="confirmPassword">
          <NInput
            v-model:value="formData.confirmPassword"
            type="password"
            :placeholder="t('auth.confirmNewPassword', 'Confirm new password')"
            show-password-on="click"
          />
        </NFormItem>

        <NFormItem>
          <NCheckbox v-model:checked="formData.resetTOTPSecret">
            {{ t('auth.resetTOTP', 'Reset two-factor authentication (2FA)') }}
          </NCheckbox>
          <div class="help-text" :class="{ 'light': !isDarkTheme }">
            {{ t('auth.resetTOTPHelp', 'Check this if you have lost access to your authenticator app')
            }}
          </div>
        </NFormItem>

        <div style="margin-top: 24px;">
          <NSpace vertical align="center">
            <NButton
              type="primary"
              block
              @click="handleResetPassword"
              :loading="loading"
            >
              {{ t('auth.resetPassword', 'Reset Password') }}
            </NButton>

            <NButton text @click="goToLogin" class="login-link">
              {{ t('auth.backToLogin', 'Back to Login') }}
            </NButton>
          </NSpace>
        </div>
      </NForm>
    </div>
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

.help-text {
  font-size: 0.85rem;
  color: #aaa;
  margin-top: 4px;
}

.help-text.light {
  color: #666;
}

.qr-code-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin: 20px 0;
}

.qr-code-title {
  color: #e0e0e0;
  margin-bottom: 12px;
}

.qr-code-title.light {
  color: #2c3e50;
}

.qr-code {
  width: 200px;
  height: 200px;
  border: 1px solid #eee;
  padding: 10px;
  background: white;
  margin-top: 12px;
}

.login-link {
  color: #41b883 !important;
}

:deep(.reset-result) {
  background: transparent !important;
}

:deep(.reset-result .n-result-header .n-result-icon) {
  color: #41b883 !important;
}

:deep(.reset-result .n-result-header .n-result-title) {
  color: v-bind('isDarkTheme ? "#e0e0e0" : "#2c3e50"');
}

:deep(.reset-result .n-result-content) {
  color: v-bind('isDarkTheme ? "#aaa" : "#555"');
}

:deep(.n-checkbox__label) {
  color: v-bind('isDarkTheme ? "#e0e0e0" : "#2c3e50"') !important;
}
</style>
