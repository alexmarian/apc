<script setup lang="ts">
import { ref, computed } from 'vue'
import { NCard, NPageHeader, NModal } from 'naive-ui'
import AccountsList from '@/components/AccountsList.vue'
import AccountForm from '@/components/AccountForm.vue'
import AssociationSelector from '@/components/AssociationSelector.vue'
import type { Account } from '@/types/api'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

// State
const associationId = ref<number | null>(null)
const showAccountModal = ref(false)
const editingAccountId = ref<number | undefined>(undefined)

// Reference to the AccountsList component
const accountsListRef = ref<InstanceType<typeof AccountsList> | null>(null)

// Computed properties
const modalTitle = computed(() => {
  return editingAccountId.value ? t('accounts.editAccount') : t('accounts.createNew')
})

const canShowAccounts = computed(() => {
  return associationId.value !== null
})

// Methods
const handleEditAccount = (accountId: number) => {
  editingAccountId.value = accountId
  showAccountModal.value = true
}

const handleCreateAccount = () => {
  editingAccountId.value = undefined
  showAccountModal.value = true
}

const handleAccountSaved = (savedAccount: Account) => {
  console.log('Account saved:', savedAccount)

  // Update or add the account in the list without reloading
  if (accountsListRef.value) {
    if (editingAccountId.value) {
      // Update existing account
      accountsListRef.value.updateAccount(savedAccount)
    } else {
      // Add new account
      accountsListRef.value.addAccount(savedAccount)
    }
  }

  // Close the modal
  showAccountModal.value = false
  editingAccountId.value = undefined
}

const handleAccountFormCancelled = () => {
  showAccountModal.value = false
  editingAccountId.value = undefined
}

const handleAssociationChanged = (newAssociationId: number | null) => {
  associationId.value = newAssociationId
  // Close any open modals when association changes
  showAccountModal.value = false
  editingAccountId.value = undefined
}
</script>

<template>
  <div class="accounts-page">
    <NPageHeader>
      <template #title>
        {{ t('accounts.title') }}
      </template>

      <template #header>
        <div style="margin-bottom: 12px;">
          <AssociationSelector
            v-model:associationId="associationId"
            @update:associationId="handleAssociationChanged"
          />
        </div>
      </template>
    </NPageHeader>

    <div v-if="!associationId">
      <NCard style="margin-top: 16px;">
        <div style="text-align: center; padding: 32px;">
          <p>{{ t('accounts.selectAssociation') }}</p>
        </div>
      </NCard>
    </div>

    <div v-else-if="canShowAccounts">
      <!-- Accounts List -->
      <AccountsList
        ref="accountsListRef"
        :association-id="associationId"
        @edit="handleEditAccount"
        @create="handleCreateAccount"
      />
    </div>

    <!-- Account Edit/Create Modal -->
    <NModal
      v-model:show="showAccountModal"
      style="width: 650px"
      preset="card"
      :title="modalTitle"
      :mask-closable="false"
      :close-on-esc="true"
    >
      <AccountForm
        v-if="showAccountModal && associationId"
        :association-id="associationId"
        :account-id="editingAccountId"
        @saved="handleAccountSaved"
        @cancelled="handleAccountFormCancelled"
      />
    </NModal>
  </div>
</template>

<style scoped>
.accounts-page {
  width: 100%;
}
</style>
