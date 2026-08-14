<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { NCard, NEmpty, NModal } from 'naive-ui'
import PageShell from '@/components/PageShell.vue'
import UnitsList from '@/components/UnitsList.vue'
import UnitForm from '@/components/UnitForm.vue'
import BuildingSelector from '@/components/BuildingSelector.vue'
import { useAssociationStore } from '@/stores/association'
import { useRouter, useRoute } from 'vue-router'
import type { Unit } from '@/types/api'
import { useI18n } from 'vue-i18n'

// Setup i18n
const { t } = useI18n()

const router = useRouter()
const route = useRoute()

const { associationId } = storeToRefs(useAssociationStore())
const buildingId = ref<number | null>(null)

// UI state
const showUnitEditModal = ref(false)
const editingUnitId = ref<number | undefined>(undefined)
const unitTypeFilter = ref<string[]>([])
const searchQuery = ref<string | null>(null)
const currentPage = ref<number>(1)

// Reference to the UnitsList component
const unitsListRef = ref<InstanceType<typeof UnitsList> | null>(null)

// For storing units loaded by the list component
const displayedUnits = ref<Unit[] | null>(null)

// Restore buildingId, editUnitId, and filters from query parameters
onMounted(() => {
  if (route.query.buildingId) {
    buildingId.value = parseInt(route.query.buildingId as string)
  }
  if (route.query.editUnitId) {
    editingUnitId.value = parseInt(route.query.editUnitId as string)
    showUnitEditModal.value = true
  }
  // Restore filter state from URL
  if (route.query.unitTypeFilter) {
    unitTypeFilter.value = (route.query.unitTypeFilter as string).split(',').filter(Boolean)
  }
  if (route.query.searchQuery) {
    searchQuery.value = route.query.searchQuery as string
  }
  if (route.query.page) {
    currentPage.value = parseInt(route.query.page as string)
  }
})

const setDisplayedUnits = (units: Unit[]) => {
  displayedUnits.value = units
}

// Computed properties
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
      editUnitId: unitId.toString(),
    },
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

  // Update URL query parameters, preserving filters
  router.replace({
    query: {
      ...route.query,
      buildingId: newBuildingId.toString(),
      unitTypeFilter: unitTypeFilter.value.length > 0 ? unitTypeFilter.value.join(',') : undefined,
      searchQuery: searchQuery.value || undefined,
    },
  })
}

// Watch for buildingId changes to close modal
watch(buildingId, () => {
  showUnitEditModal.value = false
  editingUnitId.value = undefined
})

// Watch filter/page changes and update URL
watch([unitTypeFilter, searchQuery, currentPage], ([newUnitType, newSearch, newPage]) => {
  if (buildingId.value) {
    router.replace({
      query: {
        ...route.query,
        unitTypeFilter:
          (newUnitType as string[]).length > 0 ? (newUnitType as string[]).join(',') : undefined,
        searchQuery: newSearch || undefined,
        page: (newPage as number) > 1 ? (newPage as number).toString() : undefined,
      },
    })
  }
})
</script>

<template>
  <PageShell :title="t('units.management', 'Unit Management')">
    <template #context>
      <BuildingSelector
        v-model:building-id="buildingId"
        v-model:association-id="associationId"
        @update:building-id="handleBuildingIdUpdate"
      />
    </template>

    <NCard v-if="!buildingId">
      <NEmpty
        :description="t('units.selectBuilding', 'Please select a building to manage units')"
      />
    </NCard>

    <UnitsList
      v-else-if="canShowUnits"
      ref="unitsListRef"
      :association-id="associationId"
      :building-id="buildingId"
      v-model:unit-type-filter="unitTypeFilter"
      v-model:search-query="searchQuery"
      v-model:page="currentPage"
      @edit="handleEditUnit"
      @units-rendered="setDisplayedUnits"
    />

    <!-- Unit Edit Modal -->
    <NModal
      v-model:show="showUnitEditModal"
      style="width: 650px"
      preset="card"
      :title="t('units.editUnit', 'Edit Unit')"
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
  </PageShell>
</template>
