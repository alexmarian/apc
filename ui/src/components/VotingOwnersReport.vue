<script setup lang="ts">
import { ref, computed, onMounted, watch, h } from 'vue'
import {
  NCard,
  NDataTable,
  NSpin,
  NAlert,
  NButton,
  useMessage,
  NSpace,
  NInput,
  NInputGroup,
  NSelect,
  NInputNumber,
  NFormItem,
  NForm,
  NGrid,
  NFormItemGi,
  NCheckbox
} from 'naive-ui'
import type { DataTableColumns, SelectOption } from 'naive-ui'
import { ownerApi } from '@/services/api'
import { formatPercentage } from '@/utils/formatters'
import { useI18n } from 'vue-i18n'

import type { VotingOwner, VotingUnit } from '@/types/api'

// Props
const props = defineProps<{
  associationId: number
}>()

// Define emits
const emit = defineEmits<{
  (e: 'edit-owner', ownerId: number): void
}>()

// i18n
const { t } = useI18n()

// State
const loading = ref<boolean>(false)
const error = ref<string | null>(null)
const votingOwners = ref<VotingOwner[]>([])
const sortBy = ref<'name' | 'share' | 'relative_share'>('share')
const sortOrder = ref<'asc' | 'desc'>('desc')
const message = useMessage()

// Filter state
const filterUnitType = ref<string[]>([])
const filterFloor = ref<number[]>([])
const filterEntrance = ref<number[]>([])
const showFilters = ref<boolean>(false)

// Get unique values for filter options - these should also come from server
const unitTypeOptions = computed<SelectOption[]>(() => {
  // In a real implementation, this should be fetched from a separate API endpoint
  // that returns available filter options for the association
  const commonTypes = ['apartment', 'commercial', 'office', 'parking', 'storage']
  return commonTypes.map(type => ({
    label: t(`unitTypes.${type}`, type),
    value: type
  }))
})

const floorOptions = computed<SelectOption[]>(() => {
  // This should also come from a server endpoint that returns available floors
  // For now, we'll generate a reasonable range
  const floors = Array.from({length: 20}, (_, i) => i - 2) // -2 to 17
  return floors.map(floor => ({
    label: floor < 0 ? `${t('units.basement', 'Basement')} ${Math.abs(floor)}` : `${t('units.floor', 'Floor')} ${floor}`,
    value: floor
  }))
})

const entranceOptions = computed<SelectOption[]>(() => {
  // This should also come from a server endpoint that returns available entrances
  // For now, we'll generate a reasonable range
  const entrances = Array.from({length: 10}, (_, i) => i + 1) // 1 to 10
  return entrances.map(entrance => ({
    label: `${t('units.entrance', 'Entrance')} ${entrance}`,
    value: entrance
  }))
})

// Filtered and sorted data - now only handles sorting since filtering is server-side
const sortedFilteredData = computed<VotingOwner[]>(() => {
  if (!votingOwners.value) return []

  let data = [...votingOwners.value]

  // Apply sorting only - filtering is done server-side
  data.sort((a, b) => {
    let comparison = 0

    if (sortBy.value === 'name') {
      comparison = a.name.localeCompare(b.name)
    } else if (sortBy.value === 'share') {
      comparison = a.total_condo_part - b.total_condo_part
    } else if (sortBy.value === 'relative_share') {
      const totalShare = data.reduce((sum, owner) => sum + owner.total_condo_part, 0)
      const aRelativeShare = totalShare > 0 ? a.total_condo_part / totalShare : 0
      const bRelativeShare = totalShare > 0 ? b.total_condo_part / totalShare : 0
      comparison = aRelativeShare - bRelativeShare
    }

    return sortOrder.value === 'asc' ? comparison : -comparison
  })

  return data
})

// Calculate relative share for server-filtered data
const getRelativeShare = (owner: VotingOwner): number => {
  const totalShare = votingOwners.value.reduce((sum, o) => sum + o.total_condo_part, 0)
  return totalShare > 0 ? owner.total_condo_part / totalShare : 0
}

// Columns for the data table
const columns = computed<DataTableColumns<VotingOwner>>(() => [
  {
    title: t('owners.name', 'Owner'),
    key: 'name',
    sorter: (a, b) => a.name.localeCompare(b.name)
  },
  {
    title: t('owners.identification', 'Identification'),
    key: 'identification_number'
  },
  {
    title: t('owners.contactPhone', 'Contact Phone'),
    key: 'contact_phone'
  },
  {
    title: t('owners.contactEmail', 'Contact Email'),
    key: 'contact_email'
  },
  {
    title: t('units.title', 'Units'),
    key: 'total_units',
    sorter: (a, b) => a.total_units - b.total_units
  },
  {
    title: t('owners.totalArea', 'Total Area'),
    key: 'total_area',
    sorter: (a, b) => a.total_area - b.total_area,
    render: (row: VotingOwner) => `${row.total_area.toFixed(2)} m²`
  },
  {
    title: t('owners.votingShare', 'Voting Share'),
    key: 'total_condo_part',
    sorter: (a, b) => a.total_condo_part - b.total_condo_part,
    render: (row: VotingOwner) => formatPercentage(row.total_condo_part, 4)
  },
  {
    title: t('owners.relativeShare', 'Relative Share'),
    key: 'relative_share',
    sorter: (a, b) => getRelativeShare(a) - getRelativeShare(b),
    render: (row: VotingOwner) => formatPercentage(getRelativeShare(row), 4)
  },
  {
    title: t('common.actions', 'Actions'),
    key: 'actions',
    render:(row: VotingOwner) => h(
      NSpace,
      { justify: 'center' },
      {
        default: () => [
          h(
            NButton,
            {
              size: 'small',
              onClick: () => handleViewUnits(row)
            },
            { default: () => t('common.details', 'Details') }
          ),
          h(
            NButton,
            {
              size: 'small',
              type: 'primary',
              onClick: () => handleEditOwner(row.owner_id)
            },
            { default: () => t('common.edit', 'Edit') }
          )
        ]
      }
    )
  }
])

// Columns for unit details
const unitColumns = computed(() => [
  {
    title: t('units.building', 'Building'),
    key: 'building_name'
  },
  {
    title: t('units.unitNumber', 'Unit Number'),
    key: 'unit_number'
  },
  {
    title: t('units.area', 'Area'),
    key: 'area',
    render: (row: VotingUnit) => `${row.area.toFixed(2)} m²`
  },
  {
    title: t('units.part', 'Part'),
    key: 'part',
    render: (row: VotingUnit) => formatPercentage(row.part, 4)
  },
  {
    title: t('units.type', 'Type'),
    key: 'unit_type',
    render: (row: VotingUnit) => t(`unitTypes.${row.unit_type}`, row.unit_type)
  },
  {
    title: t('units.floor', 'Floor'),
    key: 'floor',
    render: (row: VotingUnit) => (row as any).floor || '-'
  },
  {
    title: t('units.entrance', 'Entrance'),
    key: 'entrance',
    render: (row: VotingUnit) => (row as any).entrance || '-'
  }
])

// UI state
const selectedOwner = ref<VotingOwner | null>(null)

// Fetch voting owners report with filters
const fetchVotingOwnersReport = async () => {
  if (!props.associationId) return

  try {
    loading.value = true
    error.value = null

    // Build filter parameters
    const filterParams: any = {}

    if (filterUnitType.value.length > 0) {
      filterParams.unit_types = filterUnitType.value.join(',')
    }

    if (filterFloor.value.length > 0) {
      filterParams.floors = filterFloor.value.join(',')
    }

    if (filterEntrance.value.length > 0) {
      filterParams.entrances = filterEntrance.value.join(',')
    }

    const response = await ownerApi.getVotingOwners(props.associationId, filterParams)
    votingOwners.value = response.data || []
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('owners.voting.loadError', 'Failed to load voting owners')
    console.error('Error fetching voting owners:', err)
    votingOwners.value = []
  } finally {
    loading.value = false
  }
}

// Handle view units (details)
const handleViewUnits = (owner: VotingOwner) => {
  if (selectedOwner.value?.owner_id === owner.owner_id) {
    selectedOwner.value = null
  } else {
    selectedOwner.value = owner
  }
}

// Handle edit owner
const handleEditOwner = (ownerId: number) => {
  emit('edit-owner', ownerId)
}

// Apply filters and re-fetch data
const applyFilters = () => {
  fetchVotingOwnersReport()
}

// Watch for filter changes and refetch data
watch([filterUnitType, filterFloor, filterEntrance], () => {
  if (props.associationId) {
    applyFilters()
  }
}, { deep: true })

// Clear all filters
const clearFilters = () => {
  filterUnitType.value = []
  filterFloor.value = []
  filterEntrance.value = []
  // Data will be refetched automatically due to watchers
}

// Export to CSV
const exportToCSV = () => {
  if (!votingOwners.value || votingOwners.value.length === 0) {
    message.error(t('owners.voting.noDataToExport', 'No data to export'))
    return
  }

  try {
    // Create CSV headers
    const headers = [
      t('owners.voting.csvHeaders.id', 'Owner ID'),
      t('owners.voting.csvHeaders.name', 'Owner Name'),
      t('owners.voting.csvHeaders.identification', 'Identification Number'),
      t('owners.voting.csvHeaders.phone', 'Contact Phone'),
      t('owners.voting.csvHeaders.email', 'Contact Email'),
      t('owners.voting.csvHeaders.unitsCount', 'Units Count'),
      t('owners.voting.csvHeaders.area', 'Total Area (m²)'),
      t('owners.voting.csvHeaders.votingShare', 'Voting Share (%)'),
      t('owners.voting.csvHeaders.relativeShare', 'Relative Share (%)'),
      t('owners.voting.csvHeaders.units', 'Units')
    ]

    // Create CSV rows using filtered data
    const rows = sortedFilteredData.value.map(owner => {
      return [
        owner.owner_id,
        owner.name,
        owner.identification_number,
        owner.contact_phone,
        owner.contact_email,
        owner.total_units,
        owner.total_area.toFixed(2),
        (owner.total_condo_part * 100).toFixed(4),
        (getRelativeShare(owner) * 100).toFixed(4),
        owner.units.map(u => `${u.building_name} - ${u.unit_number}`).join('; ')
      ]
    })

    // Create CSV content
    let csvContent = headers.join(',') + '\n'
    rows.forEach(row => {
      const escapedRow = row.map(value => {
        if (typeof value === 'string' && (value.includes(',') || value.includes('"'))) {
          return `"${value.replace(/"/g, '""')}"`
        }
        return value
      })
      csvContent += escapedRow.join(',') + '\n'
    })

    // Add filter information
    csvContent += '\n'
    csvContent += t('owners.voting.csvHeaders.filterInfo', 'Filter Information') + '\n'
    if (filterUnitType.value.length > 0) {
      csvContent += `${t('owners.voting.csvHeaders.unitTypeFilter', 'Unit Type Filter')},${filterUnitType.value.join('; ')}\n`
    }
    if (filterFloor.value !== null) {
      csvContent += `${t('owners.voting.csvHeaders.floorFilter', 'Floor Filter')},${filterFloor.value}\n`
    }
    if (filterEntrance.value !== null) {
      csvContent += `${t('owners.voting.csvHeaders.entranceFilter', 'Entrance Filter')},${filterEntrance.value}\n`
    }

    // Add summary data for filtered results
    csvContent += '\n'
    csvContent += t('owners.voting.csvHeaders.reportSummary', 'Report Summary') + '\n'
    csvContent += `${t('owners.voting.csvHeaders.totalOwners', 'Total Voting Owners')},${sortedFilteredData.value.length}\n`

    // Calculate totals for filtered data
    const totalUnits = sortedFilteredData.value.reduce((sum, owner) => sum + owner.total_units, 0)
    const totalArea = sortedFilteredData.value.reduce((sum, owner) => sum + owner.total_area, 0)
    const totalShare = sortedFilteredData.value.reduce((sum, owner) => sum + owner.total_condo_part, 0)

    csvContent += `${t('owners.voting.csvHeaders.totalUnits', 'Total Units')},${totalUnits}\n`
    csvContent += `${t('owners.voting.csvHeaders.totalArea', 'Total Area (m²)')},${totalArea.toFixed(2)}\n`
    csvContent += `${t('owners.voting.csvHeaders.totalVotingShare', 'Total Voting Share (%)')},${(totalShare * 100).toFixed(4)}\n`
    csvContent += `${t('owners.voting.csvHeaders.totalRelativeShare', 'Total Relative Share (%)')},100.0000\n`

    // Create download link
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.setAttribute('href', url)
    link.setAttribute('download', `voting_owners_report_${new Date().toISOString().split('T')[0]}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    message.success(t('owners.voting.exportSuccess', 'CSV exported successfully'))
  } catch (err) {
    console.error('Error exporting to CSV:', err)
    message.error(t('owners.voting.exportError', 'Failed to export CSV'))
  }
}

// Summary statistics for server-filtered data
const summaryStats = computed(() => {
  if (!votingOwners.value || votingOwners.value.length === 0) return null

  const totalUnits = votingOwners.value.reduce((sum, owner) => sum + owner.total_units, 0)
  const totalArea = votingOwners.value.reduce((sum, owner) => sum + owner.total_area, 0)
  const totalShare = votingOwners.value.reduce((sum, owner) => sum + owner.total_condo_part, 0)

  return {
    totalOwners: votingOwners.value.length,
    totalUnits,
    totalArea,
    totalVotingShare: totalShare
  }
})

// Watch for associationId changes
watch(() => props.associationId, (newVal) => {
  if (newVal) {
    fetchVotingOwnersReport()
  } else {
    votingOwners.value = []
  }
}, { immediate: true })
</script>

<template>
  <div class="voting-owners-report">
    <NCard :title="t('owners.votingReport', 'Voting Owners Report')">
      <template #header-extra>
        <NSpace>
          <NButton
            type="default"
            @click="showFilters = !showFilters"
          >
            {{ showFilters ? t('common.hideFilters', 'Hide Filters') : t('common.showFilters', 'Show Filters') }}
          </NButton>
          <NButton
            type="primary"
            @click="exportToCSV"
            :disabled="!votingOwners || votingOwners.length === 0"
          >
            {{ t('owners.exportToCsv', 'Export to CSV') }}
          </NButton>
        </NSpace>
      </template>

      <!-- Filters Section -->
      <div v-if="showFilters" class="filters-section">
        <NCard size="small" :title="t('common.filters', 'Filters')">
          <NForm>
            <NGrid :cols="24" :x-gap="12" :y-gap="8">
              <NFormItemGi :span="8" :label="t('units.type', 'Unit Type')">
                <NSelect
                  v-model:value="filterUnitType"
                  :options="unitTypeOptions"
                  multiple
                  :placeholder="t('common.selectAll', 'Select All')"
                  clearable
                />
              </NFormItemGi>

              <NFormItemGi :span="8" :label="t('units.floor', 'Floor')">
                <NSelect
                  v-model:value="filterFloor"
                  :options="floorOptions"
                  multiple
                  :placeholder="t('common.selectFloors', 'Select Floors')"
                  clearable
                />
              </NFormItemGi>

              <NFormItemGi :span="8" :label="t('units.entrance', 'Entrance')">
                <NSelect
                  v-model:value="filterEntrance"
                  :options="entranceOptions"
                  multiple
                  :placeholder="t('common.selectEntrances', 'Select Entrances')"
                  clearable
                />
              </NFormItemGi>
            </NGrid>

            <div style="margin-top: 16px;">
              <NButton @click="clearFilters">
                {{ t('common.clearFilters', 'Clear Filters') }}
              </NButton>
            </div>
          </NForm>
        </NCard>
      </div>

      <!-- Controls Section -->
      <div class="report-controls">
        <NSpace align="center" justify="space-between" wrap-item>
          <div class="report-sort">
            <NInputGroup>
              <NSelect
                v-model:value="sortBy"
                :options="[
                  { label: t('owners.sortByShare', 'Sort by Voting Share'), value: 'share' },
                  { label: t('owners.sortByRelativeShare', 'Sort by Relative Share'), value: 'relative_share' },
                  { label: t('owners.sortByName', 'Sort by Name'), value: 'name' }
                ]"
                style="width: 200px"
              />
              <NSelect
                v-model:value="sortOrder"
                :options="[
                  { label: t('common.descending', 'Descending'), value: 'desc' },
                  { label: t('common.ascending', 'Ascending'), value: 'asc' }
                ]"
                style="width: 120px"
              />
            </NInputGroup>
          </div>
        </NSpace>
      </div>

      <NSpin :show="loading">
        <NAlert v-if="error" type="error" style="margin-bottom: 16px;">
          {{ error }}
          <NButton @click="fetchVotingOwnersReport">{{ t('common.retry', 'Retry') }}</NButton>
        </NAlert>

        <div v-if="votingOwners && votingOwners.length > 0" class="report-content">
          <!-- Summary Statistics -->
          <div class="summary-stats" v-if="summaryStats">
            <div class="stat-item">
              <div class="stat-label">{{ t('owners.voting.totalVotingOwners', 'Total Voting Owners') }}</div>
              <div class="stat-value">{{ summaryStats.totalOwners }}</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">{{ t('owners.voting.totalUnits', 'Total Units') }}</div>
              <div class="stat-value">{{ summaryStats.totalUnits }}</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">{{ t('owners.totalArea', 'Total Area') }}</div>
              <div class="stat-value">{{ summaryStats.totalArea.toFixed(2) }} m²</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">{{ t('owners.voting.totalVotingShare', 'Total Voting Share') }}</div>
              <div class="stat-value">{{ formatPercentage(summaryStats.totalVotingShare, 4) }}</div>
            </div>
          </div>

          <!-- Main Table -->
          <NDataTable
            :columns="columns"
            :data="sortedFilteredData"
            :pagination="{
              pageSize: 10
            }"
            :row-key="row => row.owner_id"
            :bordered="false"
          />

          <!-- Selected Owner Units -->
          <div v-if="selectedOwner" class="owner-details">
            <div class="details-header">
              <h3>{{ t('owners.voting.unitsOwnedBy', 'Units Owned by') }}: {{ selectedOwner.name }}</h3>
              <NButton size="small" @click="selectedOwner = null">{{ t('common.close', 'Close') }}</NButton>
            </div>
            <NDataTable
              :columns="unitColumns"
              :data="selectedOwner.units"
              :pagination="{
                pageSize: 5
              }"
              :row-key="row => row.unit_id"
              :bordered="false"
            />
          </div>
        </div>

        <div v-else-if="!loading && (!votingOwners || votingOwners.length === 0)" class="empty-state">
          <div style="text-align: center; padding: 32px;">
            <h3>{{ t('owners.voting.noVotingOwners', 'No Voting Owners Found') }}</h3>
            <p>{{ t('owners.voting.assignVotingRights', 'Assign voting rights to owners in the Unit Details page.') }}</p>
          </div>
        </div>
      </NSpin>
    </NCard>
  </div>
</template>

<style scoped>
.voting-owners-report {
  width: 100%;
}

.filters-section {
  margin-bottom: 16px;
}

.report-controls {
  margin-bottom: 16px;
}

.summary-stats {
  display: flex;
  justify-content: space-between;
  margin-bottom: 16px;
  padding: 12px;
  border-radius: 4px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-label {
  font-size: 0.9rem;
  color: var(--text-color-secondary);
}

.stat-value {
  font-size: 1.1rem;
  font-weight: 600;
}

.owner-details {
  margin-top: 24px;
  padding: 16px;
  border-radius: 4px;
}

.details-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.details-header h3 {
  margin: 0;
  font-size: 1.1rem;
}

.empty-state {
  padding: 24px;
  text-align: center;
  border-radius: 4px;
}
</style>
