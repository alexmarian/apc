<template>
  <div class="voting-matters-manager">
    <NCard>
      <template #header>
        <div class="matters-header">
          <h3>{{ $t('gatherings.matters.title') }}</h3>
          <NButton 
            v-if="canEdit" 
            type="primary" 
            @click="showCreateModal = true"
          >
            {{ $t('gatherings.matters.create') }}
          </NButton>
        </div>
      </template>

      <NSpin :show="loading">
        <NAlert v-if="error" type="error" closable @close="error = null">
          {{ error }}
        </NAlert>

        <div v-if="matters.length === 0" class="no-matters">
          <NEmpty :description="$t('gatherings.matters.noMatters')">
            <template #extra>
              <NButton 
                v-if="canEdit" 
                type="primary" 
                @click="showCreateModal = true"
              >
                {{ $t('gatherings.matters.create') }}
              </NButton>
            </template>
          </NEmpty>
        </div>

        <div v-else class="matters-list">
          <NDataTable
            :columns="columns"
            :data="matters"
            :loading="loading"
            :row-key="(row: VotingMatter) => row.id"
            striped
          />
        </div>
      </NSpin>
    </NCard>

    <!-- Create/Edit Modal -->
    <NModal v-model:show="showCreateModal">
      <VotingMatterForm
        :association-id="associationId"
        :gathering="gathering"
        :matter="selectedMatter"
        @saved="handleMatterSaved"
        @cancelled="handleModalCancelled"
      />
    </NModal>

    <!-- Delete Confirmation Modal -->
    <NModal v-model:show="showDeleteModal">
      <NCard style="width: 400px">
        <template #header>
          <h3>{{ $t('gatherings.matters.deleteConfirm') }}</h3>
        </template>
        
        <p>{{ $t('gatherings.matters.deleteWarning') }}</p>
        <p v-if="selectedMatter"><strong>{{ selectedMatter.title }}</strong></p>
        
        <div class="modal-actions">
          <NSpace justify="end">
            <NButton @click="showDeleteModal = false">
              {{ $t('common.cancel') }}
            </NButton>
            <NButton 
              type="error" 
              @click="handleDeleteMatter"
              :loading="deleting"
            >
              {{ $t('common.delete') }}
            </NButton>
          </NSpace>
        </div>
      </NCard>
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
  type DataTableColumns
} from 'naive-ui'
import { votingMatterApi } from '@/services/api'
import type { Gathering, VotingMatter, VotingMatterType } from '@/types/api'
import VotingMatterForm from '@/components/VotingMatterForm.vue'

interface Props {
  associationId: number
  gathering: Gathering
}

const props = defineProps<Props>()

const emit = defineEmits<{
  updated: []
}>()

const { t } = useI18n()

const matters = ref<VotingMatter[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showCreateModal = ref(false)
const showDeleteModal = ref(false)
const selectedMatter = ref<VotingMatter | undefined>(undefined)
const deleting = ref(false)

const canEdit = computed(() => {
  return props.gathering.status === 'draft' || props.gathering.status === 'published'
})

const getMatterTypeColor = (type: VotingMatterType) => {
  switch (type) {
    case 'budget':
      return 'success'
    case 'election':
      return 'warning'
    case 'policy':
      return 'info'
    case 'poll':
      return 'default'
    case 'extraordinary':
      return 'error'
    default:
      return 'default'
  }
}

const columns: DataTableColumns<VotingMatter> = [
  {
    title: t('gatherings.matters.order'),
    key: 'order_index',
    width: 80,
    sorter: (a, b) => a.order_index - b.order_index
  },
  {
    title: t('gatherings.matters.title'),
    key: 'title',
    width: 200,
    ellipsis: {
      tooltip: true
    }
  },
  {
    title: t('gatherings.matters.type'),
    key: 'matter_type',
    width: 120,
    render: (row) => h(NTag, {
      type: getMatterTypeColor(row.matter_type),
      size: 'small'
    }, {
      default: () => t(`gatherings.matters.types.${row.matter_type}`)
    })
  },
  {
    title: t('gatherings.matters.votingType'),
    key: 'voting_config',
    width: 150,
    render: (row) => t(`gatherings.matters.votingTypes.${row.voting_config.type}`)
  },
  {
    title: t('gatherings.matters.majorityType'),
    key: 'voting_config',
    width: 120,
    render: (row) => t(`gatherings.matters.majorityTypes.${row.voting_config.required_majority}`)
  },
  {
    title: t('gatherings.matters.isAnonymous'),
    key: 'voting_config',
    width: 100,
    render: (row) => row.voting_config.is_anonymous ? t('common.yes') : t('common.no')
  },
  {
    title: t('common.actions'),
    key: 'actions',
    width: 180,
    render: (row) => h(NSpace, {}, {
      default: () => [
        h(NButton, {
          size: 'small',
          onClick: () => handleEditMatter(row),
          disabled: !canEdit.value
        }, { default: () => t('common.edit') }),
        
        h(NButton, {
          size: 'small',
          type: 'error',
          onClick: () => handleDeleteConfirm(row),
          disabled: !canEdit.value
        }, { default: () => t('common.delete') })
      ]
    })
  }
]

const loadMatters = async () => {
  loading.value = true
  error.value = null
  
  try {
    const response = await votingMatterApi.getVotingMatters(props.associationId, props.gathering.id)
    matters.value = response.data.sort((a, b) => a.order_index - b.order_index)
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  } finally {
    loading.value = false
  }
}

const handleEditMatter = (matter: VotingMatter) => {
  selectedMatter.value = matter
  showCreateModal.value = true
}

const handleDeleteConfirm = (matter: VotingMatter) => {
  selectedMatter.value = matter
  showDeleteModal.value = true
}

const handleDeleteMatter = async () => {
  if (!selectedMatter.value) return
  
  deleting.value = true
  
  try {
    await votingMatterApi.deleteVotingMatter(
      props.associationId, 
      props.gathering.id, 
      selectedMatter.value.id
    )
    
    showDeleteModal.value = false
    selectedMatter.value = undefined
    await loadMatters()
    emit('updated')
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  } finally {
    deleting.value = false
  }
}

const handleMatterSaved = () => {
  showCreateModal.value = false
  selectedMatter.value = undefined
  loadMatters()
  emit('updated')
}

const handleModalCancelled = () => {
  showCreateModal.value = false
  selectedMatter.value = undefined
}

onMounted(() => {
  loadMatters()
})
</script>

<style scoped>
.voting-matters-manager {
  margin-top: 16px;
}

.matters-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.matters-header h3 {
  margin: 0;
}

.no-matters {
  padding: 32px;
}

.matters-list {
  margin-top: 16px;
}

.modal-actions {
  margin-top: 16px;
}
</style>