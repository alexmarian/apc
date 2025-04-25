<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import {
  NCard,
  NPageHeader,
  NSpin,
  NAlert,
  NDatePicker,
  NButton,
  NSpace,
  NDivider,
  NTabs,
  NTabPane,
  NGrid,
  NGridItem,
  NStatistic,
  NTooltip
} from 'naive-ui'
import { expenseApi } from '@/services/api'
import type { Expense } from '@/types/api'
import { formatCurrency } from '@/utils/formatters'
import { exportFullReportToPdf } from '@/utils/pdfExport'
import AssociationSelector from '@/components/AssociationSelector.vue'
import CategorySelector from '@/components/CategorySelector.vue'
import ExpenseCharts from '@/components/ExpenseCharts.vue'

// Association selector
const associationId = ref<number | null>(null)

// Data states
const expenses = ref<Expense[]>([])
const loading = ref<boolean>(false)
const error = ref<string | null>(null)

// Filters
const dateRange = ref<[number, number] | null>(null)
const selectedCategory = ref<number | null>(null)
const reportType = ref<string>('overview')

// Reference to the ExpenseCharts component
const expenseChartsRef = ref<any>(null)

// Reports data
const yearlyTotal = computed(() => {
  return expenses.value.reduce((sum, expense) => sum + expense.amount, 0)
})

const monthlyAverage = computed(() => {
  // Group expenses by month
  const months: Record<string, number> = {}

  expenses.value.forEach(expense => {
    const date = new Date(expense.date)
    const monthYear = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`

    if (!months[monthYear]) {
      months[monthYear] = 0
    }

    months[monthYear] += expense.amount
  })

  const monthCount = Object.keys(months).length
  return monthCount > 0 ? yearlyTotal.value / monthCount : 0
})

const expensesByType = computed(() => {
  // Group expenses by type
  const types: Record<string, number> = {}

  expenses.value.forEach(expense => {
    const type = expense.category_type || 'Uncategorized'

    if (!types[type]) {
      types[type] = 0
    }

    types[type] += expense.amount
  })

  return Object.entries(types).map(([name, value]) => ({
    name,
    value,
    percentage: (value / yearlyTotal.value) * 100
  })).sort((a, b) => b.value - a.value)
})

const expensesByMonth = computed(() => {
  // Group expenses by month
  const months: Record<string, number> = {}

  expenses.value.forEach(expense => {
    const date = new Date(expense.date)
    const monthYear = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`

    if (!months[monthYear]) {
      months[monthYear] = 0
    }

    months[monthYear] += expense.amount
  })

  return Object.entries(months).map(([month, value]) => ({
    month,
    value
  })).sort((a, b) => a.month.localeCompare(b.month))
})

// Fetch expenses
const fetchExpenses = async () => {
  if (!associationId.value) return

  try {
    loading.value = true
    error.value = null

    // Prepare date filters if set
    let startDate: string | undefined
    let endDate: string | undefined

    if (dateRange.value) {
      startDate = new Date(dateRange.value[0]).toISOString().split('T')[0]
      endDate = new Date(dateRange.value[1]).toISOString().split('T')[0]
    }

    const response = await expenseApi.getExpenses(associationId.value, startDate, endDate)
    expenses.value = response.data

    // Filter by category if selected
    if (selectedCategory.value) {
      expenses.value = expenses.value.filter(expense => expense.category_id === selectedCategory.value)
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown error occurred'
    console.error('Error fetching expenses:', err)
  } finally {
    loading.value = false
  }
}

// Reset filters
const resetFilters = () => {
  dateRange.value = null
  selectedCategory.value = null
}

// Watch for changes in filters and refresh data
watch([associationId, dateRange, selectedCategory], () => {
  fetchExpenses()
})

// Format date range for display
const formattedDateRange = computed(() => {
  if (!dateRange.value) return 'All time'

  const start = new Date(dateRange.value[0]).toLocaleDateString()
  const end = new Date(dateRange.value[1]).toLocaleDateString()

  return `${start} - ${end}`
})

// Export comprehensive report with all data
const exportReport = () => {
  try {
    // Try to find an SVG element in the ExpenseCharts component
    let chartSvg: SVGElement | null = null

    // Check if the ExpenseCharts component is accessible
    if (expenseChartsRef.value) {
      // Look for SVG elements in the component
      const svgElements = expenseChartsRef.value.$el.querySelectorAll('svg')
      if (svgElements && svgElements.length > 0) {
        // Get the first visible SVG
        for (const svg of svgElements) {
          if (svg.getBoundingClientRect().height > 0) {
            chartSvg = svg
            break
          }
        }
      }
    }

    // Prepare summary data
    const summaryData = [
      {
        label: 'Total Expenses',
        value: formatCurrency(yearlyTotal.value)
      },
      {
        label: 'Monthly Average',
        value: formatCurrency(monthlyAverage.value)
      },
      {
        label: 'Number of Expenses',
        value: expenses.value.length.toString()
      }
    ]

    // Prepare breakdown data
    const breakdownData = {
      types: expensesByType.value,
      months: expensesByMonth.value
    }

    // Export the full report
    exportFullReportToPdf(
      'Expense Analysis Report',
      summaryData,
      chartSvg,
      breakdownData,
      formattedDateRange.value
    )
  } catch (error) {
    console.error('Error exporting report:', error)
    alert('There was an error generating the report. Please try again.')
  }
}

// Initialize data
onMounted(() => {
  // Set default date range to current year
  const now = new Date()
  const startOfYear = new Date(now.getFullYear(), 0, 1)
  const endOfYear = new Date(now.getFullYear(), 11, 31)

  dateRange.value = [startOfYear.getTime(), endOfYear.getTime()]
})
</script>

<template>
  <div class="reports-view">
    <NPageHeader>
      <template #title>
        Expense Reports
      </template>

      <template #header>
        <div style="margin-bottom: 12px;">
          <AssociationSelector v-model:modelValue="associationId" />
        </div>
      </template>

      <template #extra>
        <NButton
          v-if="expenses.length > 0"
          type="primary"
          @click="exportReport"
        >
          Export Report
        </NButton>
      </template>
    </NPageHeader>

    <div v-if="!associationId">
      <NCard style="margin-top: 16px;">
        <div style="text-align: center; padding: 32px;">
          <p>Please select an association to view expense reports</p>
        </div>
      </NCard>
    </div>

    <div v-else>
      <!-- Filters -->
      <NCard style="margin-top: 16px;" title="Report Filters">
        <NSpace align="center" justify="start">
          <div>
            <label>Date Range:</label>
            <NDatePicker
              v-model:value="dateRange"
              type="daterange"
              clearable
              style="width: 240px"
            />
          </div>
          <div>
            <label>Category (Optional):</label>
            <CategorySelector
              v-model:modelValue="selectedCategory"
              :association-id="associationId"
              placeholder="All Categories"
              :include-all-option="true"
              style="width: 360px"
            />
          </div>
          <NButton @click="resetFilters">Reset Filters</NButton>
        </NSpace>
      </NCard>

      <NSpin :show="loading">
        <NAlert v-if="error" type="error" title="Error" closable style="margin-top: 16px;">
          {{ error }}
          <template #action>
            <NButton @click="fetchExpenses">Retry</NButton>
          </template>
        </NAlert>

        <template v-else-if="expenses.length === 0">
          <NCard style="margin-top: 16px;">
            <div style="text-align: center; padding: 32px;">
              <p>No expenses found for the selected filters</p>
            </div>
          </NCard>
        </template>

        <template v-else>
          <!-- Summary Statistics -->
          <NCard style="margin-top: 16px;" title="Summary">
            <div class="summary-header">
              <h3>Expense Summary for {{ formattedDateRange }}</h3>
            </div>

            <NGrid :cols="3" :x-gap="16">
              <NGridItem>
                <NStatistic label="Total Expenses" :value="formatCurrency(yearlyTotal)" />
              </NGridItem>
              <NGridItem>
                <NStatistic label="Monthly Average" :value="formatCurrency(monthlyAverage)" />
              </NGridItem>
              <NGridItem>
                <NStatistic label="Number of Expenses" :value="expenses.length" />
              </NGridItem>
            </NGrid>

            <NDivider />

            <NTabs type="line" animated>
              <NTabPane name="charts" tab="Visual Reports">
                <ExpenseCharts
                  ref="expenseChartsRef"
                  :expenses="expenses"
                />
              </NTabPane>

              <NTabPane name="breakdown" tab="Breakdown">
                <div class="breakdown-section">
                  <h3>Expense Breakdown by Type</h3>
                  <div class="breakdown-table">
                    <table>
                      <thead>
                      <tr>
                        <th>Type</th>
                        <th>Amount</th>
                        <th>Percentage</th>
                      </tr>
                      </thead>
                      <tbody>
                      <tr v-for="type in expensesByType" :key="type.name">
                        <td>{{ type.name }}</td>
                        <td>{{ formatCurrency(type.value) }}</td>
                        <td>{{ type.percentage.toFixed(1) }}%</td>
                      </tr>
                      </tbody>
                    </table>
                  </div>
                </div>

                <NDivider />

                <div class="breakdown-section">
                  <h3>Monthly Expenses</h3>
                  <div class="breakdown-table">
                    <table>
                      <thead>
                      <tr>
                        <th>Month</th>
                        <th>Amount</th>
                      </tr>
                      </thead>
                      <tbody>
                      <tr v-for="month in expensesByMonth" :key="month.month">
                        <td>{{ month.month }}</td>
                        <td>{{ formatCurrency(month.value) }}</td>
                      </tr>
                      </tbody>
                    </table>
                  </div>
                </div>
              </NTabPane>
            </NTabs>
          </NCard>
        </template>
      </NSpin>
    </div>
  </div>
</template>

<style scoped>
.reports-view {
  width: 100%;
  margin: 0 auto;
  max-width: 1600px;
}

.summary-header {
  margin-bottom: 20px;
  text-align: center;
}

.breakdown-section {
  margin: 20px 0;
}

.breakdown-section h3 {
  margin-bottom: 10px;
  font-size: 1.1rem;
  font-weight: 600;
}

.breakdown-table {
  width: 100%;
  overflow-x: auto;
}

.breakdown-table table {
  width: 100%;
  border-collapse: collapse;
}

.breakdown-table th,
.breakdown-table td {
  padding: 10px;
  text-align: left;
  border-bottom: 1px solid var(--border-color);
}

.breakdown-table th {
  font-weight: 600;
  background-color: var(--background-alt-color);
}
</style>
