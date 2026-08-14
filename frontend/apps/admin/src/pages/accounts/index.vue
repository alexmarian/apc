<script setup lang="ts">
import { ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { NButton, NModal } from 'naive-ui'
import PageShell from '@/components/PageShell.vue'
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
  // Update or add the account in the list without reloading
  if (accountsListRef.value) {
    if (editingAccountId.value) {
      accountsListRef.value.updateAccount(savedAccount)
    } else {
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
  <PageShell :title="t('accounts.title')">
    <template #actions>
      <NButton type="primary" :disabled="!associationId" @click="handleCreateAccount">
        {{ t('accounts.createNew') }}
      </NButton>
    </template>

    <AccountsList
      v-if="canShowAccounts"
      ref="accountsListRef"
      :association-id="associationId"
      @edit="handleEditAccount"
      @create="handleCreateAccount"
    />

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
  </PageShell>
</template>
