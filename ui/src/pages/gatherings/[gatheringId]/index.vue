<template>
  <div class="gathering-details-page">
    <NPageHeader>
      <template #title>
        <span v-if="gathering">{{ gathering.title }}</span>
        <span v-else>{{ $t('gatherings.title') }}</span>
      </template>
      <template #header>
        <AssociationSelector v-model:associationId="associationId" />
      </template>
      <template #extra>
        <NSpace v-if="gathering && associationId">
          <NButton @click="$router.push('/gatherings')">
            {{ $t('common.back') }}
          </NButton>
          <NButton 
            v-if="canEdit"
            type="primary" 
            @click="showEditModal = true"
          >
            {{ $t('common.edit') }}
          </NButton>
          <NButton 
            v-if="canChangeStatus"
            type="info"
            @click="showStatusModal = true"
          >
            {{ $t('gatherings.status.title') }}
          </NButton>
        </NSpace>
      </template>
    </NPageHeader>

    <div v-if="!associationId" class="no-association">
      <NCard>
        <div style="text-align: center; padding: 32px;">
          <p>{{ $t('common.selectAssociation') }}</p>
        </div>
      </NCard>
    </div>

    <div v-else class="gathering-content">
      <NSpin :show="loading">
        <NAlert v-if="error" type="error" closable @close="error = null">
          {{ error }}
        </NAlert>

        <div v-if="gathering" class="gathering-info">
          <NGrid :cols="2" :x-gap="16" :y-gap="16">
            <!-- Basic Information -->
            <NGridItem>
              <NCard>
                <template #header>
                  <h3>{{ $t('gatherings.details') }}</h3>
                </template>
                
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
                  <NDescriptionsItem :label="$t('gatherings.type')">
                    {{ $t(`gatherings.type.${gathering.type}`) }}
                  </NDescriptionsItem>
                </NDescriptions>
              </NCard>
            </NGridItem>

            <!-- Statistics -->
            <NGridItem>
              <NCard>
                <template #header>
                  <h3>{{ $t('gatherings.statistics.title') }}</h3>
                </template>
                
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
          <NTabs type="line" animated style="margin-top: 16px;">
            <NTabPane name="matters" :tab="$t('gatherings.matters.title')">
              <VotingMattersManager 
                :association-id="associationId!"
                :gathering="gathering!"
                @updated="loadGathering"
              />
            </NTabPane>
            
            <NTabPane name="participants" :tab="$t('gatherings.participants.title')">
              <ParticipantsManager 
                :association-id="associationId!"
                :gathering="gathering!"
                @updated="loadGathering"
              />
            </NTabPane>
            
            <NTabPane name="voting" :tab="$t('gatherings.voting.title')" :disabled="!canVote">
              <VotingInterface 
                :association-id="associationId!"
                :gathering="gathering!"
              />
            </NTabPane>
            
            <NTabPane name="results" :tab="$t('gatherings.results.title')" :disabled="!canViewResults">
              <ResultsDisplay 
                :association-id="associationId!"
                :gathering="gathering!"
              />
            </NTabPane>
          </NTabs>
        </div>
      </NSpin>
    </div>

    <!-- Edit Modal -->
    <NModal v-model:show="showEditModal" v-if="associationId && gathering">
      <GatheringForm
        :association-id="associationId"
        :gathering="gathering"
        @saved="handleGatheringSaved"
        @cancelled="showEditModal = false"
      />
    </NModal>

    <!-- Status Change Modal -->
    <NModal v-model:show="showStatusModal" v-if="associationId && gathering">
      <GatheringStatusForm
        :association-id="associationId"
        :gathering="gathering"
        @saved="handleStatusChanged"
        @cancelled="showStatusModal = false"
      />
    </NModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  NPageHeader,
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
  NTabPane
} from 'naive-ui'
import { gatheringApi } from '@/services/api'
import type { Gathering, GatheringStatus } from '@/types/api'
import AssociationSelector from '@/components/AssociationSelector.vue'
import GatheringForm from '@/components/GatheringForm.vue'
import GatheringStatusForm from '@/components/GatheringStatusForm.vue'
import VotingMattersManager from '@/components/VotingMattersManager.vue'
import ParticipantsManager from '@/components/ParticipantsManager.vue'
import VotingInterface from '@/components/VotingInterface.vue'
import ResultsDisplay from '@/components/ResultsDisplay.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const associationId = ref<number | null>(null)
const gathering = ref<Gathering | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
const showEditModal = ref(false)
const showStatusModal = ref(false)

const gatheringId = computed(() => {
  const id = route.params.gatheringId
  return typeof id === 'string' ? parseInt(id) : null
})

const canEdit = computed(() => {
  return gathering.value?.status === 'draft' || gathering.value?.status === 'published'
})

const canChangeStatus = computed(() => {
  return gathering.value?.status !== 'tallied'
})

const canVote = computed(() => {
  return gathering.value?.status === 'active'
})

const canViewResults = computed(() => {
  return gathering.value?.status === 'closed' || gathering.value?.status === 'tallied'
})

const participationRate = computed(() => {
  if (!gathering.value || gathering.value.qualified_units === 0) return 0
  return ((gathering.value.participating_units / gathering.value.qualified_units) * 100).toFixed(1)
})

const weightParticipationRate = computed(() => {
  if (!gathering.value || gathering.value.qualified_weight === 0) return 0
  return ((gathering.value.participating_weight / gathering.value.qualified_weight) * 100).toFixed(1)
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
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  } finally {
    loading.value = false
  }
}

const handleGatheringSaved = () => {
  showEditModal.value = false
  loadGathering()
}

const handleStatusChanged = () => {
  showStatusModal.value = false
  loadGathering()
}

watch(associationId, (newValue) => {
  if (newValue && gatheringId.value) {
    loadGathering()
  }
})

onMounted(() => {
  if (associationId.value && gatheringId.value) {
    loadGathering()
  }
})
</script>

<style scoped>
.gathering-details-page {
  padding: 16px;
}

.no-association {
  margin-top: 16px;
}

.gathering-content {
  margin-top: 16px;
}

.gathering-info {
  margin-top: 16px;
}
</style>