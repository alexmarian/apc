<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  NCard,
  NButton,
  NSpace,
  NPageHeader,
  NGrid,
  NGridItem,
  NDropdown,
  useMessage
} from 'naive-ui'
import AccountsList from '@/components/AccountsList.vue'
import AccountForm from '@/components/AccountForm.vue'
import AssociationSelector from '@/components/AssociationSelector.vue'
import { useI18n } from 'vue-i18n'

const message = useMessage()
const { t } = useI18n()
const associationId = ref<number | null>(null)

const showForm = ref(false)
const editingAccountId = ref<number | undefined>(undefined)

const formTitle = computed(() => {
  return editingAccountId.value
    ? t('accounts.editAccount', 'Edit Account')
    : t('accounts.createNew', 'Create New Account')
})

const handleCreateAccount = () => {
  if (!associationId.value) {
    message.error(t('accounts.selectAssociation', 'Please select an association first'))
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
  if (editingAccountId.value) {
    message.success(t('accounts.accountUpdated', 'Account updated successfully'))
  } else {
    message.success(t('accounts.accountCreated', 'Account created successfully'))
  }
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
        {{ t('accounts.title', 'Account Management') }}
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
          {{ t('accounts.createNew', 'Create New Account') }}
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
        <p>{{ t('accounts.selectAssociation', 'Please select an association to manage accounts') }}</p>
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
