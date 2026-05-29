<script setup lang="ts">
import { ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { NCard, NModal } from 'naive-ui'
import AccountsList from '@/components/AccountsList.vue'
import AccountForm from '@/components/AccountForm.vue'
import { useAssociationStore } from '@/stores/association'
import type { Account } from '@/types/api'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const { associationId } = storeToRefs(useAssociationStore())
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

</script>

<template>
  <div class="accounts-page">
    <div v-if="canShowAccounts">
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
