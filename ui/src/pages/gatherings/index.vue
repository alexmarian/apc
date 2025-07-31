<template>
  <div class="gatherings-page">
    <NPageHeader>
      <template #title>{{ $t('gatherings.title') }}</template>
      <template #header>
        <AssociationSelector v-model:associationId="associationId" />
      </template>
      <template #extra>
        <NButton v-if="associationId" type="primary" @click="showCreateModal = true">
          {{ $t('gatherings.create') }}
        </NButton>
      </template>
    </NPageHeader>

    <div v-if="!associationId" class="no-association">
      <NCard>
        <div style="text-align: center; padding: 32px;">
          <p>{{ $t('common.selectAssociation') }}</p>
        </div>
      </NCard>
    </div>

    <div v-else class="gatherings-content">
      <NSpin :show="loading">
        <NAlert v-if="error" type="error" closable @close="error = null">
          {{ error }}
        </NAlert>

        <NCard>
          <template #header>
            <div class="gatherings-header">
              <h3>{{ $t('gatherings.list') }}</h3>
              <NSpace>
                <NSelect
                  v-model:value="statusFilter"
                  :options="statusOptions"
                  :placeholder="$t('gatherings.filterByStatus')"
                  clearable
                  style="width: 200px"
                />
                <NButton @click="loadGatherings">
                  {{ $t('common.refresh') }}
                </NButton>
              </NSpace>
            </div>
          </template>

          <NDataTable
            :columns="columns"
            :data="filteredGatherings"
            :loading="loading"
            :pagination="pagination"
            :row-key="(row: Gathering) => row.id"
            striped
          />
        </NCard>
      </NSpin>
    </div>

    <NModal v-model:show="showCreateModal" v-if="associationId">
      <GatheringForm
        :association-id="associationId"
        @saved="handleGatheringSaved"
        @cancelled="showCreateModal = false"
      />
    </NModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, h } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  NPageHeader,
  NCard,
  NButton,
  NDataTable,
  NSpace,
  NSelect,
  NAlert,
  NSpin,
  NModal,
  NTag,
  type DataTableColumns
} from 'naive-ui'
import { gatheringApi } from '@/services/api'
import type { Gathering, GatheringStatus } from '@/types/api'
import AssociationSelector from '@/components/AssociationSelector.vue'
import GatheringForm from '@/components/GatheringForm.vue'

const { t } = useI18n()
const router = useRouter()

const associationId = ref<number | null>(null)
const gatherings = ref<Gathering[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const statusFilter = ref<GatheringStatus | null>(null)
const showCreateModal = ref(false)

const pagination = {
  pageSize: 10,
  showSizePicker: true,
  pageSizes: [10, 20, 50]
}

const statusOptions = computed(() => [
  { label: t('gatherings.status.draft'), value: 'draft' },
  { label: t('gatherings.status.published'), value: 'published' },
  { label: t('gatherings.status.active'), value: 'active' },
  { label: t('gatherings.status.closed'), value: 'closed' },
  { label: t('gatherings.status.tallied'), value: 'tallied' }
])

const filteredGatherings = computed(() => {
  if (!statusFilter.value) return gatherings.value
  return gatherings.value.filter(gathering => gathering.status === statusFilter.value)
})

const getStatusType = (status: GatheringStatus) => {
  switch (status) {
    case 'draft':
      return 'default'
    case 'published':
      return 'info'
    case 'active':
      return 'success'
    case 'closed':
      return 'warning'
    case 'tallied':
      return 'error'
    default:
      return 'default'
  }
}

const columns: DataTableColumns<Gathering> = [
  {
    title: t('gatherings.title'),
    key: 'title',
    width: 200,
    ellipsis: {
      tooltip: true
    }
  },
  {
    title: t('gatherings.scheduledDate'),
    key: 'scheduled_date',
    width: 150,
    render: (row) => new Date(row.scheduled_date).toLocaleDateString()
  },
  {
    title: t('gatherings.status.title'),
    key: 'status',
    width: 120,
    render: (row) => h(NTag, {
      type: getStatusType(row.status),
      size: 'small'
    }, {
      default: () => t(`gatherings.status.${row.status}`)
    })
  },
  {
    title: t('gatherings.type'),
    key: 'type',
    width: 100,
    render: (row) => t(`gatherings.type.${row.type}`)
  },
  {
    title: t('gatherings.participants'),
    key: 'participating_units',
    width: 120,
    render: (row) => `${row.participating_units}/${row.qualified_units}`
  },
  {
    title: t('gatherings.participation'),
    key: 'participation_rate',
    width: 120,
    render: (row) => {
      const rate = row.qualified_units > 0 ? (row.participating_units / row.qualified_units * 100) : 0
      return `${rate.toFixed(1)}%`
    }
  },
  {
    title: t('common.actions'),
    key: 'actions',
    width: 200,
    render: (row) => h(NSpace, {}, {
      default: () => [
        h(NButton, {
          size: 'small',
          onClick: () => router.push(`/gatherings/${row.id}`)
        }, { default: () => t('common.view') }),
        
        h(NButton, {
          size: 'small',
          type: 'primary',
          disabled: row.status === 'tallied',
          onClick: () => router.push(`/gatherings/${row.id}/manage`)
        }, { default: () => t('common.manage') }),
        
        h(NButton, {
          size: 'small',
          type: 'info',
          disabled: row.status === 'draft',
          onClick: () => router.push(`/gatherings/${row.id}/vote`)
        }, { default: () => t('gatherings.vote') })
      ]
    })
  }
]

const loadGatherings = async () => {
  if (!associationId.value) return
  
  loading.value = true
  error.value = null
  
  try {
    const response = await gatheringApi.getGatherings(associationId.value!)
    gatherings.value = response.data
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  } finally {
    loading.value = false
  }
}

const handleGatheringSaved = () => {
  showCreateModal.value = false
  loadGatherings()
}

watch(associationId, (newValue) => {
  if (newValue) {
    loadGatherings()
  } else {
    gatherings.value = []
  }
})

onMounted(() => {
  if (associationId.value) {
    loadGatherings()
  }
})
</script>

<style scoped>
.gatherings-page {
  padding: 16px;
}

.no-association {
  margin-top: 16px;
}

.gatherings-content {
  margin-top: 16px;
}

.gatherings-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.gatherings-header h3 {
  margin: 0;
}
</style>