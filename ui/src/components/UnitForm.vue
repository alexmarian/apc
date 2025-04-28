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
  FormRules,
  NInputNumber,
  NSelect
} from 'naive-ui'
import { unitApi } from '@/services/api'
import type { Unit } from '@/types/api'

const props = defineProps<{
  associationId: number
  buildingId: number
  unitId?: number
}>()

const emit = defineEmits<{
  (e: 'saved'): void
  (e: 'cancelled'): void
}>()

interface UnitFormData {
  unit_number: string
  address: string
  entrance: number
  area: number
  part: number
  unit_type: string
  floor: number
  room_count: number
}

// Default unit types for selection
const unitTypeOptions = [
  { label: 'Apartment', value: 'apartment' },
  { label: 'Commercial', value: 'commercial' },
  { label: 'Office', value: 'office' },
  { label: 'Parking', value: 'parking' },
  { label: 'Storage', value: 'storage' }
]

const formData = reactive<UnitFormData>({
  unit_number: '',
  address: '',
  entrance: 1,
  area: 0,
  part: 0,
  unit_type: 'apartment',
  floor: 1,
  room_count: 1
})

const rules: FormRules = {
  unit_number: [
    { required: true, message: 'Unit number is required', trigger: 'blur' }
  ],
  address: [
    { required: true, message: 'Address is required', trigger: 'blur' }
  ],
  entrance: [
    { required: true, message: 'Entrance is required', trigger: 'blur' },
    {
      type: 'number',
      min: 1,
      message: 'Entrance must be at least 1',
      trigger: 'blur'
    }
  ],
  area: [
    { required: true, message: 'Area is required', trigger: 'blur' },
    {
      type: 'number',
      min: 0.01,
      message: 'Area must be greater than 0',
      trigger: 'blur'
    }
  ],
  part: [
    { required: true, message: 'Part is required', trigger: 'blur' },
    {
      type: 'number',
      min: 0,
      max: 1,
      message: 'Part must be between 0 and 1',
      trigger: 'blur'
    }
  ],
  unit_type: [
    { required: true, message: 'Unit type is required', trigger: 'blur' }
  ],
  floor: [
    { required: true, message: 'Floor is required', trigger: 'blur' }
  ],
  room_count: [
    { required: true, message: 'Room count is required', trigger: 'blur' },
    {
      type: 'number',
      min: 0,
      message: 'Room count must be at least 0',
      trigger: 'blur'
    }
  ]
}

const loading = ref<boolean>(false)
const submitting = ref<boolean>(false)
const error = ref<string | null>(null)
const formRef = ref(null)
const dataLoaded = ref(false)

const resetValidation = async () => {
  if (formRef.value) {
    try {
      await formRef.value.resetValidation()
    } catch (err) {
      console.log('Error resetting validation:', err)
    }
  }
}

const fetchUnitDetails = async () => {
  if (!props.unitId) return

  try {
    loading.value = true
    error.value = null

    const response = await unitApi.getUnit(props.associationId, props.buildingId, props.unitId)
    const unitData = response.data

    // Update form data
    formData.unit_number = unitData.unit_number || ''
    formData.address = unitData.address || ''
    formData.entrance = unitData.entrance || 1
    formData.area = unitData.area || 0
    formData.part = unitData.part || 0
    formData.unit_type = unitData.unit_type || 'apartment'
    formData.floor = unitData.floor || 1
    formData.room_count = unitData.room_count || 1

    dataLoaded.value = true
    await nextTick()
    resetValidation()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown error occurred'
    console.error('Error fetching unit details:', err)
  } finally {
    loading.value = false
  }
}

const validateFormManually = () => {
  if (!formData.unit_number.trim()) {
    error.value = 'Unit number is required'
    return false
  }

  if (!formData.address.trim()) {
    error.value = 'Address is required'
    return false
  }

  if (formData.area <= 0) {
    error.value = 'Area must be greater than 0'
    return false
  }

  if (formData.part < 0 || formData.part > 1) {
    error.value = 'Part must be between 0 and 1'
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

    // Prepare the data to update
    const updateData = {
      unit_number: formData.unit_number,
      address: formData.address,
      entrance: formData.entrance,
      unit_type: formData.unit_type,
      floor: formData.floor,
      room_count: formData.room_count
    }

    // For now, we only have the update functionality in the API
    // We're not creating new units from the UI
    if (props.unitId) {
      await unitApi.updateUnit(
        props.associationId,
        props.buildingId,
        props.unitId,
        updateData
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
  if (props.unitId) {
    fetchUnitDetails()
  }
})
</script>

<template>
  <div class="unit-form">
    <h2>{{ props.unitId ? 'Edit Unit' : 'Create New Unit' }}</h2>

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
        <NFormItem label="Unit Number" path="unit_number">
          <NInput
            v-model:value="formData.unit_number"
            placeholder="Enter unit number"
          />
        </NFormItem>

        <NFormItem label="Address" path="address">
          <NInput
            v-model:value="formData.address"
            placeholder="Enter address"
          />
        </NFormItem>

        <NFormItem label="Entrance" path="entrance">
          <NInputNumber
            v-model:value="formData.entrance"
            :min="1"
            :precision="0"
            style="width: 100%"
          />
        </NFormItem>

        <NFormItem label="Area" path="area">
          <NInputNumber
            v-model:value="formData.area"
            :min="0.01"
            :precision="2"
            style="width: 100%"
            disabled
          />
        </NFormItem>

        <NFormItem label="Part" path="part">
          <NInputNumber
            v-model:value="formData.part"
            :min="0"
            :max="1"
            :precision="3"
            style="width: 100%"
            disabled
          />
        </NFormItem>

        <NFormItem label="Unit Type" path="unit_type">
          <NSelect
            v-model:value="formData.unit_type"
            :options="unitTypeOptions"
            placeholder="Select unit type"
          />
        </NFormItem>

        <NFormItem label="Floor" path="floor">
          <NInputNumber
            v-model:value="formData.floor"
            :precision="0"
            style="width: 100%"
          />
        </NFormItem>

        <NFormItem label="Room Count" path="room_count">
          <NInputNumber
            v-model:value="formData.room_count"
            :min="0"
            :precision="0"
            style="width: 100%"
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
              {{ props.unitId ? 'Update Unit' : 'Create Unit' }}
            </NButton>
          </NSpace>
        </div>
      </NForm>
    </NSpin>
  </div>
</template>

<style scoped>
.unit-form {
  max-width: 600px;
  margin: 0 auto;
  padding: 1.5rem;
  border-radius: 8px;
}
</style>
