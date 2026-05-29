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
  NResult
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import tokenService from '@/services/tokenService'
import AuthLayout from '@/layouts/AuthLayout.vue'
import type { FormRules } from 'naive-ui'
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
  login: '',
  password: '',
  confirmPassword: '',
  token: ''
})

// State
const formRef = ref(null)
const loading = ref(false)
const error = ref('')
const qrCode = ref('')
const registrationComplete = ref(false)

// Get token from URL if provided
onMounted(() => {
  // Check if token is in the URL parameters
  const tokenFromUrl = route.query.token as string
  if (tokenFromUrl) {
    formData.token = tokenFromUrl
    // Focus on the username field since we already have the token
    setTimeout(() => {
      document.getElementById('login-input')?.focus()
    }, 100)
  }
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
    },
    {
      min: 4,
      message: t('validation.minLength', {
        field: t('auth.username', 'Username'),
        min: 4
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
    },
    {
      min: 8,
      message: t('validation.minLength', {
        field: t('auth.password', 'Password'),
        min: 8
      }),
      trigger: 'blur'
    }
  ],
  confirmPassword: [
    {
      required: true,
      message: t('validation.required', {
        field: t('auth.confirmPassword', 'Confirm Password')
      }),
      trigger: 'blur'
    },
    {
      validator: validatePasswordMatch,
      message: t('validation.passwordMatch', 'Passwords do not match')
    }
  ],
  token: [
    {
      required: true,
      message: t('validation.required', {
        field: t('auth.token', 'Registration Token')
      }),
      trigger: 'blur'
    }, {
      min: 8,
      message: t('validation.minLength', {
        field: t('auth.token', 'Registration Token'),
        min: 8
      }),
      trigger: 'blur'
    }
  ]
}

// Validate password match
function validatePasswordMatch(_rule: any, value: string) {
  return value === formData.password
}

// Handle registration
async function handleRegister(e: MouseEvent) {
  e.preventDefault()

  if (!formRef.value) return

  try {
    // @ts-ignore - Naive UI types issue with form ref
    await formRef.value.validate()

    loading.value = true
    error.value = ''

    const response = await tokenService.registerWithToken({
      login: formData.login,
      password: formData.password,
      token: formData.token
    })

    // Store QR code for TOTP setup
    if (response && response.qrCode) {
      qrCode.value = response.qrCode
      registrationComplete.value = true
    }
  } catch (err: any) {
    error.value = err.response?.data?.msg || t('errors.registrationFailed', 'Registration failed. Please check your details and try again.')
    console.error('Registration failed:', err)
  } finally {
    loading.value = false
  }
}

// Navigate to login page
function goToLogin() {
  router.push('/login')
}
</script>

<template>
  <AuthLayout>
    <!-- Registration Form -->
    <div v-if="!registrationComplete">
      <NAlert v-if="error" type="error" style="margin-bottom: 16px;">
        {{ error }}
      </NAlert>

      <NForm
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-placement="top"
      >
        <h2 class="form-title" :class="{ 'light': !isDarkTheme }">{{ t('auth.register', 'Register') }}</h2>

        <NFormItem :label="t('auth.username', 'Username')" path="login">
          <NInput
            id="login-input"
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

        <NFormItem :label="t('auth.confirmPassword', 'Confirm Password')" path="confirmPassword">
          <NInput
            v-model:value="formData.confirmPassword"
            type="password"
            :placeholder="t('auth.confirmPassword', 'Confirm your password')"
            show-password-on="click"
          />
        </NFormItem>

        <NFormItem :label="t('auth.token', 'Registration Token')" path="token">
          <NInput
            v-model:value="formData.token"
            :placeholder="t('auth.enterToken', 'Enter your registration token')"
          />
          <template v-if="route.query.token" #feedback>
            <div class="token-info">
              {{ t('auth.tokenFromLink', 'Token was provided in the registration link') }}
            </div>
          </template>
        </NFormItem>

        <div style="margin-top: 24px;">
          <NSpace vertical align="center">
            <NButton
              type="primary"
              block
              @click="handleRegister"
              :loading="loading"
            >
              {{ t('auth.register', 'Register') }}
            </NButton>

            <NButton text @click="goToLogin" class="login-link">
              {{ t('auth.alreadyHaveAccount', 'Already have an account? Login') }}
            </NButton>
          </NSpace>
        </div>
      </NForm>
    </div>

    <!-- Registration Success -->
    <div v-else>
      <NResult
        status="success"
        :title="t('auth.registrationComplete', 'Registration Complete')"
        :description="t('auth.scanQRDescription', 'Please scan this QR code with your authenticator app to set up two-factor authentication. You will need this for logging in.')"
        class="qr-result"
      >
        <template #icon>
          <div class="qr-code-container">
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

.token-info {
  font-size: 0.85rem;
  color: #41b883;
  margin-top: 4px;
}

.qr-code-container {
  display: flex;
  justify-content: center;
  margin: 20px 0;
}

.qr-code {
  width: 200px;
  height: 200px;
  border: 1px solid #eee;
  padding: 10px;
  background: white;
}

.login-link {
  color: #41b883 !important;
}

:deep(.qr-result) {
  background: transparent !important;
}

:deep(.qr-result .n-result-header .n-result-icon) {
  color: #41b883 !important;
}

:deep(.qr-result .n-result-header .n-result-title) {
  color: v-bind('isDarkTheme ? "#e0e0e0" : "#2c3e50"');
}

:deep(.qr-result .n-result-content) {
  color: v-bind('isDarkTheme ? "#aaa" : "#555"');
}
</style>
