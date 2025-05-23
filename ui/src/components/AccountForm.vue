<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { NForm, NFormItem, NInput, NButton, NSpace, NSpin, NAlert, useMessage } from 'naive-ui'
import { accountApi } from '@/services/api'
import type { Account, AccountCreateRequest, AccountUpdateRequest } from '@/types/api'
import type { FormRules } from 'naive-ui'
import { useI18n } from 'vue-i18n'

// Props
const props = defineProps<{
  associationId: number
  accountId?: number // If provided, we're editing an existing account
}>()

// Emits
const emit = defineEmits<{
  (e: 'saved', account: Account): void  // Pass the updated/created account data
  (e: 'cancelled'): void
}>()

// I18n
const { t } = useI18n()
const message = useMessage()

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
const originalAccount = ref<Account | null>(null)

const isEditMode = computed(() => !!props.accountId)

// Fetch account details if editing
const fetchAccountDetails = async () => {
  if (!props.accountId) return

  try {
    loading.value = true
    error.value = null

    const response = await accountApi.getAccount(props.associationId, props.accountId)
    const accountData = response.data
    originalAccount.value = accountData

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

    let response: { data: Account }

    // Determine if creating or updating
    if (isEditMode.value && props.accountId) {
      const updateData: AccountUpdateRequest = {
        number: formData.number,
        destination: formData.destination,
        description: formData.description
      }
      response = await accountApi.updateAccount(props.associationId, props.accountId, updateData)
      message.success(t('accounts.accountUpdated'))
    } else {
      response = await accountApi.createAccount(props.associationId, formData)
      message.success(t('accounts.accountCreated'))
    }

    // Emit the updated/created account data so parent can update the list without reloading
    emit('saved', response.data)
  } catch (err) {
    let errorMessage: string
    if (err instanceof Error) {
      errorMessage = err.message
    } else if (typeof err === 'object' && err !== null && 'response' in err) {
      // Axios error
      const axiosError = err as any
      errorMessage = axiosError.response?.data?.msg || t('common.error')
    } else {
      errorMessage = t('common.error')
    }
    error.value = errorMessage
    message.error(errorMessage)
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
    <h2>{{ isEditMode ? t('accounts.editAccount') : t('accounts.createNew') }}</h2>

    <NSpin :show="loading">
      <NAlert v-if="error" type="error" :title="t('common.error')" style="margin-bottom: 16px;" closable @close="error = null">
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
        <!-- Readonly Account ID for edit mode -->
        <NFormItem v-if="isEditMode && originalAccount" :label="'Account ID'">
          <NInput
            :value="originalAccount.id.toString()"
            readonly
            class="account-form__readonly-input"
          />
        </NFormItem>

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
              {{ isEditMode ? t('common.update') : t('common.create') }}
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

.account-form__readonly-input {
  background-color: var(--input-color-disabled);
  color: var(--text-color-disabled);
}

.account-form__readonly-input :deep(.n-input__input-el) {
  background-color: var(--input-color-disabled) !important;
  color: var(--text-color-disabled) !important;
  cursor: not-allowed;
}
</style>
