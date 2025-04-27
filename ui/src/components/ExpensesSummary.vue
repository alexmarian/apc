<script setup lang="ts">
import { computed } from 'vue'
import { NCard, NDivider, NEmpty, NGradientText, NSpace, NStatistic } from 'naive-ui'
import type { Expense } from '@/types/api'
import { formatCurrency } from '@/utils/formatters'

// Props
const props = defineProps<{
  expenses: Expense[]
}>()

// Computed statistics
const totalExpenses = computed(() => {
  return props.expenses.reduce((sum, expense) => sum + expense.amount, 0)
})

const expenseCount = computed(() => {
  return props.expenses.length
})

const averageExpense = computed(() => {
  if (props.expenses.length === 0) return 0
  return totalExpenses.value / props.expenses.length
})

// Group expenses by category
const expensesByCategory = computed(() => {
  if (!props.expenses || props.expenses.length === 0) return []
  const grouped = props.expenses.reduce((acc, expense) => {
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

const topCategories = computed(() => {
  return expensesByCategory.value.slice(0, 3)

})
</script>

<template>
  <NCard title="Expense Summary" class="expense-summary">
    <template v-if="props.expenses.length === 0">
      <NEmpty description="No expenses found for the selected filters" />
    </template>

    <template v-else>
      <NSpace justify="space-around" align="center">
        <NStatistic label="Total Expenses" :value="formatCurrency(totalExpenses)" />
        <NStatistic label="Number of Expenses" :value="expenseCount" />
        <NStatistic label="Average Expense" :value="formatCurrency(averageExpense)" />
      </NSpace>

      <NDivider title-placement="left">Top Categories</NDivider>

      <div v-if="topCategories.length > 0" class="top-categories">
        <div v-for="category in topCategories" :key="category.id" class="category-item">
          <NText :size="16" type="success">{{ category.name }}</NText>
          <div class="category-stats">
            <span>{{ formatCurrency(category.total) }}</span>
            <span class="category-count">({{ category.count }} expenses)</span>
          </div>
        </div>
      </div>
      <div v-else class="no-data">No category data available</div>
    </template>
  </NCard>
</template>

<style scoped>
:root {
  --background-color: #f9f9f9;
  --border-color: #ddd;
  --text-color: #555;
}

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
  background-color: var(--background-color);
  border: 1px solid var(--border-color);
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
  color: var(--text-color);
  opacity: 0.7;
}
</style>
