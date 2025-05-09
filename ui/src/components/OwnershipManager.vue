<script setup lang="ts">
import { ref, onMounted, h, computed } from 'vue'
import {
  NDataTable,
  NButton,
  NSpace,
  NModal,
  NAlert,
  NEmpty,
  NSpin,
  NTabs,
  NTabPane,
  NDatePicker,
  useMessage,
  NPopconfirm,
  NTag
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { unitApi, ownershipApi } from '@/services/api'
import OwnershipForm from './OwnershipForm.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  associationId: number
  buildingId: number
  unitId: number
}>()

const emit = defineEmits<{
  (e: 'updated'): void
}>()

// Data
const ownerships = ref<any[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showOwnershipModal = ref(false)
const showDisableModal = ref(false)
const selectedOwnershipId = ref<number | null>(null)
const disableDate = ref<number>(Date.now())
const ownershipMode = ref<'create' | 'select'>('select')
const message = useMessage()

// Load ownerships
const fetchOwnerships = async () => {
  try {
    loading.value = true
    error.value = null

    const response = await unitApi.getUnitOwnerships(
      props.associationId,
      props.buildingId,
      props.unitId
    )

    ownerships.value = response.data.filter(o=>o.is_active)
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('common.error', 'Failed to load ownerships')
    console.error('Error fetching ownerships:', err)
  } finally {
    loading.value = false
  }
}

// Handle creating a new ownership
const handleAddOwnership = (mode: 'create' | 'select') => {
  ownershipMode.value = mode
  showOwnershipModal.value = true
}

// Handle saving a new ownership
const handleOwnershipSaved = () => {
  showOwnershipModal.value = false
  fetchOwnerships()
  emit('updated')
}

// Handle ownership disable/deactivation
const openDisableModal = (ownershipId: number) => {
  selectedOwnershipId.value = ownershipId
  showDisableModal.value = true
}

const handleDisableOwnership = async () => {
  if (!selectedOwnershipId.value) return

  try {
    loading.value = true
    error.value = null

    await ownershipApi.disableOwnership(
      props.associationId,
      selectedOwnershipId.value,
      new Date(disableDate.value)
    )

    message.success(t('units.ownership.deactivateSuccess', 'Ownership deactivated successfully'))
    showDisableModal.value = false
    fetchOwnerships()
    emit('updated')
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('units.ownership.deactivateError', 'Failed to deactivate ownership')
    console.error('Error deactivating ownership:', err)
  } finally {
    loading.value = false
  }
}

// Handle setting voting owner
const handleSetVotingOwner = async (ownershipId: number) => {
  try {
    loading.value = true
    error.value = null

    await ownershipApi.setOwnershipVoting(
      props.associationId,
      props.buildingId,
      props.unitId,
      ownershipId
    )

    message.success(t('units.ownership.votingUpdated', 'Voting rights updated successfully'))
    fetchOwnerships()
    emit('updated')
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('units.ownership.votingError', 'Failed to update voting rights')
    console.error('Error updating voting rights:', err)
  } finally {
    loading.value = false
  }
}

// Ownership table columns
const columns = computed<DataTableColumns<any>>(() => [
  {
    title: t('owners.name', 'Owner'),
    key: 'owner_name',
    render: (row) => row.owner_name || t('units.ownership.ownerId', 'Owner ID: {id}', { id: row.owner_id })
  },
  {
    title: t('units.ownership.startDate', 'Start Date'),
    key: 'start_date',
    render: (row) => new Date(row.start_date).toLocaleDateString()
  },
  {
    title: t('units.ownership.endDate', 'End Date'),
    key: 'end_date',
    render: (row) => row.end_date ? new Date(row.end_date).toLocaleDateString() : '-'
  },
  {
    title: t('common.status', 'Status'),
    key: 'is_active',
    render: (row) => {
      const elements = []

      // Status tag
      elements.push(
        h(NTag, {
          type: row.is_active ? 'success' : 'default',
          style: 'margin-right: 8px'
        }, {
          default: () => row.is_active ? t('common.active', 'Active') : t('common.inactive', 'Inactive')
        })
      )

      // Voting tag
      if (row.is_voting) {
        elements.push(
          h(NTag, {
            type: 'info'
          }, {
            default: () => t('units.ownership.votingOwner', 'Voting Owner')
          })
        )
      }

      return h('div', {}, elements)
    }
  },
  {
    title: t('units.ownership.registration', 'Registration'),
    key: 'registration_document',
    render: (row) => `${row.registration_document} (${new Date(row.registration_date).toLocaleDateString()})`
  },
  {
    title: t('common.actions', 'Actions'),
    key: 'actions',
    render: (row) => {
      if (!row.is_active) {
        return t('common.inactive', 'Inactive')
      }

      return h(NSpace, {}, {
        default: () => [
          !row.is_voting ?
          h(
            NButton,
            {
              type: 'primary',
              size: 'small',
              onClick: () => handleSetVotingOwner(row.id)
            },
            { default: () => t('units.ownership.setVoting', 'Set as Voting Owner')}
          ):'',

          // Deactivate button
          h(
            NPopconfirm,
            {
              onPositiveClick: () => openDisableModal(row.id),
              negativeText: t('common.cancel', 'Cancel'),
              positiveText: t('common.disable', 'Deactivate')
            },
            {
              trigger: () => h(
                NButton,
                {
                  strong: true,
                  secondary: true,
                  type: 'warning',
                  size: 'small'
                },
                { default: () => t('common.disable', 'Deactivate') }
              ),
              default: () => t('units.ownership.confirmDeactivate', 'Are you sure you want to deactivate this ownership?')
            }
          )
        ]
      })
    }
  }
])

onMounted(() => {
  fetchOwnerships()
})
</script>

<template>
  <div class="ownership-manager">
    <div class="section-header">
      <h3>{{ t('units.ownership.management', 'Unit Ownership Management') }}</h3>
      <NSpace>
        <NButton type="primary" @click="handleAddOwnership('select')">
          {{ t('units.ownership.addExisting', 'Add Existing Owner') }}
        </NButton>
        <NButton @click="handleAddOwnership('create')">
          {{ t('units.ownership.createNew', 'Create New Owner') }}
        </NButton>
      </NSpace>
    </div>

    <div class="info-box">
      <p><strong>{{ t('common.note', 'Note') }}:</strong> {{ t('units.ownership.votingNote', 'Each unit can have only one voting owner. The voting owner has the right to vote at association meetings. Setting a new voting owner will automatically remove voting rights from any previous voting owner for this unit.') }}</p>
    </div>

    <NSpin :show="loading">
      <NAlert v-if="error" type="error" style="margin-bottom: 16px;">
        {{ error }}
      </NAlert>

      <div class="ownerships-table">
        <NDataTable
          :columns="columns"
          :data="ownerships"
          :bordered="false"
          :single-line="false"
          :row-class-name="row => !row.is_active ? 'inactive-row' : ''"
        >
          <template #empty>
            <NEmpty :description="t('units.ownership.noOwnerships', 'No ownerships found for this unit')">
              <template #extra>
                <p>{{ t('units.ownership.addOwnerPrompt', 'Add an owner using the buttons above') }}</p>
              </template>
            </NEmpty>
          </template>
        </NDataTable>
      </div>
    </NSpin>

    <!-- Ownership Form Modal -->
    <NModal v-model:show="showOwnershipModal" style="width: 600px;" preset="card"
            :title="t('units.ownership.addOwnership', 'Add Ownership')">
      <OwnershipForm
        :association-id="props.associationId"
        :building-id="props.buildingId"
        :unit-id="props.unitId"
        :mode="ownershipMode"
        @saved="handleOwnershipSaved"
        @cancelled="showOwnershipModal = false"
      />
    </NModal>

    <!-- Disable Ownership Modal -->
    <NModal v-model:show="showDisableModal" style="width: 400px;" preset="card"
            :title="t('units.ownership.deactivateOwnership', 'Deactivate Ownership')">
      <div style="margin-bottom: 16px;">
        <p>{{ t('units.ownership.setEndDate', 'Set the date when ownership ends:') }}</p>
        <NDatePicker v-model:value="disableDate" type="date" clearable style="width: 100%" />
      </div>

      <div style="display: flex; justify-content: flex-end; gap: 12px;">
        <NButton @click="showDisableModal = false">{{ t('common.cancel', 'Cancel') }}</NButton>
        <NButton type="warning" @click="handleDisableOwnership">{{ t('common.disable', 'Deactivate') }}</NButton>
      </div>
    </NModal>
  </div>
</template>

<style scoped>
.ownership-manager {
  margin-top: 16px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-header h3 {
  margin: 0;
}

.info-box {
  margin-bottom: 16px;
  padding: 12px;
  border-radius: 4px;
  border-left: 4px solid #2080f0;
}

.ownerships-table {
  margin-bottom: 16px;
}

:deep(.inactive-row) {
  opacity: 0.6;
  background-color: rgba(0, 0, 0, 0.03);
}
</style>
