<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { NSelect, NSpin } from 'naive-ui'
import { buildingApi } from '@/services/api'
import type { Building } from '@/types/api'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  associationId: number | null
  buildingId: number | null
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:buildingId', id: number): void
}>()

// I18n
const { t } = useI18n()

// Data
const buildings = ref<Building[]>([])
const loading = ref<boolean>(false)
const error = ref<string | null>(null)
const selectedBuildingId = ref<number | null>(props.buildingId || null)

// Fetch buildings for the selected association
const fetchBuildings = async () => {
  if (!props.associationId) {
    buildings.value = []
    selectedBuildingId.value = null
    return
  }
  try {
    loading.value = true
    error.value = null

    const response = await buildingApi.getBuildings(props.associationId)
    buildings.value = response.data
    if (buildings.value.length > 0 && selectedBuildingId.value === null) {
      selectedBuildingId.value = buildings.value[0].id
      emit('update:buildingId', selectedBuildingId.value)
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('common.error')
    console.error('Error fetching buildings:', err)
    buildings.value = []
  } finally {
    loading.value = false
  }
}

const options = computed(() => {
  return buildings.value.map(building => ({
    label: `${building.name} - ${building.address}`,
    value: building.id
  }))
})

// Handle value changes
const handleUpdateValue = (value: number) => {
  console.log(value)
  emit('update:buildingId', value)
}

// Watch for association changes
watch(
  () => props.associationId,
  (newValue) => {
    fetchBuildings()
  },
  { immediate: true }
)
</script>

<template>
  <div class="building-selector">
    <NSpin :show="loading">
      <NSelect
        v-model:value="selectedBuildingId"
        @update:value="handleUpdateValue"
        :options="options"
        :placeholder="t('units.building', 'Select Building')"
        :disabled="!props.associationId || loading || props.disabled"
        clearable
        filterable
      />
    </NSpin>
    <p v-if="error" class="error">{{ error }}</p>
  </div>
</template>

<style scoped>
.error {
  color: #d03050;
  font-size: 0.8rem;
  margin-top: 4px;
}

.building-selector {
  width: 100%;
  max-width: 300px;
  min-width: 300px; /* Prevent shrinking */
}
</style>
