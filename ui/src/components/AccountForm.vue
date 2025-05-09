<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { NForm, NFormItem, NInput, NButton, NSpace, NSpin, NAlert } from 'naive-ui'
import { accountApi } from '@/services/api'
import type { AccountCreateRequest, AccountUpdateRequest } from '@/types/api'
import type { FormRules } from 'naive-ui'
import { useI18n } from 'vue-i18n'

// Props
const props = defineProps<{
  associationId: number
  accountId?: number // If provided, we're editing an existing account
}>()

// Emits
const emit = defineEmits<{
  (e: 'saved'): void
  (e: 'cancelled'): void
}>()

// I18n
const { t } = useI18n()

// Form data
const formData = reactive<AccountCreateRequest>({
  number: '',
  destination: '',
  description: ''
})

// Form validation rules
const rules: FormRules = {
  number: [
    { required: true, message: t('validation.required', { field: t('accounts.accountNumber') }), trigger: 'blur' }
  ],
  description: [
    { required: true, message: t('validation.required', { field: t('accounts.description') }), trigger: 'blur' }
  ]
}

// State
const loading = ref<boolean>(false)
const submitting = ref<boolean>(false)
const error = ref<string | null>(null)
const formRef = ref(null)

// Fetch account details if editing
const fetchAccountDetails = async () => {
  if (!props.accountId) return

  try {
    loading.value = true
    error.value = null

    const response = await accountApi.getAccount(props.associationId, props.accountId)
    const accountData = response.data

    // Update form data
    formData.number = accountData.number
    formData.destination = accountData.destination
    formData.description = accountData.description
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('common.error')
    console.error('Error fetching account details:', err)
  } finally {
    loading.value = false
  }
}

// Submit form
const submitForm = async (e: MouseEvent) => {
  e.preventDefault()

  if (!formRef.value) return

  try {
    // @ts-ignore - Naive UI types issue with form ref
    await formRef.value.validate()

    submitting.value = true
    error.value = null

    // Determine if creating or updating
    const isCreating = !props.accountId

    // Send request
    if (isCreating) {
      await accountApi.createAccount(props.associationId, formData)
    } else if (props.accountId) {
      const updateData: AccountUpdateRequest = {
        number: formData.number,
        destination: formData.destination,
        description: formData.description
      }
      await accountApi.updateAccount(props.associationId, props.accountId, updateData)
    }

    // Notify parent component
    emit('saved')
  } catch (err) {
    if (err instanceof Error) {
      error.value = err.message
    } else if (typeof err === 'object' && err !== null && 'response' in err) {
      // Axios error
      const axiosError = err as any
      error.value = axiosError.response?.data?.msg || t('common.error')
    } else {
      error.value = t('common.error')
    }
    console.error('Error submitting form:', err)
  } finally {
    submitting.value = false
  }
}

// Cancel form
const cancelForm = () => {
  emit('cancelled')
}

// On component mount
onMounted(() => {
  if (props.accountId) {
    fetchAccountDetails()
  }
})
</script>

<template>
  <div class="account-form">
    <h2>{{ props.accountId ? t('common.edit') : t('accounts.createNew') }}</h2>

    <NSpin :show="loading">
      <NAlert v-if="error" type="error" :title="t('common.error')" style="margin-bottom: 16px;">
        {{ error }}
      </NAlert>

      <NForm
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-placement="left"
        label-width="120px"
        require-mark-placement="right-hanging"
      >
        <NFormItem :label="t('accounts.accountNumber')" path="number">
          <NInput
            v-model:value="formData.number"
            :placeholder="t('accounts.accountNumber')"
          />
        </NFormItem>

        <NFormItem :label="t('accounts.description')" path="description">
          <NInput
            v-model:value="formData.description"
            :placeholder="t('accounts.description')"
          />
        </NFormItem>

        <NFormItem :label="t('accounts.destination')" path="destination">
          <NInput
            v-model:value="formData.destination"
            :placeholder="t('accounts.destination')"
          />
        </NFormItem>

        <div style="margin-top: 24px;">
          <NSpace justify="end">
            <NButton
              @click="cancelForm"
              :disabled="submitting"
            >
              {{ t('common.cancel') }}
            </NButton>

            <NButton
              type="primary"
              @click="submitForm"
              :loading="submitting"
            >
              {{ props.accountId ? t('common.update') : t('common.create') }}
            </NButton>
          </NSpace>
        </div>
      </NForm>
    </NSpin>
  </div>
</template>

<style scoped>
.account-form {
  max-width: 600px;
  margin: 0 auto;
  padding: 1.5rem;
  border-radius: 8px;
}
</style>
