<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { NCard, NSpin, NStatistic, NGradientText, NSpace, NDivider, NSelect } from 'naive-ui'
import { expenseApi } from '@/services/api'
import type { Expense } from '@/types/api'
import { formatCurrency } from '@/utils/formatters'
import { getCurrentMonthRange, getPreviousMonthRange, getCurrentYearRange } from '@/utils'

// Props
const props = defineProps<{
  associationId: number
}>()

// Emits
const emit = defineEmits<{
  (e: 'update:dateRange', range: [string, string]): void
}>()

// Date range options
const dateRangeOptions = [
  { label: 'Current Month', value: 'current-month' },
  { label: 'Previous Month', value: 'previous-month' },
  { label: 'Current Year', value: 'current-year' },
  { label: 'All Time', value: 'all-time' }
]

// Selected date range
const selectedDateRange = ref('current-month')

// Actual date range
const dateRange = computed<[string, string] | undefined>(() => {
  switch (selectedDateRange.value) {
    case 'current-month':
      return getCurrentMonthRange()
    case 'previous-month':
      return getPreviousMonthRange()
    case 'current-year':
      return getCurrentYearRange()
    case 'all-time':
      return undefined
    default:
      return getCurrentMonthRange()
  }
})

// Data
const expenses = ref<Expense[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

// Fetch expense data
const fetchExpenses = async () => {
  if (!props.associationId) return

  try {
    loading.value = true
    error.value = null

    // Get date range based on selection
    const [startDate, endDate] = dateRange.value || [undefined, undefined]

    const response = await expenseApi.getExpenses(
      props.associationId,
      startDate,
      endDate
    )

    expenses.value = response.data
  } catch (err) {
    console.error('Error fetching expenses for summary:', err)
    error.value = 'Failed to load expense data'
  } finally {
    loading.value = false
  }
}

// Handle date range change
const handleDateRangeChange = (value: string) => {
  selectedDateRange.value = value
}

// Computed statistics
const totalExpenses = computed(() => {
  return expenses.value.reduce((sum, expense) => sum + expense.amount, 0)
})

const expenseCount = computed(() => {
  return expenses.value.length
})

const averageExpense = computed(() => {
  if (expenses.value.length === 0) return 0
  return totalExpenses.value / expenses.value.length
})

// Group expenses by category
const expensesByCategory = computed(() => {
  const grouped = expenses.value.reduce((acc, expense) => {
    const categoryKey = expense.category_id
    if (!acc[categoryKey]) {
      acc[categoryKey] = {
        id: expense.category_id,
        name: `${expense.category_type} - ${expense.category_name}`,
        total: 0,
        count: 0
      }
    }

    acc[categoryKey].total += expense.amount
    acc[categoryKey].count += 1

    return acc
  }, {} as Record<number, { id: number, name: string, total: number, count: number }>)

  // Convert to array and sort by total amount (descending)
  return Object.values(grouped).sort((a, b) => b.total - a.total)
})

// Top 3 categories
const topCategories = computed(() => {
  return expensesByCategory.value.slice(0, 3)
})

// Watch for prop or selection changes to refresh data
watch(
  [() => props.associationId, dateRange],
  () => {
    fetchExpenses()
  },
  { immediate: true }
)
</script>

<template>
  <NCard title="Expense Summary" class="expense-summary">
    <template #header-extra>
      <NSelect
        v-model:value="selectedDateRange"
        :options="dateRangeOptions"
        @update:value="handleDateRangeChange"
        style="width: 150px"
      />
    </template>

    <NSpin :show="loading">
      <div v-if="error" class="error">{{ error }}</div>

      <template v-else>
        <NSpace justify="space-around" align="center">
          <NStatistic label="Total Expenses" :value="formatCurrency(totalExpenses)" />
          <NStatistic label="Number of Expenses" :value="expenseCount" />
          <NStatistic label="Average Expense" :value="formatCurrency(averageExpense)" />
        </NSpace>

        <NDivider title-placement="left">Top Categories</NDivider>

        <div v-if="topCategories.length > 0" class="top-categories">
          <div v-for="category in topCategories" :key="category.id" class="category-item">
            <NGradientText :size="16">{{ category.name }}</NGradientText>
            <div class="category-stats">
              <span>{{ formatCurrency(category.total) }}</span>
              <span class="category-count">({{ category.count }} expenses)</span>
            </div>
          </div>
        </div>
        <div v-else class="no-data">No category data available</div>
      </template>
    </NSpin>
  </NCard>
</template>

<style scoped>
.expense-summary {
  margin-bottom: 20px;
}

.error {
  color: #e03;
  text-align: center;
  padding: 1rem;
}

.top-categories {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.category-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background-color: rgba(0, 0, 0, 0.02);
  border-radius: 4px;
}

.category-stats {
  font-weight: 600;
}

.category-count {
  margin-left: 8px;
  font-size: 0.85em;
  opacity: 0.7;
  font-weight: normal;
}

.no-data {
  text-align: center;
  padding: 1rem;
  color: #999;
}
</style>
