<script setup lang="ts">
import { ref, onMounted, h, computed, watch } from 'vue'
import { refDebounced } from '@vueuse/core'
import {
  NDataTable,
  NButton,
  NSpace,
  NEmpty,
  NSpin,
  NAlert,
  useMessage,
  NInput,
  NSelect
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { unitApi } from '@/services/api'
import type { Unit } from '@/types/api'

// Props
const props = defineProps<{
  associationId: number,
  buildingId: number,
  unitTypeFilter?: string | null,
  searchQuery?: string | null
}>()

// Emits
const emit = defineEmits<{
  (e: 'edit', unitId: number): void
  (e: 'view-report', unitId: number): void
  (e: 'units-rendered', units: Unit[]): void
  (e: 'unit-type-changed', newUnitType: string | null): void
  (e: 'search-query-changed', newSearchQuery: string | null): void
}>()

// Data
const units = ref<Unit[]>([])
const loading = ref<boolean>(false)
const error = ref<string | null>(null)
const message = useMessage()

// Filters - use shared filters if available, otherwise use local state
const unitTypeFilter = ref<string | null>(props.unitTypeFilter || null)
const searchQuery = ref<string | null>(props.searchQuery || null)
const debouncedSearchQuery = refDebounced(searchQuery, 1000)
// Available unit types for filter (will be populated from units)
const availableUnitTypes = computed(() => {
  const types = new Set<string>()
  units.value.forEach(unit => {
    if (unit.unit_type) {
      types.add(unit.unit_type)
    }
  })
  return Array.from(types).map(type => ({
    label: type,
    value: type
  }))
})

const filteredUnits = computed(() => {
  let result = [...units.value]

  // Filter by unit type if selected
  if (unitTypeFilter.value) {
    result = result.filter(unit => unit.unit_type === unitTypeFilter.value)
  }

  // Filter by search query if provided
  if (debouncedSearchQuery.value) {
    const query = debouncedSearchQuery.value.toLowerCase()
    result = result.filter(unit =>
      unit.unit_number.toLowerCase().includes(query) ||
      unit.address.toLowerCase().includes(query) ||
      unit.cadastral_number.toLowerCase().includes(query)
    )
  }

  return result
})

// Table columns definition
const columns = ref<DataTableColumns<Unit>>([
  {
    title: 'Unit Number',
    key: 'unit_number',
    sorter: 'default'
  },
  {
    title: 'Type',
    key: 'unit_type'
  },
  {
    title: 'Floor',
    key: 'floor',
    sorter: (a, b) => a.floor - b.floor
  },
  {
    title: 'Area',
    key: 'area',
    sorter: (a, b) => a.area - b.area
  },
  {
    title: 'Part',
    key: 'part'
  },
  {
    title: 'Entrance',
    key: 'entrance'
  },
  {
    title: 'Room Count',
    key: 'room_count'
  },
  {
    title: 'Address',
    key: 'address'
  },
  {
    title: 'Actions',
    key: 'actions',
    render(row) {
      return h(
        NSpace,
        {
          justify: 'center',
          align: 'center'
        },
        {
          default: () => [
            h(
              NButton,
              {
                strong: true,
                secondary: true,
                type: 'info',
                size: 'small',
                onClick: () => emit('edit', row.id)
              },
              { default: () => 'Edit' }
            ),
            h(
              NButton,
              {
                strong: true,
                secondary: true,
                type: 'success',
                size: 'small',
                onClick: () => emit('view-report', row.id)
              },
              { default: () => 'Report' }
            )
          ]
        }
      )
    }
  }
])

// Fetch units
const fetchUnits = async () => {
  if (!props.associationId || !props.buildingId) return

  try {
    loading.value = true
    error.value = null

    const response = await unitApi.getUnits(props.associationId, props.buildingId)
    units.value = response.data

    // Emit the units for parent components
    emit('units-rendered', filteredUnits.value)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown error occurred'
    console.error('Error fetching units:', err)
  } finally {
    loading.value = false
  }
}

const searchQueryChanged = (newValue) => {
  searchQuery.value = newValue
  emit('search-query-changed', newValue)
}
const unitTypeChanged = (newValue) => {
  unitTypeFilter.value = newValue
  emit('unit-type-changed', newValue)
}
// Watch for changes in filters and refresh data
watch(unitTypeFilter, (newUnitType) => {
  emit('unit-type-changed', newUnitType)
  emit('units-rendered', filteredUnits.value)
})

watch(searchQuery, (newSearchQuery) => {
  emit('search-query-changed', newSearchQuery)
  emit('units-rendered', filteredUnits.value)
})

// Reset filters
const resetFilters = () => {
  unitTypeFilter.value = null
  searchQuery.value = null
}

onMounted(() => {
  fetchUnits()
})
</script>

<template>
  <div class="units-list">
    <h2>Units</h2>

    <div class="filters">
      <NFlex align="center">
          <NText>Search:</NText>
          <NInput
            :value="searchQuery"
            @update:value="searchQueryChanged"
            placeholder="Search by unit number, address..."
            clearable
            style="width: 300px"
          />
          <NText>Unit Type:</NText>
          <NSelect
            :value="unitTypeFilter"
            @update:value="unitTypeChanged"
            :options="availableUnitTypes"
            placeholder="All Types"
            clearable
            style="width: 200px"
          />
        <NButton @click="resetFilters">Reset Filters</NButton>
      </NFlex>
    </div>

    <NSpin :show="loading">
      <NAlert v-if="error" type="error" title="Error" closable>
        {{ error }}
        <template #action>
          <NButton @click="fetchUnits">Retry</NButton>
        </template>
      </NAlert>

      <div v-if="filteredUnits.length > 0" class="summary">
        <div>
          <span class="unit-count">{{ filteredUnits.length }} units found</span>
        </div>
      </div>

      <NDataTable
        :columns="columns"
        :data="filteredUnits"
        :bordered="false"
        :single-line="false"
        :pagination="{
          pageSize: 10
        }"
      >
        <template #empty>
          <NEmpty description="No units found">
            <template #extra>
              <p>Try changing your search filters.</p>
            </template>
          </NEmpty>
        </template>
      </NDataTable>
    </NSpin>
  </div>
</template>

<style scoped>
.units-list {
  margin: 1rem 0;
}

.filters {
  margin-bottom: 1.5rem;
  padding: 1rem;
  border-radius: 4px;
  background-color: var(--background-color);
  border: 1px solid var(--border-color);
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  align-items: center;
}

.filters > div {
  display: flex;
  flex-direction: column; /* Ensure label is above the input */
  gap: 0.25rem; /* Add spacing between label and input */
  flex: 1;
  min-width: 300px;
}

.summary {
  margin: 1rem 0;
  padding: 0.5rem 1rem;
  font-size: 1.1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: var(--background-color);
  border-radius: 4px;
  border: 1px solid var(--border-color);
}

.unit-count {
  font-size: 0.9rem;
  color: var(--text-color);
  opacity: 0.8;
}
</style>
