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
  NInput,
  NSelect,
  NCard,
  NText
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { unitApi } from '@/services/api'
import type { Unit } from '@/types/api'
import FilterBar from '@/components/FilterBar.vue'
import FilterField from '@/components/FilterField.vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const router = useRouter()

// Props
const props = defineProps<{
  associationId: number | null,
  buildingId: number | null
}>()

// Emits
const emit = defineEmits<{
  (e: 'edit', unitId: number): void
  (e: 'units-rendered', units: Unit[]): void
}>()

// Filters are owned by the page (which persists them to the URL) and bound
// two-way, so there is one copy of each rather than a page ref shadowed by a
// list ref that only ever syncs upward.
const unitTypeFilter = defineModel<string[]>('unitTypeFilter', { default: () => [] })
const searchQuery = defineModel<string | null>('searchQuery', { default: null })
const currentPage = defineModel<number>('page', { default: 1 })

const debouncedSearchQuery = refDebounced(searchQuery, 500)

// Data
const units = ref<Unit[]>([])
const loading = ref<boolean>(false)
const error = ref<string | null>(null)
const hasInitialized = ref(false)

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
  if (unitTypeFilter.value.length > 0) {
    result = result.filter(unit => unitTypeFilter.value.includes(unit.unit_type))
  }

  // Filter by search query if provided
  if (debouncedSearchQuery.value) {
    const query = debouncedSearchQuery.value.toLowerCase()
    result = result.filter(unit =>
      unit.unit_number.toLowerCase().includes(query) ||
      (unit.owner_names || '').toLowerCase().includes(query) ||
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
    sorter: 'default',
    titleColSpan: 1,
    render: (row) => h('span', { style: 'white-space: nowrap' }, row.cadastral_number)
  },
  {
    title: t('units.unitNumber', 'Unit Number'),
    key: 'unit_number',
    sorter: 'default',
    width: 80
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
    title: t('units.owners', 'Owners'),
    key: 'owner_names'
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
                tertiary: true,
                size: 'small',
                onClick: () => router.push({
                  path: `/units/${row.id}`,
                  query: {
                    associationId: props.associationId!.toString(),
                    buildingId: props.buildingId?.toString(),
                    unitTypeFilter: unitTypeFilter.value.length > 0 ? unitTypeFilter.value.join(',') : undefined,
                    searchQuery: searchQuery.value || undefined,
                    page: currentPage.value.toString()
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

// Narrowing the result set can leave the current page out of range, so any
// filter change returns to page 1.
const searchQueryChanged = (newValue: string | null) => {
  searchQuery.value = newValue
  currentPage.value = 1
}

const unitTypeChanged = (newValue: string[]) => {
  unitTypeFilter.value = newValue
  currentPage.value = 1
}

const resetFilters = () => {
  unitTypeFilter.value = []
  searchQuery.value = null
  currentPage.value = 1
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
  <div>
    <FilterBar resettable @reset="resetFilters">
      <FilterField :label="t('common.search', 'Search')" :grow="2">
        <NInput
          :value="searchQuery"
          @update:value="searchQueryChanged"
          :placeholder="t('units.searchPlaceholder', 'Search by unit number, owner...')"
          clearable
        />
      </FilterField>
      <FilterField :label="t('units.type', 'Unit Type')" :grow="1">
        <NSelect
          :value="unitTypeFilter"
          @update:value="unitTypeChanged"
          :options="availableUnitTypes"
          :placeholder="t('units.allTypes', 'All Types')"
          multiple
          clearable
          :max-tag-count="2"
        />
      </FilterField>
    </FilterBar>

    <NCard>
      <NSpin :show="loading">
        <NAlert v-if="error" type="error" :title="t('common.error', 'Error')" closable>
          {{ error }}
          <NButton @click="fetchUnits">{{ t('common.retry', 'Retry') }}</NButton>
        </NAlert>

        <NText v-if="filteredUnits.length > 0" depth="3" class="units-list__count">
          {{ t('units.unitsFound', { count: filteredUnits.length }) }}
        </NText>

        <NDataTable
          :columns="columns"
          :data="filteredUnits"
          :bordered="false"
          :single-line="false"
          :pagination="{
            page: currentPage,
            pageSize: 10,
            onUpdatePage: (page: number) => currentPage = page
          }"
        >
          <template #empty>
            <NEmpty :description="t('units.noUnitsFound', 'No units found')">
              <template #extra>
                {{ t('units.tryChangingFilters', 'Try changing your search filters.') }}
              </template>
            </NEmpty>
          </template>
        </NDataTable>
      </NSpin>
    </NCard>
  </div>
</template>

<style scoped>
.units-list__count {
  display: block;
  margin-bottom: 12px;
}
</style>
