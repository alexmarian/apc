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
import type { OwnerReportItem, Owner, OwnerUnit, OwnerCoOwner } from '@/types/api'
import { UnitType } from '@/types/api'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

// i18n
const { t } = useI18n()
const router = useRouter()

// Props
const props = defineProps<{
  associationId: number | null
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
// Multi-select unit-type filter for the owner list (OR semantics)
const unitTypeFilter = ref<string[]>([])
// Selecting a row in the per-type summary narrows the units table to that type
const selectedUnitType = ref<string | null>(null)

// Unit data is required whenever we filter by type, drill into an owner, or the
// user explicitly asks for it. Drives whether the report is fetched with units.
const unitsNeeded = computed<boolean>(() =>
  includeUnits.value || unitTypeFilter.value.length > 0 || ownerFilter.value !== null
)

// Select options for the unit-type filters
const unitTypeOptions = computed(() =>
  Object.values(UnitType).map(value => ({
    label: t(`unitTypes.${value}`, value),
    value
  }))
)

// Computed property for selected owner data
const selectedOwnerData = computed<OwnerReportItem | undefined>(()=>{
  return filteredSortedData.value?.find(item => item.owner.id === ownerFilter.value)
})

// Per-unit-type breakdown for the selected owner (all types they hold)
const unitTypeSummary = computed(() => {
  const units = selectedOwnerData.value?.units
  if (!units) return []

  const byType = new Map<string, { unit_type: string; count: number; area: number; part: number }>()
  for (const unit of units) {
    const entry = byType.get(unit.unit_type) ?? { unit_type: unit.unit_type, count: 0, area: 0, part: 0 }
    entry.count += 1
    entry.area += unit.area
    entry.part += unit.part
    byType.set(unit.unit_type, entry)
  }

  return Array.from(byType.values())
})

// Selected owner's units, narrowed to the unit type picked in the summary (if any)
const filteredOwnerUnits = computed<OwnerUnit[]>(() => {
  const units = selectedOwnerData.value?.units ?? []
  if (selectedUnitType.value === null) return units
  return units.filter(unit => unit.unit_type === selectedUnitType.value)
})

// Toggle the units-table filter when a summary row is clicked
const handleSelectUnitType = (unitType: string): void => {
  selectedUnitType.value = selectedUnitType.value === unitType ? null : unitType
}

// CSV-escape a single value
const csvEscape = (value: string | number): string => {
  if (typeof value === 'string' && (value.includes(',') || value.includes('"') || value.includes('\n'))) {
    return `"${value.replace(/"/g, '""')}"`
  }
  return String(value)
}

// Export the currently shown units of the selected owner (respects the type filter)
const exportUnitsToCSV = (): void => {
  const units = filteredOwnerUnits.value
  if (units.length === 0) {
    message.error(t('owners.noDataToExport', 'No data to export'))
    return
  }

  try {
    const headers = [
      t('units.building', 'Building'),
      t('units.cadastralNumber', 'Cadastral Number'),
      t('units.address', 'Address'),
      t('units.unit', 'Unit'),
      t('owners.csvHeaders.areaShort', 'Area (m²)'),
      t('owners.csvHeaders.partShort', 'Part (%)'),
      t('units.type', 'Type')
    ]

    const rows = units.map(unit => [
      unit.building_name,
      unit.unit_cadastral_number,
      unit.unit_address,
      unit.unit_number,
      parseFloat(unit.area.toFixed(2)),
      parseFloat((unit.part * 100).toFixed(4)),
      t(`unitTypes.${unit.unit_type}`, unit.unit_type)
    ])

    let csvContent = headers.map(csvEscape).join(',') + '\n'
    rows.forEach(row => {
      csvContent += row.map(csvEscape).join(',') + '\n'
    })

    const ownerName = selectedOwnerData.value?.owner.name?.replace(/\s+/g, '_') ?? 'owner'
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.setAttribute('href', url)
    link.setAttribute('download', `${ownerName}_units_${new Date().toISOString().split('T')[0]}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    message.success(t('owners.exportSuccess', 'CSV exported successfully'))
  } catch (err) {
    console.error('Error exporting units to CSV:', err)
    message.error(t('owners.exportError', 'Failed to export CSV'))
  }
}

// Column definitions for the data table
const columns = computed<DataTableColumns<OwnerReportItem>>(() => {
  const cols: DataTableColumns<OwnerReportItem> = [
    {
      title: t('owners.name', 'Owner'),
      key: 'owner.name'
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
      render: (row: OwnerReportItem) => `${row.statistics.total_area.toFixed(2)} m²`
    },
    {
      title: t('owners.totalPart', 'Condo Part'),
      key: 'statistics.total_condo_part',
      render: (row: OwnerReportItem) => formatPercentage(row.statistics.total_condo_part, 4)
    },
    {
      title: t('units.title', 'Units'),
      key: 'statistics.total_units'
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

type UnitTypeSummaryRow = { unit_type: string; count: number; area: number; part: number }

// Make summary rows clickable to filter the units table by that type
const summaryRowProps = (row: UnitTypeSummaryRow) => ({
  style: 'cursor: pointer',
  onClick: () => handleSelectUnitType(row.unit_type)
})
const summaryRowClass = (row: UnitTypeSummaryRow) =>
  row.unit_type === selectedUnitType.value ? 'selected-type-row' : ''

// Columns for the selected owner's per-unit-type breakdown header
const unitTypeSummaryColumns = computed<DataTableColumns<UnitTypeSummaryRow>>(() => [
  {
    title: t('units.type', 'Type'),
    key: 'unit_type',
    sorter: (a, b) => t(`unitTypes.${a.unit_type}`, a.unit_type).localeCompare(t(`unitTypes.${b.unit_type}`, b.unit_type)),
    render: (row) => t(`unitTypes.${row.unit_type}`, row.unit_type)
  },
  {
    title: t('owners.unitCount', 'Count'),
    key: 'count',
    sorter: (a, b) => a.count - b.count
  },
  {
    title: t('owners.surface', 'Surface'),
    key: 'area',
    sorter: (a, b) => a.area - b.area,
    render: (row) => `${row.area.toFixed(2)} m²`
  },
  {
    title: t('units.part', 'Part'),
    key: 'part',
    sorter: (a, b) => a.part - b.part,
    render: (row) => formatPercentage(row.part, 4)
  }
])

// TOTAL row for the breakdown table, taken from the owner's official statistics
const unitTypeSummaryRow = () => {
  const stats = selectedOwnerData.value?.statistics
  return {
    unit_type: { value: h(NText, { strong: true }, { default: () => t('common.total', 'Total') }) },
    count: { value: h(NText, { strong: true }, { default: () => stats?.total_units ?? 0 }) },
    area: { value: h(NText, { strong: true }, { default: () => `${(stats?.total_area ?? 0).toFixed(2)} m²` }) },
    part: { value: h(NText, { strong: true }, { default: () => formatPercentage(stats?.total_condo_part ?? 0, 4) }) }
  }
}

// Deep-link to the unit edit flow (opens the edit modal on the Units page)
const handleEditUnit = (unit: OwnerUnit): void => {
  router.push({
    path: '/units',
    query: {
      buildingId: unit.building_id.toString(),
      editUnitId: unit.unit_id.toString()
    }
  })
}

// Columns for the selected owner's units detail table (sortable + edit action)
const ownerUnitsColumns = computed<DataTableColumns<OwnerUnit>>(() => [
  {
    title: t('units.building', 'Building'),
    key: 'building_name',
    sorter: (a, b) => a.building_name.localeCompare(b.building_name)
  },
  { title: t('units.cadastralNumber', 'Cadastral Number'), key: 'unit_cadastral_number' },
  { title: t('units.address', 'Address'), key: 'unit_address' },
  {
    title: t('units.unit', 'Unit'),
    key: 'unit_number',
    sorter: (a, b) => a.unit_number.localeCompare(b.unit_number, undefined, { numeric: true })
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
    sorter: (a, b) => a.part - b.part,
    render: (row) => formatPercentage(row.part, 4)
  },
  {
    title: t('units.type', 'Type'),
    key: 'unit_type',
    sorter: (a, b) => t(`unitTypes.${a.unit_type}`, a.unit_type).localeCompare(t(`unitTypes.${b.unit_type}`, b.unit_type)),
    render: (row) => t(`unitTypes.${row.unit_type}`, row.unit_type)
  },
  {
    title: t('common.actions', 'Actions'),
    key: 'actions',
    render: (row) => h(
      NButton,
      { size: 'small', onClick: () => handleEditUnit(row) },
      { default: () => t('common.edit', 'Edit') }
    )
  }
])

// Fetch owners report
const fetchOwnersReport = async (forceUnits = false): Promise<void> => {
  if (!props.associationId) return

  try {
    loading.value = true
    error.value = null

    const response = await ownerApi.getOwnerReport(
      props.associationId,
      forceUnits || unitsNeeded.value,
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
  // Reset the per-type units filter whenever the selected owner changes
  selectedUnitType.value = null
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

// Aggregate an owner's units into per-type { area, part } keyed by unit type
const unitTypeBreakdown = (item: OwnerReportItem): Record<string, { area: number; part: number }> => {
  const breakdown: Record<string, { area: number; part: number }> = {}
  for (const type of Object.values(UnitType)) {
    breakdown[type] = { area: 0, part: 0 }
  }
  for (const unit of item.units ?? []) {
    if (!breakdown[unit.unit_type]) breakdown[unit.unit_type] = { area: 0, part: 0 }
    breakdown[unit.unit_type].area += unit.area
    breakdown[unit.unit_type].part += unit.part
  }
  return breakdown
}

// Export to CSV
const exportToCSV = async (): Promise<void> => {
  if (!ownersData.value || filteredSortedData.value.length === 0) {
    message.error(t('owners.noDataToExport', 'No data to export'))
    return
  }

  // The per-type breakdown columns need unit data; load it if not already present.
  if (!unitsNeeded.value && !ownersData.value.some(item => item.units !== undefined)) {
    await fetchOwnersReport(true)
  }

  try {
    const unitTypes = Object.values(UnitType)

    // Create CSV headers
    const headers: string[] = [
      t('owners.csvHeaders.id', 'Owner ID'),
      t('owners.csvHeaders.name', 'Owner Name'),
      t('owners.csvHeaders.identification', 'Identification Number'),
      t('owners.csvHeaders.phone', 'Contact Phone'),
      t('owners.csvHeaders.email', 'Contact Email'),
      t('owners.csvHeaders.units', 'Total Units'),
      t('owners.csvHeaders.area', 'Total Area (m²)'),
      t('owners.csvHeaders.part', 'Total Condo Part (%)'),
      // Per-unit-type surface columns followed by per-unit-type part columns
      ...unitTypes.map(type => `${t(`unitTypes.${type}`, type)} ${t('owners.csvHeaders.areaShort', 'Area (m²)')}`),
      ...unitTypes.map(type => `${t(`unitTypes.${type}`, type)} ${t('owners.csvHeaders.partShort', 'Part (%)')}`)
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
      const breakdown = unitTypeBreakdown(item)
      const row: (string | number)[] = [
        item.owner.id,
        item.owner.name,
        item.owner.identification_number,
        item.owner.contact_phone,
        item.owner.contact_email,
        item.statistics.total_units,
        parseFloat(item.statistics.total_area.toFixed(2)),
        parseFloat((item.statistics.total_condo_part * 100).toFixed(4)),
        // Per-type surfaces, then per-type parts — same order as the headers
        ...unitTypes.map(type => parseFloat(breakdown[type].area.toFixed(2))),
        ...unitTypes.map(type => parseFloat((breakdown[type].part * 100).toFixed(4)))
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

  // Apply unit-type filter: keep owners holding at least one unit of any selected
  // type. Rows are narrowed but each owner's stat columns keep their full totals.
  if (unitTypeFilter.value.length > 0) {
    data = data.filter(item =>
      item.units?.some(unit => unitTypeFilter.value.includes(unit.unit_type))
    )
  }

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

// Refetch only when the data we need from the server changes (units presence or
// co-owners). Owner/type filtering is applied client-side, but unitsNeeded flips
// to true the first time a type filter or owner drill-down requires unit data.
watch([unitsNeeded, includeCoOwners], () => {
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
              <NSelect
                v-model:value="unitTypeFilter"
                :options="unitTypeOptions"
                :placeholder="t('owners.filterByUnitType', 'Filter by unit type')"
                multiple
                clearable
                style="min-width: 220px"
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
          <NButton @click="() => fetchOwnersReport()">{{ t('common.retry', 'Retry') }}</NButton>
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

          <!-- Per-unit-type summary header (visible when an owner is selected) -->
          <div v-if="ownerFilter !== null && unitTypeSummary.length > 0" class="details-section">
            <h3>{{ t('owners.holdingsByType', 'Holdings by unit type') }}</h3>
            <p class="details-hint">{{ t('owners.selectTypeHint', 'Click a row to show only that unit type below.') }}</p>
            <NDataTable
              :columns="unitTypeSummaryColumns"
              :data="unitTypeSummary"
              :summary="unitTypeSummaryRow"
              :row-key="(row: any) => row.unit_type"
              :row-props="summaryRowProps"
              :row-class-name="summaryRowClass"
              :bordered="false"
            />
          </div>

          <!-- Units Details (visible when an owner is selected and unit data is loaded) -->
          <div v-if="ownerFilter !== null && selectedOwnerData?.units && selectedOwnerData.units.length > 0">
            <div class="details-section">
              <div class="details-header">
                <div class="details-title">
                  <h3>{{ t('owners.unitsDetails', "Owner's Units") }}</h3>
                  <NTag
                    v-if="selectedUnitType !== null"
                    type="info"
                    closable
                    @close="selectedUnitType = null"
                  >
                    {{ t(`unitTypes.${selectedUnitType}`, selectedUnitType) }}
                  </NTag>
                </div>
                <NButton
                  size="small"
                  type="primary"
                  @click="exportUnitsToCSV"
                  :disabled="filteredOwnerUnits.length === 0"
                >
                  {{ t('owners.exportToCsv', 'Export to CSV') }}
                </NButton>
              </div>
              <NDataTable
                :columns="ownerUnitsColumns"
                :data="filteredOwnerUnits"
                :pagination="{
                  pageSize: 5
                }"
                :row-key="(row: OwnerUnit) => row.unit_id"
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

.report-filters {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
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

.details-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  margin-bottom: 12px;
}

.details-header h3 {
  margin-bottom: 0;
}

.details-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.details-hint {
  margin: 0 0 8px;
  font-size: 0.85rem;
  opacity: 0.7;
}

:deep(.selected-type-row td) {
  background-color: rgba(32, 128, 240, 0.12);
  font-weight: 600;
}
</style>
