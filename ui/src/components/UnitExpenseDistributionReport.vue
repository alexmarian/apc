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
  useMessage
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { expenseApi, categoryApi } from '@/services/api'
import AssociationSelector from '@/components/AssociationSelector.vue'
import BuildingSelector from '@/components/BuildingSelector.vue'
import CategorySelector from '@/components/CategorySelector.vue'
import { formatCurrency } from '@/utils/formatters'
import type { ExpenseDistributionResponse } from '@/types/api.ts'

// Message provider
const message = useMessage()

// Props
const props = defineProps<{
  associationId?: number
  buildingId?: number
}>()

// Emits
const emit = defineEmits<{
  (e: 'update:associationId', id: number): void
  (e: 'update:buildingId', id: number): void
}>()

// State
const loading = ref(false)
const metadataLoading = ref(false)
const error = ref<string | null>(null)
const dateRange = ref<[number, number] | null>([
  new Date(new Date().getFullYear(), new Date().getMonth() - 1, 1).getTime(),
  new Date().getTime()
])
const selectedCategoryId = ref<number | null>(null)
const selectedCategoryType = ref<string | null>(null)
const selectedCategoryFamily = ref<string | null>(null)
const distributionMethod = ref<'area' | 'count' | 'equal'>('area')
const unitType = ref<string | null>(null)
const distributionData = ref<ExpenseDistributionResponse | null>(null)
const categories = ref<Category[] | null>(null)
const categoryOptions = ref<{ label: string, value: string }[]>([])
const categoryTypeOptions = ref<{ label: string, value: string }[]>([])
const categoryFamilyOptions = ref<{ label: string, value: string }[]>([])
const initialLoadComplete = ref(false)

// Static unit type options as requested
const unitTypeOptions = [
  { label: 'Apartment', value: 'apartment' },
  { label: 'Commercial', value: 'commercial' },
  { label: 'Office', value: 'office' },
  { label: 'Parking', value: 'parking' },
  { label: 'Storage', value: 'storage' }
]

// Set local association and building IDs if provided as props
const associationId = computed({
  get: () => props.associationId || null,
  set: (value) => emit('update:associationId', value as number)
})

const buildingId = computed({
  get: () => props.buildingId || null,
  set: (value) => emit('update:buildingId', value as number)
})

// Computed properties
const startDate = computed(() => {
  return dateRange.value ? new Date(dateRange.value[0]) : null
})

const endDate = computed(() => {
  return dateRange.value ? new Date(dateRange.value[1]) : null
})

const formattedDateRange = computed(() => {
  if (!dateRange.value) return 'All time'
  const start = new Date(dateRange.value[0]).toLocaleDateString()
  const end = new Date(dateRange.value[1]).toLocaleDateString()
  return `${start} - ${end}`
})

const totalExpenses = computed(() => {
  return distributionData.value?.total_expenses || 0
})

const totalUnits = computed(() => {
  return distributionData.value?.total_units || 0
})

// Distribution table columns
const distributionColumns = computed(() => {
  const columns: DataTableColumns<any> = [
    {
      title: 'Unit',
      key: 'unit_info',
      render: (row) => row.unit_number
    },
    {
      title: 'Type',
      key: 'unit_type',
      render: (row) => row.unit_type
    },
    {
      title: 'Area',
      key: 'area',
      render: (row) => `${row.area} m²`
    },
    {
      title: 'Factor',
      key: 'distribution_factor',
      render: (row) => (row.distribution_factor * 100).toFixed(2) + '%'
    },
    {
      title: 'Total Share',
      key: 'total_share',
      render: (row) => formatCurrency(row.total_share),
      sorter: (a, b) => a.total_share - b.total_share
    }
  ]

  // Add category shares if we have category data
  if (distributionData.value && distributionData.value.category_totals) {
    const categoryKeys = Object.keys(distributionData.value.category_totals)

    for (const category of categoryKeys) {
      columns.push({
        title: category,
        key: category,
        render: (row) => formatCurrency(row.expenses_share[category] || 0)
      })
    }
  }

  return columns
})

// Methods
// Fetch data from the expense distribution API
const fetchDistributionReport = async () => {
  if (!associationId.value) {
    error.value = 'Please select an association'
    return
  }

  try {
    loading.value = true
    error.value = null

    // Format dates for API
    const startDateStr = startDate.value ? startDate.value.toISOString().split('T')[0] : undefined
    const endDateStr = endDate.value ? endDate.value.toISOString().split('T')[0] : undefined

    // Call the expense distribution API
    const response = await expenseApi.getExpenseDistribution(
      associationId.value,
      {
        start_date: startDateStr,
        end_date: endDateStr,
        category_id: selectedCategoryId.value,
        category_type: selectedCategoryType.value,
        category_family: selectedCategoryFamily.value,
        distribution_method: distributionMethod.value,
        unit_type: unitType.value
      }
    )

    distributionData.value = response.data
    console.log('Distribution data:', distributionData.value)
    message.success('Report generated successfully')
  } catch (err) {
    console.error('Error fetching expense distribution:', err)
    error.value = err instanceof Error ? err.message : 'An error occurred while fetching data'
  } finally {
    loading.value = false
  }
}

// Fetch category types and families from API
const fetchCategoryMetadata = async () => {
  if (!associationId.value) return

  try {
    metadataLoading.value = true

    // Fetch all categories from the existing endpoint
    const response = await categoryApi.getCategories(associationId.value)
    categories.value = response.data

    // Extract unique types and families
    const types = new Set<string>()
    const families = new Set<string>()

    categories.value.forEach(category => {
      if (category.type) types.add(category.type)
      if (category.family) families.add(category.family)
    })

    categoryOptions.value = Array.from(categories).map(cat => ({
      label: cat.name,
      value: cat.id
    }))
    // Update options
    categoryTypeOptions.value = Array.from(types).map(type => ({
      label: type,
      value: type
    }))

    categoryFamilyOptions.value = Array.from(families).map(family => ({
      label: family,
      value: family
    }))

  } catch (err) {
    console.error('Error fetching category metadata:', err)
  } finally {
    metadataLoading.value = false
  }
}

// Update family options when category type changes
const updateCategoryFamilies = () => {
  if (!associationId.value) return

  try {
    metadataLoading.value = true

    // Use the categories API to get all categories and filter them
    categoryApi.getCategories(associationId.value).then(response => {
      const categories = response.data
      const families = new Set<string>()

      categories.forEach(category => {
        // Only include families that match the selected type
        if (selectedCategoryType.value && category.type !== selectedCategoryType.value) {
          return
        }

        if (category.family) {
          families.add(category.family)
        }
      })

      categoryFamilyOptions.value = Array.from(families).map(family => ({
        label: family,
        value: family
      }))

      metadataLoading.value = false
    })
  } catch (err) {
    console.error('Error updating category families:', err)
    metadataLoading.value = false
  }
}

// Reset all filters
const resetFilters = () => {
  dateRange.value = [
    new Date(new Date().getFullYear(), new Date().getMonth() - 1, 1).getTime(),
    new Date().getTime()
  ]
  selectedCategoryId.value = null
  selectedCategoryType.value = null
  selectedCategoryFamily.value = null
  unitType.value = null
  distributionMethod.value = 'area'
}

// Export data to CSV
const exportToCSV = () => {
  if (!distributionData.value) {
    message.error('No data to export')
    return
  }

  try {
    // Create CSV header row
    const headers = ['Unit Number', 'Building', 'Type', 'Area (m²)', 'Distribution Factor', 'Total Share']

    // Add category headers if available
    const categoryHeaders = distributionData.value.category_totals
      ? Object.keys(distributionData.value.category_totals)
      : []

    headers.push(...categoryHeaders)

    // Create CSV rows
    const rows = distributionData.value.unit_distributions.map(unit => {
      const row = [
        unit.unit_number,
        unit.building_name,
        unit.unit_type,
        unit.area,
        (unit.distribution_factor * 100).toFixed(2) + '%',
        unit.total_share.toFixed(2)
      ]

      // Add category values
      categoryHeaders.forEach(category => {
        row.push((unit.expenses_share[category] || 0).toFixed(2))
      })

      return row
    })

    // Create CSV content
    let csvContent = headers.join(',') + '\n'
    rows.forEach(row => {
      csvContent += row.join(',') + '\n'
    })

    // Add summary data
    csvContent += '\n'
    csvContent += 'Report Period:,' + formattedDateRange.value + '\n'
    csvContent += 'Distribution Method:,' + distributionMethod.value + '\n'
    csvContent += 'Total Units:,' + totalUnits.value + '\n'
    csvContent += 'Total Expenses:,' + totalExpenses.value.toFixed(2) + '\n'

    // Add category totals
    if (distributionData.value.category_totals) {
      csvContent += '\nCategory Totals:\n'
      csvContent += 'Category,Amount\n'
      Object.entries(distributionData.value.category_totals).forEach(([category, data]: [string, any]) => {
        csvContent += `${category},${data.amount.toFixed(2)}\n`
      })
    }

    // Create a CSV blob and download it
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.setAttribute('href', url)
    link.setAttribute('download', `expense_distribution_${new Date().toISOString().split('T')[0]}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    message.success('CSV exported successfully')
  } catch (err) {
    console.error('Error exporting to CSV:', err)
    message.error('Failed to export CSV')
  }
}

// Watch for property changes to reload metadata
watch([associationId], () => {
  if (associationId.value) {
    fetchCategoryMetadata()
    initialLoadComplete.value = true
  }
})

// When category type changes, update family options
watch([selectedCategoryType], () => {
  if (associationId.value) {
    // If we select a category type, reset the family selection
    if (selectedCategoryType.value !== null) {
      selectedCategoryFamily.value = null
    }

    // Update family options based on the selected type
    updateCategoryFamilies()
  }
})

// Load category metadata on component mount
onMounted(() => {
  if (associationId.value) {
    fetchCategoryMetadata()
    initialLoadComplete.value = true
  }
})
</script>

<template>
  <div class="expense-distribution-report">
    <NCard title="Expense Distribution Report">
      <div class="filters-section">
        <div class="selectors">
          <div class="selector-group">
            <label>Association:</label>
            <AssociationSelector v-model:associationId="associationId" />
          </div>

          <div class="selector-group">
            <label>Building (Optional):</label>
            <BuildingSelector
              v-model:building-id="buildingId"
              v-model:association-id="associationId"
            />
          </div>
        </div>

        <NDivider />

        <div class="filter-grid">
          <div class="filter-group">
            <label>Date Range:</label>
            <NDatePicker
              v-model:value="dateRange"
              type="daterange"
              clearable
              style="width: 100%"
            />
          </div>

          <div class="filter-group">
            <label>Unit Type:</label>
            <NSelect
              v-model:value="unitType"
              :options="unitTypeOptions"
              placeholder="All Types"
              clearable
              style="width: 100%"
            />
          </div>

          <div class="filter-group">
            <label>Category Type:</label>
            <NSelect
              v-model:value="selectedCategoryType"
              :options="categoryTypeOptions"
              placeholder="All Types"
              clearable
              :loading="metadataLoading"
              style="width: 100%"
            />
          </div>

          <div class="filter-group">
            <label>Category Family:</label>
            <NSelect
              v-model:value="selectedCategoryFamily"
              :options="categoryFamilyOptions"
              placeholder="All Families"
              clearable
              :loading="metadataLoading"
              style="width: 100%"
            />
          </div>

          <div class="filter-group">
            <label>Specific Category:</label>
            <CategorySelector
              v-model:modelValue="selectedCategoryId"
              :association-id="associationId || 0"
              placeholder="All Categories"
              :options="categoryOptions"
              include-all-option
              :disabled="!associationId"
            />
          </div>

          <div class="filter-group">
            <label>Distribution Method:</label>
            <NRadioGroup v-model:value="distributionMethod">
              <NSpace>
                <NRadio value="area">By Area</NRadio>
                <NRadio value="count">By Count</NRadio>
                <NRadio value="equal">Equal</NRadio>
              </NSpace>
            </NRadioGroup>
          </div>
        </div>

        <div class="actions">
          <NSpace>
            <NButton @click="resetFilters">Reset Filters</NButton>
            <NButton type="primary" @click="fetchDistributionReport" :loading="loading">
              Generate Report
            </NButton>
            <NButton type="info" @click="exportToCSV" :disabled="!distributionData">
              Export to CSV
            </NButton>
          </NSpace>
        </div>
      </div>

      <NSpin :show="loading">
        <NAlert v-if="error" type="error" style="margin-bottom: 16px;">
          {{ error }}
        </NAlert>

        <div v-if="distributionData" class="report-content">
          <div class="report-summary">
            <div class="summary-item">
              <div class="summary-label">Period:</div>
              <div class="summary-value">{{ formattedDateRange }}</div>
            </div>

            <div class="summary-item">
              <div class="summary-label">Distribution:</div>
              <div class="summary-value">
                {{ distributionMethod === 'area' ? 'By Area' : distributionMethod === 'count' ? 'By Count' : 'Equal'
                }}
              </div>
            </div>

            <div class="summary-item">
              <div class="summary-label">Total Units:</div>
              <div class="summary-value">{{ totalUnits }}</div>
            </div>

            <div class="summary-item">
              <div class="summary-label">Total Expenses:</div>
              <div class="summary-value">{{ formatCurrency(totalExpenses) }}</div>
            </div>
          </div>

          <div v-if="distributionData.category_totals" class="category-totals">
            <h4>Category Totals</h4>
            <div class="category-totals-grid">
              <div
                v-for="(category, name) in distributionData.category_totals"
                :key="name"
                class="category-total-item"
              >
                <div class="category-name">{{ name }}</div>
                <div class="category-amount">{{ formatCurrency(category.amount) }}</div>
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
                <NEmpty description="No distribution data found for the selected filters" />
              </template>
            </NDataTable>
          </div>
        </div>

        <div v-else-if="!loading && !error" class="no-data">
          <NEmpty
            description="No report data available yet. Click 'Generate Report' to run the calculation.">
            <template #extra>
              <NButton type="primary" @click="fetchDistributionReport" :loading="loading">
                Generate Report
              </NButton>
            </template>
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
