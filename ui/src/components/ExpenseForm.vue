<script setup lang="ts">
import { ref, reactive, onMounted, nextTick, computed } from 'vue'
import {
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NSpin,
  NAlert,
  NDatePicker,
  NInputNumber,
  useMessage
} from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'
import { expenseApi } from '@/services/api'
import type { ExpenseCreateRequest, Expense } from '@/types/api'
import CategorySelector from '@/components/CategorySelector.vue'
import AccountSelector from '@/components/AccountSelector.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const message = useMessage()

const props = defineProps<{
  associationId: number
  expenseId?: number
}>()

const emit = defineEmits<{
  saved: [expense: Expense]  // Pass the updated/created expense data
  cancelled: []
}>()

const formData = reactive<ExpenseCreateRequest>({
  amount: 0.01,
  description: '',
  destination: '',
  date: new Date().toISOString().split('T')[0],
  category_id: 0,
  account_id: 0,
  document_ref: ''
})

const rules: FormRules = {
  amount: [
    {
      required: true,
      type: 'number',
      min: 0.01,
      message: t('expenses.amountPositive'),
      trigger: ['blur', 'change']
    }
  ],
  description: [
    {
      required: true,
      message: t('validation.required', { field: t('expenses.description') }),
      trigger: 'blur'
    },
    {
      type: 'string',
      max: 255,
      message: t('validation.maxLength', { field: t('expenses.description'), max: 255 }),
      trigger: 'blur'
    }
  ],
  date: [
    {
      required: true,
      message: t('validation.required', { field: t('expenses.date') }),
      trigger: 'blur'
    }
  ],
  category_id: [
    {
      type: 'number',
      min: 1,
      message: t('validation.required', { field: t('expenses.category') }),
      trigger: ['blur', 'change']
    }
  ],
  document_ref: [
    {
      type: 'string',
      max: 50,
      message: t('validation.maxLength', { field: t('expenses.documentRef'), max: 50 }),
      trigger: 'blur'
    }
  ]
}

const loading = ref(false)
const submitting = ref(false)
const error = ref<string | null>(null)
const formRef = ref<FormInst | null>(null)
const originalExpense = ref<Expense | null>(null)

const isEditMode = computed(() => !!props.expenseId)

const formatDateForInput = (timestamp: number | string): string => {
  return new Date(timestamp).toISOString().split('T')[0]
}

const resetValidation = async () => {
  if (formRef.value) {
    try {
      await formRef.value.restoreValidation()
    } catch (err) {
      console.warn('Error resetting validation:', err)
    }
  }
}

const resetForm = () => {
  Object.assign(formData, {
    amount: 0.01,
    description: '',
    destination: '',
    date: new Date().toISOString().split('T')[0],
    category_id: 0,
    account_id: 0,
    document_ref: ''
  })
  error.value = null
}

const fetchExpenseDetails = async () => {
  if (!props.expenseId) return

  try {
    loading.value = true
    error.value = null

    const response = await expenseApi.getExpense(props.associationId, props.expenseId)
    const expenseData: Expense = response.data
    originalExpense.value = expenseData

    Object.assign(formData, {
      amount: Number(expenseData.amount) || 0.01,
      description: expenseData.description || '',
      destination: expenseData.destination || '',
      date: expenseData.date ? formatDateForInput(expenseData.date) : new Date().toISOString().split('T')[0],
      category_id: Number(expenseData.category_id) || 0,
      account_id: Number(expenseData.account_id) || 0,
      document_ref: expenseData.document_ref || ''
    })

    await nextTick()
    await resetValidation()
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : t('common.error')
    error.value = errorMessage
    console.error('Error fetching expense details:', err)
  } finally {
    loading.value = false
  }
}

const handleDateChange = (timestamp: number | null) => {
  if (timestamp) {
    formData.date = formatDateForInput(timestamp)
  }
}

const handleSubmit = async () => {
  error.value = null

  // Validate form
  try {
    await formRef.value?.validate()
  } catch (validationError) {
    console.warn('Form validation failed:', validationError)
    return
  }

  try {
    submitting.value = true

    const submitData: ExpenseCreateRequest = {
      amount: Number(formData.amount),
      description: formData.description.trim(),
      destination: formData.destination?.trim() || '',
      date: formData.date + 'T00:00:00Z',
      category_id: Number(formData.category_id),
      account_id: Number(formData.account_id),
      document_ref: formData.document_ref?.trim() || ''
    }

    let response: { data: Expense }

    if (isEditMode.value && props.expenseId) {
      response = await expenseApi.updateExpense(props.associationId, props.expenseId, submitData)
      message.success(t('expenses.expenseUpdated'))
    } else {
      response = await expenseApi.createExpense(props.associationId, submitData)
      message.success(t('expenses.expenseCreated'))
    }

    // Emit the updated/created expense data so parent can update the list without reloading
    emit('saved', response.data)
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : t('common.error')
    error.value = errorMessage
    message.error(errorMessage)
    console.error('Error submitting form:', err)
  } finally {
    submitting.value = false
  }
}

const handleCancel = () => {
  emit('cancelled')
}

onMounted(() => {
  if (isEditMode.value) {
    fetchExpenseDetails()
  } else {
    resetForm()
  }
})
</script>

<template>
  <div class="expense-form">
    <h2 class="expense-form__title">
      {{ isEditMode ? t('expenses.editExpense') : t('expenses.createNew') }}
    </h2>

    <NSpin :show="loading">
      <NAlert
        v-if="error"
        type="error"
        :title="t('common.error')"
        class="expense-form__error"
        closable
        @close="error = null"
      >
        {{ error }}
      </NAlert>

      <NForm
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-placement="left"
        label-width="120px"
        require-mark-placement="right-hanging"
        class="expense-form__form"
      >
        <!-- Readonly Expense ID for edit mode -->
        <NFormItem v-if="isEditMode && originalExpense" :label="'Expense ID'">
          <NInput
            :value="originalExpense.id.toString()"
            readonly
            class="expense-form__readonly-input"
          />
        </NFormItem>

        <NFormItem :label="t('expenses.amount')" path="amount">
          <NInputNumber
            v-model:value="formData.amount"
            :min="0.01"
            :precision="2"
            class="expense-form__number-input"
            clearable
          />
        </NFormItem>

        <NFormItem :label="t('expenses.date')" path="date">
          <NDatePicker
            :value="formData.date ? new Date(formData.date).getTime() : null"
            type="date"
            clearable
            @update:value="handleDateChange"
            class="expense-form__date-picker"
          />
        </NFormItem>

        <NFormItem :label="t('expenses.description')" path="description">
          <NInput
            v-model:value="formData.description"
            :placeholder="t('expenses.enterDescription')"
            type="textarea"
            :rows="3"
            :maxlength="255"
            :show-count="true"
            clearable
          />
        </NFormItem>

        <NFormItem :label="t('expenses.destination')" path="destination">
          <NInput
            v-model:value="formData.destination"
            :placeholder="t('expenses.enterDestination')"
            clearable
          />
        </NFormItem>

        <NFormItem :label="t('expenses.category')" path="category_id">
          <CategorySelector
            v-model:modelValue="formData.category_id"
            :association-id="props.associationId"
            :placeholder="t('expenses.selectCategory')"
            :disabled="submitting"
          />
        </NFormItem>

        <NFormItem :label="t('expenses.account')" path="account_id">
          <AccountSelector
            v-model:modelValue="formData.account_id"
            :association-id="props.associationId"
            :active-only="true"
            :placeholder="t('expenses.selectAccount')"
            :disabled="submitting"
          />
        </NFormItem>

        <NFormItem :label="t('expenses.documentRef')" path="document_ref">
          <NInput
            v-model:value="formData.document_ref"
            :placeholder="t('expenses.enterDocumentRef')"
            clearable
          />
        </NFormItem>

        <div class="expense-form__actions">
          <NSpace justify="end">
            <NButton
              @click="handleCancel"
              :disabled="submitting"
            >
              {{ t('common.cancel') }}
            </NButton>

            <NButton
              type="primary"
              @click="handleSubmit"
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
.expense-form {
  max-width: 600px;
  margin: 0 auto;
  padding: 1.5rem;
  border-radius: 8px;
}

.expense-form__title {
  margin-bottom: 1.5rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-color-1);
}

.expense-form__error {
  margin-bottom: 1rem;
}

.expense-form__form {
  margin-bottom: 1.5rem;
}

.expense-form__readonly-input {
  background-color: var(--input-color-disabled);
  color: var(--text-color-disabled);
}

.expense-form__readonly-input :deep(.n-input__input-el) {
  background-color: var(--input-color-disabled) !important;
  color: var(--text-color-disabled) !important;
  cursor: not-allowed;
}

.expense-form__number-input,
.expense-form__date-picker {
  width: 100%;
}

.expense-form__actions {
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border-color);
}
</style>
