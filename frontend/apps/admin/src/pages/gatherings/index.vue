<template>
  <PageShell :title="$t('gatherings.title')">
    <template #actions>
      <NSpace>
        <NButton @click="loadGatherings">
          {{ $t('common.refresh') }}
        </NButton>
        <NButton v-if="associationId" type="primary" @click="showCreateModal = true">
          {{ $t('gatherings.create') }}
        </NButton>
      </NSpace>
    </template>

    <FilterBar resettable @reset="statusFilter = null">
      <FilterField :label="$t('gatherings.status.title')">
        <NSelect
          v-model:value="statusFilter"
          :options="statusOptions"
          :placeholder="$t('gatherings.filterByStatus')"
          clearable
          style="width: 200px"
        />
      </FilterField>
    </FilterBar>

    <NSpin :show="loading">
      <NAlert v-if="error" type="error" closable @close="error = null">
        {{ error }}
      </NAlert>

      <NCard>
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

    <NModal
      v-if="associationId"
      v-model:show="showCreateModal"
      style="width: 650px"
      preset="card"
      :title="$t('gatherings.create')"
      :mask-closable="false"
      :close-on-esc="true"
    >
      <GatheringForm
        :association-id="associationId"
        @saved="handleGatheringSaved"
        @cancelled="showCreateModal = false"
      />
    </NModal>
  </PageShell>
</template>

<script setup lang="ts">
import { ref, computed, watch, h } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
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
import { storeToRefs } from 'pinia'
import { gatheringApi } from '@/services/api'
import type { Gathering, GatheringStatus } from '@/types/api'
import { useAssociationStore } from '@/stores/association'
import GatheringForm from '@/components/GatheringForm.vue'
import PageShell from '@/components/PageShell.vue'
import FilterBar from '@/components/FilterBar.vue'
import FilterField from '@/components/FilterField.vue'

const { t } = useI18n()
const router = useRouter()

const { associationId } = storeToRefs(useAssociationStore())
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
    title: t('gatherings.type.title'),
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
    width: 100,
    render: (row) => h(NButton, {
      size: 'small',
      onClick: () => router.push(`/gatherings/${row.id}`)
    }, { default: () => t('common.view') })
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

watch(associationId, (val) => {
  if (val) loadGatherings()
  else gatherings.value = []
}, { immediate: true })
</script>

