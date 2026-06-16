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
  NInputGroup,
  NInput,
  NSelect,
  NText
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { ownerApi } from '@/services/api'
import { formatPercentage } from '@/utils/formatters'
import type { OwnerReportItem } from '@/types/api'
import { UnitType } from '@/types/api'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

// i18n
const { t } = useI18n()
const route = useRoute()
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

// State — initialised from URL query so filters survive navigation
const loading = ref<boolean>(false)
const error = ref<string | null>(null)
const ownersData = ref<OwnerReportItem[] | null>(null)
const message = useMessage()
const searchQuery = ref<string>((route.query.search as string) ?? '')
const sortBy = ref<'name' | 'part'>((route.query.sortBy as 'name' | 'part') ?? 'part')
const sortOrder = ref<'asc' | 'desc'>((route.query.sortOrder as 'asc' | 'desc') ?? 'desc')
const unitTypeFilter = ref<string[]>(
  route.query.unitTypes ? (route.query.unitTypes as string).split(',').filter(Boolean) : []
)
const currentPage = ref<number>(route.query.page ? parseInt(route.query.page as string) : 1)
const pagination = computed(() => ({
  pageSize: 10,
  page: currentPage.value,
  onUpdatePage: (p: number) => { currentPage.value = p }
}))

// Unit data is required whenever filtering by type (client-side narrowing needs it).
const unitsNeeded = computed<boolean>(() => unitTypeFilter.value.length > 0)

// Select options for the unit-type filters
const unitTypeOptions = computed(() =>
  Object.values(UnitType).map(value => ({
    label: t(`unitTypes.${value}`, value),
    value
  }))
)

// CSV-escape a single value
const csvEscape = (value: string | number): string => {
  if (typeof value === 'string' && (value.includes(',') || value.includes('"') || value.includes('\n'))) {
    return `"${value.replace(/"/g, '""')}"`
  }
  return String(value)
}

// When a unit-type filter is active, compute area/part/count from filtered units.
// Falls back to server-side statistics when no filter is applied.
const filteredStatsForRow = (row: OwnerReportItem): { area: number; part: number; count: number } => {
  if (unitTypeFilter.value.length === 0 || !row.units) {
    return {
      area: row.statistics.total_area,
      part: row.statistics.total_condo_part,
      count: row.statistics.total_units
    }
  }
  const matching = row.units.filter(u => unitTypeFilter.value.includes(u.unit_type))
  return {
    area: matching.reduce((s, u) => s + u.area, 0),
    part: matching.reduce((s, u) => s + u.part, 0),
    count: matching.length
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
      render: (row: OwnerReportItem) => `${filteredStatsForRow(row).area.toFixed(2)} m²`
    },
    {
      title: t('owners.totalPart', 'Condo Part'),
      key: 'statistics.total_condo_part',
      render: (row: OwnerReportItem) => formatPercentage(filteredStatsForRow(row).part, 4)
    },
    {
      title: t('units.title', 'Units'),
      key: 'statistics.total_units',
      render: (row: OwnerReportItem) => String(filteredStatsForRow(row).count)
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
const fetchOwnersReport = async (forceUnits = false): Promise<void> => {
  if (!props.associationId) return

  try {
    loading.value = true
    error.value = null

    const response = await ownerApi.getOwnerReport(
      props.associationId,
      forceUnits || unitsNeeded.value,
      false
    )

    ownersData.value = response.data
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('owners.loadError', 'Failed to load owners report')
    console.error('Error fetching owners report:', err)
  } finally {
    loading.value = false
  }
}

// Handle viewing owner details — navigate to dedicated detail page
const handleViewOwnerDetails = (ownerId: number): void => {
  router.push({
    path: `/owners/${ownerId}`,
    query: { from: route.fullPath }
  })
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

  // CSV needs units + co-owners; fetch them if not already loaded.
  if (!ownersData.value.some(item => item.units !== undefined)) {
    const response = await ownerApi.getOwnerReport(props.associationId!, true, true)
    ownersData.value = response.data
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

    headers.push(t('owners.csvHeaders.coOwners', 'Co-Owners'))
    headers.push(t('owners.csvHeaders.unitsList', 'Units'))

    // Create CSV rows
    const rows: (string | number)[][] = filteredSortedData.value.map(item => {
      const breakdown = unitTypeBreakdown(item)
      const stats = filteredStatsForRow(item)
      const row: (string | number)[] = [
        item.owner.id,
        item.owner.name,
        item.owner.identification_number,
        item.owner.contact_phone,
        item.owner.contact_email,
        stats.count,
        parseFloat(stats.area.toFixed(2)),
        parseFloat((stats.part * 100).toFixed(4)),
        // Per-type surfaces, then per-type parts — same order as the headers
        ...unitTypes.map(type => parseFloat(breakdown[type].area.toFixed(2))),
        ...unitTypes.map(type => parseFloat((breakdown[type].part * 100).toFixed(4)))
      ]

      row.push(item.co_owners && item.co_owners.length > 0
        ? item.co_owners.map(co => co.name).join(', ')
        : '')
      row.push(item.units && item.units.length > 0
        ? item.units.map(unit => `${unit.building_name} - ${unit.unit_number} (${unit.area} m²)`).join('; ')
        : '')

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

  // Apply unit-type filter: keep owners holding at least one unit of the selected types.
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
      comparison = filteredStatsForRow(a).part - filteredStatsForRow(b).part
    }

    return sortOrder.value === 'asc' ? comparison : -comparison
  })

  return data
})

// Sync filter/page state to URL so navigating away and back restores it
watch([searchQuery, sortBy, sortOrder, unitTypeFilter, currentPage], () => {
  router.replace({
    query: {
      ...route.query,
      search: searchQuery.value || undefined,
      sortBy: sortBy.value !== 'part' ? sortBy.value : undefined,
      sortOrder: sortOrder.value !== 'desc' ? sortOrder.value : undefined,
      unitTypes: unitTypeFilter.value.length > 0 ? unitTypeFilter.value.join(',') : undefined,
      page: currentPage.value > 1 ? currentPage.value.toString() : undefined,
    }
  })
})

// Refetch when the association changes or the unit-type filter crosses the
// needs-units boundary (empty→some or some→empty).
watch(() => props.associationId, (id) => {
  if (id) fetchOwnersReport()
}, { immediate: true })

watch(unitsNeeded, () => {
  if (props.associationId) fetchOwnersReport()
})
</script>

<template>
  <div class="owners-report">
    <NCard>

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
            <NSpace align="center">
              <NText depth="3" style="white-space: nowrap;">
                {{ filteredSortedData.length }} {{ t('owners.totalOwners', 'owners') }}
              </NText>
              <NButton
                type="primary"
                @click="exportToCSV"
                :disabled="!ownersData || filteredSortedData.length === 0"
              >
                {{ t('owners.exportToCsv', 'Export to CSV') }}
              </NButton>
            </NSpace>
          </div>
        </NSpace>
      </div>

      <NSpin :show="loading">
        <NAlert v-if="error" type="error" style="margin-bottom: 16px;">
          {{ error }}
          <NButton @click="() => fetchOwnersReport()">{{ t('common.retry', 'Retry') }}</NButton>
        </NAlert>

        <div v-if="ownersData && filteredSortedData.length > 0" class="owners-table">
          <NDataTable
            :columns="columns"
            :data="filteredSortedData"
            :pagination="pagination"
            :row-key="(row: OwnerReportItem) => row.owner.id"
            :bordered="false"
          />
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



</style>
