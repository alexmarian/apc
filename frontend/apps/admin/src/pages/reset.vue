<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  NForm,
  NFormItem,
  NInput,
  NButton,
  NAlert,
  NSpace,
  NCheckbox,
  NResult,
  NText
} from 'naive-ui'
import AuthLayout from '@/layouts/AuthLayout.vue'
import { useI18n } from 'vue-i18n'
import { authApi } from '@/services/api.ts'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()

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
  <AuthLayout :title="success ? undefined : t('auth.resetPassword', 'Reset Password')">
    <!-- Reset Success -->
    <div v-if="success">
      <NResult
        status="success"
        :title="t('auth.passwordResetComplete', 'Password Reset Complete')"
        :description="t('auth.passwordResetSuccess', 'Your password has been reset successfully.')"
      >
        <template #icon v-if="qrCode">
          <div class="qr-code-container">
            <NText>
              {{ t('auth.scanQRDescription', 'Please scan this QR code with your authenticator app') }}
            </NText>
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
          <div class="help-text">
            {{ t('auth.resetTOTPHelp', 'Check this if you have lost access to your authenticator app')
            }}
          </div>
        </NFormItem>

        <div style="margin-top: 24px;">
          <NSpace vertical align="center">
            <NButton
              type="primary"
              size="large"
              block
              @click="handleResetPassword"
              :loading="loading"
            >
              {{ t('auth.resetPassword', 'Reset Password') }}
            </NButton>

            <NButton text type="primary" @click="goToLogin">
              {{ t('auth.backToLogin', 'Back to Login') }}
            </NButton>
          </NSpace>
        </div>
      </NForm>
    </div>
  </AuthLayout>
</template>

<style scoped>
.help-text {
  font-size: var(--text-md);
  color: var(--text-color-3);
  margin-top: 4px;
}

.qr-code-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin: 20px 0;
}

/* Stays white in both themes: scanners need the light quiet zone, and the
 * code itself is rendered as black-on-transparent. */
.qr-code {
  width: 200px;
  height: 200px;
  border: 1px solid var(--border-color);
  padding: 10px;
  background: white;
  margin-top: 12px;
}
</style>
