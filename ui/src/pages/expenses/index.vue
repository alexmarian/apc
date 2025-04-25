<script setup lang="ts">
import { computed, ref } from 'vue'
import { NButton, NCard, NPageHeader, NSpace, useMessage } from 'naive-ui'
import ExpensesList from '@/components/ExpensesList.vue'
import ExpenseForm from '@/components/ExpenseForm.vue'
import ExpensesSummary from '@/components/ExpensesSummary.vue'
import AssociationSelector from '@/components/AssociationSelector.vue'
import type { Expense } from '@/types/api.ts'

// Setup Naive UI message system
const message = useMessage()

// Association selector
const associationId = ref<number | null>(null)

// UI state
const showForm = ref(false)
const editingExpenseId = ref<number | undefined>(undefined)
const showSummary = ref(true)

const displayedExpenses = ref<Expense[] | null>(null)

const setDisplayedExpenses = (expenses: Expense[]) => {
  displayedExpenses.value = expenses
}

// Computed properties
const formTitle = computed(() => {
  return editingExpenseId.value ? 'Edit Expense' : 'Create New Expense'
})

// Methods
const handleCreateExpense = () => {
  if (!associationId.value) {
    message.error('Please select an association first')
    return
  }

  editingExpenseId.value = undefined
  showForm.value = true
}

const handleEditExpense = (expenseId: number) => {
  editingExpenseId.value = expenseId
  showForm.value = true
}

const handleFormSaved = () => {
  showForm.value = false
  // Show success message
  message.success(`Expense ${editingExpenseId.value ? 'updated' : 'created'} successfully`)
  // check on how to trigger reload
}

const handleFormCancelled = () => {
  showForm.value = false
}

const toggleSummary = () => {
  showSummary.value = !showSummary.value
}
</script>

<template>
  <div class="expenses-view">
    <NPageHeader>
      <template #title>
        Expense Management
      </template>

      <template #header>
        <div style="margin-bottom: 12px;">
          <AssociationSelector v-model:modelValue="associationId" />
        </div>
      </template>

      <template #extra>
        <NSpace>
          <NButton
            v-if="!showForm && associationId"
            secondary
            @click="toggleSummary"
          >
            {{ showSummary ? 'Hide Summary' : 'Show Summary' }}
          </NButton>
          <NButton
            v-if="!showForm"
            type="primary"
            @click="handleCreateExpense"
            :disabled="!associationId"
          >
            Create New Expense
          </NButton>
        </NSpace>
      </template>
    </NPageHeader>

    <div v-if="!associationId">
      <NCard style="margin-top: 16px;">
        <div style="text-align: center; padding: 32px;">
          <p>Please select an association to manage expenses</p>
        </div>
      </NCard>
    </div>

    <div v-else-if="showForm">
      <NCard style="margin-top: 16px;">
        <ExpenseForm
          :association-id="associationId"
          :expense-id="editingExpenseId"
          @saved="handleFormSaved"
          @cancelled="handleFormCancelled"
        />
      </NCard>
    </div>
    <div v-else>
      <!-- List comes first in vertical layout -->
      <NCard style="margin-top: 16px;">
        <ExpensesList
          :association-id="associationId"
          @edit="handleEditExpense"
          @expenses-rendered="setDisplayedExpenses"
        />
      </NCard>

      <!-- Summary is below the list and can be toggled -->
      <div v-if="showSummary" style="margin-top: 16px;">
        <ExpensesSummary v-if="displayedExpenses"
          :expenses="displayedExpenses"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.expenses-view {
  width: 100%;
  margin: 0 auto;
}
</style>
