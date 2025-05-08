<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import {
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NSpin,
  NAlert
} from 'naive-ui'
import { ownerApi } from '@/services/api'
import type { Owner } from '@/types/api'
import type { FormInst, FormRules } from 'naive-ui'

const props = defineProps<{
  associationId: number
  ownerId?: number
}>()

const emit = defineEmits<{
  (e: 'saved'): void
  (e: 'cancelled'): void
}>()

interface OwnerFormData {
  name: string
  identification_number: string
  contact_phone: string
  contact_email: string
}

const formData = reactive<OwnerFormData>({
  name: '',
  identification_number: '',
  contact_phone: '',
  contact_email: ''
})

const rules: FormRules = {
  name: [
    { required: true, message: 'Name is required', trigger: 'blur' }
  ],
  identification_number: [
    { required: true, message: 'Identification number is required', trigger: 'blur' }
  ],
  contact_phone: [
    { required: true, message: 'Contact phone is required', trigger: 'blur' }
  ],
  contact_email: [
    { required: true, message: 'Contact email is required', trigger: 'blur' },
    { type: 'email', message: 'Invalid email format', trigger: 'blur' }
  ]
}

const loading = ref<boolean>(false)
const submitting = ref<boolean>(false)
const error = ref<string | null>(null)
const formRef = ref<FormInst | null>(null)

const resetValidation = async () => {
  if (formRef.value) {
    try {
      await formRef.value.restoreValidation()
    } catch (err) {
      console.log('Error resetting validation:', err)
    }
  }
}

const fetchOwnerDetails = async () => {
  if (!props.ownerId) return

  try {
    loading.value = true
    error.value = null

    // Add an API method to get an owner by ID
    const response = await ownerApi.getOwner(props.associationId, props.ownerId)
    const ownerData = response.data

    // Update form data
    formData.name = ownerData.name || ''
    formData.identification_number = ownerData.identification_number || ''
    formData.contact_phone = ownerData.contact_phone || ''
    formData.contact_email = ownerData.contact_email || ''

    await resetValidation()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown error occurred'
    console.error('Error fetching owner details:', err)
  } finally {
    loading.value = false
  }
}

const validateFormManually = () => {
  if (!formData.name.trim()) {
    error.value = 'Name is required'
    return false
  }

  if (!formData.identification_number.trim()) {
    error.value = 'Identification number is required'
    return false
  }

  if (!formData.contact_phone.trim()) {
    error.value = 'Contact phone is required'
    return false
  }

  if (!formData.contact_email.trim()) {
    error.value = 'Contact email is required'
    return false
  }

  // Simple email validation
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(formData.contact_email)) {
    error.value = 'Invalid email format'
    return false
  }

  return true
}

const submitForm = async (e: MouseEvent) => {
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

    // Prepare the data to update/create
    const ownerData = {
      name: formData.name,
      identification_number: formData.identification_number,
      contact_phone: formData.contact_phone,
      contact_email: formData.contact_email
    }

    if (props.ownerId) {
      // Update an existing owner
      await ownerApi.updateOwner(
        props.associationId,
        props.ownerId,
        ownerData
      )
    } else {
      // Create a new owner
      await ownerApi.createOwner(
        props.associationId,
        ownerData
      )
    }

    emit('saved')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'An error occurred while submitting the form'
    console.error('Error submitting form:', err)
  } finally {
    submitting.value = false
  }
}

const cancelForm = () => {
  emit('cancelled')
}

onMounted(() => {
  if (props.ownerId) {
    fetchOwnerDetails()
  }
})
</script>

<template>
  <div class="owner-form">
    <h2>{{ props.ownerId ? 'Edit Owner' : 'Create New Owner' }}</h2>

    <NSpin :show="loading">
      <NAlert v-if="error" type="error" title="Error" style="margin-bottom: 16px;">
        {{ error }}
      </NAlert>

      <NForm
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-placement="left"
        label-width="180px"
        require-mark-placement="right-hanging"
      >
        <NFormItem label="Name" path="name">
          <NInput
            v-model:value="formData.name"
            placeholder="Enter owner's full name"
          />
        </NFormItem>

        <NFormItem label="Identification Number" path="identification_number">
          <NInput
            v-model:value="formData.identification_number"
            placeholder="Enter identification number"
          />
        </NFormItem>

        <NFormItem label="Contact Phone" path="contact_phone">
          <NInput
            v-model:value="formData.contact_phone"
            placeholder="Enter contact phone"
          />
        </NFormItem>

        <NFormItem label="Contact Email" path="contact_email">
          <NInput
            v-model:value="formData.contact_email"
            placeholder="Enter contact email"
            type="text"
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
              {{ props.ownerId ? 'Update Owner' : 'Create Owner' }}
            </NButton>
          </NSpace>
        </div>
      </NForm>
    </NSpin>
  </div>
</template>

<style scoped>
.owner-form {
  max-width: 600px;
  margin: 0 auto;
  padding: 1.5rem;
  border-radius: 8px;
}
</style>
