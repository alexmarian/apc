<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  NCard,
  NSpace,
  NSpin,
  NAlert,
  NButton,
  NDatePicker,
  NSelect,
  NRadioGroup,
  NRadio,
  NDataTable,
  NDivider,
  NEmpty,
  NFlex,
  NText,
  NSkeleton,
  useMessage
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useRouter, useRoute } from 'vue-router'
import { expenseApi, categoryApi } from '@/services/api'
import AssociationSelector from '@/components/AssociationSelector.vue'
import BuildingSelector from '@/components/BuildingSelector.vue'
import CategorySelector from '@/components/CategorySelector.vue'
import { formatCurrency } from '@/utils/formatters'
import { arrayToCsv, downloadCsv } from '@/utils/csvUtils'
import { getDefaultLastMonthRange, formatDateRange } from '@/utils/expenseUtils'
import type { ExpenseDistributionResponse, Category, UnitDistribution } from '@/types/api'
import { UnitType } from '@/types/api'
import { useI18n } from 'vue-i18n'
// Message provider
const message = useMessage()
const router = useRouter()
const route = useRoute()

// Props
const props = defineProps<{
  associationId: number | null;
  buildingId: number | null;
}>()

// Emits
const emit = defineEmits<{
  (e: 'update:associationId', id: number): void;
  (e: 'update:buildingId', id: number): void;
}>()
const { t } = useI18n()
// State
const loading = ref<boolean>(false)
const metadataLoading = ref<boolean>(false)
const error = ref<string | null>(null)
const dateRange = ref<[number, number] | null>(getDefaultLastMonthRange())
const selectedCategoryId = ref<number | null>(null)
const selectedCategoryType = ref<string | null>(null)
const selectedCategoryFamily = ref<string | null>(null)
const distributionMethod = ref<'area' | 'count' | 'equal'>('area')
const unitType = ref<string | null>(null)
const distributionData = ref<ExpenseDistributionResponse | null>(null)
const categories = ref<Category[] | null>(null)
const categoryOptions = ref<{ label: string; value: string }[]>([])
const categoryTypeOptions = ref<{ label: string; value: string }[]>([])
const categoryFamilyOptions = ref<{ label: string; value: string }[]>([])
const initialLoadComplete = ref<boolean>(false)

const unitTypeOptions = computed(() =>
  Object.entries(UnitType).map(([key, value]) => ({
    label: t(`unitTypes.${value}`),
    value: value as string
  }))
)


// Computed properties
const associationId = computed<number | null>({
  get: () => props.associationId || null,
  set: (value) => emit('update:associationId', value as number)
})

const buildingId = computed<number | null>({
  get: () => props.buildingId || null,
  set: (value) => emit('update:buildingId', value as number)
})

const startDate = computed<Date | null>(() => {
  return dateRange.value ? new Date(dateRange.value[0]) : null
})

const endDate = computed<Date | null>(() => {
  return dateRange.value ? new Date(dateRange.value[1]) : null
})

const formattedDateRange = computed<string>(() => {
  if (!dateRange.value) return t('expenses.allTime', 'All time')
  return formatDateRange(dateRange.value[0], dateRange.value[1])
})

const totalExpenses = computed<number>(() => {
  return distributionData.value?.total_expenses || 0
})

const totalUnits = computed<number>(() => {
  return distributionData.value?.total_units || 0
})

const distributionColumns = computed<DataTableColumns<any>>(() => {
  const columns: DataTableColumns<any> = [
    {
      title: t('units.unit'),
      key: 'unit_info',
      render: (row) => row.unit_number
    },
    {
      title: t('units.type'),
      key: 'unit_type',
      render: (row) => row.unit_type
    },
    {
      title: t('units.area'),
      key: 'area',
      render: (row) => `${row.area} m²`
    },
    {
      title: t('units.part'),
      key: 'distribution_factor',
      render: (row) => (row.distribution_factor * 100).toFixed(2) + '%'
    },
    {
      title: t('distribution.total_share'),
      key: 'total_share',
      render: (row) => formatCurrency(row.total_share),
      sorter: (a, b) => a.total_share - b.total_share
    }
  ]

  if (distributionData.value?.category_totals) {
    const categoryKeys = Object.keys(distributionData.value.category_totals)

    for (const category of categoryKeys) {
      columns.push({
        title: t(`categories.names.${category}`),
        key: category,
        render: (row) => formatCurrency(row.expenses_share[category] || 0)
      })
    }
  }

  return columns
})

// Methods
const fetchDistributionReport = async (): Promise<void> => {
  if (!associationId.value) {
    error.value = t('association.noAssociationsMessage', 'Please select an association')
    return
  }

  try {
    loading.value = true
    error.value = null

    const startDateStr = startDate.value?.toISOString().split('T')[0]
    const endDateStr = endDate.value?.toISOString().split('T')[0]

    const response = await expenseApi.getExpenseDistribution(associationId.value, {
      start_date: startDateStr,
      end_date: endDateStr,
      category_id: selectedCategoryId.value,
      category_type: selectedCategoryType.value,
      category_family: selectedCategoryFamily.value,
      distribution_method: distributionMethod.value,
      unit_type: unitType.value
    })

    distributionData.value = response.data
    message.success(t('distribution.generation_success'))
  } catch (err) {
    console.error('Error fetching expense distribution:', err)
    error.value = err instanceof Error ? err.message : t('common.error')
  } finally {
    loading.value = false
  }
}

const fetchCategoryMetadata = async (): Promise<void> => {
  if (!associationId.value) return

  try {
    metadataLoading.value = true

    const response = await categoryApi.getCategories(associationId.value)
    categories.value = response.data

    updateCategoryOptions()
  } catch (err) {
    console.error('Error fetching category metadata:', err)
  } finally {
    metadataLoading.value = false
  }
}

// Consolidated function to update all category options
const updateCategoryOptions = (): void => {
  if (!categories.value) return

  const types = new Set<string>()
  const families = new Set<string>()

  categories.value.forEach((category) => {
    if (category.type) types.add(category.type)

    // Only include families that match the selected type (or all if no type selected)
    if (!selectedCategoryType.value || category.type === selectedCategoryType.value) {
      if (category.family) {
        families.add(category.family)
      }
    }
  })

  categoryOptions.value = categories.value.map((cat) => ({
    label: cat.name,
    value: cat.id.toString()
  }))

  categoryTypeOptions.value = Array.from(types).map((type) => ({
    label: t(`categories.types.${type}`),
    value: type
  }))

  categoryFamilyOptions.value = Array.from(families).map((family) => ({
    label: t(`categories.families.${family}`),
    value: family
  }))
}

const resetFilters = (): void => {
  dateRange.value = getDefaultLastMonthRange()
  selectedCategoryId.value = null
  selectedCategoryType.value = null
  selectedCategoryFamily.value = null
  unitType.value = null
  distributionMethod.value = 'area'

  // Update URL to reflect reset state
  updateUrlFromFilters()
}

const exportToCSV = (): void => {
  if (!distributionData.value) {
    message.error(t('distribution.no_data_to_export', 'No data to export'))
    return
  }

  try {
    // Build all rows for the CSV
    const allRows: (string | number)[][] = []

    // Header row
    const headers: (string | number)[] = [
      t('units.unit'),
      t('units.type'),
      t('units.area') + ' (m²)',
      t('units.part'),
      t('distribution.total_share')
    ]

    // Get category names (keys from category_totals)
    const categoryKeys = distributionData.value.category_totals
      ? Object.keys(distributionData.value.category_totals)
      : []

    // Add translated category headers
    categoryKeys.forEach(categoryKey => {
      headers.push(t(`categories.names.${categoryKey}`))
    })

    allRows.push(headers)

    // Data rows
    distributionData.value.unit_distributions.forEach((unit: UnitDistribution) => {
      const row: (string | number)[] = [
        unit.unit_number,
        unit.unit_type,
        unit.area.toFixed(2),
        (unit.distribution_factor * 100).toFixed(2) + '%',
        unit.total_share.toFixed(2)
      ]

      // Add expense shares for each category
      categoryKeys.forEach(categoryKey => {
        row.push((unit.expenses_share[categoryKey] || 0).toFixed(2))
      })

      allRows.push(row)
    })

    // Add empty row
    allRows.push([])

    // Add summary section
    allRows.push([t('distribution.report_period'), formattedDateRange.value])
    allRows.push([t('distribution.distribution_method'), t(`distribution.method.${distributionMethod.value}`)])
    allRows.push([t('distribution.total_units'), totalUnits.value])
    allRows.push([t('distribution.total_expenses'), totalExpenses.value.toFixed(2)])

    // Add category totals section
    if (distributionData.value.category_totals) {
      allRows.push([])
      allRows.push([t('distribution.category_totals')])
      allRows.push([t('categories.names.title'), t('charts.amount')])

      Object.entries(distributionData.value.category_totals).forEach(([categoryKey, data]: [string, any]) => {
        allRows.push([
          t(`categories.names.${categoryKey}`),
          data.amount.toFixed(2)
        ])
      })
    }

    // Convert to CSV with proper escaping
    const csvContent = arrayToCsv(allRows)

    // Download
    const filename = `expense_distribution_${new Date().toISOString().split('T')[0]}.csv`
    downloadCsv(csvContent, filename)

    message.success(t('distribution.export_success'))
  } catch (err) {
    console.error('Error exporting to CSV:', err)
    message.error(t('distribution.export_error', 'Failed to export CSV'))
  }
}

// Watchers
watch([associationId], () => {
  if (associationId.value) {
    fetchCategoryMetadata()
    initialLoadComplete.value = true
    updateUrlFromFilters()
  }
})

watch([selectedCategoryType], () => {
  if (associationId.value) {
    // Clear family when type changes
    if (selectedCategoryType.value !== null) {
      selectedCategoryFamily.value = null
    }

    // Update family options based on new type
    updateCategoryOptions()
    updateUrlFromFilters()
  }
})

// Watch all filter changes for URL sync and error recovery
watch([buildingId, dateRange, selectedCategoryId, selectedCategoryFamily, unitType, distributionMethod], () => {
  // Clear error when filters change
  error.value = null
  updateUrlFromFilters()
})

// Update URL with current filter state
const updateUrlFromFilters = (): void => {
  const query: Record<string, string> = {}

  if (associationId.value) query.associationId = associationId.value.toString()
  if (buildingId.value) query.buildingId = buildingId.value.toString()
  if (dateRange.value) {
    query.startDate = dateRange.value[0].toString()
    query.endDate = dateRange.value[1].toString()
  }
  if (selectedCategoryId.value) query.categoryId = selectedCategoryId.value.toString()
  if (selectedCategoryType.value) query.categoryType = selectedCategoryType.value
  if (selectedCategoryFamily.value) query.categoryFamily = selectedCategoryFamily.value
  if (unitType.value) query.unitType = unitType.value
  if (distributionMethod.value !== 'area') query.distributionMethod = distributionMethod.value

  router.replace({ query })
}

// Restore state from URL
const restoreFromUrl = (): void => {
  if (route.query.startDate && route.query.endDate) {
    dateRange.value = [
      parseInt(route.query.startDate as string),
      parseInt(route.query.endDate as string)
    ]
  }
  if (route.query.categoryId) {
    selectedCategoryId.value = parseInt(route.query.categoryId as string)
  }
  if (route.query.categoryType) {
    selectedCategoryType.value = route.query.categoryType as string
  }
  if (route.query.categoryFamily) {
    selectedCategoryFamily.value = route.query.categoryFamily as string
  }
  if (route.query.unitType) {
    unitType.value = route.query.unitType as string
  }
  if (route.query.distributionMethod) {
    distributionMethod.value = route.query.distributionMethod as 'area' | 'count' | 'equal'
  }
}

// Lifecycle
onMounted(() => {
  restoreFromUrl()

  if (associationId.value) {
    fetchCategoryMetadata()
    initialLoadComplete.value = true
  }
})
</script>

<template>
  <div class="expense-distribution-report">
    <NCard :loading="loading">
      <div class="filters-section">
        <div class="selectors">
          <NSpace justify-center="center" align="center">
            <NText>{{t('associations.one')}}:</NText>
            <AssociationSelector v-model:associationId="associationId" />
          </NSpace>

          <NSpace justify-center="center" align="center">
            <NText>{{ t('units.building') }}:</NText>
            <BuildingSelector
              v-model:building-id="buildingId"
              v-model:association-id="associationId"
            />
          </NSpace>
        </div>

        <NDivider />

        <div class="filter-grid">
          <NFlex vertical>
            <NText>{{ t('common.dateRange') }}:</NText>
            <NDatePicker
              v-model:value="dateRange"
              type="daterange"
              clearable
              style="width: 100%"
            />
          </NFlex>

          <NFlex vertical>
            <NText>{{ t('units.type') }}:</NText>
            <NSelect
              v-model:value="unitType"
              :options="unitTypeOptions"
              placeholder="All Types"
              clearable
              style="width: 100%"
            />
          </NFlex>

          <NFlex vertical>
            <NText>{{ t('categories.types.title') }}</NText>
            <NSelect
              v-model:value="selectedCategoryType"
              :options="categoryTypeOptions"
              placeholder="All Types"
              clearable
              :loading="metadataLoading"
              style="width: 100%"
            />
          </NFlex>

          <NFlex vertical>
            <NText>{{ t('categories.families.title') }}: </NText>
            <NSelect
              v-model:value="selectedCategoryFamily"
              :options="categoryFamilyOptions"
              placeholder="All Families"
              clearable
              :loading="metadataLoading"
              style="width: 100%"
            />
          </NFlex>

          <NFlex vertical>
            <NText>{{ t('categories.names.title') }}</NText>
            <CategorySelector
              v-model:modelValue="selectedCategoryId"
              :association-id="associationId || 0"
              placeholder="All Categories"
              :options="categoryOptions"
              include-all-option
              :disabled="!associationId"
            />
          </NFlex>

          <NFlex vertical>
            <NText>{{t('distribution.distribution_method')}}</NText>
            <NRadioGroup v-model:value="distributionMethod">
              <NSpace>
                <NRadio value="area">{{t('distribution.method.area')}}</NRadio>
                <NRadio value="count">{{t('distribution.method.count')}}</NRadio>
                <NRadio value="equal">{{t('distribution.method.equal')}}</NRadio>
              </NSpace>
            </NRadioGroup>
          </NFlex>
        </div>

        <div class="actions">
          <NSpace>
            <NButton @click="resetFilters">{{ t('common.reset_filters') }}</NButton>
            <NButton type="primary" @click="fetchDistributionReport" :loading="loading">
              {{ t('distribution.generate_report') }}
            </NButton>
            <NButton type="info" @click="exportToCSV" :disabled="!distributionData">
              {{ t('distribution.export_to_csv') }}
            </NButton>
          </NSpace>
        </div>
      </div>

      <NSpin :show="loading">
        <NAlert v-if="error" type="error" style="margin-bottom: 16px;">
          {{ error }}
        </NAlert>

        <!-- Loading Skeleton -->
        <div v-if="loading && !distributionData" style="margin-top: 20px;">
          <NSkeleton height="100px" style="margin-bottom: 20px;" />
          <NSkeleton height="400px" />
        </div>

        <div v-else-if="distributionData" class="report-content">
          <div class="report-summary">
            <div class="summary-item">
              <div class="summary-label">{{ t('distribution.report_period') }}:</div>
              <div class="summary-value">{{ formattedDateRange }}</div>
            </div>

            <div class="summary-item">
              <div class="summary-label">{{ t('distribution.distribution_method') }}:</div>
              <div class="summary-value">
                {{ t(`distribution.method.${distributionMethod}`) }}
              </div>
            </div>

            <div class="summary-item">
              <div class="summary-label">{{ t('distribution.total_units') }}:</div>
              <div class="summary-value">{{ totalUnits }}</div>
            </div>

            <div class="summary-item">
              <div class="summary-label">{{ t('distribution.total_expenses') }}:</div>
              <div class="summary-value">{{ formatCurrency(totalExpenses) }}</div>
            </div>
          </div>

          <div v-if="distributionData.category_totals" class="category-totals">
            <h4>{{ t('distribution.category_totals') }}</h4>
            <div class="category-totals-grid">
              <div
                v-for="(category, name) in distributionData.category_totals"
                :key="name"
                class="category-total-item"
              >
                <div class="summary-label">{{ t(`categories.names.${name}`) }}</div>
                <div class="summary-value">{{ formatCurrency(category.amount) }}</div>
              </div>
            </div>
          </div>

          <NDivider />
          <div class="distribution-table">
            <NDataTable
              :columns="distributionColumns"
              :data="distributionData.unit_distributions"
              :bordered="false"
              :single-line="false"
              :pagination="{
                pageSize: 25
              }"
            >
              <template #empty>
                <NEmpty :description="t('distribution.no_data_found')" />
              </template>
            </NDataTable>
          </div>
        </div>

        <div v-else-if="!loading && !error" class="no-data">
          <NEmpty
            :description="t('distribution.no_report_data')">
          </NEmpty>
        </div>
      </NSpin>
    </NCard>
  </div>
</template>

<style scoped>
.expense-distribution-report {
  width: 100%;
}

.filters-section {
  margin-bottom: 20px;
}

.selectors {
  display: flex;
  gap: 20px;
  margin-bottom: 16px;
}

.selector-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 200px;
}

.filter-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.report-content {
  margin-top: 20px;
}

.report-summary {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
  padding: 16px;
  border-radius: 8px;
}

.summary-item {
  display: flex;
  flex-direction: column;
}

.summary-label {
  font-size: 0.9rem;
  color: #666;
}

.summary-value {
  font-size: 1.1rem;
  font-weight: 600;
}

.category-totals {
  margin: 20px 0;
}

.category-totals h4 {
  margin-bottom: 10px;
}

.category-totals-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 10px;
}

.category-total-item {
  padding: 8px 12px;
  border-radius: 4px;
}

.category-name {
  font-size: 0.9rem;
}

.category-amount {
  font-weight: 600;
}

.distribution-table {
  margin-top: 20px;
}

.no-data {
  margin: 40px 0;
  text-align: center;
}
</style>
