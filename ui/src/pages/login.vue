<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, NAlert, NSpace } from 'naive-ui'
import { useAuthStore } from '../stores/auth'
import type { FormRules } from 'naive-ui'
// Get auth store
const auth = useAuthStore()
const router = useRouter()

// Form data
const formData = reactive({
  login: '',
  password: '',
  totp: ''
})

// Form validation rules
const rules: FormRules = {
  login: [
    { required: true, message: 'Username is required', trigger: 'blur' }
  ],
  password: [
    { required: true, message: 'Password is required', trigger: 'blur' }
  ],
  totp: [
    { required: true, message: 'TOTP code is required', trigger: 'blur' },
    { type: 'string', len: 6, message: 'TOTP code must be 6 digits', trigger: 'blur' }
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
  <div class="login-container">
    <NCard title="Login" class="login-card">
      <NAlert v-if="auth.error" type="error" style="margin-bottom: 16px;">
        {{ auth.error }}
      </NAlert>

      <NForm
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-placement="top"
      >
        <NFormItem label="Username" path="login">
          <NInput
            v-model:value="formData.login"
            placeholder="Enter your username"
            autofocus
          />
        </NFormItem>

        <NFormItem label="Password" path="password">
          <NInput
            v-model:value="formData.password"
            type="password"
            placeholder="Enter your password"
            show-password-on="click"
          />
        </NFormItem>

        <NFormItem label="TOTP Code" path="totp">
          <NInput
            v-model:value="formData.totp"
            placeholder="Enter your 6-digit code"
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
              Login
            </NButton>
          </NSpace>
        </div>
      </NForm>
    </NCard>
  </div>
</template>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 80vh;
}

.login-card {
  width: 100%;
  max-width: 400px;
}
</style>
