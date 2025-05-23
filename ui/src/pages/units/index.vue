<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { NCard, NButton, NSpace, NPageHeader, NDivider, useMessage, NModal } from 'naive-ui'
import UnitsList from '@/components/UnitsList.vue'
import UnitForm from '@/components/UnitForm.vue'
import AssociationSelector from '@/components/AssociationSelector.vue'
import BuildingSelector from '@/components/BuildingSelector.vue'
import { useRouter, useRoute } from 'vue-router'
import type { Unit } from '@/types/api'
import { useI18n } from 'vue-i18n'

// Setup i18n
const { t } = useI18n()

// Setup Naive UI message system
const message = useMessage()
const router = useRouter()
const route = useRoute()

// Selectors
const associationId = ref<number | null>(null)
const buildingId = ref<number | null>(null)

// UI state
const showUnitEditModal = ref(false)
const editingUnitId = ref<number | undefined>(undefined)
const unitTypeFilter = ref<string | null>(null)
const searchQuery = ref<string | null>(null)

// Reference to the UnitsList component
const unitsListRef = ref<InstanceType<typeof UnitsList> | null>(null)

// For storing units loaded by the list component
const displayedUnits = ref<Unit[] | null>(null)

// Try to get associationId, buildingId, and editUnitId from query parameters
onMounted(() => {
  if (route.query.associationId) {
    associationId.value = parseInt(route.query.associationId as string)
  }
  if (route.query.buildingId) {
    buildingId.value = parseInt(route.query.buildingId as string)
  }
  if (route.query.editUnitId) {
    editingUnitId.value = parseInt(route.query.editUnitId as string)
    showUnitEditModal.value = true
  }
})

const setDisplayedUnits = (units: Unit[]) => {
  displayedUnits.value = units
}

// Computed properties
const formTitle = computed(() => {
  return editingUnitId.value ? t('units.editUnit', 'Edit Unit') : t('units.createUnit', 'Create New Unit')
})

const canShowUnits = computed(() => {
  return associationId.value !== null && buildingId.value !== null
})

// Methods
const handleEditUnit = (unitId: number) => {
  editingUnitId.value = unitId
  showUnitEditModal.value = true

  // Update URL to reflect edit mode
  router.replace({
    query: {
      ...route.query,
      editUnitId: unitId.toString()
    }
  })
}

const handleUnitSaved = (updatedUnit: Unit) => {
  console.log('Unit saved in parent:', updatedUnit)

  // Update the unit in the list without reloading the entire page
  if (unitsListRef.value) {
    unitsListRef.value.updateUnit(updatedUnit)
  }

  // Close the modal
  showUnitEditModal.value = false
  editingUnitId.value = undefined

  // Remove editUnitId from query params
  const newQuery = { ...route.query }
  delete newQuery.editUnitId
  router.replace({ query: newQuery })
}

const handleUnitFormCancelled = () => {
  showUnitEditModal.value = false
  editingUnitId.value = undefined

  // Remove editUnitId from query params
  const newQuery = { ...route.query }
  delete newQuery.editUnitId
  router.replace({ query: newQuery })
}

const handleBuildingIdUpdate = (newBuildingId: number) => {
  buildingId.value = newBuildingId

  // Update URL query parameters
  router.replace({
    query: {
      ...route.query,
      buildingId: newBuildingId.toString()
    }
  })
}

const handleAssociationIdUpdate = (newAssociationId: number) => {
  associationId.value = newAssociationId
  buildingId.value = null

  // Update URL query parameters
  router.replace({
    query: {
      ...route.query,
      associationId: newAssociationId.toString(),
      buildingId: undefined
    }
  })
}

// Clear buildingId when associationId changes
watch(associationId, () => {
  buildingId.value = null
  showUnitEditModal.value = false
  editingUnitId.value = undefined
})

// Watch for buildingId changes to close modal
watch(buildingId, () => {
  showUnitEditModal.value = false
  editingUnitId.value = undefined
})
</script>

<template>
  <div class="units-view">
    <NPageHeader>
      <template #title>
        {{ t('units.management', 'Unit Management') }}
      </template>

      <template #header>
        <div style="margin-bottom: 12px;">
          <NSpace align="center">
            <AssociationSelector
              v-model:associationId="associationId"
              @update:associationId="handleAssociationIdUpdate"
            />
            <BuildingSelector
              v-model:building-id="buildingId"
              v-model:association-id="associationId"
              @update:building-id="handleBuildingIdUpdate"
            />
          </NSpace>
        </div>
      </template>
    </NPageHeader>

    <div v-if="!associationId || !buildingId">
      <NCard style="margin-top: 16px;">
        <div style="text-align: center; padding: 32px;">
          <p>{{ t('units.selectAssociationAndBuilding', 'Please select an association and building to manage units') }}</p>
        </div>
      </NCard>
    </div>

    <div v-else-if="canShowUnits">
      <!-- Units list -->
      <UnitsList
        ref="unitsListRef"
        :association-id="associationId"
        :building-id="buildingId"
        :unit-type-filter="unitTypeFilter"
        :search-query="searchQuery"
        @edit="handleEditUnit"
        @units-rendered="setDisplayedUnits"
        @unit-type-changed="newUnitType => unitTypeFilter = newUnitType"
        @search-query-changed="newQuery => searchQuery = newQuery"
      />
    </div>

    <!-- Unit Edit Modal -->
    <NModal
      v-model:show="showUnitEditModal"
      style="width: 650px"
      preset="card"
      :title="formTitle"
      :mask-closable="false"
      :close-on-esc="true"
    >
      <UnitForm
        v-if="showUnitEditModal && editingUnitId && associationId && buildingId"
        :association-id="associationId"
        :building-id="buildingId"
        :unit-id="editingUnitId"
        @saved="handleUnitSaved"
        @cancelled="handleUnitFormCancelled"
      />
    </NModal>
  </div>
</template>

<style scoped>
.units-view {
  width: 100%;
  margin: 0 auto;
}
</style>
