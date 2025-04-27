<script setup lang="ts">
import { computed } from 'vue'
import { NDescriptions, NDescriptionsItem, NTag } from 'naive-ui'
import type { Unit } from '@/types/api'

const props = defineProps<{
  unit: Unit
  showDetails?: boolean
}>()

// Compute a display string for the unit
const unitDisplayString = computed(() => {
  return `${props.unit.unit_number} (${props.unit.unit_type}, ${props.unit.area} m²)`
})

// Get appropriate color for unit type
const unitTypeColor = computed(() => {
  switch (props.unit.unit_type.toLowerCase()) {
    case 'apartment':
      return 'success'
    case 'commercial':
      return 'info'
    case 'office':
      return 'warning'
    case 'parking':
      return 'error'
    case 'storage':
      return 'default'
    default:
      return 'default'
  }
})
</script>

<template>
  <div class="unit-info">
    <div v-if="!showDetails" class="unit-info-compact">
      <strong>{{ unit.unit_number }}</strong>
      <NTag :type="unitTypeColor" size="small">{{ unit.unit_type }}</NTag>
      <span>{{ unit.area }} m²</span>
    </div>

    <NDescriptions v-else bordered size="small">
      <NDescriptionsItem label="Unit Number">
        {{ unit.unit_number }}
      </NDescriptionsItem>
      <NDescriptionsItem label="Type">
        <NTag :type="unitTypeColor">{{ unit.unit_type }}</NTag>
      </NDescriptionsItem>
      <NDescriptionsItem label="Area">
        {{ unit.area }} m²
      </NDescriptionsItem>
      <NDescriptionsItem label="Floor">
        {{ unit.floor }}
      </NDescriptionsItem>
      <NDescriptionsItem label="Entrance">
        {{ unit.entrance }}
      </NDescriptionsItem>
      <NDescriptionsItem label="Rooms">
        {{ unit.room_count }}
      </NDescriptionsItem>
      <NDescriptionsItem label="Address" span="2">
        {{ unit.address }}
      </NDescriptionsItem>
    </NDescriptions>
  </div>
</template>

<style scoped>
.unit-info-compact {
  display: flex;
  align-items: center;
  gap: 8px;
}

.unit-info {
  margin: 4px 0;
}
</style>
