<script setup lang="ts">
import { computed, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { NButton, NSpace, useMessage, NModal } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import PageShell from '@/components/PageShell.vue'
import ExpensesList from '@/components/ExpensesList.vue'
import ExpenseForm from '@/components/ExpenseForm.vue'
import ExpensesSummary from '@/components/ExpensesSummary.vue'
import { useAssociationStore } from '@/stores/association'
import type { Expense } from '@/types/api.ts'

// Setup i18n
const { t } = useI18n()

// Setup Naive UI message system
const message = useMessage()

const { associationId } = storeToRefs(useAssociationStore())

// UI state
const showExpenseModal = ref(false)
const editingExpenseId = ref<number | undefined>(undefined)
const showSummary = ref(true)

// Reference to the ExpensesList component
const expensesListRef = ref<InstanceType<typeof ExpensesList> | null>(null)

// Filters persistence
const dateRange = ref<[number, number] | null>(null)
const selectedCategory = ref<number | null>(null)

const displayedExpenses = ref<Expense[] | null>(null)

const setDisplayedExpenses = (expenses: Expense[]) => {
  displayedExpenses.value = expenses
}

// Computed properties
const modalTitle = computed(() => {
  return editingExpenseId.value
    ? t('expenses.editExpense', 'Edit Expense')
    : t('expenses.createNew', 'Create New Expense')
})

const canShowExpenses = computed(() => {
  return associationId.value !== null
})

// Methods
const handleCreateExpense = () => {
  if (!associationId.value) {
    message.error(t('expenses.selectAssociation', 'Please select an association first'))
    return
  }

  editingExpenseId.value = undefined
  showExpenseModal.value = true
}

const handleEditExpense = (expenseId: number) => {
  editingExpenseId.value = expenseId
  showExpenseModal.value = true
}

const handleExpenseSaved = (savedExpense: Expense) => {
  console.log('Expense saved:', savedExpense)

  // Update or add the expense in the list without reloading
  if (expensesListRef.value) {
    if (editingExpenseId.value) {
      // Update existing expense
      expensesListRef.value.updateExpense(savedExpense)
    } else {
      // Add new expense
      expensesListRef.value.addExpense(savedExpense)
    }
  }

  // Close the modal
  showExpenseModal.value = false
  editingExpenseId.value = undefined
}

const handleExpenseFormCancelled = () => {
  showExpenseModal.value = false
  editingExpenseId.value = undefined
}

const toggleSummary = () => {
  showSummary.value = !showSummary.value
}
</script>

<template>
  <PageShell :title="t('expenses.title', 'Expense Management')">
    <template #actions>
      <NSpace>
        <NButton
          v-if="associationId"
          secondary
          @click="toggleSummary"
        >
          {{ showSummary ? t('common.hide', 'Hide') + ' ' + t('expenses.summary', 'Summary') : t('common.show', 'Show') + ' ' + t('expenses.summary', 'Summary') }}
        </NButton>
        <NButton
          type="primary"
          @click="handleCreateExpense"
          :disabled="!associationId"
        >
          {{ t('expenses.createNew', 'Create New Expense') }}
        </NButton>
      </NSpace>
    </template>

    <div v-if="canShowExpenses">
      <!-- Expenses List -->
      <ExpensesList
        ref="expensesListRef"
        :association-id="associationId"
        v-model:date-range="dateRange"
        v-model:selected-category="selectedCategory"
        @edit="handleEditExpense"
        @create="handleCreateExpense"
        @expenses-rendered="setDisplayedExpenses"
      />

      <!-- Summary is below the list and can be toggled -->
      <div v-if="showSummary" class="expenses-view__summary">
        <ExpensesSummary
          v-if="displayedExpenses"
          :expenses="displayedExpenses"
        />
      </div>
    </div>

    <!-- Expense Edit/Create Modal -->
    <NModal
      v-model:show="showExpenseModal"
      style="width: 650px"
      preset="card"
      :title="modalTitle"
      :mask-closable="false"
      :close-on-esc="true"
    >
      <ExpenseForm
        v-if="showExpenseModal && associationId"
        :association-id="associationId"
        :expense-id="editingExpenseId"
        @saved="handleExpenseSaved"
        @cancelled="handleExpenseFormCancelled"
      />
    </NModal>
  </PageShell>
</template>

<style scoped>
.expenses-view__summary {
  margin-top: 16px;
}
</style>
