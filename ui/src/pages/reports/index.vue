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
  NTooltip,
  NText,
  NFlex,
  NSkeleton
} from 'naive-ui'
import { useRouter, useRoute } from 'vue-router'
import { expenseApi } from '@/services/api'
import type { Expense } from '@/types/api'
import { formatCurrency } from '@/utils/formatters'
import { groupExpensesByMonth, calculateMonthlyAverage } from '@/utils/expenseUtils'
import AssociationSelector from '@/components/AssociationSelector.vue'
import CategorySelector from '@/components/CategorySelector.vue'
import ExpenseCharts from '@/components/ExpenseCharts.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()

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

// Reports data
const yearlyTotal = computed(() => {
  return expenses.value.reduce((sum, expense) => sum + expense.amount, 0)
})

const monthlyAverage = computed(() => {
  return calculateMonthlyAverage(expenses.value)
})

const expensesByType = computed(() => {
  // Group expenses by type
  const types: Record<string, number> = {}

  expenses.value.forEach(expense => {
    const type = expense.category_type || t('expenses.uncategorized', 'Uncategorized')

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
  return groupExpensesByMonth(expenses.value)
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

    const response = await expenseApi.getExpenses(associationId.value, startDate, endDate, selectedCategory.value || undefined)
    expenses.value = response.data
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('common.error', 'Unknown error occurred')
    console.error('Error fetching expenses:', err)
  } finally {
    loading.value = false
  }
}

// Date preset helpers
const selectPreviousMonth = () => {
  const now = new Date()
  const start = new Date(now.getFullYear(), now.getMonth() - 1, 1)
  const end = new Date(now.getFullYear(), now.getMonth(), 0)
  dateRange.value = [start.getTime(), end.getTime()]
}

const selectCurrentMonth = () => {
  const now = new Date()
  const start = new Date(now.getFullYear(), now.getMonth(), 1)
  const end = new Date(now.getFullYear(), now.getMonth() + 1, 0)
  dateRange.value = [start.getTime(), end.getTime()]
}

const selectCurrentYear = () => {
  const now = new Date()
  const start = new Date(now.getFullYear(), 0, 1)
  const end = new Date(now.getFullYear(), 11, 31)
  dateRange.value = [start.getTime(), end.getTime()]
}

// Reset filters
const resetFilters = () => {
  dateRange.value = null
  selectedCategory.value = null

  // Update URL
  if (associationId.value) {
    router.replace({
      query: {
        associationId: associationId.value.toString()
      }
    })
  }
}

// Update URL when filters change
const updateUrlFromFilters = () => {
  if (!associationId.value) return

  const query: Record<string, string> = {
    associationId: associationId.value.toString()
  }

  if (dateRange.value) {
    query.startDate = dateRange.value[0].toString()
    query.endDate = dateRange.value[1].toString()
  }

  if (selectedCategory.value) {
    query.categoryId = selectedCategory.value.toString()
  }

  router.replace({ query })
}

// Watch for changes in filters and refresh data
watch([associationId, dateRange, selectedCategory], () => {
  // Clear error when filters change to allow retry
  error.value = null
  updateUrlFromFilters()
  fetchExpenses()
})


// Format date range for display
const formattedDateRange = computed(() => {
  if (!dateRange.value) return t('expenses.allTime', 'All time')

  const start = new Date(dateRange.value[0]).toLocaleDateString()
  const end = new Date(dateRange.value[1]).toLocaleDateString()

  return `${start} - ${end}`
})

// Initialize from URL query params
onMounted(() => {
  // Restore association ID from URL
  if (route.query.associationId) {
    associationId.value = parseInt(route.query.associationId as string)
  }

  // Restore date range from URL, or set default to current year
  if (route.query.startDate && route.query.endDate) {
    dateRange.value = [
      parseInt(route.query.startDate as string),
      parseInt(route.query.endDate as string)
    ]
  } else if (associationId.value) {
    // Only set default date range if association is selected
    const now = new Date()
    const startOfYear = new Date(now.getFullYear(), 0, 1)
    const endOfYear = new Date(now.getFullYear(), 11, 31)
    dateRange.value = [startOfYear.getTime(), endOfYear.getTime()]
  }

  // Restore category filter from URL
  if (route.query.categoryId) {
    selectedCategory.value = parseInt(route.query.categoryId as string)
  }
})
</script>

<template>
  <div class="reports-view">
    <NPageHeader>
      <template #title>
        {{ t('reports.title', 'Expense Reports') }}
      </template>

      <template #header>
        <div style="margin-bottom: 12px;">
          <AssociationSelector v-model:associationId="associationId" />
        </div>
      </template>
    </NPageHeader>

    <div v-if="!associationId">
      <NCard style="margin-top: 16px;">
        <div style="text-align: center; padding: 32px;">
          <p>{{ t('expenses.selectAssociation', 'Please select an association to view expense reports') }}</p>
        </div>
      </NCard>
    </div>

    <div v-else>
      <!-- Filters -->
      <NCard style="margin-top: 16px;">
        <NFlex align="center" justify="start">
          <NText>{{ t('expenses.dateRange', 'Date Range') }}:</NText>
          <NDatePicker
            v-model:value="dateRange"
            type="daterange"
            clearable
            style="width: 240px"
          />
          <NButton size="small" @click="selectPreviousMonth">{{ t('reports.previousMonth', 'Previous Month') }}</NButton>
          <NButton size="small" @click="selectCurrentMonth">{{ t('reports.currentMonth', 'Current Month') }}</NButton>
          <NButton size="small" @click="selectCurrentYear">{{ t('reports.currentYear', 'Current Year') }}</NButton>
          <NText>{{ t('expenses.category', 'Category') }}:</NText>
          <CategorySelector
            v-model:modelValue="selectedCategory"
            :association-id="associationId"
            :placeholder="t('categories.allCategories', 'All Categories')"
            :include-all-option="true"
            style="width: 360px"
          />
          <NButton @click="resetFilters">{{ t('expenses.resetFilters', 'Reset Filters') }}</NButton>
        </NFlex>
      </NCard>

      <NSpin :show="loading">
        <NAlert v-if="error" type="error" :title="t('common.error', 'Error')" closable style="margin-top: 16px;">
          {{ error }}
          <NButton @click="fetchExpenses">{{ t('common.retry', 'Retry') }}</NButton>
        </NAlert>

        <template v-else-if="expenses.length === 0">
          <NCard style="margin-top: 16px;">
            <div style="text-align: center; padding: 32px;">
              <p>{{ t('expenses.noExpensesFilters', 'No expenses found for the selected filters') }}</p>
            </div>
          </NCard>
        </template>

        <template v-else>
          <!-- Summary Statistics -->
          <NCard style="margin-top: 16px;" :title="t('expenses.summary', 'Summary')">
            <div class="summary-header">
              <h3>{{ t('expenses.expenseAnalysis', 'Expense Analysis') }} {{ formattedDateRange }}</h3>
            </div>

            <NGrid :cols="3" :x-gap="16">
              <NGridItem>
                <NStatistic :label="t('expenses.totalExpenses', 'Total Expenses')" :value="formatCurrency(yearlyTotal)" />
              </NGridItem>
              <NGridItem>
                <NStatistic :label="t('expenses.averageExpense', 'Monthly Average')" :value="formatCurrency(monthlyAverage)" />
              </NGridItem>
              <NGridItem>
                <NStatistic :label="t('expenses.numberOfExpenses', 'Number of Expenses')" :value="expenses.length" />
              </NGridItem>
            </NGrid>

            <NDivider />

            <NTabs type="line" animated>
              <NTabPane name="charts" :tab="t('charts.pieChart', 'Visual Reports')">
                <template v-if="loading">
                  <div style="padding: 20px;">
                    <NSkeleton height="300px" style="margin-bottom: 20px;" />
                    <NSkeleton height="300px" style="margin-bottom: 20px;" />
                    <NSkeleton height="300px" />
                  </div>
                </template>
                <ExpenseCharts
                  v-else
                  :expenses="expenses"
                  :dateRange="dateRange"
                />
              </NTabPane>

              <NTabPane name="breakdown" :tab="t('expenses.expenseTypeBreakdown', 'Breakdown')">
                <div class="breakdown-section">
                  <h3>{{ t('expenses.expensesByType', 'Expense Breakdown by Type') }}</h3>
                  <div class="breakdown-table">
                    <table>
                      <thead>
                      <tr>
                        <th>{{ t('categories.type') }}</th>
                        <th>{{ t('charts.amount', 'Amount') }}</th>
                        <th>{{ t('charts.percentage', 'Percentage') }}</th>
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
                  <h3>{{ t('expenses.monthlyTrends', 'Monthly Expenses') }}</h3>
                  <div class="breakdown-table">
                    <table>
                      <thead>
                      <tr>
                        <th>{{ t('expenses.month', 'Month') }}</th>
                        <th>{{ t('charts.amount', 'Amount') }}</th>
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
