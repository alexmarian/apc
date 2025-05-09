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
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

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
    { required: true, message: t('validation.required', '{field} is required', { field: t('owners.name', 'Name') }), trigger: 'blur' }
  ],
  identification_number: [
    { required: true, message: t('validation.required', '{field} is required', { field: t('owners.identification', 'Identification number') }), trigger: 'blur' }
  ],
  contact_phone: [
    { required: true, message: t('validation.required', '{field} is required', { field: t('owners.contactPhone', 'Contact phone') }), trigger: 'blur' }
  ],
  contact_email: [
    { required: true, message: t('validation.required', '{field} is required', { field: t('owners.contactEmail', 'Contact email') }), trigger: 'blur' },
    { type: 'email', message: t('validation.email', 'Please enter a valid email address'), trigger: 'blur' }
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
    error.value = err instanceof Error ? err.message : t('common.error', 'Unknown error occurred')
    console.error('Error fetching owner details:', err)
  } finally {
    loading.value = false
  }
}

const validateFormManually = () => {
  if (!formData.name.trim()) {
    error.value = t('validation.required', '{field} is required', { field: t('owners.name', 'Name') })
    return false
  }

  if (!formData.identification_number.trim()) {
    error.value = t('validation.required', '{field} is required', { field: t('owners.identification', 'Identification number') })
    return false
  }

  if (!formData.contact_phone.trim()) {
    error.value = t('validation.required', '{field} is required', { field: t('owners.contactPhone', 'Contact phone') })
    return false
  }

  if (!formData.contact_email.trim()) {
    error.value = t('validation.required', '{field} is required', { field: t('owners.contactEmail', 'Contact email') })
    return false
  }

  // Simple email validation
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(formData.contact_email)) {
    error.value = t('validation.email', 'Invalid email format')
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
    error.value = err instanceof Error ? err.message : t('common.error', 'An error occurred while submitting the form')
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
    <h2>{{ props.ownerId ? t('owners.editOwner', 'Edit Owner') : t('owners.createOwner', 'Create New Owner') }}</h2>

    <NSpin :show="loading">
      <NAlert v-if="error" type="error" :title="t('common.error', 'Error')" style="margin-bottom: 16px;">
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
        <NFormItem :label="t('owners.name', 'Name')" path="name">
          <NInput
            v-model:value="formData.name"
            :placeholder="t('owners.enterName', 'Enter owner\'s full name')"
          />
        </NFormItem>

        <NFormItem :label="t('owners.identification', 'Identification Number')" path="identification_number">
          <NInput
            v-model:value="formData.identification_number"
            :placeholder="t('owners.enterIdentification', 'Enter identification number')"
          />
        </NFormItem>

        <NFormItem :label="t('owners.contactPhone', 'Contact Phone')" path="contact_phone">
          <NInput
            v-model:value="formData.contact_phone"
            :placeholder="t('owners.enterPhone', 'Enter contact phone')"
          />
        </NFormItem>

        <NFormItem :label="t('owners.contactEmail', 'Contact Email')" path="contact_email">
          <NInput
            v-model:value="formData.contact_email"
            :placeholder="t('owners.enterEmail', 'Enter contact email')"
            type="text"
          />
        </NFormItem>

        <div style="margin-top: 24px;">
          <NSpace justify="end">
            <NButton
              @click="cancelForm"
              :disabled="submitting"
            >
              {{ t('common.cancel', 'Cancel') }}
            </NButton>

            <NButton
              type="primary"
              @click="submitForm"
              :loading="submitting"
            >
              {{ props.ownerId ? t('common.update', 'Update Owner') : t('common.create', 'Create Owner') }}
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
