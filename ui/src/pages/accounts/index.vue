<script setup lang="ts">
import { ref, computed } from 'vue'
import { NCard, NButton, NSpace, NPageHeader, NGrid, NGridItem, NDropdown, useMessage } from 'naive-ui'
import AccountsList from '@/components/AccountsList.vue'
import AccountForm from '@/components/AccountForm.vue'
import AssociationSelector from '@/components/AssociationSelector.vue'

// Setup Naive UI message system
const message = useMessage()

// Association selector
const associationId = ref<number | null>(null)

// UI state
const showForm = ref(false)
const editingAccountId = ref<number | undefined>(undefined)

// Computed properties
const formTitle = computed(() => {
  return editingAccountId.value ? 'Edit Account' : 'Create New Account'
})

// Methods
const handleCreateAccount = () => {
  if (!associationId.value) {
    message.error('Please select an association first')
    return
  }

  editingAccountId.value = undefined
  showForm.value = true
}

const handleEditAccount = (accountId: number) => {
  editingAccountId.value = accountId
  showForm.value = true
}

const handleFormSaved = () => {
  showForm.value = false
  // Show success message
  message.success(`Account ${editingAccountId.value ? 'updated' : 'created'} successfully`)
  // In a real app, you would refresh the accounts list here or update the local state
  // For now, just reload the page after a short delay
  setTimeout(() => {
    location.reload()
  }, 1000)
}

const handleFormCancelled = () => {
  showForm.value = false
}
</script>

<template>
  <div class="accounts-view">
    <NPageHeader>
      <template #title>
        Account Management
      </template>

      <template #header>
        <div style="margin-bottom: 12px;">
          <AssociationSelector v-model:associationId="associationId" />
        </div>
      </template>

      <template #extra>
        <NButton
          v-if="!showForm"
          type="primary"
          @click="handleCreateAccount"
          :disabled="!associationId"
        >
          Create New Account
        </NButton>
      </template>
    </NPageHeader>

    <NCard v-if="showForm && associationId">
      <AccountForm
        :association-id="associationId"
        :account-id="editingAccountId"
        @saved="handleFormSaved"
        @cancelled="handleFormCancelled"
      />
    </NCard>

    <NCard v-else-if="associationId" style="margin-top: 16px;">
      <AccountsList
        :association-id="associationId"
        @edit="handleEditAccount"
      />
    </NCard>

    <NCard v-else style="margin-top: 16px;">
      <div style="text-align: center; padding: 32px;">
        <p>Please select an association to manage accounts</p>
      </div>
    </NCard>
  </div>
</template>

<style scoped>
.accounts-view {
  max-width: 100%; /* Change from 1200px to 100% */
  margin: 0 auto;
}
</style>
