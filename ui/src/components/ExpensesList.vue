<script setup lang="ts">
import { ref, onMounted, h, computed, watch, inject } from 'vue'
import {
  NDataTable,
  NButton,
  NSpace,
  NEmpty,
  NSpin,
  NAlert,
  useMessage,
  NDatePicker,
  NSelect,
  NCard,
  NFlex,
  NText
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { expenseApi } from '@/services/api'
import type { Expense, ApiResponse } from '@/types/api'
import { formatCurrency } from '@/utils/formatters'
import CategorySelector from '@/components/CategorySelector.vue'
import { useI18n } from 'vue-i18n'
import LocalizedCategoryDisplay from './LocalizedCategoryDisplay.vue'

// Props
const props = defineProps<{
  associationId: number,
  dateRange?: [number, number] | null,
  selectedCategory?: number | null
}>()

// Emits
const emit = defineEmits<{
  (e: 'edit', expenseId: number): void
  (e: 'expenses-rendered', expenses: Expense[]): void
  (e: 'date-range-changed', newDateRange: [number, number] | null): void
  (e: 'category-changed', newCategory: number | null): void
}>()

// I18n
const { t } = useI18n()

// Data
const expenses = ref<Expense[]>([])
const loading = ref<boolean>(false)
const error = ref<string | null>(null)
const message = useMessage()

// Filters - use shared filters if available, otherwise use local state
const dateRange = ref<[number, number] | null>(props.dateRange ?? null)
const selectedCategory = ref<number | null>(props.selectedCategory ?? null)

const filteredExpenses = computed<Expense[]>(() => {
  if (selectedCategory.value) {
    return expenses.value.filter(expense => expense.category_id === selectedCategory.value)
  } else {
    return expenses.value
  }
})

// Table columns definition
const columns = ref<DataTableColumns<Expense>>([
  {
    title: t('expenses.date'),
    key: 'date',
    sorter: 'default',
    render(row: Expense) {
      return new Date(row.date).toLocaleDateString()
    }
  },
  {
    title: t('expenses.amount'),
    key: 'amount',
    sorter: (a: Expense, b: Expense) => a.amount - b.amount,
    render(row: Expense) {
      return formatCurrency(row.amount)
    }
  },
  {
    title: t('expenses.documentRef'),
    key: 'document_ref'
  },
  {
    title: t('expenses.description'),
    key: 'description'
  },
  {
    title: t('expenses.destination'),
    key: 'destination'
  },
  {
    title: t('categories.types.title', 'Type'),
    key: 'category_type',
    render(row: Expense) {
      return h('span', {}, t(`categories.types.${row.category_type}`))
    }
  },
  {
    title: t('categories.families.title', 'Family'),
    key: 'category_family',
    render(row: Expense) {
      return h('span', {}, t(`categories.families.${row.category_family}`))
    }
  },
  {
    title: t('categories.names.title', 'Category'),
    key: 'category_name',
    render(row: Expense) {
      return h('span', {}, t(`categories.names.${row.category_name}`))
    }
  },
  {
    title: t('expenses.account'),
    key: 'account_number',
    render(row: Expense) {
      return `${row.account_number} - ${row.account_name || ''}`
    }
  },
  {
    title: t('common.actions'),
    key: 'actions',
    render(row: Expense) {
      return h(
        NSpace,
        {
          justify: 'center',
          align: 'center'
        },
        {
          default: () => [
            h(
              NButton,
              {
                strong: true,
                secondary: true,
                type: 'info',
                size: 'small',
                onClick: () => emit('edit', row.id)
              },
              { default: () => t('common.edit') }
            ),
            h(
              NButton,
              {
                strong: true,
                secondary: true,
                type: 'error',
                size: 'small',
                onClick: () => confirmDeleteExpense(row.id)
              },
              { default: () => t('common.delete') }
            )
          ]
        }
      )
    }
  }
])

// Format date range for display
const formattedDateRange = computed<string>(() => {
  if (!dateRange.value) return t('expenses.allDates', 'All dates')

  const start = new Date(dateRange.value[0]).toLocaleDateString()
  const end = new Date(dateRange.value[1]).toLocaleDateString()

  return `${start} - ${end}`
})

// Fetch expenses
const fetchExpenses = async (): Promise<void> => {
  if (!props.associationId) return

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

    const response = await expenseApi.getExpenses(props.associationId, startDate, endDate)
    expenses.value = response.data
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('common.error')
    console.error('Error fetching expenses:', err)
  } finally {
    loading.value = false
  }
}

// Delete expense
const confirmDeleteExpense = (expenseId: number): void => {
  if (window.confirm(t('expenses.confirmDelete'))) {
    deleteExpense(expenseId)
  }
}

const deleteExpense = async (expenseId: number): Promise<void> => {
  try {
    await expenseApi.deleteExpense(props.associationId, expenseId)

    // Remove from local array
    expenses.value = expenses.value.filter(expense => expense.id !== expenseId)

    message.success(t('expenses.expenseDeleted'))
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : t('common.error')
    error.value = errorMessage
    console.error('Error deleting expense:', err)
    message.error(t('common.error') + ': ' + errorMessage)
  }
}

// Calculate total amount for the filtered expenses
const totalAmount = computed<number>(() => {
  return filteredExpenses.value.reduce((sum, expense) => sum + expense.amount, 0)
})

// Watch for changes in filters and refresh data
watch(dateRange, (newDateRange) => {
  fetchExpenses()
  emit('date-range-changed', newDateRange)
})

watch(selectedCategory, (newCategory) => {
  emit('category-changed', newCategory)
})

watch(filteredExpenses, () => {
  emit('expenses-rendered', filteredExpenses.value)
})

// Reset filters
const resetFilters = (): void => {
  dateRange.value = null
  selectedCategory.value = null
}

onMounted(() => {
  fetchExpenses()
})
</script>

<template>
  <div class="expenses-list">
    <NCard style="margin-top: 16px;">
      <NFlex align="center" justify="start">
        <NText>{{ t('expenses.dateRange') }}:</NText>
        <NDatePicker
          v-model:value="dateRange"
          type="daterange"
          clearable
          style="width: 240px"
        />
        <NText>{{ t('expenses.category') }}:</NText>
        <CategorySelector
          v-model:modelValue="selectedCategory"
          :association-id="props.associationId"
          :placeholder="t('expenses.category')"
          :include-all-option="true"
          style="width: 360px"
        />
        <NButton @click="resetFilters">{{ t('expenses.resetFilters') }}</NButton>
      </NFlex>
    </NCard>
    <NCard style="margin-top: 16px;">
      <NSpin :show="loading">
        <NAlert v-if="error" type="error" :title="t('common.error')" closable>
          {{ error }}
          <NButton @click="fetchExpenses">{{ t('common.retry') }}</NButton>
        </NAlert>

        <div v-if="expenses.length > 0" class="summary">
          <div>
            <span class="date-range-label">{{ t('expenses.period', 'Period') }}: {{ formattedDateRange }}</span>
          </div>
          <strong>{{ t('expenses.totalAmount') }}: {{ formatCurrency(totalAmount) }}</strong>
        </div>

        <NDataTable
          :columns="columns"
          :data="filteredExpenses"
          :bordered="false"
          :single-line="false"
          :pagination="{
            pageSize: 10
          }"
        >
          <template #empty>
            <NEmpty :description="t('expenses.noExpenses')">
              <template #extra>
                <p>{{ t('expenses.createToStart') }}</p>
              </template>
            </NEmpty>
          </template>
        </NDataTable>
      </NSpin>
    </NCard>
  </div>
</template>

<style scoped>
.expenses-list {
  margin: 1rem 0;
}

.filters {
  margin-bottom: 1.5rem;
  padding: 1rem;
  border-radius: 4px;
  background-color: var(--background-color);
  border: 1px solid var(--border-color);
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  align-items: center;
}

.filters > div {
  flex: 1;
  min-width: 200px;
}

.summary {
  margin: 1rem 0;
  padding: 0.5rem 1rem;
  font-size: 1.1rem;
  text-align: right;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: var(--background-color);
  border-radius: 4px;
  border: 1px solid var(--border-color);
}

.date-range-label {
  font-size: 0.9rem;
  color: var(--text-color);
  opacity: 0.8;
}
</style>
