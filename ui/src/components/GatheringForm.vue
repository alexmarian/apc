<template>
  <div class="gathering-form">
    <NCard style="width: 600px">
      <template #header>
        <h2>{{ isEditMode ? $t('gatherings.edit') : $t('gatherings.create') }}</h2>
      </template>

      <NSpin :show="loading">
        <NAlert v-if="error" type="error" closable @close="error = null">
          {{ error }}
        </NAlert>

        <NForm ref="formRef" :model="formData" :rules="rules" label-placement="top">
          <NFormItem :label="$t('gatherings.title')" path="title">
            <NInput v-model:value="formData.title" :placeholder="$t('gatherings.titlePlaceholder')" />
          </NFormItem>

          <NFormItem :label="$t('gatherings.description')" path="description">
            <NInput
              v-model:value="formData.description"
              type="textarea"
              :placeholder="$t('gatherings.descriptionPlaceholder')"
              :rows="3"
            />
          </NFormItem>

          <NFormItem :label="$t('gatherings.location')" path="location">
            <NInput v-model:value="formData.location" :placeholder="$t('gatherings.locationPlaceholder')" />
          </NFormItem>

          <NFormItem :label="$t('gatherings.scheduledDate')" path="scheduled_date">
            <NDatePicker
              v-model:value="formData.scheduled_date"
              type="datetime"
              :placeholder="$t('gatherings.scheduledDatePlaceholder')"
              style="width: 100%"
            />
          </NFormItem>

          <NFormItem :label="$t('gatherings.type')" path="type">
            <NSelect v-model:value="formData.type" :options="typeOptions" />
          </NFormItem>

          <NFormItem :label="$t('gatherings.votingMode.title')" path="voting_mode">
            <NSelect v-model:value="formData.voting_mode" :options="votingModeOptions" />
          </NFormItem>

          <NFormItem :label="$t('gatherings.qualificationCriteria')" path="qualification_criteria">
            <NCard size="small">
              <template #header>
                <h4>{{ $t('gatherings.qualificationSettings') }}</h4>
              </template>

              <NSpace vertical>
                <NFormItem :label="$t('gatherings.unitTypes')" label-placement="left">
                  <NSelect
                    v-model:value="formData.qualification_criteria.unit_types"
                    :options="unitTypeOptions"
                    multiple
                    :placeholder="$t('gatherings.allUnitTypes')"
                    clearable
                  />
                </NFormItem>

                <NFormItem :label="$t('gatherings.floors')" label-placement="left">
                  <NSelect
                    v-model:value="formData.qualification_criteria.floors"
                    :options="floorOptions"
                    multiple
                    :placeholder="$t('gatherings.allFloors')"
                    clearable
                  />
                </NFormItem>

                <NFormItem :label="$t('gatherings.entrances')" label-placement="left">
                  <NSelect
                    v-model:value="formData.qualification_criteria.entrances"
                    :options="entranceOptions"
                    multiple
                    :placeholder="$t('gatherings.allEntrances')"
                    clearable
                  />
                </NFormItem>

              </NSpace>
            </NCard>
          </NFormItem>
        </NForm>

        <div class="form-actions">
          <NSpace justify="end">
            <NButton @click="handleCancel">
              {{ $t('common.cancel') }}
            </NButton>
            <NButton type="primary" @click="handleSubmit" :loading="submitting">
              {{ isEditMode ? $t('common.update') : $t('common.create') }}
            </NButton>
          </NSpace>
        </div>
      </NSpin>
    </NCard>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NCard,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NButton,
  NSpace,
  NDatePicker,
  NAlert,
  NSpin,
  type FormInst,
  type FormRules
} from 'naive-ui'
import { gatheringApi, unitApi, buildingApi } from '@/services/api'
import type {
  GatheringCreateRequest,
  GatheringUpdateRequest,
  Gathering,
  GatheringType,
  QualificationCriteria,
  Unit,
  Building
} from '@/types/api'

interface Props {
  associationId: number
  gathering?: Gathering
}

const props = defineProps<Props>()

const emit = defineEmits<{
  saved: []
  cancelled: []
}>()

const { t } = useI18n()

const formRef = ref<FormInst | null>(null)
const loading = ref(false)
const submitting = ref(false)
const error = ref<string | null>(null)
const buildings = ref<Building[]>([])
const units = ref<Unit[]>([])

const isEditMode = computed(() => !!props.gathering)

const formData = reactive<{
  title: string
  description: string
  location: string
  scheduled_date: number | null
  type: GatheringType
  voting_mode: 'by_weight' | 'by_unit'
  qualification_criteria: QualificationCriteria
}>({
  title: '',
  description: '',
  location: '',
  scheduled_date: null,
  type: 'initial' as GatheringType,
  voting_mode: 'by_weight',
  qualification_criteria: {
    unit_types: [],
    floors: [],
    entrances: []
  }
})

const typeOptions = computed(() => [
  { label: t('gatherings.type.initial'), value: 'initial' },
  { label: t('gatherings.type.repeated'), value: 'repeated' },
  { label: t('gatherings.type.remote'), value: 'remote' }
])

const votingModeOptions = computed(() => [
  { label: t('gatherings.votingMode.byWeight'), value: 'by_weight' },
  { label: t('gatherings.votingMode.byUnit'), value: 'by_unit' }
])

const unitTypeOptions = computed(() => [
  { label: t('unitTypes.apartment'), value: 'apartment' },
  { label: t('unitTypes.commercial'), value: 'commercial' },
  { label: t('unitTypes.office'), value: 'office' },
  { label: t('unitTypes.parking'), value: 'parking' },
  { label: t('unitTypes.storage'), value: 'storage' }
])

const floorOptions = computed(() => {
  const floors = new Set<number>()
  units.value.forEach(unit => floors.add(unit.floor))
  return Array.from(floors)
    .sort((a, b) => a - b)
    .map(floor => ({ label: `${t('units.floor')} ${floor}`, value: floor }))
})

const entranceOptions = computed(() => {
  const entrances = new Set<number>()
  units.value.forEach(unit => entrances.add(unit.entrance))
  return Array.from(entrances)
    .sort((a, b) => a - b)
    .map(entrance => ({ label: `${t('units.entrance')} ${entrance}`, value: entrance }))
})


const rules: FormRules = {
  title: [
    { required: true, message: t('gatherings.titleRequired') }
  ],
  description: [
    { required: true, message: t('gatherings.descriptionRequired') }
  ],
  location: [
    { required: true, message: t('gatherings.locationRequired') }
  ],
  scheduled_date: [
    { required: true, message: t('gatherings.scheduledDateRequired'), type: 'number' }
  ],
  type: [
    { required: true, message: t('gatherings.typeRequired') }
  ]
}

const loadBuildings = async () => {
  try {
    const response = await buildingApi.getBuildings(props.associationId)
    buildings.value = response.data
  } catch (err: unknown) {
    console.error('Failed to load buildings:', err)
  }
}

const loadUnits = async () => {
  try {
    const allUnits: Unit[] = []
    for (const building of buildings.value) {
      const response = await unitApi.getUnits(props.associationId, building.id)
      allUnits.push(...response.data)
    }
    units.value = allUnits
  } catch (err: unknown) {
    console.error('Failed to load units:', err)
  }
}

const loadData = async () => {
  loading.value = true
  try {
    await loadBuildings()
    await loadUnits()
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : t('common.error')
    error.value = errorMessage
  } finally {
    loading.value = false
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    submitting.value = true
    error.value = null

    // Clean up qualification criteria - remove empty arrays
    const qualificationCriteria: QualificationCriteria = {}
    if (formData.qualification_criteria.unit_types?.length) {
      qualificationCriteria.unit_types = formData.qualification_criteria.unit_types
    }
    if (formData.qualification_criteria.floors?.length) {
      qualificationCriteria.floors = formData.qualification_criteria.floors
    }
    if (formData.qualification_criteria.entrances?.length) {
      qualificationCriteria.entrances = formData.qualification_criteria.entrances
    }

    const requestData = {
      title: formData.title,
      description: formData.description,
      intent: formData.description, // Use description as intent
      location: formData.location,
      gathering_date: new Date(formData.scheduled_date!).toISOString(),
      gathering_type: formData.type,
      voting_mode: formData.voting_mode,
      qualification_unit_types: qualificationCriteria.unit_types || [],
      qualification_floors: qualificationCriteria.floors || [],
      qualification_entrances: qualificationCriteria.entrances || [],
      qualification_custom_rule: ""
    }

    if (isEditMode.value) {
      await gatheringApi.updateGathering(props.associationId, props.gathering!.id, requestData as GatheringUpdateRequest)
    } else {
      await gatheringApi.createGathering(props.associationId, requestData as GatheringCreateRequest)
    }

    emit('saved')
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : t('common.error')
    error.value = errorMessage
  } finally {
    submitting.value = false
  }
}

const handleCancel = () => {
  emit('cancelled')
}

watch(() => props.gathering, (newGathering) => {
  if (newGathering) {
    Object.assign(formData, {
      title: newGathering.title,
      description: newGathering.description,
      location: newGathering.location,
      scheduled_date: new Date(newGathering.scheduled_date).getTime(),
      type: newGathering.type,
      voting_mode: newGathering.voting_mode || 'by_weight',
      qualification_criteria: {
        unit_types: newGathering.qualification_criteria?.unit_types || [],
        floors: newGathering.qualification_criteria?.floors || [],
        entrances: newGathering.qualification_criteria?.entrances || []
      }
    })
  }
}, { immediate: true })

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.gathering-form {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 16px;
}

.form-actions {
  margin-top: 24px;
}
</style>
