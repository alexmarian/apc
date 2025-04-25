<script setup lang="ts">
import { ref, computed } from 'vue'
import { NCard, NButton, NPageHeader, NSpace, useMessage, NGrid, NGridItem } from 'naive-ui'
import ExpensesList from '@/components/ExpensesList.vue'
import ExpenseForm from '@/components/ExpenseForm.vue'
import ExpensesSummary from '@/components/ExpensesSummary.vue'
import AssociationSelector from '@/components/AssociationSelector.vue'

// Setup Naive UI message system
const message = useMessage()

// Association selector
const associationId = ref<number | null>(null)

// UI state
const showForm = ref(false)
const editingExpenseId = ref<number | undefined>(undefined)

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
  // Reload the expenses list
  setTimeout(() => {
    location.reload()
  }, 1000)
}

const handleFormCancelled = () => {
  showForm.value = false
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
        <NButton
          v-if="!showForm"
          type="primary"
          @click="handleCreateExpense"
          :disabled="!associationId"
        >
          Create New Expense
        </NButton>
      </template>
    </NPageHeader>

    <div v-if="!associationId">
      <NCard style="margin-top: 16px;">
        <div style="text-align: center; padding: 32px;">
          <p>Please select an association to manage expenses</p>
        </div>
      </NCard>
    </div>

    <div v-else>
      <div v-if="showForm">
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
        <NCard style="margin-top: 16px;">
          <NGrid :cols="1" :x-gap="16" :y-gap="16">
            <NGridItem>
              <ExpensesSummary
                :association-id="associationId"
              />
            </NGridItem>
            <NGridItem>
              <ExpensesList
                :association-id="associationId"
                @edit="handleEditExpense"
              />
            </NGridItem>
          </NGrid>
        </NCard>
      </div>
    </div>
  </div>
</template>

<style scoped>
.expenses-view {
  width: 100%;
  margin: 0 auto;
}

@media (min-width: 1200px) {
  :deep(.n-grid) {
    grid-template-columns: 1fr 2fr !important;
  }
}
</style>
