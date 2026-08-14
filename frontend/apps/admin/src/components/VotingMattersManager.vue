<template>
  <div class="voting-matters-manager">
    <NCard :title="$t('gatherings.matters.title')">
      <template #header-extra>
        <NSpace>
          <NButton
            v-if="canEdit"
            type="primary"
            @click="showCreateModal = true"
          >
            {{ $t('gatherings.matters.create') }}
          </NButton>
        </NSpace>
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
    <NModal
      v-model:show="showCreateModal"
      style="width: 750px"
      preset="card"
      :title="selectedMatter ? $t('gatherings.matters.edit') : $t('gatherings.matters.create')"
      :mask-closable="false"
      :close-on-esc="true"
    >
      <VotingMatterForm
        :association-id="associationId"
        :gathering="gathering"
        :matter="selectedMatter"
        @saved="handleMatterSaved"
        @cancelled="handleModalCancelled"
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
  NIcon,
  type DataTableColumns
} from 'naive-ui'
import { ArrowUpOutline, ArrowDownOutline } from '@vicons/ionicons5'
import { votingMatterApi } from '@/services/api'
import type { Gathering, VotingMatter, VotingMatterType } from '@/types/api'
import VotingMatterForm from '@/components/VotingMatterForm.vue'
import { useConfirm } from '@/composables/useConfirm'

interface Props {
  associationId: number
  gathering: Gathering
}

const props = defineProps<Props>()

const emit = defineEmits<{
  updated: []
}>()

const { t } = useI18n()
const confirm = useConfirm()

const matters = ref<VotingMatter[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showCreateModal = ref(false)
const selectedMatter = ref<VotingMatter | undefined>(undefined)

const canEdit = computed(() => {
  return props.gathering.status === 'draft'
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

const handleMoveUp = async (matter: VotingMatter) => {
  const index = matters.value.findIndex(m => m.id === matter.id)
  if (index === 0) return

  const targetMatter = matters.value[index - 1]

  try {
    // Swap order_index values
    await votingMatterApi.updateVotingMatter(
      props.associationId,
      props.gathering.id,
      matter.id,
      {
        title: matter.title,
        description: matter.description,
        matter_type: matter.matter_type,
        order_index: targetMatter.order_index,
        voting_config: matter.voting_config
      }
    )

    await votingMatterApi.updateVotingMatter(
      props.associationId,
      props.gathering.id,
      targetMatter.id,
      {
        title: targetMatter.title,
        description: targetMatter.description,
        matter_type: targetMatter.matter_type,
        order_index: matter.order_index,
        voting_config: targetMatter.voting_config
      }
    )

    await loadMatters()
    emit('updated')
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  }
}

const handleMoveDown = async (matter: VotingMatter) => {
  const index = matters.value.findIndex(m => m.id === matter.id)
  if (index === matters.value.length - 1) return

  const targetMatter = matters.value[index + 1]

  try {
    // Swap order_index values
    await votingMatterApi.updateVotingMatter(
      props.associationId,
      props.gathering.id,
      matter.id,
      {
        title: matter.title,
        description: matter.description,
        matter_type: matter.matter_type,
        order_index: targetMatter.order_index,
        voting_config: matter.voting_config
      }
    )

    await votingMatterApi.updateVotingMatter(
      props.associationId,
      props.gathering.id,
      targetMatter.id,
      {
        title: targetMatter.title,
        description: targetMatter.description,
        matter_type: targetMatter.matter_type,
        order_index: matter.order_index,
        voting_config: targetMatter.voting_config
      }
    )

    await loadMatters()
    emit('updated')
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  }
}

const columns: DataTableColumns<VotingMatter> = [
  {
    title: t('gatherings.matters.order'),
    key: 'order_index',
    width: 100,
    render: (row) => {
      const index = matters.value.findIndex(m => m.id === row.id)
      return h(NSpace, { align: 'center' }, {
        default: () => [
          h(NButton, {
            size: 'small',
            circle: true,
            disabled: !canEdit.value || index === 0,
            onClick: () => handleMoveUp(row)
          }, {
            default: () => h(NIcon, {}, {
              default: () => h(ArrowUpOutline)
            })
          }),
          h('span', { style: { margin: '0 8px' } }, row.order_index),
          h(NButton, {
            size: 'small',
            circle: true,
            disabled: !canEdit.value || index === matters.value.length - 1,
            onClick: () => handleMoveDown(row)
          }, {
            default: () => h(NIcon, {}, {
              default: () => h(ArrowDownOutline)
            })
          })
        ]
      })
    }
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
  confirm({
    title: t('gatherings.matters.deleteConfirm'),
    content: `${t('gatherings.matters.deleteWarning')}\n\n${matter.title}`,
    positiveText: t('common.delete'),
    onConfirm: () => deleteMatter(matter)
  })
}

const deleteMatter = async (matter: VotingMatter) => {
  try {
    await votingMatterApi.deleteVotingMatter(
      props.associationId,
      props.gathering.id,
      matter.id
    )

    await loadMatters()
    emit('updated')
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
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
