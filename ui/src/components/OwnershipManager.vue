<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
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
  NPopconfirm
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { unitApi, ownershipApi } from '@/services/api'
import OwnershipForm from './OwnershipForm.vue'

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

    ownerships.value = response.data
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load ownerships'
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

    message.success('Ownership deactivated successfully')
    showDisableModal.value = false
    fetchOwnerships()
    emit('updated')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to deactivate ownership'
    console.error('Error deactivating ownership:', err)
  } finally {
    loading.value = false
  }
}

// Ownership table columns
const columns: DataTableColumns<any> = [
  {
    title: 'Owner',
    key: 'owner_name',
    render: (row) => row.owner_name || `Owner ID: ${row.owner_id}`
  },
  {
    title: 'Start Date',
    key: 'start_date',
    render: (row) => new Date(row.start_date).toLocaleDateString()
  },
  {
    title: 'End Date',
    key: 'end_date',
    render: (row) => row.end_date ? new Date(row.end_date).toLocaleDateString() : '-'
  },
  {
    title: 'Status',
    key: 'is_active',
    render: (row) => row.is_active ? 'Active' : 'Inactive'
  },
  {
    title: 'Registration',
    key: 'registration_document',
    render: (row) => `${row.registration_document} (${new Date(row.registration_date).toLocaleDateString()})`
  },
  {
    title: 'Actions',
    key: 'actions',
    render: (row) => {
      if (!row.is_active) {
        return 'Inactive'
      }

      return h(
        NPopconfirm,
        {
          onPositiveClick: () => openDisableModal(row.id),
          negativeText: 'Cancel',
          positiveText: 'Deactivate'
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
            { default: () => 'Deactivate' }
          ),
          default: () => 'Are you sure you want to deactivate this ownership?'
        }
      )
    }
  }
]

onMounted(() => {
  fetchOwnerships()
})
</script>

<template>
  <div class="ownership-manager">
    <div class="section-header">
      <h3>Unit Ownership Management</h3>
      <NSpace>
        <NButton type="primary" @click="handleAddOwnership('select')">
          Add Existing Owner
        </NButton>
        <NButton @click="handleAddOwnership('create')">
          Create New Owner
        </NButton>
      </NSpace>
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
            <NEmpty description="No ownerships found for this unit">
              <template #extra>
                <p>Add an owner using the buttons above</p>
              </template>
            </NEmpty>
          </template>
        </NDataTable>
      </div>
    </NSpin>

    <!-- Ownership Form Modal -->
    <NModal v-model:show="showOwnershipModal" style="width: 600px;" preset="card"
            title="Add Ownership">
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
            title="Deactivate Ownership">
      <div style="margin-bottom: 16px;">
        <p>Set the date when ownership ends:</p>
        <NDatePicker v-model:value="disableDate" type="date" clearable style="width: 100%" />
      </div>

      <div style="display: flex; justify-content: flex-end; gap: 12px;">
        <NButton @click="showDisableModal = false">Cancel</NButton>
        <NButton type="warning" @click="handleDisableOwnership">Deactivate</NButton>
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

.ownerships-table {
  margin-bottom: 16px;
}

:deep(.inactive-row) {
  opacity: 0.6;
  background-color: rgba(0, 0, 0, 0.03);
}
</style>
