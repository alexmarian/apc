<script setup lang="ts">
import { ref, onMounted, h, computed, watch, nextTick } from 'vue'
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
  NSelect,
  NCard,
  NFlex,
  NText
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { unitApi } from '@/services/api'
import type { Unit } from '@/types/api'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const router = useRouter()

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
const hasInitialized = ref(false)

// Filters - use shared filters if available, otherwise use local state
const unitTypeFilter = ref<string | null>(props.unitTypeFilter || null)
const searchQuery = ref<string | null>(props.searchQuery || null)
const debouncedSearchQuery = refDebounced(searchQuery, 500)

// Available unit types for filter (will be populated from units)
const availableUnitTypes = computed(() => {
  const types = new Set<string>()
  units.value.forEach(unit => {
    if (unit.unit_type) {
      types.add(unit.unit_type)
    }
  })
  return Array.from(types).map(type => ({
    label: t(`unitTypes.${type}`, type),
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
const columns = computed<DataTableColumns<Unit>>(() => [
  {
    title: t('units.cadastralNumber', 'Cadastral Number'),
    key: 'cadastral_number',
    sorter: 'default'
  },
  {
    title: t('units.unitNumber', 'Unit Number'),
    key: 'unit_number',
    sorter: 'default'
  },
  {
    title: t('units.type', 'Type'),
    key: 'unit_type',
    render: (row) => t(`unitTypes.${row.unit_type}`, row.unit_type)
  },
  {
    title: t('units.floor', 'Floor'),
    key: 'floor',
    sorter: (a, b) => a.floor - b.floor
  },
  {
    title: t('units.area', 'Area'),
    key: 'area',
    sorter: (a, b) => a.area - b.area,
    render: (row) => `${row.area.toFixed(2)} m²`
  },
  {
    title: t('units.part', 'Part'),
    key: 'part',
    render: (row) => `${(row.part * 100).toFixed(3)}%`
  },
  {
    title: t('units.entrance', 'Entrance'),
    key: 'entrance'
  },
  {
    title: t('units.roomCount', 'Room Count'),
    key: 'room_count'
  },
  {
    title: t('units.address', 'Address'),
    key: 'address'
  },
  {
    title: t('common.actions', 'Actions'),
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
              { default: () => t('common.edit', 'Edit') }
            ),
            h(NButton,
              {
                type: 'primary',
                size: 'small',
                onClick: () => router.push({
                  path: `/units/${row.id}`,
                  query: {
                    associationId: props.associationId.toString(),
                    buildingId: props.buildingId.toString(),
                    unitTypeFilter: unitTypeFilter.value || undefined,
                    searchQuery: searchQuery.value || undefined
                  }
                })
              },
              { default: () => t('common.details', 'View Details') }
            )
          ]
        }
      )
    }
  }
])

// Fetch units - simplified to prevent loops
const fetchUnits = async () => {
  // Prevent multiple simultaneous calls
  if (loading.value) {
    console.log('Already loading, skipping fetch')
    return
  }

  if (!props.associationId || !props.buildingId) {
    console.log('Missing associationId or buildingId, skipping fetch')
    return
  }

  try {
    loading.value = true
    error.value = null
    console.log('Fetching units for association:', props.associationId, 'building:', props.buildingId)

    const response = await unitApi.getUnits(props.associationId, props.buildingId)
    units.value = response.data
    hasInitialized.value = true

    console.log('Units fetched successfully:', response.data.length, 'units')
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('common.error', 'Unknown error occurred')
    console.error('Error fetching units:', err)
  } finally {
    loading.value = false
  }
}

// Method to update a single unit in the list (called when unit is updated)
const updateUnit = (updatedUnit: Unit) => {
  const index = units.value.findIndex(unit => unit.id === updatedUnit.id)
  if (index !== -1) {
    // Replace the unit in the array while maintaining reactivity
    units.value.splice(index, 1, updatedUnit)
    console.log('Unit updated in list:', updatedUnit)
  }
}

// Method to add a new unit to the list (for future create functionality)
const addUnit = (newUnit: Unit) => {
  units.value.push(newUnit)
}

// Expose methods for parent components
defineExpose({
  updateUnit,
  addUnit,
  refreshData: fetchUnits
})

const searchQueryChanged = (newValue: string | null) => {
  searchQuery.value = newValue
  emit('search-query-changed', newValue)
}

const unitTypeChanged = (newValue: string | null) => {
  unitTypeFilter.value = newValue
  emit('unit-type-changed', newValue)
}

// Reset filters
const resetFilters = () => {
  unitTypeFilter.value = null
  searchQuery.value = null
}

// Watch for prop changes and refetch only when necessary
watch(() => [props.associationId, props.buildingId],
  ([newAssocId, newBuildId], [oldAssocId, oldBuildId]) => {
    console.log('Props changed:', { newAssocId, newBuildId, oldAssocId, oldBuildId })

    // Only fetch if both IDs are present and at least one has changed
    if (newAssocId && newBuildId && (newAssocId !== oldAssocId || newBuildId !== oldBuildId)) {
      hasInitialized.value = false
      units.value = []
      fetchUnits()
    }
  },
  { immediate: false }
)

// Watch filtered units and emit changes, but only after initialization
watch(filteredUnits, (newFilteredUnits) => {
  if (hasInitialized.value) {
    emit('units-rendered', newFilteredUnits)
  }
}, { deep: true })

// Initialize on mount
onMounted(() => {
  console.log('UnitsList mounted with props:', props)
  if (props.associationId && props.buildingId) {
    fetchUnits()
  }
})
</script>

<template>
  <div class="units-list">
    <NCard style="margin-top: 16px;">
      <NFlex align="center">
        <NText>{{ t('common.search', 'Search') }}:</NText>
        <NInput
          :value="searchQuery"
          @update:value="searchQueryChanged"
          :placeholder="t('units.searchPlaceholder', 'Search by unit number, address...')"
          clearable
          style="width: 300px"
        />
        <NText>{{ t('units.type', 'Unit Type') }}:</NText>
        <NSelect
          :value="unitTypeFilter"
          @update:value="unitTypeChanged"
          :options="availableUnitTypes"
          :placeholder="t('units.allTypes', 'All Types')"
          clearable
          style="width: 200px"
        />
        <NButton @click="resetFilters">{{ t('common.reset_filters', 'Reset Filters') }}</NButton>
      </NFlex>
    </NCard>

    <NCard style="margin-top: 16px;">
      <NSpin :show="loading">
        <NAlert v-if="error" type="error" :title="t('common.error', 'Error')" closable>
          {{ error }}
          <NButton @click="fetchUnits">{{ t('common.retry', 'Retry') }}</NButton>
        </NAlert>

        <div v-if="filteredUnits.length > 0" class="summary">
          <div>
            <span class="unit-count">
              {{ t('units.unitsFound', { count: filteredUnits.length }) }}
            </span>
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
            <NEmpty :description="t('units.noUnitsFound', 'No units found')">
              <template #extra>
                <p>{{ t('units.tryChangingFilters', 'Try changing your search filters.') }}</p>
              </template>
            </NEmpty>
          </template>
        </NDataTable>
      </NSpin>
    </NCard>
  </div>
</template>

<style scoped>
.units-list {
  margin: 1rem 0;
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
