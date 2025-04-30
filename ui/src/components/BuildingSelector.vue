<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { NSelect, NSpin } from 'naive-ui'
import { buildingApi } from '@/services/api'
import type { Building } from '@/types/api'

const props = defineProps<{
  associationId: number | null
  buildingId: number | null
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:buildingId', id: number): void
}>()

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
    error.value = err instanceof Error ? err.message : 'Unknown error occurred'
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
        :placeholder="'Select Building'"
        :disabled="!props.associationId || loading || props.disabled"
        clearable
        filterable
      />
    </NSpin>
    <p v-if="error" class="error">{{ error }}</p>
  </div>
</template>

<style scoped>
.error-text {
  color: #e03;
  font-size: 0.85rem;
  margin-top: 0.5rem;
}

.building-selector {
  width: 100%;
  max-width: 300px;
  min-width: 300px; /* Prevent shrinking */
}
</style>
