<script setup lang="ts">
import { ref, onMounted, h, watch } from 'vue'
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
  NDataTable,
  NTabs,
  NTabPane
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { unitApi } from '@/services/api'
import OwnershipManager from './OwnershipManager.vue'
import type {
  UnitReportDetails,
  ApiResponse
} from '@/types/api'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

// Use interfaces directly from the imported types
type Owner = UnitReportDetails['current_owners'][number];
type OwnershipRecord = UnitReportDetails['ownership_history'][number];

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
  (e: 'edit-unit'): void
}>()

// Data
const loading = ref<boolean>(false)
const error = ref<string | null>(null)
const unitReport = ref<UnitReportDetails | null>(null)
const activeTab = ref<string>('info')

// Define a more specific column type that ensures key property exists
interface OwnerTableColumn {
  title: string;
  key: string;
  render?: (row: Owner) => any;
}

// Owners table columns
const ownersColumns = ref<OwnerTableColumn[]>([
  {
    title: t('owners.name', 'Name'),
    key: 'name'
  },
  {
    title: t('owners.identification', 'Identification'),
    key: 'identification_number'
  },
  {
    title: t('owners.contactPhone', 'Contact Phone'),
    key: 'contact_phone'
  },
  {
    title: t('owners.contactEmail', 'Contact Email'),
    key: 'contact_email'
  },
  {
    title: t('common.status', 'Status'),
    key: 'is_active',
    render(row: Owner) {
      return row.is_active
        ? h(NTag, { type: 'success' }, { default: () => t('common.active', 'Active') })
        : h(NTag, { type: 'warning' }, { default: () => t('common.inactive', 'Inactive') })
    }
  },
  {
    title: t('common.actions', 'Actions'),
    key: 'actions',
    render(row: Owner) {
      return h(
        NButton,
        {
          size: 'small',
          onClick: () => emit('edit-owner', row.id)
        },
        { default: () => t('common.edit', 'Edit') }
      )
    }
  }
])

// Define a more specific column type for ownership history
interface OwnershipTableColumn {
  title: string;
  key: string;
  render?: (row: OwnershipRecord) => any;
}

// Ownership history columns
const ownershipColumns = ref<OwnershipTableColumn[]>([
  {
    title: t('owners.name', 'Owner'),
    key: 'owner_name'
  },
  {
    title: t('units.ownership.startDate', 'Start Date'),
    key: 'start_date',
    render(row: OwnershipRecord) {
      return new Date(row.start_date).toLocaleDateString()
    }
  },
  {
    title: t('units.ownership.endDate', 'End Date'),
    key: 'end_date',
    render(row: OwnershipRecord) {
      return row.end_date ? new Date(row.end_date).toLocaleDateString() : '-'
    }
  },
  {
    title: t('common.status', 'Status'),
    key: 'is_active',
    render(row: OwnershipRecord) {
      return row.is_active
        ? h(NTag, { type: 'success' }, { default: () => t('common.active', 'Active') })
        : h(NTag, { type: 'warning' }, { default: () => t('common.inactive', 'Inactive') })
    }
  },
  {
    title: t('units.ownership.registrationDoc', 'Registration Doc'),
    key: 'registration_document'
  }
])

// Fetch unit report
const fetchUnitReport = async (): Promise<void> => {
  try {
    loading.value = true
    error.value = null

    const response: ApiResponse<UnitReportDetails> = await unitApi.getUnitReport(
      props.associationId,
      props.buildingId,
      props.unitId
    )

    unitReport.value = response.data
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('common.error', 'Unknown error occurred')
    console.error('Error fetching unit report:', err)
  } finally {
    loading.value = false
  }
}

// Handle ownership updates
const handleOwnershipUpdated = (): void => {
  fetchUnitReport()
}

// Handle edit unit button click
const handleEditUnit = (): void => {
  emit('edit-unit')
}

// Watch for property changes to reload data
watch(
  () => [props.associationId, props.buildingId, props.unitId],
  () => {
    fetchUnitReport()
  }
)

onMounted(() => {
  fetchUnitReport()
})
</script>

<template>
  <div class="unit-details" :class="{ 'excerpt': props.showExcerpt }">
    <NSpin :show="loading">
      <NAlert v-if="error" type="error" style="margin-bottom: 16px;">
        {{ error }}
        <NButton @click="fetchUnitReport">{{ t('common.retry', 'Retry') }}</NButton>
      </NAlert>

      <template v-if="unitReport">
        <div v-if="!props.showExcerpt" class="actions-bar">
          <NSpace>
            <NButton type="primary" @click="handleEditUnit">
              {{ t('units.editUnit', 'Edit Unit') }}
            </NButton>
          </NSpace>
        </div>

        <NTabs v-if="!props.showExcerpt" v-model:value="activeTab" type="line" animated>
          <NTabPane name="info" :tab="t('units.information', 'Unit Information')">
            <!-- Unit Details -->
            <NCard :title="t('units.details', 'Unit Details')" class="report-section">
              <NDescriptions bordered>
                <NDescriptionsItem :label="t('units.unit', 'Unit Number')">
                  {{ unitReport.unit_details.unit_number }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="t('units.cadastralNumber', 'Cadastral Number')">
                  {{ unitReport.unit_details.cadastral_number }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="t('units.address', 'Address')">
                  {{ unitReport.unit_details.address }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="t('units.entrance', 'Entrance')">
                  {{ unitReport.unit_details.entrance }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="t('units.floor', 'Floor')">
                  {{ unitReport.unit_details.floor }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="t('units.type', 'Unit Type')">
                  {{ t(`unitTypes.${unitReport.unit_details.unit_type}`, unitReport.unit_details.unit_type) }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="t('units.area', 'Area')">
                  {{ unitReport.unit_details.area }} m²
                </NDescriptionsItem>
                <NDescriptionsItem :label="t('units.part', 'Part')">
                  {{ unitReport.unit_details.part }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="t('units.roomCount', 'Room Count')">
                  {{ unitReport.unit_details.room_count }}
                </NDescriptionsItem>
              </NDescriptions>
            </NCard>

            <!-- Building Information (only in full view) -->
            <NCard :title="t('units.buildingInformation', 'Building Information')" class="report-section">
              <NDescriptions bordered>
                <NDescriptionsItem :label="t('units.buildingName', 'Building Name')">
                  {{ unitReport.building_details.name }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="t('units.buildingAddress', 'Building Address')">
                  {{ unitReport.building_details.address }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="t('units.buildingCadastralNumber', 'Cadastral Number')">
                  {{ unitReport.building_details.cadastral_number }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="t('units.buildingTotalArea', 'Total Area')">
                  {{ unitReport.building_details.total_area }} m²
                </NDescriptionsItem>
              </NDescriptions>
            </NCard>
          </NTabPane>

          <NTabPane name="owners" :tab="t('units.currentOwners', 'Current Owners')">
            <!-- Current Owners -->
            <NCard :title="t('units.currentOwners', 'Current Owners')" class="report-section">
              <template
                v-if="unitReport.current_owners && unitReport.current_owners.filter(owner => owner.is_active).length > 0">
                <NDataTable
                  :columns="ownersColumns"
                  :data="unitReport.current_owners.filter(owner => owner.is_active)"
                  :bordered="false"
                  :single-line="false"
                  :pagination="{ pageSize: 5 }"
                />
              </template>
              <template v-else>
                <NEmpty :description="t('units.noActiveOwners', 'No active owners found for this unit')" />
              </template>
            </NCard>
          </NTabPane>

          <NTabPane name="ownership" :tab="t('units.ownershipHistory', 'Ownership History')">
            <!-- Ownership History -->
            <NCard :title="t('units.ownershipHistory', 'Ownership History')" class="report-section">
              <template
                v-if="unitReport.ownership_history && unitReport.ownership_history.length > 0">
                <NDataTable
                  :columns="ownershipColumns"
                  :data="unitReport.ownership_history"
                  :bordered="false"
                  :single-line="false"
                  :pagination="{ pageSize: 5 }"
                />
              </template>
              <template v-else>
                <NEmpty :description="t('units.noOwnershipHistory', 'No ownership history found for this unit')" />
              </template>
            </NCard>
          </NTabPane>

          <NTabPane name="manage" :tab="t('units.manageOwnership', 'Manage Ownership')">
            <!-- Ownership Management -->
            <OwnershipManager
              :association-id="props.associationId"
              :building-id="props.buildingId"
              :unit-id="props.unitId"
              @updated="handleOwnershipUpdated"
            />
          </NTabPane>
        </NTabs>

        <!-- Excerpt mode (simplified view) -->
        <template v-else>
          <!-- Current Owners (always shown in excerpt mode) -->
          <NCard :title="t('units.unitOwners', 'Unit Owners')" class="report-section">
            <template
              v-if="unitReport.current_owners && unitReport.current_owners.filter(owner => owner.is_active).length > 0">
              <NDataTable
                :columns="ownersColumns.filter(col => col.key !== 'actions')"
                :data="unitReport.current_owners.filter(owner => owner.is_active)"
                :bordered="false"
                :single-line="false"
                :pagination="false"
              />
            </template>
            <template v-else>
              <NEmpty :description="t('units.noActiveOwners', 'No active owners found for this unit')" />
            </template>
          </NCard>
        </template>

      </template>

      <NEmpty v-else-if="!loading && !error" :description="t('units.noUnitReport', 'No unit report found')" />
    </NSpin>
  </div>
</template>

<style scoped>
.unit-details {
  width: 100%;
}

.report-section {
  margin-bottom: 1.5rem;
}

.excerpt .report-section {
  margin-bottom: 0.5rem;
}

.actions-bar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}
</style>
