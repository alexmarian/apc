<template>
  <PageShell
    :title="gathering ? gathering.title : $t('gatherings.title')"
    back
    @back="$router.push('/gatherings')"
  >
    <template #actions>
      <NSpace v-if="gathering && associationId">
        <NButton v-if="canEdit" type="primary" @click="showEditModal = true">
          {{ $t('common.edit') }}
        </NButton>
        <NButton
          v-for="targetStatus in availableTransitions"
          :key="targetStatus"
          :type="getTransitionButtonType(targetStatus)"
          :loading="transitionSubmitting"
          @click="handleTransition(targetStatus)"
        >
          {{ $t(`gatherings.status.transitions.${gathering.status}_to_${targetStatus}`) }}
        </NButton>
      </NSpace>
    </template>

    <NSpin :show="loading">
      <NAlert v-if="error" type="error" closable @close="error = null">
        {{ error }}
      </NAlert>

      <div v-if="gathering">
        <NGrid :cols="2" :x-gap="16" :y-gap="16">
          <!-- Basic Information -->
          <NGridItem>
            <NCard :title="$t('gatherings.details')">
              <NDescriptions :column="1" label-placement="left">
                <NDescriptionsItem :label="$t('gatherings.title')">
                  {{ gathering.title }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="$t('gatherings.description')">
                  {{ gathering.description }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="$t('gatherings.location')">
                  {{ gathering.location }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="$t('gatherings.scheduledDate')">
                  {{ formatDate(gathering.scheduled_date) }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="$t('gatherings.status.title')">
                  <NTag :type="getStatusType(gathering.status)">
                    {{ $t(`gatherings.status.${gathering.status}`) }}
                  </NTag>
                </NDescriptionsItem>
                <NDescriptionsItem :label="$t('gatherings.type.title')">
                  {{ $t(`gatherings.type.${gathering.type}`) }}
                </NDescriptionsItem>
              </NDescriptions>
            </NCard>
          </NGridItem>

          <!-- Statistics -->
          <NGridItem>
            <NCard :title="$t('gatherings.statistics.title')">
              <NDescriptions :column="1" label-placement="left">
                <NDescriptionsItem :label="$t('gatherings.statistics.qualified')">
                  {{ gathering.qualified_units }} {{ $t('gatherings.statistics.units') }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="$t('gatherings.statistics.participating')">
                  {{ gathering.participating_units }} {{ $t('gatherings.statistics.units') }}
                </NDescriptionsItem>
                <NDescriptionsItem :label="$t('gatherings.statistics.participationRate')">
                  {{ participationRate }}%
                </NDescriptionsItem>
                <NDescriptionsItem :label="$t('gatherings.statistics.weightParticipationRate')">
                  {{ weightParticipationRate }}%
                </NDescriptionsItem>
              </NDescriptions>
            </NCard>
          </NGridItem>
        </NGrid>

        <!-- Tabs for different sections -->
        <NTabs type="line" animated style="margin-top: 16px">
          <NTabPane name="matters" :tab="$t('gatherings.matters.title')">
            <VotingMattersManager
              :association-id="associationId!"
              :gathering="gathering!"
              @updated="loadGathering"
            />
          </NTabPane>

          <NTabPane name="voting" :tab="$t('gatherings.voting.title')" :disabled="!canVote">
            <VotingWizard
              :association-id="associationId!"
              :gathering="gathering!"
              @completed="loadGathering"
            />
          </NTabPane>

          <NTabPane
            name="results"
            :tab="$t('gatherings.results.title')"
            :disabled="!canViewResults"
          >
            <ResultsDisplay :association-id="associationId!" :gathering="gathering!" />
          </NTabPane>

          <NTabPane
            name="ballots"
            :tab="$t('gatherings.ballots.title')"
            :disabled="!canViewResults"
          >
            <BallotsList :association-id="associationId!" :gathering="gathering!" />
          </NTabPane>

          <NTabPane name="invitations" :tab="$t('gatherings.invitations.title')">
            <MemberInvitationsPanel :association-id="associationId!" :gathering="gathering!" />
          </NTabPane>
        </NTabs>
      </div>
    </NSpin>

    <!-- Edit Modal -->
    <NModal
      v-if="associationId && gathering"
      v-model:show="showEditModal"
      style="width: 650px"
      preset="card"
      :title="$t('common.edit')"
      :mask-closable="false"
      :close-on-esc="true"
    >
      <GatheringForm
        :association-id="associationId"
        :gathering="gathering"
        @saved="handleGatheringSaved"
        @cancelled="showEditModal = false"
      />
    </NModal>
  </PageShell>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  NCard,
  NButton,
  NSpace,
  NAlert,
  NSpin,
  NModal,
  NGrid,
  NGridItem,
  NDescriptions,
  NDescriptionsItem,
  NTag,
  NTabs,
  NTabPane,
} from 'naive-ui'
import { storeToRefs } from 'pinia'
import { gatheringApi } from '@/services/api'
import { apiErrorMessage } from '@/utils/apiError'
import type { Gathering, GatheringStatus } from '@/types/api'
import { useAssociationStore } from '@/stores/association'
import GatheringForm from '@/components/GatheringForm.vue'
import VotingMattersManager from '@/components/VotingMattersManager.vue'
import VotingWizard from '@/components/VotingWizard.vue'
import ResultsDisplay from '@/components/ResultsDisplay.vue'
import BallotsList from '@/components/BallotsList.vue'
import MemberInvitationsPanel from '@/components/MemberInvitationsPanel.vue'
import PageShell from '@/components/PageShell.vue'
import { useConfirm } from '@/composables/useConfirm'

const { t } = useI18n()
const route = useRoute()
const confirm = useConfirm()

const { associationId } = storeToRefs(useAssociationStore())
const gathering = ref<Gathering | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
const showEditModal = ref(false)
const transitionSubmitting = ref(false)

const gatheringId = computed(() => {
  const id = route.params.gatheringId
  return typeof id === 'string' ? parseInt(id) : null
})

const canEdit = computed(() => {
  return gathering.value?.status === 'draft' || gathering.value?.status === 'published'
})

const statusTransitions: Record<GatheringStatus, GatheringStatus[]> = {
  draft: ['published' as GatheringStatus],
  published: ['active' as GatheringStatus, 'draft' as GatheringStatus],
  active: ['closed' as GatheringStatus],
  closed: ['tallied' as GatheringStatus, 'active' as GatheringStatus],
  tallied: [],
}

const availableTransitions = computed<GatheringStatus[]>(() => {
  if (!gathering.value) return []
  return statusTransitions[gathering.value.status] || []
})

const getTransitionButtonType = (targetStatus: GatheringStatus) => {
  switch (targetStatus) {
    case 'published':
      return 'info'
    case 'active':
      return gathering.value?.status === 'closed' ? 'error' : 'success'
    case 'closed':
      return 'warning'
    case 'tallied':
      return 'primary'
    case 'draft':
      return 'warning'
    default:
      return 'default'
  }
}

const handleTransition = (targetStatus: GatheringStatus) => {
  if (!gathering.value || !associationId.value) return
  const isDestructive =
    targetStatus === 'draft' || (targetStatus === 'active' && gathering.value.status === 'closed')
  const gatheringId = gathering.value.id

  confirm({
    title: t('gatherings.status.confirmTitle'),
    content: t(`gatherings.status.descriptions.${targetStatus}`),
    tone: isDestructive ? 'destructive' : 'neutral',
    onConfirm: async () => {
      transitionSubmitting.value = true
      error.value = null
      try {
        await gatheringApi.updateGatheringStatus(associationId.value!, gatheringId, {
          status: targetStatus,
        })
        loadGathering()
      } catch (err) {
        error.value = apiErrorMessage(err, t('common.error'))
      } finally {
        transitionSubmitting.value = false
      }
    },
  })
}

const canVote = computed(() => {
  return gathering.value?.status === 'active'
})

const canViewResults = computed(() => {
  return (
    gathering.value?.status === 'active' ||
    gathering.value?.status === 'closed' ||
    gathering.value?.status === 'tallied'
  )
})

const participationRate = computed(() => {
  if (!gathering.value || gathering.value.qualified_units === 0) return 0
  return ((gathering.value.participating_units / gathering.value.qualified_units) * 100).toFixed(1)
})

const weightParticipationRate = computed(() => {
  if (!gathering.value || gathering.value.qualified_weight === 0) return 0
  return ((gathering.value.participating_weight / gathering.value.qualified_weight) * 100).toFixed(
    1,
  )
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

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

const loadGathering = async () => {
  if (!associationId.value || !gatheringId.value) return

  loading.value = true
  error.value = null

  try {
    const response = await gatheringApi.getGathering(associationId.value!, gatheringId.value!)
    gathering.value = response.data
  } catch (err) {
    error.value = apiErrorMessage(err, t('common.error'))
  } finally {
    loading.value = false
  }
}

const handleGatheringSaved = () => {
  showEditModal.value = false
  loadGathering()
}

watch(
  [associationId, gatheringId],
  ([assocId, gId]) => {
    if (assocId && gId) loadGathering()
  },
  { immediate: true },
)
</script>
