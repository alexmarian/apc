<script setup lang="ts">
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import { NPageHeader, NCard, NModal } from 'naive-ui'
import OwnersReport from '@/components/OwnersReport.vue'
import OwnerForm from '@/components/OwnerForm.vue'
import { useAssociationStore } from '@/stores/association'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const { associationId } = storeToRefs(useAssociationStore())

// Owner editing state
const showOwnerForm = ref(false)
const editingOwnerId = ref<number | null>(null)
const refreshKey = ref(0) // Key to force OwnersReport to re-render

// Methods for owner editing
const handleEditOwner = (ownerId: number) => {
  editingOwnerId.value = ownerId
  showOwnerForm.value = true
}

const handleOwnerFormSaved = () => {
  // Close the owner form
  showOwnerForm.value = false
  editingOwnerId.value = null

  // Force a re-render of OwnersReport to refresh data
  refreshKey.value++
}

const handleOwnerFormCancelled = () => {
  // Just close the form without refreshing
  showOwnerForm.value = false
  editingOwnerId.value = null
}
</script>

<template>
  <div class="owners-report-page">
    <NPageHeader>
      <template #title>
        {{ t('owners.report', 'Owners Report') }}
      </template>

    </NPageHeader>

    <div class="content">
      <div>
        <OwnersReport
          :key="refreshKey"
          :association-id="associationId"
          @edit-owner="handleEditOwner"
        />
      </div>
    </div>

    <!-- Owner Edit Modal -->
    <NModal
      v-model:show="showOwnerForm"
      style="width: 650px"
      preset="card"
      :title="t('owners.editOwner', 'Edit Owner')"
      :mask-closable="false"
    >
      <OwnerForm
        v-if="associationId && editingOwnerId"
        :association-id="associationId"
        :owner-id="editingOwnerId"
        @saved="handleOwnerFormSaved"
        @cancelled="handleOwnerFormCancelled"
      />
    </NModal>
  </div>
</template>

<style scoped>
.owners-report-page {
  width: 100%;
}

.content {
  margin-top: 16px;
}
</style>
