<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import {
  NCard,
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
  NText,
  NEmpty,
  NDataTable,
  NSkeleton,
  type DataTableColumns,
} from 'naive-ui'
import { useRouter, useRoute } from 'vue-router'
import { expenseApi } from '@/services/api'
import type { Expense } from '@/types/api'
import { formatCurrency } from '@/utils/formatters'
import { groupExpensesByMonth, calculateMonthlyAverage } from '@/utils/expenseUtils'
import { storeToRefs } from 'pinia'
import { useAssociationStore } from '@/stores/association'
import CategorySelector from '@/components/CategorySelector.vue'
import PageShell from '@/components/PageShell.vue'
import FilterBar from '@/components/FilterBar.vue'
import FilterField from '@/components/FilterField.vue'
import ExpenseCharts from '@/components/ExpenseCharts.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()

const { associationId } = storeToRefs(useAssociationStore())

// Data states
const expenses = ref<Expense[]>([])
const loading = ref<boolean>(false)
const error = ref<string | null>(null)

// Filters
const dateRange = ref<[number, number] | null>(null)
const selectedCategory = ref<number | null>(null)

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

  expenses.value.forEach((expense) => {
    const type = expense.category_type || t('expenses.uncategorized', 'Uncategorized')

    if (!types[type]) {
      types[type] = 0
    }

    types[type] += expense.amount
  })

  return Object.entries(types)
    .map(([name, value]) => ({
      name,
      value,
      percentage: (value / yearlyTotal.value) * 100,
    }))
    .sort((a, b) => b.value - a.value)
})

const expensesByMonth = computed(() => {
  return groupExpensesByMonth(expenses.value)
})

const byTypeColumns = computed<
  DataTableColumns<{ name: string; value: number; percentage: number }>
>(() => [
  { title: t('categories.type'), key: 'name' },
  { title: t('charts.amount', 'Amount'), key: 'value', render: (row) => formatCurrency(row.value) },
  {
    title: t('charts.percentage', 'Percentage'),
    key: 'percentage',
    render: (row) => `${row.percentage.toFixed(1)}%`,
  },
])

const byMonthColumns = computed<DataTableColumns<{ month: string; value: number }>>(() => [
  { title: t('expenses.month', 'Month'), key: 'month' },
  { title: t('charts.amount', 'Amount'), key: 'value', render: (row) => formatCurrency(row.value) },
])

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

    const response = await expenseApi.getExpenses(
      associationId.value,
      startDate,
      endDate,
      selectedCategory.value || undefined,
    )
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
        associationId: associationId.value.toString(),
      },
    })
  }
}

// Update URL when filters change
const updateUrlFromFilters = () => {
  if (!associationId.value) return

  const query: Record<string, string> = {
    associationId: associationId.value.toString(),
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

// Initialize filters from URL query params or set defaults
onMounted(() => {
  if (route.query.startDate && route.query.endDate) {
    dateRange.value = [
      parseInt(route.query.startDate as string),
      parseInt(route.query.endDate as string),
    ]
  } else {
    const now = new Date()
    dateRange.value = [
      new Date(now.getFullYear(), 0, 1).getTime(),
      new Date(now.getFullYear(), 11, 31).getTime(),
    ]
  }

  if (route.query.categoryId) {
    selectedCategory.value = parseInt(route.query.categoryId as string)
  }
})
</script>

<template>
  <PageShell :title="t('reports.title', 'Expense Reports')">
    <FilterBar resettable @reset="resetFilters">
      <FilterField :label="t('expenses.dateRange', 'Date Range')">
        <NDatePicker v-model:value="dateRange" type="daterange" clearable style="width: 240px" />
      </FilterField>
      <NSpace :size="8">
        <NButton size="small" @click="selectPreviousMonth">{{
          t('reports.previousMonth', 'Previous Month')
        }}</NButton>
        <NButton size="small" @click="selectCurrentMonth">{{
          t('reports.currentMonth', 'Current Month')
        }}</NButton>
        <NButton size="small" @click="selectCurrentYear">{{
          t('reports.currentYear', 'Current Year')
        }}</NButton>
      </NSpace>
      <FilterField :label="t('expenses.category', 'Category')">
        <CategorySelector
          v-model:modelValue="selectedCategory"
          :association-id="associationId"
          :placeholder="t('categories.allCategories', 'All Categories')"
          :include-all-option="true"
          style="width: 360px"
        />
      </FilterField>
    </FilterBar>

    <div>
      <NSpin :show="loading">
        <NAlert v-if="error" type="error" :title="t('common.error', 'Error')" closable>
          {{ error }}
          <NButton @click="fetchExpenses">{{ t('common.retry', 'Retry') }}</NButton>
        </NAlert>

        <template v-else-if="expenses.length === 0">
          <NCard>
            <NEmpty
              :description="
                t('expenses.noExpensesFilters', 'No expenses found for the selected filters')
              "
            />
          </NCard>
        </template>

        <template v-else>
          <!-- Summary Statistics -->
          <NCard :title="t('expenses.summary', 'Summary')">
            <NText strong class="summary-header">
              {{ t('expenses.expenseAnalysis', 'Expense Analysis') }} {{ formattedDateRange }}
            </NText>

            <NGrid :cols="3" :x-gap="16">
              <NGridItem>
                <NStatistic
                  :label="t('expenses.totalExpenses', 'Total Expenses')"
                  :value="formatCurrency(yearlyTotal)"
                />
              </NGridItem>
              <NGridItem>
                <NStatistic
                  :label="t('expenses.averageExpense', 'Monthly Average')"
                  :value="formatCurrency(monthlyAverage)"
                />
              </NGridItem>
              <NGridItem>
                <NStatistic
                  :label="t('expenses.numberOfExpenses', 'Number of Expenses')"
                  :value="expenses.length"
                />
              </NGridItem>
            </NGrid>

            <NDivider />

            <NTabs type="line" animated>
              <NTabPane name="charts" :tab="t('charts.pieChart', 'Visual Reports')">
                <template v-if="loading">
                  <div style="padding: 20px">
                    <NSkeleton height="300px" style="margin-bottom: 20px" />
                    <NSkeleton height="300px" style="margin-bottom: 20px" />
                    <NSkeleton height="300px" />
                  </div>
                </template>
                <ExpenseCharts v-else :expenses="expenses" :dateRange="dateRange" />
              </NTabPane>

              <NTabPane name="breakdown" :tab="t('expenses.expenseTypeBreakdown', 'Breakdown')">
                <NText strong class="breakdown-title">
                  {{ t('expenses.expensesByType', 'Expense Breakdown by Type') }}
                </NText>
                <NDataTable
                  :columns="byTypeColumns"
                  :data="expensesByType"
                  :row-key="(row) => row.name"
                  :bordered="false"
                  size="small"
                />

                <NDivider />

                <NText strong class="breakdown-title">
                  {{ t('expenses.monthlyTrends', 'Monthly Expenses') }}
                </NText>
                <NDataTable
                  :columns="byMonthColumns"
                  :data="expensesByMonth"
                  :row-key="(row) => row.month"
                  :bordered="false"
                  size="small"
                />
              </NTabPane>
            </NTabs>
          </NCard>
        </template>
      </NSpin>
    </div>
  </PageShell>
</template>

<style scoped>
.summary-header {
  display: block;
  margin-bottom: 20px;
  text-align: center;
}

.breakdown-title {
  display: block;
  margin-bottom: 12px;
}
</style>
