<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NCard, NButton, NSpace, NPageHeader, NDivider, useMessage } from 'naive-ui'
import UnitsList from '@/components/UnitsList.vue'
import UnitForm from '@/components/UnitForm.vue'
import UnitReport from '@/components/UnitReport.vue'
import AssociationSelector from '@/components/AssociationSelector.vue'
import BuildingSelector from '@/components/BuildingSelector.vue'
import type { Unit } from '@/types/api'

// Setup Naive UI message system
const message = useMessage()

// Selectors
const associationId = ref<number | null>(null)
const buildingId = ref<number | null>(null)

// UI state
const showForm = ref(false)
const showFullReport = ref(false)
const editingUnitId = ref<number | undefined>(undefined)
const selectedUnitId = ref<number | null>(null)

// Filters persistence
const unitTypeFilter = ref<string | null>(null)
const searchQuery = ref<string | null>(null)

// For storing units loaded by the list component
const displayedUnits = ref<Unit[] | null>(null)

const setDisplayedUnits = (units: Unit[]) => {
  displayedUnits.value = units
}

// Computed properties
const formTitle = computed(() => {
  return editingUnitId.value ? 'Edit Unit' : 'Create New Unit'
})

const canShowUnits = computed(() => {
  return associationId.value !== null && buildingId.value !== null
})

// Methods
const handleEditUnit = (unitId: number) => {
  editingUnitId.value = unitId
  showForm.value = true
  showFullReport.value = false
}

const handleViewReport = (unitId: number) => {
  selectedUnitId.value = unitId
  showFullReport.value = true
  showForm.value = false
}

const handleFormSaved = () => {
  showForm.value = false
  // Show success message
  message.success(`Unit ${editingUnitId.value ? 'updated' : 'created'} successfully`)
  // Reload the units list
  // In a real implementation, you might want to update the list without a full reload
  window.location.reload()
}

const handleFormCancelled = () => {
  showForm.value = false
}

const handleBackFromReport = () => {
  showFullReport.value = false
  selectedUnitId.value = null
}
const handleBuildingIdUpdate = (newBuildingId: number) => {
  buildingId.value = newBuildingId
}

const handleEditOwner = (ownerId: number) => {
  // This is a placeholder for navigating to the owner edit page
  // You would implement this based on your application's routing
  message.info(`Navigate to edit owner with ID: ${ownerId}`)
}

// Clear buildingId when associationId changes
watch(associationId, () => {
  buildingId.value = null
  showForm.value = false
  showFullReport.value = false
  selectedUnitId.value = null
})
</script>

<template>
  <div class="units-view">
    <NPageHeader>
      <template #title>
        Unit Management
      </template>

      <template #header>
        <div style="margin-bottom: 12px;">
          <NSpace align="center">
            <AssociationSelector v-model:associationId="associationId" />
            <BuildingSelector
              v-model:building-id="buildingId"
              v-model:association-id="associationId"
              @update:building-id="handleBuildingIdUpdate"
            />
          </NSpace>
        </div>
      </template>

      <template #extra>
        <NSpace>
          <NButton
            v-if="showFullReport"
            @click="handleBackFromReport"
          >
            Back to Units List
          </NButton>
        </NSpace>
      </template>
    </NPageHeader>

    <div v-if="!associationId || !buildingId">
      <NCard style="margin-top: 16px;">
        <div style="text-align: center; padding: 32px;">
          <p>Please select an association and building to manage units</p>
        </div>
      </NCard>
    </div>

    <div v-else-if="showForm">
      <NCard style="margin-top: 16px;">
        <UnitForm
          :association-id="associationId"
          :building-id="buildingId"
          :unit-id="editingUnitId"
          @saved="handleFormSaved"
          @cancelled="handleFormCancelled"
        />
      </NCard>
    </div>

    <div v-else-if="showFullReport && selectedUnitId">
      <NCard style="margin-top: 16px;">
        <UnitReport
          :association-id="associationId"
          :building-id="buildingId"
          :unit-id="selectedUnitId"
          @edit-owner="handleEditOwner"
        />
      </NCard>
    </div>

    <div v-else-if="canShowUnits">
      <!-- Units list -->
        <UnitsList
          :association-id="associationId"
          :building-id="buildingId"
          :unit-type-filter="unitTypeFilter"
          :search-query="searchQuery"
          @edit="handleEditUnit"
          @view-report="handleViewReport"
          @units-rendered="setDisplayedUnits"
          @unit-type-changed="newUnitType => unitTypeFilter = newUnitType"
          @search-query-changed="newQuery => searchQuery = newQuery"
        />

      <!-- Unit Report Preview (shows when a unit is selected) -->
      <div v-if="selectedUnitId" style="margin-top: 16px;">
        <NCard title="Unit Summary">
          <UnitReport
            :association-id="associationId"
            :building-id="buildingId"
            :unit-id="selectedUnitId"
            :show-excerpt="true"
            @edit-owner="handleEditOwner"
          />
          <template #footer>
            <NSpace justify="end">
              <NButton
                type="primary"
                @click="showFullReport = true"
              >
                View Full Report
              </NButton>
            </NSpace>
          </template>
        </NCard>
      </div>
    </div>
  </div>
</template>

<style scoped>
.units-view {
  width: 100%;
  margin: 0 auto;
}
</style>
