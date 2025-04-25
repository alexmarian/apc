// Format date range for display
const formattedDateRange = computed(() => {
if (!dateRange.value) return 'All dates'

const start = new Date(dateRange.value[0]).toLocaleDateString()
const end = new Date(dateRange.value[1]).toLocaleDateString()

return `${start} - ${end}`
})<script setup lang="ts">
import { ref, onMounted, h, computed, watch } from 'vue'
import { NDataTable, NButton, NSpace, NEmpty, NSpin, NAlert, useMessage, NDatePicker } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { expenseApi } from '@/services/api'
import type { Expense } from '@/types/api'
import { formatCurrency } from '@/utils/formatters'
import CategorySelector from '@/components/CategorySelector.vue'

// Props
const props = defineProps<{
  associationId: number
}>()

// Emits
const emit = defineEmits<{
  (e: 'edit', expenseId: number): void
}>()

// Data
const expenses = ref<Expense[]>([])
const loading = ref<boolean>(true)
const error = ref<string | null>(null)
const message = useMessage()

// Filters
const dateRange = ref<[number, number] | null>(null)
const selectedCategory = ref<number | null>(null)

// Table columns definition
const columns = ref<DataTableColumns<Expense>>([
  {
    title: 'Date',
    key: 'date',
    sorter: 'default',
    render(row) {
      return new Date(row.date).toLocaleDateString()
    }
  },
  {
    title: 'Amount',
    key: 'amount',
    sorter: (a, b) => a.amount - b.amount,
    render(row) {
      return formatCurrency(row.amount)
    }
  },
  {
    title: 'Description',
    key: 'description'
  },
  {
    title: 'Destination',
    key: 'destination'
  },
  {
    title: 'Category',
    key: 'category_name',
    render(row) {
      return `${row.category_type} - ${row.category_family} - ${row.category_name}`
    }
  },
  {
    title: 'Account',
    key: 'account_number',
    render(row) {
      return `${row.account_number} - ${row.account_name || ''}`
    }
  },
  {
    title: 'Actions',
    key: 'actions',
    render(row) {
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
              { default: () => 'Edit' }
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
              { default: () => 'Delete' }
            )
          ]
        }
      )
    }
  }
])

// Fetch expenses
const fetchExpenses = async () => {
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

    // Filter by category if selected
    if (selectedCategory.value) {
      expenses.value = response.data.filter(expense => expense.category_id === selectedCategory.value)
    } else {
      expenses.value = response.data
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown error occurred'
    console.error('Error fetching expenses:', err)
  } finally {
    loading.value = false
  }
}

// Delete expense
const confirmDeleteExpense = (expenseId: number) => {
  if (window.confirm('Are you sure you want to delete this expense?')) {
    deleteExpense(expenseId)
  }
}

const deleteExpense = async (expenseId: number) => {
  try {
    await expenseApi.deleteExpense(props.associationId, expenseId)

    // Remove from local array
    expenses.value = expenses.value.filter(expense => expense.id !== expenseId)

    message.success('Expense deleted successfully')
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : 'Unknown error occurred'
    error.value = errorMessage
    console.error('Error deleting expense:', err)
    message.error('Failed to delete expense: ' + errorMessage)
  }
}

// Calculate total amount for the filtered expenses
const totalAmount = computed(() => {
  return expenses.value.reduce((sum, expense) => sum + expense.amount, 0)
})

// Watch for changes in filters and refresh data
watch([dateRange, selectedCategory], () => {
  fetchExpenses()
})

// Reset filters
const resetFilters = () => {
  dateRange.value = null
  selectedCategory.value = null
  fetchExpenses()
}

// Load data on component mount
onMounted(() => {
  fetchExpenses()
})
</script>

<template>
  <div class="expenses-list">
    <h2>Expenses</h2>

    <!-- Filters -->
    <div class="filters">
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
          <label>Category:</label>
          <CategorySelector
            v-model:modelValue="selectedCategory"
            :association-id="props.associationId"
            placeholder="Select Category"
            :include-all-option="true"
            style="width: 240px"
          />
        </div>
        <NButton @click="resetFilters">Reset Filters</NButton>
      </NSpace>
    </div>

    <NSpin :show="loading">
      <NAlert v-if="error" type="error" title="Error" closable>
        {{ error }}
        <template #action>
          <NButton @click="fetchExpenses">Retry</NButton>
        </template>
      </NAlert>

      <div v-if="expenses.length > 0" class="summary">
        <div>
          <span class="date-range-label">Period: {{ formattedDateRange }}</span>
        </div>
        <strong>Total Amount: {{ formatCurrency(totalAmount) }}</strong>
      </div>

      <NDataTable
        :columns="columns"
        :data="expenses"
        :bordered="false"
        :single-line="false"
        :pagination="{
          pageSize: 10
        }"
      >
        <template #empty>
          <NEmpty description="No expenses found">
            <template #extra>
              <p>Create a new expense to get started.</p>
            </template>
          </NEmpty>
        </template>
      </NDataTable>
    </NSpin>
  </div>
</template>

<style scoped>
.expenses-list {
  margin: 2rem 0;
}

.filters {
  margin-bottom: 1.5rem;
  padding: 1rem;
  background-color: #f9f9f9;
  border-radius: 4px;
}

.summary {
  margin: 1rem 0;
  padding: 0.5rem;
  font-size: 1.1rem;
  text-align: right;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.date-range-label {
  font-size: 0.9rem;
  color: #666;
}
</style>
