<script setup lang="ts">
import { ref, reactive, onMounted, nextTick } from 'vue'
import {
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NSpin,
  NAlert,
  NDatePicker,
  NInputNumber
} from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'
import { expenseApi } from '@/services/api'
import type { ExpenseCreateRequest, Expense, ApiResponse } from '@/types/api'
import CategorySelector from '@/components/CategorySelector.vue'
import AccountSelector from '@/components/AccountSelector.vue'

const props = defineProps<{
  associationId: number
  expenseId?: number
}>()

const emit = defineEmits<{
  (e: 'saved'): void
  (e: 'cancelled'): void
}>()

const formData = reactive<ExpenseCreateRequest>({
  amount: 0.01,
  description: '',
  destination: '',
  date: new Date().toISOString().split('T')[0],
  category_id: 0,
  account_id: 0
})

const rules: FormRules = {
  amount: [
    { required: true, message: 'Amount is required', trigger: 'blur' },
    {
      validator: (rule: any, value: number) => {
        return Number(value) > 0
      },
      message: 'Amount must be greater than 0',
      trigger: 'blur'
    }
  ],
  description: [
    { required: true, message: 'Description is required', trigger: 'blur' },
    {
      type: 'string',
      max: 255,
      message: 'Description cannot exceed 255 characters',
      trigger: 'blur'
    }
  ],
  date: [
    { required: true, message: 'Date is required', trigger: 'blur' }
  ],
  category_id: [
    {
      validator: (rule: any, value: number) => {
        return value > 0
      },
      message: 'Category is required',
      trigger: 'blur'
    }
  ],
  account_id: [
    {
      validator: (rule: any, value: number) => {
        return value > 0
      },
      message: 'Account is required',
      trigger: 'blur'
    }
  ]
}

const loading = ref<boolean>(false)
const submitting = ref<boolean>(false)
const error = ref<string | null>(null)
const formRef = ref<FormInst | null>(null)
const dataLoaded = ref<boolean>(false)

const formatDateForInput = (timestamp: number | string): string => {
  return new Date(timestamp).toISOString().split('T')[0]
}

const resetValidation = async (): Promise<void> => {
  if (formRef.value) {
    try {
      await formRef.value.restoreValidation()
    } catch (err) {
      console.log('Error resetting validation:', err)
    }
  }
}

const fetchExpenseDetails = async (): Promise<void> => {
  if (!props.expenseId) return

  try {
    loading.value = true
    error.value = null

    const response = await expenseApi.getExpense(props.associationId, props.expenseId)
    const expenseData: Expense = response.data

    formData.amount = Number(expenseData.amount) || 0.01
    formData.description = expenseData.description || ''
    formData.destination = expenseData.destination || ''
    formData.date = expenseData.date ? formatDateForInput(expenseData.date) : new Date().toISOString().split('T')[0]
    formData.category_id = Number(expenseData.category_id) || 0
    formData.account_id = Number(expenseData.account_id) || 0

    dataLoaded.value = true
    await nextTick()
    resetValidation()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown error occurred'
    console.error('Error fetching expense details:', err)
  } finally {
    loading.value = false
  }
}

const handleDateChange = (timestamp: number): void => {
  if (timestamp) {
    formData.date = new Date(timestamp).toISOString().split('T')[0]
  }
}

const validateFormManually = (): boolean => {
  if (Number(formData.amount) <= 0) {
    error.value = 'Amount must be greater than 0'
    return false
  }

  if (!formData.description.trim()) {
    error.value = 'Description is required'
    return false
  }

  if (!formData.date) {
    error.value = 'Date is required'
    return false
  }

  if (formData.category_id <= 0) {
    error.value = 'Category is required'
    return false
  }

  if (formData.account_id <= 0) {
    error.value = 'Account is required'
    return false
  }

  return true
}

const submitForm = async (e: MouseEvent): Promise<void> => {
  e.preventDefault()
  error.value = null

  let isValid = true
  if (formRef.value) {
    try {
      await formRef.value.validate()
    } catch (err) {
      console.log('Form validation failed, using manual validation')
      isValid = false
    }
  }

  if (!isValid) {
    isValid = validateFormManually()
    if (!isValid) return
  }

  try {
    submitting.value = true

    const formDataToSubmit: ExpenseCreateRequest = {
      amount: Number(formData.amount),
      description: formData.description,
      destination: formData.destination,
      date: formData.date + 'T00:00:00Z',
      category_id: Number(formData.category_id),
      account_id: Number(formData.account_id)
    }

    const isCreating = !props.expenseId

    if (isCreating) {
      await expenseApi.createExpense(props.associationId, formDataToSubmit)
    } else if (props.expenseId) {
      await expenseApi.updateExpense(props.associationId, props.expenseId, formDataToSubmit)
    }

    emit('saved')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'An error occurred while submitting the form'
    console.error('Error submitting form:', err)
  } finally {
    submitting.value = false
  }
}

const cancelForm = (): void => {
  emit('cancelled')
}

onMounted(() => {
  if (props.expenseId) {
    fetchExpenseDetails()
  }
})
</script>

<template>
  <div class="expense-form">
    <h2>{{ props.expenseId ? 'Edit Expense' : 'Create New Expense' }}</h2>

    <NSpin :show="loading">
      <NAlert v-if="error" type="error" title="Error" style="margin-bottom: 16px;">
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
        <NFormItem label="Amount" path="amount">
          <NInputNumber
            v-model:value="formData.amount"
            :min="0.01"
            :precision="2"
            style="width: 100%"
          />
        </NFormItem>

        <NFormItem label="Date" path="date">
          <NDatePicker
            :value="formData.date ? new Date(formData.date).getTime() : null"
            type="date"
            clearable
            @update:value="handleDateChange"
            style="width: 100%"
          />
        </NFormItem>

        <NFormItem label="Description" path="description">
          <NInput
            v-model:value="formData.description"
            placeholder="Enter description"
          />
        </NFormItem>

        <NFormItem label="Destination" path="destination">
          <NInput
            v-model:value="formData.destination"
            placeholder="Enter destination (payee)"
          />
        </NFormItem>

        <NFormItem label="Category" path="category_id">
          <CategorySelector
            v-model:modelValue="formData.category_id"
            :association-id="props.associationId"
            placeholder="Select category"
            :disabled="submitting"
          />
        </NFormItem>

        <NFormItem label="Account" path="account_id">
          <AccountSelector
            v-model:modelValue="formData.account_id"
            :association-id="props.associationId"
            :active-only="true"
            placeholder="Select account"
            :disabled="submitting"
          />
        </NFormItem>

        <div style="margin-top: 24px;">
          <NSpace justify="end">
            <NButton
              @click="cancelForm"
              :disabled="submitting"
            >
              Cancel
            </NButton>

            <NButton
              type="primary"
              @click="submitForm"
              :loading="submitting"
            >
              {{ props.expenseId ? 'Update Expense' : 'Create Expense' }}
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
</style>
