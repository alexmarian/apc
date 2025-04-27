<script setup lang="ts">
import { ref, onMounted,h } from 'vue'
import {
  NCard,
  NSpin,
  NAlert,
  NButton,
  NDescriptions,
  NDescriptionsItem,
  NSpace,
  NDivider,
  NEmpty,
  NTag,
  NDataTable
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { unitApi } from '@/services/api'

// Props
const props = defineProps<{
  associationId: number
  buildingId: number
  unitId: number
  showExcerpt?: boolean
}>()

// Emit
const emit = defineEmits<{
  (e: 'edit-owner', ownerId: number): void
}>()

// Data
const loading = ref<boolean>(false)
const error = ref<string | null>(null)
const unitReport = ref<any | null>(null)

// Owners table columns
const ownersColumns = ref<DataTableColumns<any>>([
  {
    title: 'Name',
    key: 'name'
  },
  {
    title: 'Identification',
    key: 'identification_number'
  },
  {
    title: 'Contact Phone',
    key: 'contact_phone'
  },
  {
    title: 'Contact Email',
    key: 'contact_email'
  },
  {
    title: 'Status',
    key: 'is_active',
    render(row) {
      return row.is_active
        ? h(NTag, { type: 'success' }, { default: () => 'Active' })
        : h(NTag, { type: 'warning' }, { default: () => 'Inactive' })
    }
  }
])

// Ownership history columns
const ownershipColumns = ref<DataTableColumns<any>>([
  {
    title: 'Owner',
    key: 'owner_name'
  },
  {
    title: 'Start Date',
    key: 'start_date',
    render(row) {
      return new Date(row.start_date).toLocaleDateString()
    }
  },
  {
    title: 'End Date',
    key: 'end_date',
    render(row) {
      return row.end_date ? new Date(row.end_date).toLocaleDateString() : '-'
    }
  },
  {
    title: 'Status',
    key: 'is_active',
    render(row) {
      return row.is_active
        ? h(NTag, { type: 'success' }, { default: () => 'Active' })
        : h(NTag, { type: 'warning' }, { default: () => 'Inactive' })
    }
  },
  {
    title: 'Registration Doc',
    key: 'registration_document'
  }
])

// Fetch unit report
const fetchUnitReport = async () => {
  try {
    loading.value = true
    error.value = null

    const response = await unitApi.getUnitReport(
      props.associationId,
      props.buildingId,
      props.unitId
    )

    unitReport.value = response.data
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown error occurred'
    console.error('Error fetching unit report:', err)
  } finally {
    loading.value = false
  }
}

// Handle edit owner button click
const handleEditOwner = (ownerId: number) => {
  emit('edit-owner', ownerId)
}

onMounted(() => {
  fetchUnitReport()
})
</script>

<template>
  <div class="unit-report" :class="{ 'excerpt': props.showExcerpt }">
    <NSpin :show="loading">
      <NAlert v-if="error" type="error" title="Error" style="margin-bottom: 16px;">
        {{ error }}
        <template #action>
          <NButton @click="fetchUnitReport">Retry</NButton>
        </template>
      </NAlert>

      <template v-if="unitReport">
        <!-- Unit Details -->
        <NCard v-if="!props.showExcerpt" title="Unit Details" class="report-section">
          <NDescriptions bordered>
            <NDescriptionsItem label="Unit Number">
              {{ unitReport.unit_details.unit_number }}
            </NDescriptionsItem>
            <NDescriptionsItem label="Cadastral Number">
              {{ unitReport.unit_details.cadastral_number }}
            </NDescriptionsItem>
            <NDescriptionsItem label="Address">
              {{ unitReport.unit_details.address }}
            </NDescriptionsItem>
            <NDescriptionsItem label="Entrance">
              {{ unitReport.unit_details.entrance }}
            </NDescriptionsItem>
            <NDescriptionsItem label="Floor">
              {{ unitReport.unit_details.floor }}
            </NDescriptionsItem>
            <NDescriptionsItem label="Unit Type">
              {{ unitReport.unit_details.unit_type }}
            </NDescriptionsItem>
            <NDescriptionsItem label="Area">
              {{ unitReport.unit_details.area }} m²
            </NDescriptionsItem>
            <NDescriptionsItem label="Part">
              {{ unitReport.unit_details.part }}
            </NDescriptionsItem>
            <NDescriptionsItem label="Room Count">
              {{ unitReport.unit_details.room_count }}
            </NDescriptionsItem>
          </NDescriptions>
        </NCard>

        <!-- Building Information (only in full view) -->
        <NCard v-if="!props.showExcerpt" title="Building Information" class="report-section">
          <NDescriptions bordered>
            <NDescriptionsItem label="Building Name">
              {{ unitReport.building_details.name }}
            </NDescriptionsItem>
            <NDescriptionsItem label="Building Address">
              {{ unitReport.building_details.address }}
            </NDescriptionsItem>
            <NDescriptionsItem label="Cadastral Number">
              {{ unitReport.building_details.cadastral_number }}
            </NDescriptionsItem>
            <NDescriptionsItem label="Total Area">
              {{ unitReport.building_details.total_area }} m²
            </NDescriptionsItem>
          </NDescriptions>
        </NCard>

        <!-- Current Owners (always shown) -->
        <NCard
          :title="props.showExcerpt ? 'Unit Owners' : 'Current Owners'"
          class="report-section"
        >
          <template v-if="unitReport.current_owners && unitReport.current_owners.length > 0">
            <NDataTable
              :columns="ownersColumns"
              :data="unitReport.current_owners"
              :bordered="false"
              :single-line="false"
              :pagination="props.showExcerpt ? false : { pageSize: 5 }"
            />
          </template>
          <template v-else>
            <NEmpty description="No owners found for this unit" />
          </template>
        </NCard>

        <!-- Ownership History (only in full view) -->
        <NCard v-if="!props.showExcerpt" title="Ownership History" class="report-section">
          <template v-if="unitReport.ownership_history && unitReport.ownership_history.length > 0">
            <NDataTable
              :columns="ownershipColumns"
              :data="unitReport.ownership_history"
              :bordered="false"
              :single-line="false"
              :pagination="{ pageSize: 5 }"
            />
          </template>
          <template v-else>
            <NEmpty description="No ownership history found for this unit" />
          </template>
        </NCard>
      </template>

      <NEmpty v-else-if="!loading && !error" description="No unit report found" />
    </NSpin>
  </div>
</template>

<style scoped>
.unit-report {
  width: 100%;
}

.report-section {
  margin-bottom: 1.5rem;
}

.excerpt .report-section {
  margin-bottom: 0.5rem;
}

.ownership-tag {
  margin-right: 8px;
}
</style>
