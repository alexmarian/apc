<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { NForm, NFormItem, NInput, NButton, NSpace, NSpin, NAlert, FormRules, NDatePicker, NInputNumber } from 'naive-ui'
import { expenseApi } from '@/services/api'
import type { ExpenseCreateRequest } from '@/types/api'
import CategorySelector from '@/components/CategorySelector.vue'
import AccountSelector from '@/components/AccountSelector.vue'

// Props
const props = defineProps<{
  associationId: number
  expenseId?: number // If provided, we're editing an existing expense
}>()

// Emits
const emit = defineEmits<{
  (e: 'saved'): void
  (e: 'cancelled'): void
}>()

// Form data
const formData = reactive<ExpenseCreateRequest>({
  amount: 0,
  description: '',
  destination: '',
  date: new Date().toISOString().split('T')[0], // Today's date in ISO format
  category_id: 0,
  account_id: 0
})

// Form validation rules
const rules: FormRules = {
  amount: [
    { required: true, message: 'Amount is required', trigger: 'blur' },
    { type: 'number', min: 0.01, message: 'Amount must be greater than 0', trigger: 'blur' }
  ],
  description: [
    { required: true, message: 'Description is required', trigger: 'blur' },
    { type: 'string', max: 255, message: 'Description cannot exceed 255 characters', trigger: 'blur' }
  ],
  date: [
    { required: true, message: 'Date is required', trigger: 'blur' }
  ],
  category_id: [
    { required: true, type: 'number', min: 1, message: 'Category is required', trigger: 'blur' }
  ],
  account_id: [
    { required: true, type: 'number', min: 1, message: 'Account is required', trigger: 'blur' }
  ]
}

// State
const loading = ref<boolean>(false)
const submitting = ref<boolean>(false)
const error = ref<string | null>(null)
const formRef = ref(null)

// Format timestamp to date string (YYYY-MM-DD)
const formatDateForInput = (timestamp: number) => {
  return new Date(timestamp).toISOString().split('T')[0]
}

// Fetch expense details if editing
const fetchExpenseDetails = async () => {
  if (!props.expenseId) return

  try {
    loading.value = true
    error.value = null

    const response = await expenseApi.getExpense(props.associationId, props.expenseId)
    const expenseData = response.data

    // Update form data
    formData.amount = expenseData.amount
    formData.description = expenseData.description
    formData.destination = expenseData.destination
    formData.date = expenseData.date
    formData.category_id = expenseData.category_id
    formData.account_id = expenseData.account_id
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown error occurred'
    console.error('Error fetching expense details:', err)
  } finally {
    loading.value = false
  }
}

// Handle date change
const handleDateChange = (timestamp: number) => {
  if (timestamp) {
    formData.date = formatDateForInput(timestamp)
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
    const isCreating = !props.expenseId

    // Send request
    if (isCreating) {
      await expenseApi.createExpense(props.associationId, formData)
    } else if (props.expenseId) {
      await expenseApi.updateExpense(props.associationId, props.expenseId, formData)
    }

    // Notify parent component
    emit('saved')
  } catch (err) {
    if (err instanceof Error) {
      error.value = err.message
    } else if (typeof err === 'object' && err !== null && 'response' in err) {
      // Axios error
      const axiosError = err as any
      error.value = axiosError.response?.data?.msg || 'An error occurred while submitting the form'
    } else {
      error.value = 'An unknown error occurred'
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
            :min="0"
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
