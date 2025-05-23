<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NButton, NSpace, NPageHeader, NSpin, NAlert, NModal } from 'naive-ui'
import UnitDetails from '@/components/UnitDetails.vue'
import UnitForm from '@/components/UnitForm.vue'
import OwnerForm from '@/components/OwnerForm.vue'
import { useRoute, useRouter } from 'vue-router'
import { unitApi } from '@/services/api'
import type { Unit, Owner } from '@/types/api'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const refreshKey = ref(0)

// Extract route params and query params
const unitId = ref<number>(parseInt(route.params.unitId as string))
const associationId = ref<number | null>(
  route.query.associationId ? parseInt(route.query.associationId as string) : null
)
const buildingId = ref<number | null>(
  route.query.buildingId ? parseInt(route.query.buildingId as string) : null
)

// UI state
const loading = ref(true)
const error = ref<string | null>(null)
const showEditForm = ref(false)
const showOwnerForm = ref(false)
const editingOwnerId = ref<number | null>(null)

const verifyParams = async () => {
  // Only fetch if we have all required IDs
  if (!associationId.value || !buildingId.value) {
    error.value = t('units.missingIds', 'Missing association or building ID. Please ensure these are provided in the URL.')
    loading.value = false
    return
  }

  loading.value = true
  error.value = null

  try {
    // Use the existing API method that requires all three IDs
    await unitApi.getUnit(
      associationId.value,
      buildingId.value,
      unitId.value
    )

    // If the API call is successful, then we have valid IDs

  } catch (err) {
    error.value = err instanceof Error ? err.message : t('common.error', 'Error fetching unit information')
    console.error('Error fetching unit data:', err)
  } finally {
    loading.value = false
  }
}

// Methods for unit editing
const handleEditUnit = () => {
  // Toggle the edit form instead of navigating away
  showEditForm.value = true
}

const handleUnitSaved = (savedUnit: Unit) => {
  console.log('Unit saved:', savedUnit)

  // Close the edit form
  showEditForm.value = false

  // Refresh the unit details to show updated info
  refreshKey.value++
}

const handleFormCancelled = () => {
  // Return to details view without refreshing
  showEditForm.value = false
}

// Methods for owner editing
const handleEditOwner = (ownerId: number) => {
  editingOwnerId.value = ownerId
  showOwnerForm.value = true
}

const handleOwnerFormSaved = (savedOwner: Owner) => {
  console.log('Owner saved:', savedOwner)

  // Close the modal
  showOwnerForm.value = false
  editingOwnerId.value = null

  // Refresh the unit details to show updated owner info
  refreshKey.value++
}

const handleOwnerFormCancelled = () => {
  // Just close the form without refreshing
  showOwnerForm.value = false
  editingOwnerId.value = null
}

const handleBackToUnits = () => {
  // Navigate back to units list
  if (associationId.value && buildingId.value) {
    router.push({
      path: '/units',
      query: {
        associationId: associationId.value.toString(),
        buildingId: buildingId.value.toString()
      }
    })
  } else {
    router.push('/units')
  }
}

onMounted(() => {
  verifyParams()
})
</script>

<template>
  <div class="unit-details-page">
    <NPageHeader>
      <template #title>
        {{ showEditForm ? t('units.editUnit', 'Edit Unit') : t('units.details', 'Unit Details') }}
      </template>
      <template #extra>
        <NSpace>
          <NButton @click="handleBackToUnits">
            {{ t('units.backToList', 'Back to Units List') }}
          </NButton>
          <NButton
            v-if="!showEditForm"
            type="primary"
            @click="handleEditUnit"
            :disabled="loading || error !== null"
          >
            {{ t('units.editUnit', 'Edit Unit') }}
          </NButton>
        </NSpace>
      </template>
    </NPageHeader>

    <NSpin :show="loading">
      <NCard v-if="error" style="margin-top: 16px;">
        <NAlert type="error" :title="t('common.error', 'Error')">
          {{ error }}
        </NAlert>
        <div style="text-align: center; padding: 16px;">
          <NButton @click="verifyParams">{{ t('common.retry', 'Retry') }}</NButton>
          <NButton @click="handleBackToUnits" style="margin-left: 16px;">
            {{ t('units.backToList', 'Return to Units List') }}
          </NButton>
        </div>
      </NCard>

      <div v-else-if="associationId && buildingId" style="margin-top: 16px;">
        <!-- Show edit form or details based on state -->
        <NCard v-if="showEditForm">
          <UnitForm
            :association-id="associationId"
            :building-id="buildingId"
            :unit-id="unitId"
            @saved="handleUnitSaved"
            @cancelled="handleFormCancelled"
          />
        </NCard>
        <NCard v-else>
          <UnitDetails
            :key="refreshKey"
            :association-id="associationId"
            :building-id="buildingId"
            :unit-id="unitId"
            @edit-owner="handleEditOwner"
            @edit-unit="handleEditUnit"
          />
        </NCard>
      </div>
    </NSpin>

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
.unit-details-page {
  width: 100%;
  margin: 0 auto;
}
</style>
