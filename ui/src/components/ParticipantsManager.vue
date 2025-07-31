<template>
  <div class="participants-manager">
    <NCard>
      <template #header>
        <div class="participants-header">
          <h3>{{ $t('gatherings.participants.title') }}</h3>
          <NSpace>
            <NButton 
              v-if="canEdit" 
              type="primary" 
              @click="showAddModal = true"
            >
              {{ $t('gatherings.participants.add') }}
            </NButton>
            <NButton @click="loadParticipants">
              {{ $t('common.refresh') }}
            </NButton>
          </NSpace>
        </div>
      </template>

      <NSpin :show="loading">
        <NAlert v-if="error" type="error" closable @close="error = null">
          {{ error }}
        </NAlert>

        <div v-if="participants.length === 0" class="no-participants">
          <NEmpty :description="$t('gatherings.participants.noParticipants')">
            <template #extra>
              <NButton 
                v-if="canEdit" 
                type="primary" 
                @click="showAddModal = true"
              >
                {{ $t('gatherings.participants.add') }}
              </NButton>
            </template>
          </NEmpty>
        </div>

        <div v-else class="participants-list">
          <NDataTable
            :columns="columns"
            :data="participants"
            :loading="loading"
            :row-key="(row: GatheringParticipant) => row.id"
            striped
          />
        </div>

        <div class="participants-stats">
          <NStatistic
            :label="$t('gatherings.participants.total')"
            :value="participants.length"
          />
          <NStatistic
            :label="$t('gatherings.participants.checkedIn')"
            :value="checkedInCount"
          />
          <NStatistic
            :label="$t('gatherings.participants.checkInRate')"
            :value="checkInRate"
            suffix="%"
          />
        </div>
      </NSpin>
    </NCard>

    <!-- Add Participant Modal -->
    <NModal v-model:show="showAddModal">
      <ParticipantForm
        :association-id="associationId"
        :gathering="gathering"
        @saved="handleParticipantSaved"
        @cancelled="showAddModal = false"
      />
    </NModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NCard,
  NButton,
  NSpace,
  NAlert,
  NSpin,
  NModal,
  NDataTable,
  NEmpty,
  NTag,
  NStatistic,
  type DataTableColumns
} from 'naive-ui'
import { participantApi } from '@/services/api'
import type { Gathering, GatheringParticipant, ParticipantType } from '@/types/api'
import ParticipantForm from '@/components/ParticipantForm.vue'

interface Props {
  associationId: number
  gathering: Gathering
}

const props = defineProps<Props>()

const emit = defineEmits<{
  updated: []
}>()

const { t } = useI18n()

const participants = ref<GatheringParticipant[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showAddModal = ref(false)

const canEdit = computed(() => {
  return props.gathering.status !== 'tallied'
})

const checkedInCount = computed(() => {
  return participants.value.filter(p => p.checked_in_at).length
})

const checkInRate = computed(() => {
  if (participants.value.length === 0) return 0
  return Math.round((checkedInCount.value / participants.value.length) * 100)
})

const getParticipantTypeColor = (type: ParticipantType) => {
  return type === 'owner' ? 'success' : 'warning'
}

const columns: DataTableColumns<GatheringParticipant> = [
  {
    title: t('gatherings.participants.owner'),
    key: 'owner_name',
    width: 200,
    ellipsis: { tooltip: true }
  },
  {
    title: t('gatherings.participants.participantType'),
    key: 'type',
    width: 120,
    render: (row) => h(NTag, {
      type: getParticipantTypeColor(row.type),
      size: 'small'
    }, {
      default: () => t(`gatherings.participants.types.${row.type}`)
    })
  },
  {
    title: t('gatherings.participants.units'),
    key: 'unit_ids',
    width: 120,
    render: (row) => row.unit_ids.length
  },
  {
    title: t('gatherings.participants.contact'),
    key: 'contact_phone',
    width: 150
  },
  {
    title: t('gatherings.participants.status'),
    key: 'checked_in_at',
    width: 120,
    render: (row) => h(NTag, {
      type: row.checked_in_at ? 'success' : 'default',
      size: 'small'
    }, {
      default: () => row.checked_in_at ? t('gatherings.participants.checkedIn') : t('gatherings.participants.checkedOut')
    })
  },
  {
    title: t('common.actions'),
    key: 'actions',
    width: 200,
    render: (row) => h(NSpace, {}, {
      default: () => [
        h(NButton, {
          size: 'small',
          type: row.checked_in_at ? 'default' : 'primary',
          onClick: () => handleCheckIn(row),
          disabled: !canEdit.value
        }, { 
          default: () => row.checked_in_at ? t('gatherings.participants.checkOut') : t('gatherings.participants.checkIn')
        }),
        
        h(NButton, {
          size: 'small',
          type: 'error',
          onClick: () => handleRemoveParticipant(row),
          disabled: !canEdit.value
        }, { default: () => t('common.delete') })
      ]
    })
  }
]

const loadParticipants = async () => {
  loading.value = true
  error.value = null
  
  try {
    const response = await participantApi.getParticipants(props.associationId, props.gathering.id)
    participants.value = response.data
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  } finally {
    loading.value = false
  }
}

const handleCheckIn = async (participant: GatheringParticipant) => {
  try {
    const checkInData = {
      checked_in_at: participant.checked_in_at ? '' : new Date().toISOString()
    }
    
    await participantApi.checkInParticipant(
      props.associationId, 
      props.gathering.id, 
      participant.id,
      checkInData
    )
    
    await loadParticipants()
    emit('updated')
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  }
}

const handleRemoveParticipant = async (participant: GatheringParticipant) => {
  try {
    await participantApi.removeParticipant(
      props.associationId, 
      props.gathering.id, 
      participant.id
    )
    
    await loadParticipants()
    emit('updated')
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  }
}

const handleParticipantSaved = () => {
  showAddModal.value = false
  loadParticipants()
  emit('updated')
}

onMounted(() => {
  loadParticipants()
})
</script>

<style scoped>
.participants-manager {
  margin-top: 16px;
}

.participants-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.participants-header h3 {
  margin: 0;
}

.no-participants {
  padding: 32px;
}

.participants-list {
  margin-top: 16px;
}

.participants-stats {
  margin-top: 16px;
  display: flex;
  gap: 24px;
}
</style>