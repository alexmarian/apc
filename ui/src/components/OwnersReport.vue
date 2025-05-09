<script setup lang="ts">
import { ref, onMounted, computed, watch, h } from 'vue'
import {
  NCard,
  NDataTable,
  NSpace,
  NSpin,
  NAlert,
  NButton,
  useMessage,
  NSwitch,
  NInputGroup,
  NInput,
  NSelect,
  NRadioGroup,
  NRadio,
  NText
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { ownerApi } from '@/services/api'
import { formatPercentage } from '@/utils/formatters'
import type { OwnerReportItem, Owner, OwnerUnit, OwnerCoOwner, VotingUnit } from '@/types/api'
import { useI18n } from 'vue-i18n'

// i18n
const { t } = useI18n()

// Props
const props = defineProps<{
  associationId: number
  buildingId?: number | null
}>()

// Define emits
const emit = defineEmits<{
  (e: 'edit-owner', ownerId: number): void
}>()

// State
const loading = ref<boolean>(false)
const error = ref<string | null>(null)
const ownersData = ref<OwnerReportItem[] | null>(null)
const message = useMessage()
const includeUnits = ref<boolean>(false)
const includeCoOwners = ref<boolean>(false)
const searchQuery = ref<string>('')
const sortBy = ref<'name' | 'part'>('part')
const sortOrder = ref<'asc' | 'desc'>('desc')
const ownerFilter = ref<number | null>(null)

// Computed property for selected owner data
const selectedOwnerData = computed<OwnerReportItem | undefined>(()=>{
  return filteredSortedData.value?.find(item => item.owner.id === ownerFilter.value)
})

// Column definitions for the data table
const columns = computed<DataTableColumns<OwnerReportItem>>(() => {
  const cols: DataTableColumns<OwnerReportItem> = [
    {
      title: t('owners.name', 'Owner'),
      key: 'owner.name',
      sorter: (a: OwnerReportItem, b: OwnerReportItem) => a.owner.name.localeCompare(b.owner.name)
    },
    {
      title: t('owners.identification', 'Identification'),
      key: 'owner.identification_number'
    },
    {
      title: t('owners.contactPhone', 'Contact Phone'),
      key: 'owner.contact_phone'
    },
    {
      title: t('owners.contactEmail', 'Contact Email'),
      key: 'owner.contact_email'
    },
    {
      title: t('owners.totalArea', 'Total Area'),
      key: 'statistics.total_area',
      sorter: (a: OwnerReportItem, b: OwnerReportItem) => a.statistics.total_area - b.statistics.total_area,
      render: (row: OwnerReportItem) => `${row.statistics.total_area.toFixed(2)} m²`
    },
    {
      title: t('owners.totalPart', 'Condo Part'),
      key: 'statistics.total_condo_part',
      sorter: (a: OwnerReportItem, b: OwnerReportItem) => a.statistics.total_condo_part - b.statistics.total_condo_part,
      render: (row: OwnerReportItem) => formatPercentage(row.statistics.total_condo_part, 4)
    },
    {
      title: t('units.title', 'Units'),
      key: 'statistics.total_units',
      sorter: (a: OwnerReportItem, b: OwnerReportItem) => a.statistics.total_units - b.statistics.total_units
    },
    {
      title: t('common.actions', 'Actions'),
      key: 'actions',
      render: (row: OwnerReportItem) => {
        return h(
          NSpace,
          { justify: 'center', align: 'center' },
          {
            default: () => [
              h(
                NButton,
                {
                  size: 'small',
                  onClick: () => handleViewOwnerDetails(row.owner.id)
                },
                { default: () => t('common.details', 'Details') }
              ),
              h(
                NButton,
                {
                  size: 'small',
                  type: 'primary',
                  onClick: () => handleEditOwner(row.owner.id)
                },
                { default: () => t('common.edit', 'Edit') }
              )
            ]
          }
        )
      }
    }
  ]

  return cols
})

// Fetch owners report
const fetchOwnersReport = async (): Promise<void> => {
  if (!props.associationId) return

  try {
    loading.value = true
    error.value = null

    const response = await ownerApi.getOwnerReport(
      props.associationId,
      includeUnits.value,
      includeCoOwners.value
    )

    ownersData.value = response.data
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('owners.loadError', 'Failed to load owners report')
    console.error('Error fetching owners report:', err)
  } finally {
    loading.value = false
  }
}

// Handle viewing owner details
const handleViewOwnerDetails = (ownerId: number): void => {
  if (ownerId === ownerFilter.value) {
    // If already filtered by this owner, clear filter
    ownerFilter.value = null
  } else {
    // Filter to show only this owner
    ownerFilter.value = ownerId
  }
}

// Handle edit owner - emit event to parent
const handleEditOwner = (ownerId: number): void => {
  emit('edit-owner', ownerId)
}

// Export to CSV
const exportToCSV = (): void => {
  if (!ownersData.value || filteredSortedData.value.length === 0) {
    message.error(t('owners.noDataToExport', 'No data to export'))
    return
  }

  try {
    // Create CSV headers
    const headers: string[] = [
      t('owners.csvHeaders.id', 'Owner ID'),
      t('owners.csvHeaders.name', 'Owner Name'),
      t('owners.csvHeaders.identification', 'Identification Number'),
      t('owners.csvHeaders.phone', 'Contact Phone'),
      t('owners.csvHeaders.email', 'Contact Email'),
      t('owners.csvHeaders.units', 'Total Units'),
      t('owners.csvHeaders.area', 'Total Area (m²)'),
      t('owners.csvHeaders.part', 'Total Condo Part (%)')
    ]

    // Add co-owner headers if included
    if (includeCoOwners.value) {
      headers.push(t('owners.csvHeaders.coOwners', 'Co-Owners'))
    }

    // Add unit headers if included
    if (includeUnits.value) {
      headers.push(t('owners.csvHeaders.unitsList', 'Units'))
    }

    // Create CSV rows
    const rows: (string | number)[][] = filteredSortedData.value.map(item => {
      const row: (string | number)[] = [
        item.owner.id,
        item.owner.name,
        item.owner.identification_number,
        item.owner.contact_phone,
        item.owner.contact_email,
        item.statistics.total_units,
        parseFloat(item.statistics.total_area.toFixed(2)),
        parseFloat((item.statistics.total_condo_part * 100).toFixed(4))
      ]

      // Add co-owners if included
      if (includeCoOwners.value) {
        if (item.co_owners && item.co_owners.length > 0) {
          row.push(item.co_owners.map(co => co.name).join(', '))
        } else {
          row.push('')
        }
      }

      // Add units if included
      if (includeUnits.value) {
        if (item.units && item.units.length > 0) {
          row.push(item.units.map(unit =>
            `${unit.building_name} - ${unit.unit_number} (${unit.area} m²)`
          ).join('; '))
        } else {
          row.push('')
        }
      }

      return row
    })

    // Create CSV content
    let csvContent = headers.join(',') + '\n'
    rows.forEach(row => {
      // Properly escape values that might contain commas
      const escapedRow = row.map(value => {
        if (typeof value === 'string' && (value.includes(',') || value.includes('"'))) {
          return `"${value.replace(/"/g, '""')}"`
        }
        return value
      })
      csvContent += escapedRow.join(',') + '\n'
    })

    // Add summary data
    csvContent += '\n'
    csvContent += t('owners.csvHeaders.reportSummary', 'Report Summary') + '\n'
    csvContent += `${t('owners.csvHeaders.totalOwners', 'Total Owners')},${filteredSortedData.value.length}\n`

    const totalArea = filteredSortedData.value.reduce((sum, item) =>
      sum + item.statistics.total_area, 0
    )
    csvContent += `${t('owners.csvHeaders.totalArea', 'Total Area (m²)')},${totalArea.toFixed(2)}\n`

    const totalPart = filteredSortedData.value.reduce((sum, item) =>
      sum + item.statistics.total_condo_part, 0
    )
    csvContent += `${t('owners.csvHeaders.totalPart', 'Total Condo Part (%)')},${(totalPart * 100).toFixed(4)}\n`

    // Create a CSV blob and download it
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.setAttribute('href', url)
    link.setAttribute('download', `owners_report_${new Date().toISOString().split('T')[0]}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    message.success(t('owners.exportSuccess', 'CSV exported successfully'))
  } catch (err) {
    console.error('Error exporting to CSV:', err)
    message.error(t('owners.exportError', 'Failed to export CSV'))
  }
}

// Filter and sort data
const filteredSortedData = computed<OwnerReportItem[]>(() => {
  if (!ownersData.value) return []

  let data = [...ownersData.value]

  // Apply search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    data = data.filter(item =>
      item.owner.name.toLowerCase().includes(query) ||
      item.owner.identification_number.toLowerCase().includes(query) ||
      item.owner.contact_phone.toLowerCase().includes(query) ||
      item.owner.contact_email.toLowerCase().includes(query)
    )
  }

  // Apply sorting
  data.sort((a, b) => {
    let comparison = 0

    if (sortBy.value === 'name') {
      comparison = a.owner.name.localeCompare(b.owner.name)
    } else if (sortBy.value === 'part') {
      comparison = a.statistics.total_condo_part - b.statistics.total_condo_part
    }

    return sortOrder.value === 'asc' ? comparison : -comparison
  })

  return data
})

// Watch for changes in filters and refresh data
watch([includeUnits, includeCoOwners, ownerFilter], () => {
  if (props.associationId) {
    fetchOwnersReport()
  }
})

// Initialize component
onMounted(() => {
  if (props.associationId) {
    fetchOwnersReport()
  }
})
</script>

<template>
  <div class="owners-report">
    <NCard :title="t('owners.report', 'Owners Report')">
      <template #header-extra>
        <NSpace>
          <NButton
            type="primary"
            @click="exportToCSV"
            :disabled="!ownersData || filteredSortedData.length === 0"
          >
            {{ t('owners.exportToCsv', 'Export to CSV') }}
          </NButton>
        </NSpace>
      </template>

      <div class="report-controls">
        <NSpace align="center" justify="space-between" wrap-item>
          <div class="report-filters">
            <NInputGroup>
              <NInput
                v-model:value="searchQuery"
                :placeholder="t('owners.searchPlaceholder', 'Search owners...')"
                clearable
              />
              <NSelect
                v-model:value="sortBy"
                :options="[
                  { label: t('owners.sortByPart', 'Sort by Ownership Part'), value: 'part' },
                  { label: t('owners.sortByName', 'Sort by Name'), value: 'name' }
                ]"
                style="width: 180px"
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

          <div class="report-options">
            <NSpace>
              <div class="option">
                <span>{{ t('owners.includeUnits', 'Include Units') }}: </span>
                <NSwitch v-model:value="includeUnits" />
              </div>
              <div class="option">
                <span>{{ t('owners.includeCoOwners', 'Include Co-Owners') }}: </span>
                <NSwitch v-model:value="includeCoOwners" />
              </div>
            </NSpace>
          </div>
        </NSpace>
      </div>

      <div v-if="ownerFilter !== null" class="filter-banner">
        <div>
          {{ t('owners.viewingDetails', 'Viewing details for selected owner') }}
        </div>
        <NButton size="small" @click="ownerFilter = null">{{ t('common.clear', 'Clear Filter') }}</NButton>
      </div>

      <NSpin :show="loading">
        <NAlert v-if="error" type="error" style="margin-bottom: 16px;">
          {{ error }}
          <NButton @click="fetchOwnersReport">{{ t('common.retry', 'Retry') }}</NButton>
        </NAlert>

        <div v-if="ownersData && filteredSortedData.length > 0" class="owners-table">
          <div class="summary-stats">
            <div class="stat-item">
              <div class="stat-label">{{ t('owners.totalOwners', 'Total Owners') }}:</div>
              <div class="stat-value">{{ filteredSortedData.length }}</div>
            </div>
          </div>

          <NDataTable
            :columns="columns"
            :data="filteredSortedData"
            :pagination="{
              pageSize: 10
            }"
            :row-key="(row: OwnerReportItem) => row.owner.id"
            :bordered="false"
          />

          <!-- Units Details (visible when owner is filtered and includeUnits is true) -->
          <div v-if="ownerFilter !== null && includeUnits && filteredSortedData.length > 0 && selectedOwnerData?.units">
            <div class="details-section">
              <h3>{{ t('owners.unitsDetails', "Owner's Units") }}</h3>
              <NDataTable
                :columns="[
                  { title: t('units.building', 'Building'), key: 'building_name' },
                  { title: t('units.unit', 'Unit'), key: 'unit_number' },
                  { title: t('units.area', 'Area'), key: 'area', render: (row: OwnerUnit) => `${row.area.toFixed(2)} m²` },
                  { title: t('units.part', 'Part'), key: 'part', render: (row: OwnerUnit) => formatPercentage(row.part, 4) },
                  { title: t('units.type', 'Type'), key:'unit_type', render: (row: VotingUnit) => t(`unitTypes.${row.unit_type}`, row.unit_type) }
                ]"
                :data="selectedOwnerData.units"
                :pagination="{
                  pageSize: 5
                }"
                :bordered="false"
              />
            </div>
          </div>

          <!-- Co-Owners Details (visible when owner is filtered and includeCoOwners is true) -->
          <div v-if="ownerFilter !== null && includeCoOwners && filteredSortedData.length > 0 && selectedOwnerData?.co_owners && selectedOwnerData.co_owners.length > 0">
            <div class="details-section">
              <h3>{{ t('owners.coOwners', 'Co-Owners') }}</h3>
              <NDataTable
                :columns="[
                  { title: t('owners.name', 'Name'), key: 'name' },
                  { title: t('owners.identification', 'Identification'), key: 'identification_number' },
                  { title: t('owners.contactPhone', 'Contact Phone'), key: 'contact_phone' },
                  { title: t('owners.contactEmail', 'Contact Email'), key: 'contact_email' },
                  { title: t('owners.sharedUnits', 'Shared Units'), key: 'shared_unit_nums', render: (row: OwnerCoOwner) => row.shared_unit_nums.join(', ') }
                ]"
                :data="selectedOwnerData.co_owners"
                :pagination="{
                  pageSize: 5
                }"
                :bordered="false"
              />
            </div>
          </div>
        </div>

        <NEmpty v-else-if="!loading && (!ownersData || filteredSortedData.length === 0)" :description="t('owners.noOwnersFound', 'No owners found')">
          <template #extra>
            <p>{{ t('owners.tryChangingFilters', 'Try changing your search filters or selecting a different association.') }}</p>
          </template>
        </NEmpty>
      </NSpin>
    </NCard>
  </div>
</template>

<style scoped>
.owners-report {
  width: 100%;
}

.report-controls {
  margin-bottom: 16px;
}

.report-options {
  display: flex;
  gap: 16px;
}

.option {
  display: flex;
  align-items: center;
  gap: 8px;
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
  color: var(--text-color);
  opacity: 0.8;
}

.stat-value {
  font-size: 1.1rem;
  font-weight: 600;
}

.filter-banner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  margin-bottom: 16px;
  border-radius: 4px;
  border-left: 4px solid #2080f0;
}

.details-section {
  margin-top: 24px;
  padding-top: 16px;
}

.details-section h3 {
  margin-bottom: 12px;
  font-size: 1.1rem;
  font-weight: 600;
}
</style>
