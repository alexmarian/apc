<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  NSteps,
  NStep,
  NCard,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NButton,
  NSpace,
  NCheckbox,
  NCheckboxGroup,
  NRadioGroup,
  NRadio,
  NSpin,
  NAlert,
  NDataTable,
  useMessage,
  NTag
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { gatheringApi, votingMatterApi, votingApi, ownerApi, participantApi } from '@/services/api'
import type { Gathering, VotingMatter, QualifiedUnit, Owner } from '@/types/api'
import { useI18n } from 'vue-i18n'

// Props
const props = defineProps<{
  associationId: number
  gathering: Gathering
}>()

// Emits
const emit = defineEmits<{
  (e: 'completed'): void
}>()

const { t } = useI18n()
const message = useMessage()

// State
const currentStep = ref(0)
const loading = ref(false)
const error = ref<string | null>(null)

// Step 1: Participant identification and unit selection
const participantType = ref<'owner' | 'delegate'>('owner')
const ownerSearch = ref('')
const selectedOwner = ref<Owner | null>(null)
const delegateName = ref('')
const delegateIdentification = ref('')
const delegatingOwnerSearch = ref('')
const selectedDelegatingOwner = ref<Owner | null>(null)
const availableOwners = ref<Owner[]>([])
const qualifiedUnits = ref<QualifiedUnit[]>([])
const selectedUnitIds = ref<number[]>([])

// Step 2: Voting matters
const votingMatters = ref<VotingMatter[]>([])
const votes = ref<Record<number, any>>({})

// Computed
const ownerOptions = computed(() => {
  return availableOwners.value.map(o => ({
    label: `${o.name} (${o.identification_number})`,
    value: o.id
  }))
})

const eligibleUnits = computed(() => {
  // Show all non-participating units, regardless of owner
  // This is simpler and allows flexibility in case ownership data doesn't match perfectly
  return qualifiedUnits.value.filter(u => !u.is_participating)
})

const unitColumns = computed<DataTableColumns<QualifiedUnit>>(() => [
  {
    type: 'selection',
    disabled: (row: QualifiedUnit) => row.is_participating
  },
  {
    title: t('owners.owner', 'Owner'),
    key: 'owner_name',
    width: 150,
    ellipsis: {
      tooltip: true
    }
  },
  {
    title: t('units.building', 'Building'),
    key: 'building_name',
    width: 150,
    ellipsis: {
      tooltip: true
    }
  },
  {
    title: t('units.unitNumber', 'Unit Number'),
    key: 'unit_number',
    width: 100
  },
  {
    title: t('units.area', 'Area'),
    key: 'area',
    width: 100,
    render: (row: QualifiedUnit) => `${row.area.toFixed(2)} m²`
  },
  {
    title: t('units.votingWeight', 'Voting Weight'),
    key: 'part',
    width: 120,
    render: (row: QualifiedUnit) => `${(row.part * 100).toFixed(4)}%`
  }
])

const canProceedStep1 = computed(() => {
  // For owner: just need to have units selected (owner will be inferred from units)
  if (participantType.value === 'owner') {
    return selectedUnitIds.value.length > 0
  } else {
    // For delegate: need delegate name and units selected
    return delegateName.value.trim() !== '' &&
           selectedUnitIds.value.length > 0
  }
})

const canSubmitVotes = computed(() => {
  // Check that all required voting matters have been answered
  return votingMatters.value.every(matter => {
    const vote = votes.value[matter.id]
    if (matter.voting_config.allow_abstention && vote === 'abstain') {
      return true
    }
    if (matter.voting_config.type === 'multiple_choice') {
      return Array.isArray(vote) && vote.length > 0
    }
    return vote !== undefined && vote !== null && vote !== ''
  })
})

// Methods
const fetchOwners = async () => {
  try {
    const response = await ownerApi.getOwners(props.associationId)
    availableOwners.value = response.data || []
  } catch (err) {
    console.error('Error fetching owners:', err)
    message.error(t('owners.loadError', 'Failed to load owners'))
  }
}

const fetchQualifiedUnits = async () => {
  try {
    const response = await gatheringApi.getQualifiedUnits(props.associationId, props.gathering.id)
    qualifiedUnits.value = response.data || []
  } catch (err) {
    console.error('Error fetching qualified units:', err)
    message.error(t('gatherings.qualifiedUnitsError', 'Failed to load qualified units'))
  }
}

const fetchVotingMatters = async () => {
  try {
    const response = await votingMatterApi.getVotingMatters(props.associationId, props.gathering.id)
    votingMatters.value = response.data || []

    // Initialize votes object
    votingMatters.value.forEach(matter => {
      if (matter.voting_config.type === 'multiple_choice') {
        votes.value[matter.id] = []
      } else {
        votes.value[matter.id] = null
      }
    })
  } catch (err) {
    console.error('Error fetching voting matters:', err)
    message.error(t('gatherings.matters.loadError', 'Failed to load voting matters'))
  }
}

const handleOwnerSelect = (ownerId: number) => {
  selectedOwner.value = availableOwners.value.find(o => o.id === ownerId) || null
  selectedUnitIds.value = []
}

const handleDelegatingOwnerSelect = (ownerId: number) => {
  selectedDelegatingOwner.value = availableOwners.value.find(o => o.id === ownerId) || null
  selectedUnitIds.value = []
}

const handleUnitSelection = (keys: Array<string | number>) => {
  selectedUnitIds.value = keys.map(k => typeof k === 'string' ? parseInt(k) : k)
}

const nextStep = async () => {
  if (currentStep.value === 0) {
    // Create participant and proceed to voting
    try {
      loading.value = true
      error.value = null

      // Infer the owner from the selected units
      const selectedUnitsData = qualifiedUnits.value.filter(u => selectedUnitIds.value.includes(u.id))
      const inferredOwnerId = selectedUnitsData.length > 0 ? selectedUnitsData[0].owner_id : null

      const participantData = {
        participant_type: participantType.value,
        owner_id: participantType.value === 'owner' ? inferredOwnerId : undefined,
        unit_ids: selectedUnitIds.value,
        delegating_owner_id: participantType.value === 'delegate' ? inferredOwnerId : undefined,
        delegation_document_ref: participantType.value === 'delegate' ? delegateIdentification.value : undefined
      }

      const response = await participantApi.addParticipant(props.associationId, props.gathering.id, participantData)

      // Auto check-in the participant
      await participantApi.checkInParticipant(
        props.associationId,
        props.gathering.id,
        response.data.id,
        { checked_in_at: new Date().toISOString() }
      )

      // Store participant ID for ballot submission
      currentParticipantId.value = response.data.id

      // Load voting matters
      await fetchVotingMatters()

      currentStep.value = 1
      message.success(t('gatherings.participants.added', 'Participant added successfully'))
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message || t('gatherings.participants.addError', 'Failed to add participant')
      message.error(error.value)
    } finally {
      loading.value = false
    }
  }
}

const currentParticipantId = ref<number | null>(null)

const submitBallot = async () => {
  try {
    loading.value = true
    error.value = null

    // Transform votes to match API format
    const ballotContent: Record<string, any> = {}

    votingMatters.value.forEach(matter => {
      const voteValue = votes.value[matter.id]

      if (matter.voting_config.type === 'multiple_choice') {
        // For multiple choice, store as array of option IDs
        ballotContent[matter.id.toString()] = {
          matter_id: matter.id,
          option_id: Array.isArray(voteValue) ? voteValue.join(',') : '',
          vote_value: 'multiple'
        }
      } else if (matter.voting_config.type === 'yes_no') {
        ballotContent[matter.id.toString()] = {
          matter_id: matter.id,
          vote_value: voteValue
        }
      } else {
        // single_choice or other types
        ballotContent[matter.id.toString()] = {
          matter_id: matter.id,
          option_id: voteValue,
          vote_value: voteValue
        }
      }
    })

    await votingApi.submitBallot(
      props.associationId,
      props.gathering.id,
      currentParticipantId.value!,
      { ballot_content: ballotContent }
    )

    message.success(t('gatherings.voting.submitted', 'Ballot submitted successfully'))
    emit('completed')

    // Reset wizard
    resetWizard()
  } catch (err: any) {
    error.value = err.response?.data?.error || err.message || t('gatherings.voting.submitError', 'Failed to submit ballot')
    message.error(error.value)
  } finally {
    loading.value = false
  }
}

const previousStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

const resetWizard = () => {
  currentStep.value = 0
  participantType.value = 'owner'
  selectedOwner.value = null
  delegateName.value = ''
  delegateIdentification.value = ''
  selectedDelegatingOwner.value = null
  selectedUnitIds.value = []
  votes.value = {}
  currentParticipantId.value = null

  // Refresh the qualified units list to update is_participating status
  fetchQualifiedUnits()
}

// Initialize
onMounted(() => {
  fetchOwners()
  fetchQualifiedUnits()
})
</script>

<template>
  <div class="voting-wizard">
    <NSpin :show="loading">
      <NAlert v-if="error" type="error" closable @close="error = null" style="margin-bottom: 16px;">
        {{ error }}
      </NAlert>

      <NSteps :current="currentStep" style="margin-bottom: 24px;">
        <NStep :title="t('gatherings.voting.step1', 'Participant & Units')" />
        <NStep :title="t('gatherings.voting.step2', 'Cast Votes')" />
      </NSteps>

      <!-- Step 1: Participant Identification and Unit Selection -->
      <div v-if="currentStep === 0" class="step-content">
        <NCard :title="t('gatherings.voting.participantInfo', 'Participant Information')">
          <NForm label-placement="left" label-width="150">
            <NFormItem :label="t('gatherings.participants.type', 'Participant Type')">
              <NRadioGroup v-model:value="participantType">
                <NSpace>
                  <NRadio value="owner">{{ t('gatherings.participants.owner', 'Owner') }}</NRadio>
                  <NRadio value="delegate">{{ t('gatherings.participants.delegate', 'Delegate') }}</NRadio>
                </NSpace>
              </NRadioGroup>
            </NFormItem>

            <!-- Owner Selection - Info only -->
            <template v-if="participantType === 'owner'">
              <NAlert type="info" style="margin-bottom: 16px;">
                {{ t('gatherings.voting.ownerInstructions', 'Select the units below that you want to vote for. The owner will be automatically identified from the selected units.') }}
              </NAlert>
            </template>

            <!-- Delegate Information -->
            <template v-if="participantType === 'delegate'">
              <NFormItem :label="t('gatherings.participants.delegateName', 'Delegate Name')">
                <NInput
                  v-model:value="delegateName"
                  :placeholder="t('gatherings.participants.enterDelegateName', 'Enter delegate name')"
                />
              </NFormItem>

              <NFormItem :label="t('gatherings.participants.delegateId', 'Delegate ID')">
                <NInput
                  v-model:value="delegateIdentification"
                  :placeholder="t('gatherings.participants.enterDelegateId', 'Enter identification/document number')"
                />
              </NFormItem>

              <NAlert type="info" style="margin-top: 16px;">
                {{ t('gatherings.voting.delegateInstructions', 'Enter your delegate information above, then select the units you are voting for below. The delegating owner will be automatically identified from the selected units.') }}
              </NAlert>
            </template>
          </NForm>
        </NCard>

        <!-- Unit Selection -->
        <NCard
          :title="t('gatherings.voting.selectUnits', 'Select Units to Vote For')"
          style="margin-top: 16px;"
        >
          <NAlert v-if="eligibleUnits.length === 0" type="warning" style="margin-bottom: 16px;">
            {{ t('gatherings.voting.noEligibleUnits', 'No eligible units available. All units may already be participating.') }}
          </NAlert>

          <NDataTable
            v-else
            :columns="unitColumns"
            :data="eligibleUnits"
            :row-key="(row: QualifiedUnit) => row.id"
            :checked-row-keys="selectedUnitIds"
            @update:checked-row-keys="handleUnitSelection"
            :pagination="{ pageSize: 10 }"
          />

          <div v-if="selectedUnitIds.length > 0" style="margin-top: 16px;">
            <NAlert type="info">
              {{ t('gatherings.voting.selectedUnitsCount', 'Selected units') }}: {{ selectedUnitIds.length }}
              <br>
              {{ t('gatherings.voting.totalVotingWeight', 'Total voting weight') }}:
              {{ (eligibleUnits.filter(u => selectedUnitIds.includes(u.id)).reduce((sum, u) => sum + u.part, 0) * 100).toFixed(4) }}%
            </NAlert>
          </div>
        </NCard>

        <div style="margin-top: 24px; display: flex; justify-content: flex-end;">
          <NButton
            type="primary"
            :disabled="!canProceedStep1"
            @click="nextStep"
          >
            {{ t('common.next', 'Next') }} →
          </NButton>
        </div>
      </div>

      <!-- Step 2: Voting Matters -->
      <div v-if="currentStep === 1" class="step-content">
        <div v-for="matter in votingMatters" :key="matter.id" style="margin-bottom: 16px;">
          <NCard>
            <template #header>
              <h3>{{ matter.title }}</h3>
              <p style="font-weight: normal; margin-top: 8px;">{{ matter.description }}</p>
              <NTag :type="getMatterTypeColor(matter.matter_type)" size="small" style="margin-top: 8px;">
                {{ t(`gatherings.matters.type.${matter.matter_type}`, matter.matter_type) }}
              </NTag>
            </template>

            <!-- Yes/No Vote -->
            <div v-if="matter.voting_config.type === 'yes_no'">
              <NRadioGroup v-model:value="votes[matter.id]">
                <NSpace>
                  <NRadio value="yes">
                    <span style="font-size: 16px;">{{ t('common.yes', 'Yes') }}</span>
                  </NRadio>
                  <NRadio value="no">
                    <span style="font-size: 16px;">{{ t('common.no', 'No') }}</span>
                  </NRadio>
                  <NRadio v-if="matter.voting_config.allow_abstention" value="abstain">
                    <span style="font-size: 16px;">{{ t('gatherings.voting.abstain', 'Abstain') }}</span>
                  </NRadio>
                </NSpace>
              </NRadioGroup>
            </div>

            <!-- Single Choice -->
            <div v-else-if="matter.voting_config.type === 'single_choice'">
              <NRadioGroup v-model:value="votes[matter.id]">
                <NSpace vertical>
                  <NRadio
                    v-for="option in matter.voting_config.options"
                    :key="option.id"
                    :value="option.id"
                  >
                    <span style="font-size: 16px;">{{ option.text }}</span>
                  </NRadio>
                  <NRadio v-if="matter.voting_config.allow_abstention" value="abstain">
                    <span style="font-size: 16px;">{{ t('gatherings.voting.abstain', 'Abstain') }}</span>
                  </NRadio>
                </NSpace>
              </NRadioGroup>
            </div>

            <!-- Multiple Choice -->
            <div v-else-if="matter.voting_config.type === 'multiple_choice'">
              <NCheckboxGroup v-model:value="votes[matter.id]">
                <NSpace vertical>
                  <NCheckbox
                    v-for="option in matter.voting_config.options"
                    :key="option.id"
                    :value="option.id"
                  >
                    <span style="font-size: 16px;">{{ option.text }}</span>
                  </NCheckbox>
                </NSpace>
              </NCheckboxGroup>
            </div>

            <!-- Voting Config Info -->
            <div style="margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--n-border-color);">
              <NSpace size="small">
                <NTag size="small">
                  {{ t('gatherings.matters.requiredMajority', 'Required') }}:
                  {{ matter.voting_config.required_majority }}
                  <template v-if="matter.voting_config.required_majority_value">
                    ({{ matter.voting_config.required_majority_value }}%)
                  </template>
                </NTag>
                <NTag size="small" v-if="matter.voting_config.quorum > 0">
                  {{ t('gatherings.matters.quorum', 'Quorum') }}: {{ matter.voting_config.quorum }}%
                </NTag>
                <NTag size="small" v-if="matter.voting_config.is_anonymous">
                  {{ t('gatherings.matters.anonymous', 'Anonymous') }}
                </NTag>
              </NSpace>
            </div>
          </NCard>
        </div>

        <div style="margin-top: 24px; display: flex; justify-content: space-between;">
          <NButton @click="previousStep">
            ← {{ t('common.previous', 'Previous') }}
          </NButton>
          <NButton
            type="primary"
            :disabled="!canSubmitVotes"
            @click="submitBallot"
          >
            {{ t('gatherings.voting.submitBallot', 'Submit Ballot') }}
          </NButton>
        </div>
      </div>
    </NSpin>
  </div>
</template>

<script lang="ts">
export default {
  name: 'VotingWizard'
}

function getMatterTypeColor(type: string) {
  const colors: Record<string, 'default' | 'success' | 'warning' | 'error' | 'info'> = {
    budget: 'success',
    election: 'info',
    policy: 'warning',
    poll: 'default',
    extraordinary: 'error'
  }
  return colors[type] || 'default'
}
</script>

<style scoped>
.voting-wizard {
  width: 100%;
}

.step-content {
  animation: fadeIn 0.3s ease-in;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.matter-section {
  margin-bottom: 16px;
}
</style>
