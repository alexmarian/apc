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
  NInputNumber,
  NSelect
} from 'naive-ui'
import { unitApi } from '@/services/api'
import type { Unit } from '@/types/api'
import { UnitType } from '@/types/api'
import type { FormInst, FormRules } from 'naive-ui'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  associationId: number
  buildingId: number
  unitId?: number
}>()

const emit = defineEmits<{
  saved: []
  cancelled: []
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

const unitTypeOptions = computed(() =>
  Object.entries(UnitType).map(([key, value]) => ({
    label: t(`unitTypes.${value}`),
    value
  }))
)

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
    {
      required: true,
      message: t('validation.required', { field: t('units.unit') }),
      trigger: 'blur'
    }
  ],
  address: [
    {
      required: true,
      message: t('validation.required', { field: t('units.address') }),
      trigger: 'blur'
    }
  ],
  entrance: [
    {
      required: true,
      type: 'number',
      min: 1,
      message: t('units.entranceMin'),
      trigger: 'blur'
    }
  ],
  area: [
    {
      required: true,
      type: 'number',
      min: 0.01,
      message: t('units.areaPositive'),
      trigger: 'blur'
    }
  ],
  part: [
    {
      required: true,
      type: 'number',
      min: 0,
      max: 1,
      message: t('units.partRange'),
      trigger: 'blur'
    }
  ],
  unit_type: [
    {
      required: true,
      message: t('validation.required', { field: t('units.type') }),
      trigger: 'blur'
    }
  ],
  floor: [
    {
      required: true,
      type: 'number',
      message: t('validation.required', { field: t('units.floor') }),
      trigger: ['blur', 'change']
    }
  ],
  room_count: [
    {
      required: true,
      type: 'number',
      min: 0,
      message: t('units.roomCountMin'),
      trigger: 'blur'
    }
  ]
}

const loading = ref(false)
const submitting = ref(false)
const error = ref<string | null>(null)
const formRef = ref<FormInst | null>(null)

const isEditMode = computed(() => !!props.unitId)

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
    unit_number: '',
    address: '',
    entrance: 1,
    area: 0,
    part: 0,
    unit_type: 'apartment',
    floor: 1,
    room_count: 1
  })
  error.value = null
}

const fetchUnitDetails = async () => {
  if (!props.unitId) return

  try {
    loading.value = true
    error.value = null

    const response = await unitApi.getUnit(props.associationId, props.buildingId, props.unitId)
    const unitData = response.data

    // Update form data with fallback values
    Object.assign(formData, {
      unit_number: unitData.unit_number || '',
      address: unitData.address || '',
      entrance: unitData.entrance || 1,
      area: unitData.area || 0,
      part: unitData.part || 0,
      unit_type: unitData.unit_type || 'apartment',
      floor: unitData.floor || 1,
      room_count: unitData.room_count || 1
    })

    await nextTick()
    await resetValidation()
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : t('common.error')
    error.value = errorMessage
    console.error('Error fetching unit details:', err)
  } finally {
    loading.value = false
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

    const updateData = {
      unit_number: formData.unit_number.trim(),
      address: formData.address.trim(),
      entrance: formData.entrance,
      unit_type: formData.unit_type,
      floor: formData.floor,
      room_count: formData.room_count,
      // Include area and part if they're editable in the future
      area: formData.area,
      part: formData.part
    }

    if (isEditMode.value) {
      await unitApi.updateUnit(
        props.associationId,
        props.buildingId,
        props.unitId!,
        updateData
      )
    } else {
      //never happens
    }

    emit('saved')
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : t('common.error')
    error.value = errorMessage
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
    fetchUnitDetails()
  } else {
    resetForm()
  }
})
</script>

<template>
  <div class="unit-form">
    <h2 class="unit-form__title">
      {{ isEditMode ? t('units.editUnit') : t('units.createUnit') }}
    </h2>

    <NSpin :show="loading">
      <NAlert
        v-if="error"
        type="error"
        :title="t('common.error')"
        class="unit-form__error"
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
        class="unit-form__form"
      >
        <NFormItem :label="t('units.unit')" path="unit_number">
          <NInput
            v-model:value="formData.unit_number"
            :placeholder="t('units.enterUnitNumber')"
            clearable
          />
        </NFormItem>

        <NFormItem :label="t('units.address')" path="address">
          <NInput
            v-model:value="formData.address"
            :placeholder="t('units.enterAddress')"
            clearable
          />
        </NFormItem>

        <NFormItem :label="t('units.entrance')" path="entrance">
          <NInputNumber
            v-model:value="formData.entrance"
            :min="1"
            :precision="0"
            class="unit-form__number-input"
          />
        </NFormItem>

        <NFormItem :label="t('units.area')" path="area">
          <NInputNumber
            v-model:value="formData.area"
            :min="0.01"
            :precision="2"
            class="unit-form__number-input unit-form__number-input--disabled"
            disabled
          />
        </NFormItem>

        <NFormItem :label="t('units.part')" path="part">
          <NInputNumber
            v-model:value="formData.part"
            :min="0"
            :max="1"
            :precision="3"
            class="unit-form__number-input unit-form__number-input--disabled"
            disabled
          />
        </NFormItem>

        <NFormItem :label="t('units.type')" path="unit_type">
          <NSelect
            v-model:value="formData.unit_type"
            :options="unitTypeOptions"
            :placeholder="t('units.selectType')"
            clearable
          />
        </NFormItem>

        <NFormItem :label="t('units.floor')" path="floor">
          <NInputNumber
            v-model:value="formData.floor"
            :precision="0"
            class="unit-form__number-input"
          />
        </NFormItem>

        <NFormItem :label="t('units.roomCount')" path="room_count">
          <NInputNumber
            v-model:value="formData.room_count"
            :min="0"
            :precision="0"
            class="unit-form__number-input"
          />
        </NFormItem>

        <div class="unit-form__actions">
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
.unit-form {
  max-width: 600px;
  margin: 0 auto;
  padding: 1.5rem;
  border-radius: 8px;
}

.unit-form__title {
  margin-bottom: 1.5rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-color-1);
}

.unit-form__error {
  margin-bottom: 1rem;
}

.unit-form__form {
  margin-bottom: 1.5rem;
}

.unit-form__number-input {
  width: 100%;
}

.unit-form__number-input--disabled {
  --n-text-color-disabled: var(--text-color-2);
}

.unit-form__number-input--disabled :deep(.n-input-number-input) {
  text-decoration: none !important;
}

.unit-form__number-input--disabled :deep(.n-input__input-el[disabled]) {
  text-decoration: none !important;
  -webkit-text-decoration: none !important;
}

.unit-form__actions {
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border-color);
}
</style>
