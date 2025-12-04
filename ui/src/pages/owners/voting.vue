<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { NPageHeader, NCard, NModal, NSelect, NSpace, NButton, NAlert } from 'naive-ui'
import { useRouter } from 'vue-router'
import VotingOwnersReport from '@/components/VotingOwnersReport.vue'
import OwnerForm from '@/components/OwnerForm.vue'
import AssociationSelector from '@/components/AssociationSelector.vue'
import { gatheringApi } from '@/services/api'
import type { Gathering } from '@/types/api'
import { useI18n } from 'vue-i18n'

// State
const associationId = ref<number | null>(null)
const { t } = useI18n()
const router = useRouter()

// Gathering state for ballot submission
const gatherings = ref<Gathering[]>([])
const selectedGatheringId = ref<number | null>(null)
const loadingGatherings = ref(false)
const gatheringError = ref<string | null>(null)

// Owner editing state
const showOwnerForm = ref(false)
const editingOwnerId = ref<number | null>(null)
const refreshKey = ref(0) // Key to force VotingOwnersReport to re-render

// Computed properties
const activeGatherings = computed(() => {
  return gatherings.value.filter(g => g.status === 'active' || g.status === 'published')
})

const gatheringOptions = computed(() => {
  return activeGatherings.value.map(g => ({
    label: `${g.title} - ${new Date(g.scheduled_date).toLocaleDateString()}`,
    value: g.id
  }))
})

const selectedGathering = computed(() => {
  return gatherings.value.find(g => g.id === selectedGatheringId.value)
})

// Methods for gathering selection
const fetchGatherings = async () => {
  if (!associationId.value) return

  try {
    loadingGatherings.value = true
    gatheringError.value = null
    const response = await gatheringApi.getGatherings(associationId.value)
    gatherings.value = response.data || []
  } catch (err) {
    gatheringError.value = err instanceof Error ? err.message : t('gatherings.loadError', 'Failed to load gatherings')
    console.error('Error fetching gatherings:', err)
    gatherings.value = []
  } finally {
    loadingGatherings.value = false
  }
}

const navigateToVoting = () => {
  if (selectedGatheringId.value && associationId.value) {
    router.push(`/gatherings/${selectedGatheringId.value}`)
  }
}

// Methods for owner editing
const handleEditOwner = (ownerId: number) => {
  editingOwnerId.value = ownerId
  showOwnerForm.value = true
}

const handleOwnerFormSaved = () => {
  // Close the owner form
  showOwnerForm.value = false
  editingOwnerId.value = null

  // Force a re-render of VotingOwnersReport to refresh data
  refreshKey.value++
}

const handleOwnerFormCancelled = () => {
  // Just close the form without refreshing
  showOwnerForm.value = false
  editingOwnerId.value = null
}

// Watch for association changes
import { watch } from 'vue'
watch(associationId, (newVal) => {
  if (newVal) {
    selectedGatheringId.value = null
    fetchGatherings()
  } else {
    gatherings.value = []
    selectedGatheringId.value = null
  }
})


<template>
  <div class="voting-report-page">
    <NPageHeader>
      <template #title>
        {{ t('owners.votingReport', 'Voting Owners Report') }}
      </template>

      <template #header>
        <div style="margin-bottom: 12px;">
          <AssociationSelector v-model:associationId="associationId" />
        </div>
      </template>
    </NPageHeader>

    <div class="content">
      <div v-if="!associationId">
        <NCard style="margin-top: 16px;">
          <div style="text-align: center; padding: 32px;">
            <p>{{ t('owners.selectAssociation', 'Please select an association to view the voting owners report') }}</p>
          </div>
        </NCard>
      </div>

      <div v-else>
        <!-- Gathering Selection Card for Ballot Submission -->
        <NCard style="margin-top: 16px; margin-bottom: 16px;" v-if="associationId">
          <template #header>
            <div style="display: flex; align-items: center; justify-content: space-between;">
              <span>{{ t('owners.voting.ballotSubmission', 'Ballot Submission') }}</span>
            </div>
          </template>

          <NAlert
            v-if="gatheringError"
            type="error"
            closable
            @close="gatheringError = null"
            style="margin-bottom: 16px;"
          >
            {{ gatheringError }}
          </NAlert>

          <NSpace vertical size="large">
            <div>
              <p style="margin-bottom: 8px;">
                {{ t('owners.voting.selectGatheringInfo', 'To submit ballots for voting owners, select an active gathering and navigate to its voting interface. Once there, switch to the Voting tab to submit ballots.') }}
              </p>

              <NSpace align="center">
                <NSelect
                  v-model:value="selectedGatheringId"
                  :options="gatheringOptions"
                  :placeholder="t('owners.voting.selectGathering', 'Select a gathering for ballot submission')"
                  :loading="loadingGatherings"
                  :disabled="loadingGatherings || activeGatherings.length === 0"
                  style="min-width: 350px;"
                  clearable
                />

                <NButton
                  type="primary"
                  :disabled="!selectedGatheringId"
                  @click="navigateToVoting"
                >
                  {{ t('owners.voting.goToVoting', 'Go to Voting Interface') }}
                </NButton>
              </NSpace>
            </div>

            <!-- Show gathering info if selected -->
            <div v-if="selectedGathering" class="gathering-info">
              <h4>{{ t('owners.voting.selectedGathering', 'Selected Gathering') }}: {{ selectedGathering.title }}</h4>
              <p>
                <strong>{{ t('gatherings.date', 'Date') }}:</strong>
                {{ new Date(selectedGathering.scheduled_date).toLocaleString() }}
              </p>
              <p>
                <strong>{{ t('gatherings.status.title', 'Status') }}:</strong>
                {{ t(`gatherings.status.${selectedGathering.status}`, selectedGathering.status) }}
              </p>
              <p v-if="selectedGathering.description">
                <strong>{{ t('gatherings.description', 'Description') }}:</strong>
                {{ selectedGathering.description }}
              </p>
            </div>

            <!-- Info message if no active gatherings -->
            <NAlert
              v-if="!loadingGatherings && activeGatherings.length === 0"
              type="info"
            >
              {{ t('owners.voting.noActiveGatherings', 'No active gatherings found. Create a gathering in the Gatherings section to enable ballot submission.') }}
            </NAlert>
          </NSpace>
        </NCard>

        <!-- Voting Owners Report -->
        <VotingOwnersReport
          :key="refreshKey"
          :association-id="associationId"
          @edit-owner="handleEditOwner"
        />
      </div>
    </div>

    <!-- Owner Edit Modal -->
    <NModal
      v-model:show="showOwnerForm"
      style="width: 650px"
      preset="card"
      :title="t('owners.editOwner', 'Edit Owner')"
      :mask-closable="false"
    >
      <OwnerForm
        v-if="associationId && editingOwnerId"
        :association-id="associationId"
        :owner-id="editingOwnerId"
        @saved="handleOwnerFormSaved"
        @cancelled="handleOwnerFormCancelled"
      />
    </NModal>
  </div>
</template>

<style scoped>
.voting-report-page {
  width: 100%;
}

.content {
  margin-top: 16px;
}

.gathering-info {
  padding: 16px;
  background-color: var(--n-color-target);
  border-radius: 4px;
  border: 1px solid var(--n-border-color);
}

.gathering-info h4 {
  margin: 0 0 12px 0;
  font-size: 1.1rem;
}

.gathering-info p {
  margin: 8px 0;
  line-height: 1.6;
}
</style>
