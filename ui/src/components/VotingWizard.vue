<template>
  <div class="voting-wizard">
    <NSpin :show="loading">
      <NAlert v-if="error" type="error" closable @close="error = null" style="margin-bottom: 16px;">
        {{ error }}
      </NAlert>

      <NSteps :current="currentStep" style="margin-bottom: 24px;">
        <NStep :title="$t('gatherings.voting.selectOwnerUnits')" />
        <NStep v-if="isDelegateVoting" :title="$t('gatherings.voting.delegateDetails')" />
        <NStep :title="$t('gatherings.voting.castBallot')" />
      </NSteps>

      <!-- Step 1: Select Owner and Units -->
      <div v-if="currentStep === 0" class="step-content">
        <NCard :title="$t('gatherings.voting.selectOwnerUnits')">
          <NAlert type="info" style="margin-bottom: 16px;">
            {{ $t('gatherings.voting.step1Instructions') }}
          </NAlert>

          <!-- Delegation Checkbox -->
          <NFormItem>
            <NCheckbox v-model:checked="isDelegateVoting">
              {{ $t('gatherings.voting.voteByDelegation') }}
            </NCheckbox>
          </NFormItem>

          <!-- Owner Selection -->
          <NFormItem :label="$t('gatherings.participants.selectOwner')">
            <NSelect
              v-model:value="selectedOwnerId"
              :options="ownerOptions"
              :placeholder="$t('gatherings.participants.selectOwnerPlaceholder')"
              filterable
              @update:value="handleOwnerChange"
            />
          </NFormItem>

          <!-- Owner Units -->
          <div v-if="selectedOwnerId && selectedOwner" style="margin-top: 16px;">
            <h4>{{ $t('gatherings.voting.availableUnits') }}</h4>

            <NAlert v-if="selectedOwner.available_units_count === 0" type="warning" style="margin: 16px 0;">
              {{ $t('gatherings.voting.noAvailableUnits') }}
            </NAlert>

            <div v-else style="margin-top: 16px;">
              <!-- Select All Checkbox -->
              <div style="margin-bottom: 12px; padding-bottom: 12px; border-bottom: 1px solid var(--n-border-color);">
                <NCheckbox
                  :checked="allUnitsSelected"
                  @update:checked="handleSelectAllUnits"
                >
                  <strong>{{ $t('common.selectAll') }}</strong>
                  ({{ selectedOwner.units.filter(u => u.is_available).length }} {{ $t('gatherings.voting.availableUnits').toLowerCase() }})
                </NCheckbox>
              </div>

              <NCheckboxGroup v-model:value="selectedUnitIds">
                <NSpace vertical>
                  <NCheckbox
                    v-for="unit in selectedOwner.units.filter(u => u.is_available)"
                    :key="unit.id"
                    :value="unit.id"
                  >
                    <div style="display: flex; justify-content: space-between; min-width: 500px;">
                      <span>
                        <strong>{{ unit.unit_number }}</strong>
                        - {{ unit.floor !== null ? $t('units.floor') + ': ' + unit.floor : $t('units.basement') }}
                        ({{ $t('units.entrance') }}: {{ unit.entrance }})
                      </span>
                      <span style="margin-left: 16px;">
                        {{ unit.area.toFixed(2) }} m² ({{ (unit.voting_weight * 100).toFixed(4) }}%)
                      </span>
                    </div>
                  </NCheckbox>
                </NSpace>
              </NCheckboxGroup>

              <NAlert v-if="selectedUnitIds.length > 0" type="info" style="margin-top: 16px;">
                {{ $t('gatherings.voting.selectedUnitsCount') }}: {{ selectedUnitIds.length }}
                <br>
                {{ $t('gatherings.voting.totalVotingWeight') }}: {{ totalSelectedWeight.toFixed(4) }}%
                <br>
                {{ $t('gatherings.voting.totalArea') }}: {{ totalSelectedArea.toFixed(2) }} m²
              </NAlert>
            </div>
          </div>

          <div style="margin-top: 24px; display: flex; justify-content: flex-end;">
            <NButton
              type="primary"
              :disabled="!canProceedStep1"
              @click="nextStep"
            >
              {{ $t('common.next') }} →
            </NButton>
          </div>
        </NCard>
      </div>

      <!-- Step 2: Delegate Details (conditional) -->
      <div v-if="currentStep === 1 && isDelegateVoting" class="step-content">
        <NCard :title="$t('gatherings.voting.delegateDetails')">
          <NAlert type="info" style="margin-bottom: 16px;">
            {{ $t('gatherings.voting.step2Instructions') }}
          </NAlert>

          <NForm ref="delegateFormRef" :model="delegateData" :rules="delegateRules" label-placement="top">
            <NFormItem :label="$t('gatherings.participants.delegateName')" path="name">
              <NInput
                v-model:value="delegateData.name"
                :placeholder="$t('gatherings.participants.delegateNamePlaceholder')"
              />
            </NFormItem>

            <NFormItem :label="$t('gatherings.participants.delegateId')" path="identification">
              <NInput
                v-model:value="delegateData.identification"
                :placeholder="$t('gatherings.participants.enterDelegateId')"
              />
            </NFormItem>

            <NFormItem :label="$t('gatherings.participants.delegationDocument')" path="document">
              <NInput
                v-model:value="delegateData.document"
                :placeholder="$t('gatherings.participants.delegationDocumentPlaceholder')"
              />
            </NFormItem>
          </NForm>

          <div style="margin-top: 24px; display: flex; justify-content: space-between;">
            <NButton @click="previousStep">
              ← {{ $t('common.back') }}
            </NButton>
            <NButton
              type="primary"
              :disabled="!canProceedStep2"
              @click="nextStep"
            >
              {{ $t('common.next') }} →
            </NButton>
          </div>
        </NCard>
      </div>

      <!-- Step 3: Cast Ballot -->
      <div v-if="(currentStep === 1 && !isDelegateVoting) || (currentStep === 2 && isDelegateVoting)" class="step-content">
        <NCard :title="$t('gatherings.voting.castBallot')">
          <NAlert type="info" style="margin-bottom: 16px;">
            {{ $t('gatherings.voting.step3Instructions') }}
          </NAlert>

          <!-- Voting Summary -->
          <NCard size="small" style="margin-bottom: 16px;">
            <template #header>
              <h4>{{ $t('gatherings.voting.summary') }}</h4>
            </template>
            <NDescriptions :column="1" size="small">
              <NDescriptionsItem :label="$t('gatherings.participants.owner')">
                {{ selectedOwner?.owner.Name }}
              </NDescriptionsItem>
              <NDescriptionsItem v-if="isDelegateVoting" :label="$t('gatherings.participants.delegate')">
                {{ delegateData.name }}
              </NDescriptionsItem>
              <NDescriptionsItem :label="$t('gatherings.voting.selectedUnits')">
                {{ selectedUnitIds.length }}
              </NDescriptionsItem>
              <NDescriptionsItem :label="$t('gatherings.voting.totalWeight')">
                {{ totalSelectedWeight.toFixed(4) }}%
              </NDescriptionsItem>
            </NDescriptions>
          </NCard>

          <!-- Voting Matters -->
          <div v-for="(matter, index) in votingMatters" :key="matter.id" style="margin-bottom: 16px;">
            <NCard>
              <template #header>
                <div>
                  <NTag size="small" style="margin-right: 8px;">{{ index + 1 }}</NTag>
                  <strong>{{ matter.title }}</strong>
                </div>
                <p style="font-weight: normal; margin-top: 8px; font-size: 14px;">{{ matter.description }}</p>
              </template>

              <!-- Yes/No Vote -->
              <div v-if="matter.voting_config.type === 'yes_no'">
                <NRadioGroup v-model:value="votes[matter.id]">
                  <NSpace>
                    <NRadio value="yes">
                      <span style="font-size: 16px;">✓ {{ $t('common.yes') }}</span>
                    </NRadio>
                    <NRadio value="no">
                      <span style="font-size: 16px;">✗ {{ $t('common.no') }}</span>
                    </NRadio>
                    <NRadio v-if="matter.voting_config.allow_abstention" value="abstain">
                      <span style="font-size: 16px;">○ {{ $t('gatherings.voting.abstain') }}</span>
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
                      <span style="font-size: 16px;">○ {{ $t('gatherings.voting.abstain') }}</span>
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
                  <NTag size="small" type="info">
                    {{ $t('gatherings.matters.type') }}: {{ $t(`gatherings.matters.types.${matter.matter_type}`) }}
                  </NTag>
                  <NTag size="small">
                    {{ $t('gatherings.matters.majorityType') }}: {{ $t(`gatherings.matters.majorityTypes.${matter.voting_config.required_majority}`) }}
                  </NTag>
                  <NTag v-if="matter.voting_config.is_anonymous" size="small" type="warning">
                    {{ $t('gatherings.matters.isAnonymous') }}
                  </NTag>
                </NSpace>
              </div>
            </NCard>
          </div>

          <div style="margin-top: 24px; display: flex; justify-content: space-between;">
            <NButton @click="previousStep">
              ← {{ $t('common.back') }}
            </NButton>
            <NButton
              type="primary"
              :disabled="!canSubmitBallot"
              :loading="submitting"
              @click="submitBallot"
            >
              {{ $t('gatherings.voting.submitBallot') }}
            </NButton>
          </div>
        </NCard>
      </div>
    </NSpin>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
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
  NTag,
  NDescriptions,
  NDescriptionsItem,
  useMessage,
  type FormInst,
  type FormRules
} from 'naive-ui'
import { gatheringApi, votingMatterApi, votingApi } from '@/services/api'
import type { Gathering, VotingMatter } from '@/types/api'

interface EligibleVoterUnit {
  id: number
  unit_number: string
  cadastral_number: string
  floor: number | null
  entrance: number
  area: number
  voting_weight: number
  unit_type: string
  building_name: string
  building_address: string
  is_available: boolean
}

interface EligibleVoter {
  owner: {
    ID: number
    Name: string
    NormalizedName: string
    IdentificationNumber: string
    ContactPhone: string
    ContactEmail: string
  }
  units: EligibleVoterUnit[]
  total_available_weight: number
  total_available_area: number
  total_weight: number
  total_area: number
  has_available_units: boolean
  available_units_count: number
}

const props = defineProps<{
  associationId: number
  gathering: Gathering
}>()

const emit = defineEmits<{
  completed: []
}>()

const { t } = useI18n()
const message = useMessage()

// State
const loading = ref(false)
const submitting = ref(false)
const error = ref<string | null>(null)
const currentStep = ref(0)

// Step 1: Owner and Units
const isDelegateVoting = ref(false)
const eligibleVoters = ref<EligibleVoter[]>([])
const selectedOwnerId = ref<number | null>(null)
const selectedUnitIds = ref<number[]>([])

// Step 2: Delegate Details
const delegateFormRef = ref<FormInst | null>(null)
const delegateData = ref({
  name: '',
  identification: '',
  document: ''
})

const delegateRules: FormRules = {
  name: [
    { required: true, message: t('gatherings.participants.delegateNameRequired'), trigger: 'blur' }
  ],
  identification: [
    { required: true, message: t('gatherings.participants.delegateIdRequired'), trigger: 'blur' }
  ]
}

// Step 3: Voting
const votingMatters = ref<VotingMatter[]>([])
const votes = ref<Record<number, any>>({})

// Computed
const selectedOwner = computed(() => {
  if (!selectedOwnerId.value) return null
  return eligibleVoters.value.find(v => v.owner.ID === selectedOwnerId.value) || null
})

const ownerOptions = computed(() => {
  return eligibleVoters.value
    .filter(v => v.has_available_units)
    .map(v => ({
      label: `${v.owner.Name} (${v.owner.IdentificationNumber})`,
      value: v.owner.ID
    }))
})

const totalSelectedWeight = computed(() => {
  if (!selectedOwner.value) return 0
  return selectedOwner.value.units
    .filter(u => selectedUnitIds.value.includes(u.id))
    .reduce((sum, u) => sum + u.voting_weight, 0) * 100
})

const totalSelectedArea = computed(() => {
  if (!selectedOwner.value) return 0
  return selectedOwner.value.units
    .filter(u => selectedUnitIds.value.includes(u.id))
    .reduce((sum, u) => sum + u.area, 0)
})

const canProceedStep1 = computed(() => {
  return selectedOwnerId.value !== null && selectedUnitIds.value.length > 0
})

const canProceedStep2 = computed(() => {
  return delegateData.value.name.trim() !== '' && delegateData.value.identification.trim() !== ''
})

const canSubmitBallot = computed(() => {
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
const loadEligibleVoters = async () => {
  try {
    loading.value = true
    const response = await gatheringApi.getEligibleVoters(props.associationId, props.gathering.id)
    eligibleVoters.value = response.data || []
  } catch (err: any) {
    error.value = err.response?.data?.error || err.message || t('gatherings.voting.loadError')
    message.error(error.value)
  } finally {
    loading.value = false
  }
}

const loadVotingMatters = async () => {
  try {
    const response = await votingMatterApi.getVotingMatters(props.associationId, props.gathering.id)
    votingMatters.value = response.data.sort((a, b) => a.order_index - b.order_index)

    // Initialize votes
    votingMatters.value.forEach(matter => {
      if (matter.voting_config.type === 'multiple_choice') {
        votes.value[matter.id] = []
      } else {
        votes.value[matter.id] = null
      }
    })
  } catch (err: any) {
    error.value = err.response?.data?.error || err.message || t('gatherings.matters.loadError')
    message.error(error.value)
  }
}

const handleOwnerChange = () => {
  selectedUnitIds.value = []
}

const handleSelectAllUnits = (checked: boolean) => {
  if (!selectedOwner.value) return

  if (checked) {
    selectedUnitIds.value = selectedOwner.value.units
      .filter(u => u.is_available)
      .map(u => u.id)
  } else {
    selectedUnitIds.value = []
  }
}

const allUnitsSelected = computed(() => {
  if (!selectedOwner.value) return false
  const availableUnits = selectedOwner.value.units.filter(u => u.is_available)
  return availableUnits.length > 0 && selectedUnitIds.value.length === availableUnits.length
})

const nextStep = async () => {
  if (currentStep.value === 0) {
    // Moving from step 1 to step 2 or 3
    if (isDelegateVoting.value) {
      // Load voting matters and go to delegate details
      await loadVotingMatters()
      currentStep.value = 1
    } else {
      // Load voting matters and go directly to voting
      await loadVotingMatters()
      currentStep.value = 1
    }
  } else if (currentStep.value === 1 && isDelegateVoting.value) {
    // Moving from step 2 (delegate) to step 3 (voting)
    currentStep.value = 2
  }
}

const previousStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

const submitBallot = async () => {
  try {
    submitting.value = true
    error.value = null

    // Build ballot content
    const ballotContent: Record<string, any> = {}
    votingMatters.value.forEach(matter => {
      const voteValue = votes.value[matter.id]

      if (matter.voting_config.type === 'yes_no') {
        ballotContent[matter.id.toString()] = { vote_value: voteValue }
      } else if (matter.voting_config.type === 'single_choice') {
        ballotContent[matter.id.toString()] = {
          vote_value: voteValue === 'abstain' ? 'abstain' : voteValue,
          option_id: voteValue === 'abstain' ? null : voteValue
        }
      } else if (matter.voting_config.type === 'multiple_choice') {
        ballotContent[matter.id.toString()] = {
          vote_value: 'multiple',
          option_ids: Array.isArray(voteValue) ? voteValue : []
        }
      }
    })

    // Build request payload
    const payload: any = {
      voter_type: isDelegateVoting.value ? 'delegate' : 'owner',
      owner_id: selectedOwnerId.value,
      unit_ids: selectedUnitIds.value,
      ballot_content: ballotContent
    }

    if (isDelegateVoting.value) {
      payload.delegating_owner_id = selectedOwnerId.value
      payload.delegation_document_ref = delegateData.value.document
      // Note: The backend will use the delegate info from the participant creation
    }

    await votingApi.submitBallot(props.associationId, props.gathering.id, payload)

    message.success(t('gatherings.voting.ballotSubmitted'))
    emit('completed')
    resetWizard()
  } catch (err: any) {
    error.value = err.response?.data?.error || err.message || t('gatherings.voting.submitError')
    message.error(error.value)
  } finally {
    submitting.value = false
  }
}

const resetWizard = () => {
  currentStep.value = 0
  isDelegateVoting.value = false
  selectedOwnerId.value = null
  selectedUnitIds.value = []
  delegateData.value = {
    name: '',
    identification: '',
    document: ''
  }
  votes.value = {}

  // Refresh eligible voters
  loadEligibleVoters()
}

onMounted(() => {
  loadEligibleVoters()
})
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
</style>
