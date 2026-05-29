<template>
  <div class="ballots-list">
    <NSpin :show="loading">
      <NAlert v-if="error" type="error" closable @close="error = null">
        {{ error }}
      </NAlert>

      <NCard>
        <template #header>
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <h3>{{ $t('gatherings.ballots.title') }}</h3>
            <NButton
              type="primary"
              @click="handleDownloadBallots"
              :loading="downloading"
            >
              <template #icon>
                <NIcon><DownloadOutlined /></NIcon>
              </template>
              {{ $t('gatherings.ballots.download') }}
            </NButton>
          </div>
        </template>

        <div v-if="!ballots || ballots.length === 0" class="no-ballots">
          <NAlert type="info">
            {{ $t('gatherings.ballots.noBallots') }}
          </NAlert>
        </div>

        <div v-else>
          <NDataTable
            :columns="columns"
            :data="ballots"
            :pagination="pagination"
            :row-key="row => row.id"
          />
        </div>
      </NCard>
    </NSpin>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NCard,
  NButton,
  NIcon,
  NAlert,
  NSpin,
  NDataTable,
  NTag,
  type DataTableColumns
} from 'naive-ui'
import { DownloadOutlined } from '@vicons/antd'
import { votingApi } from '@/services/api'
import type { Gathering } from '@/types/api'

interface Ballot {
  id: number
  gathering_id: number
  participant_id: number
  participant_name: string
  units_info: number[]
  units_area: number
  units_part: number
  ballot_hash: string
  submitted_at: string
  is_valid: boolean
}

interface Props {
  associationId: number
  gathering: Gathering
}

const props = defineProps<Props>()

const { t } = useI18n()

const loading = ref(false)
const downloading = ref(false)
const error = ref<string | null>(null)
const ballots = ref<Ballot[]>([])

const pagination = {
  pageSize: 20
}

const columns: DataTableColumns<Ballot> = [
  {
    title: t('gatherings.ballots.participant'),
    key: 'participant_name',
    width: 200
  },
  {
    title: t('gatherings.ballots.unitsCount'),
    key: 'units_info',
    width: 100,
    render: (row) => row.units_info.length
  },
  {
    title: t('gatherings.ballots.weight'),
    key: 'units_part',
    width: 120,
    render: (row) => row.units_part.toFixed(4)
  },
  {
    title: t('gatherings.ballots.area'),
    key: 'units_area',
    width: 120,
    render: (row) => `${row.units_area.toFixed(2)} m²`
  },
  {
    title: t('gatherings.ballots.submittedAt'),
    key: 'submitted_at',
    width: 180,
    render: (row) => new Date(row.submitted_at).toLocaleString()
  },
  {
    title: t('gatherings.ballots.status'),
    key: 'is_valid',
    width: 100,
    render: (row) => h(
      NTag,
      {
        type: row.is_valid ? 'success' : 'error',
        size: 'small'
      },
      { default: () => row.is_valid ? t('common.valid') : t('common.invalid') }
    )
  },
  {
    title: t('gatherings.ballots.hash'),
    key: 'ballot_hash',
    width: 150,
    ellipsis: {
      tooltip: true
    },
    render: (row) => row.ballot_hash.substring(0, 16) + '...'
  }
]

const loadBallots = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await votingApi.getBallots(props.associationId, props.gathering.id)
    ballots.value = response.data
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  } finally {
    loading.value = false
  }
}

const handleDownloadBallots = async () => {
  downloading.value = true
  try {
    const response = await votingApi.downloadBallots(props.associationId, props.gathering.id)

    // Create a blob from the response data
    const blob = new Blob([response.data], { type: 'text/markdown' })

    // Create a download link
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `voting-ballots-${props.gathering.title}-${new Date().toISOString().split('T')[0]}.md`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  } finally {
    downloading.value = false
  }
}

onMounted(() => {
  loadBallots()
})
</script>

<style scoped>
.ballots-list {
  margin-top: 16px;
}

.no-ballots {
  margin: 16px 0;
}
</style>
