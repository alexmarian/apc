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

          <!-- Invitation link -->
          <div v-if="selectedOwnerId" style="margin-top: 8px; margin-bottom: 16px;">
            <NSpin :show="invitationLinkLoading" size="small">
              <template v-if="!invitationLinkLoading">
                <NAlert v-if="ownerHasActiveInvitation && !invitationLink" type="info" style="margin-bottom: 0">
                  {{ $t('gatherings.voting.memberLinkAlreadyActive') }}
                </NAlert>
                <NInputGroup v-else-if="invitationLink">
                  <NInput :value="invitationLink" readonly style="font-size: 12px; font-family: monospace" />
                  <NButton @click="copyInvitationLink" style="min-width: 80px">
                    {{ invitationLinkCopied ? $t('gatherings.voting.linkCopied') : $t('gatherings.voting.copyLink') }}
                  </NButton>
                </NInputGroup>
                <NButton
                  v-else
                  size="small"
                  :loading="invitationLinkGenerating"
                  @click="generateInvitationLink"
                >
                  {{ $t('gatherings.voting.generateMemberLink') }}
                </NButton>
              </template>
            </NSpin>
          </div>

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

          <!-- Voting Matters via shared BallotForm -->
          <BallotForm :matters="votingMatters" v-model="ballotVotes" />

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
  NInputGroup,
  NSelect,
  NButton,
  NSpace,
  NCheckbox,
  NCheckboxGroup,
  NSpin,
  NAlert,
  NTag,
  NDescriptions,
  NDescriptionsItem,
  useMessage,
  type FormInst,
  type FormRules
} from 'naive-ui'
import { BallotForm } from '@apc/voting-widgets'
import { gatheringApi, votingMatterApi, votingApi, invitationApi } from '@/services/api'
import type { Gathering, VotingMatter, MemberInvitation } from '@/types/api'

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

// Step 1: Invitation link
const MEMBER_BASE_URL = import.meta.env.VITE_MEMBER_BASE_URL || 'https://member.blocul-nostru.online'
const ownerHasActiveInvitation = ref(false)
const invitationLink = ref<string | null>(null)
const invitationLinkLoading = ref(false)
const invitationLinkGenerating = ref(false)
const invitationLinkCopied = ref(false)

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
const ballotVotes = ref<Record<string, string[]>>({})

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
  if (!selectedOwner.value || !props.gathering.qualified_area) return 0
  return (totalSelectedArea.value / props.gathering.qualified_area) * 100
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
  return votingMatters.value.every(m => (ballotVotes.value[String(m.id)]?.length ?? 0) > 0)
})

// Methods
const loadEligibleVoters = async () => {
  try {
    loading.value = true
    const response = await gatheringApi.getEligibleVoters(props.associationId, props.gathering.id)
    eligibleVoters.value = response.data || []
  } catch (err: any) {
    error.value = err.response?.data?.error || err.message || t('gatherings.voting.loadError')
    message.error(error.value ?? 'An error occurred')
  } finally {
    loading.value = false
  }
}

const loadVotingMatters = async () => {
  try {
    const response = await votingMatterApi.getVotingMatters(props.associationId, props.gathering.id)
    votingMatters.value = response.data.sort((a, b) => a.order_index - b.order_index)
    ballotVotes.value = {}
  } catch (err: any) {
    error.value = err.response?.data?.error || err.message || t('gatherings.matters.loadError')
    message.error(error.value ?? 'An error occurred')
  }
}

const handleOwnerChange = async () => {
  selectedUnitIds.value = []
  invitationLink.value = null
  invitationLinkCopied.value = false
  ownerHasActiveInvitation.value = false
  if (!selectedOwnerId.value) return
  invitationLinkLoading.value = true
  try {
    const res = await invitationApi.list(props.associationId, props.gathering.id)
    const active = (res.data as MemberInvitation[]).find(
      inv => inv.owner_id === selectedOwnerId.value && inv.status === 'active'
    )
    ownerHasActiveInvitation.value = !!active
  } catch {
    // non-critical, ignore
  } finally {
    invitationLinkLoading.value = false
  }
}

function defaultInvitationExpiry(): string {
  const base = props.gathering.scheduled_date
    ? new Date(props.gathering.scheduled_date)
    : new Date()
  base.setMonth(base.getMonth() + 3)
  if (base.getTime() <= Date.now()) {
    const fallback = new Date()
    fallback.setMonth(fallback.getMonth() + 1)
    return fallback.toISOString()
  }
  return base.toISOString()
}

const generateInvitationLink = async () => {
  if (!selectedOwnerId.value) return
  invitationLinkGenerating.value = true
  try {
    const res = await invitationApi.create(props.associationId, props.gathering.id, {
      owner_id: selectedOwnerId.value,
      expires_at: defaultInvitationExpiry()
    })
    invitationLink.value = `${MEMBER_BASE_URL}/${res.data.token}`
  } catch (err: any) {
    error.value = err.response?.data?.error ?? err.message ?? 'Failed to generate invitation link'
  } finally {
    invitationLinkGenerating.value = false
  }
}

const copyInvitationLink = async () => {
  if (!invitationLink.value) return
  await navigator.clipboard.writeText(invitationLink.value)
  invitationLinkCopied.value = true
  setTimeout(() => { invitationLinkCopied.value = false }, 2000)
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
    for (const matter of votingMatters.value) {
      ballotContent[String(matter.id)] = {
        matter_id: matter.id,
        values: ballotVotes.value[String(matter.id)] ?? []
      }
    }

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
    message.error(error.value ?? 'An error occurred')
  } finally {
    submitting.value = false
  }
}

const resetWizard = () => {
  currentStep.value = 0
  isDelegateVoting.value = false
  selectedOwnerId.value = null
  selectedUnitIds.value = []
  ownerHasActiveInvitation.value = false
  invitationLink.value = null
  invitationLinkCopied.value = false
  delegateData.value = {
    name: '',
    identification: '',
    document: ''
  }
  ballotVotes.value = {}

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
