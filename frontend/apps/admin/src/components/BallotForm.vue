<template>
  <div class="ballot-form">
    <NCard>
      <template #header>
        <h3>{{ $t('gatherings.voting.ballotInfo') }}</h3>
        <p v-if="isOfflineMode">{{ $t('gatherings.voting.offlineBallotDescription') }}</p>
      </template>

      <NSpin :show="loading">
        <NAlert v-if="error" type="error" closable @close="error = null">
          {{ error }}
        </NAlert>

        <NForm ref="formRef" :model="formData" :rules="rules">
          <div v-for="matter in matters" :key="matter.id" class="matter-section">
            <NCard size="small">
              <template #header>
                <h4>{{ matter.title }}</h4>
                <p>{{ matter.description }}</p>
              </template>

              <NFormItem
                :label="$t('gatherings.voting.choice')"
                :path="`votes.${matter.id}`"
              >
                <div v-if="matter.voting_config.type === 'yes_no'" class="vote-options">
                  <NRadioGroup v-model:value="(formData.votes[matter.id] as string[])[0]" @update:value="val => setVote(matter.id, [val])">
                    <NSpace>
                      <NRadio value="yes">{{ $t('common.yes') }}</NRadio>
                      <NRadio value="no">{{ $t('common.no') }}</NRadio>
                      <NRadio v-if="matter.voting_config.allow_abstention" value="abstain">
                        {{ $t('gatherings.voting.abstain') }}
                      </NRadio>
                    </NSpace>
                  </NRadioGroup>
                </div>

                <div v-else-if="matter.voting_config.type === 'single_choice'" class="vote-options">
                  <NRadioGroup v-model:value="(formData.votes[matter.id] as string[])[0]" @update:value="val => setVote(matter.id, [val])">
                    <NSpace vertical>
                      <NRadio
                        v-for="option in matter.voting_config.options"
                        :key="option.id"
                        :value="option.id"
                      >
                        {{ option.text }}
                      </NRadio>
                      <NRadio v-if="matter.voting_config.allow_abstention" value="abstain">
                        {{ $t('gatherings.voting.abstain') }}
                      </NRadio>
                    </NSpace>
                  </NRadioGroup>
                </div>

                <div v-else-if="matter.voting_config.type === 'multiple_choice'" class="vote-options">
                  <NCheckboxGroup
                    :value="formData.votes[matter.id] as string[]"
                    @update:value="val => setVote(matter.id, val)"
                  >
                    <NSpace vertical>
                      <NCheckbox
                        v-for="option in matter.voting_config.options"
                        :key="option.id"
                        :value="option.id"
                      >
                        {{ option.text }}
                      </NCheckbox>
                    </NSpace>
                  </NCheckboxGroup>
                </div>

                <div v-else-if="matter.voting_config.type === 'ranking'" class="vote-options">
                  <p class="ranking-hint">{{ $t('gatherings.voting.rankingHint') }}</p>
                  <div
                    v-for="(optId, idx) in (formData.votes[matter.id] as string[])"
                    :key="optId"
                    class="ranking-item"
                  >
                    <span class="ranking-pos">{{ idx + 1 }}.</span>
                    <span class="ranking-label">{{ getOptionText(matter, optId) }}</span>
                    <NSpace size="small">
                      <NButton size="tiny" :disabled="idx === 0" @click="moveRankUp(matter.id, idx)">↑</NButton>
                      <NButton size="tiny" :disabled="idx === (formData.votes[matter.id] as string[]).length - 1" @click="moveRankDown(matter.id, idx)">↓</NButton>
                    </NSpace>
                  </div>
                </div>
              </NFormItem>
            </NCard>
          </div>

          <div v-if="isOfflineMode" class="offline-info">
            <NCard size="small">
              <template #header>
                <h4>{{ $t('gatherings.voting.offlineInfo') }}</h4>
              </template>

              <NFormItem :label="$t('gatherings.voting.ballotNumber')" path="ballot_number">
                <NInput
                  v-model:value="formData.ballot_number"
                  :placeholder="$t('gatherings.voting.ballotNumberPlaceholder')"
                />
              </NFormItem>

              <NFormItem :label="$t('gatherings.voting.submittedAt')" path="submitted_at">
                <NDatePicker
                  v-model:value="formData.submitted_at"
                  type="datetime"
                  :placeholder="$t('gatherings.voting.submittedAtPlaceholder')"
                />
              </NFormItem>

              <NFormItem :label="$t('gatherings.voting.notes')" path="notes">
                <NInput
                  v-model:value="formData.notes"
                  type="textarea"
                  :placeholder="$t('gatherings.voting.notesPlaceholder')"
                />
              </NFormItem>
            </NCard>
          </div>
        </NForm>

        <div class="ballot-actions">
          <NSpace justify="end">
            <NButton @click="handleReset">
              {{ $t('common.reset') }}
            </NButton>
            <NButton @click="handleCancel">
              {{ $t('common.cancel') }}
            </NButton>
            <NButton
              type="primary"
              @click="handleSubmit"
              :loading="submitting"
              :disabled="!canSubmit"
            >
              {{ $t('gatherings.voting.submit') }}
            </NButton>
          </NSpace>
        </div>
      </NSpin>
    </NCard>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NCard,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NAlert,
  NSpin,
  NRadioGroup,
  NRadio,
  NCheckboxGroup,
  NCheckbox,
  NDatePicker,
  type FormInst,
  type FormRules
} from 'naive-ui'
import { votingMatterApi, votingApi } from '@/services/api'
import type {
  Gathering,
  VotingMatter,
  GatheringParticipant,
  BallotSubmissionRequest,
  BallotVote
} from '@/types/api'

interface Props {
  associationId: number
  gathering: Gathering
  participant: GatheringParticipant
  isOfflineMode?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  isOfflineMode: false
})

const emit = defineEmits<{
  saved: []
  cancelled: []
}>()

const { t } = useI18n()

const formRef = ref<FormInst | null>(null)
const loading = ref(false)
const submitting = ref(false)
const error = ref<string | null>(null)
const matters = ref<VotingMatter[]>([])

const formData = reactive<{
  votes: Record<number, string[]>
  ballot_number: string
  submitted_at: number | null
  notes: string
}>({
  votes: {},
  ballot_number: '',
  submitted_at: null,
  notes: ''
})

const canSubmit = computed(() => {
  const allAnswered = matters.value.every(matter => {
    const v = formData.votes[matter.id]
    return v && v.length > 0
  })

  if (props.isOfflineMode) {
    return allAnswered && formData.ballot_number.trim() !== ''
  }

  return allAnswered
})

const rules: FormRules = {
  ballot_number: [
    { required: true, message: t('gatherings.voting.ballotNumberRequired'), trigger: 'blur' }
  ]
}

const setVote = (matterId: number, values: string[]) => {
  formData.votes[matterId] = values
}

const getOptionText = (matter: VotingMatter, optId: string): string => {
  return matter.voting_config.options?.find(o => o.id === optId)?.text ?? optId
}

const moveRankUp = (matterId: number, idx: number) => {
  const arr = [...formData.votes[matterId]]
  ;[arr[idx - 1], arr[idx]] = [arr[idx], arr[idx - 1]]
  formData.votes[matterId] = arr
}

const moveRankDown = (matterId: number, idx: number) => {
  const arr = [...formData.votes[matterId]]
  ;[arr[idx], arr[idx + 1]] = [arr[idx + 1], arr[idx]]
  formData.votes[matterId] = arr
}

const initVotes = () => {
  matters.value.forEach(matter => {
    const type = matter.voting_config.type
    if (type === 'ranking') {
      formData.votes[matter.id] = matter.voting_config.options?.map(o => o.id) ?? []
    } else if (type === 'multiple_choice') {
      formData.votes[matter.id] = []
    } else {
      formData.votes[matter.id] = []
    }
  })
}

const loadMatters = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await votingMatterApi.getVotingMatters(props.associationId, props.gathering.id)
    matters.value = response.data.sort((a, b) => a.order_index - b.order_index)
    initVotes()
  } catch (err: unknown) {
    error.value = (err as Error).message || t('common.error')
  } finally {
    loading.value = false
  }
}

const handleReset = () => {
  initVotes()
  formData.ballot_number = ''
  formData.submitted_at = null
  formData.notes = ''
}

const handleCancel = () => {
  emit('cancelled')
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    submitting.value = true
    error.value = null

    const ballotContent: Record<string, BallotVote> = {}
    matters.value
      .filter(matter => formData.votes[matter.id]?.length > 0)
      .forEach(matter => {
        ballotContent[String(matter.id)] = {
          matter_id: matter.id,
          values: formData.votes[matter.id]
        }
      })

    const payload: BallotSubmissionRequest = {
      voter_type: props.participant.participant_type,
      owner_id: props.participant.owner_id ?? 0,
      delegating_owner_id: props.participant.delegating_owner_id,
      delegation_document_ref: props.participant.delegation_document_ref,
      unit_ids: props.participant.units_info,
      ballot_content: ballotContent
    }

    await votingApi.submitBallot(props.associationId, props.gathering.id, payload)

    emit('saved')

  } catch (err: unknown) {
    error.value = (err as Error).message || t('common.error')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadMatters()
})
</script>

<style scoped>
.ballot-form {
  margin-top: 16px;
}

.matter-section {
  margin-bottom: 16px;
}

.vote-options {
  margin-top: 8px;
}

.offline-info {
  margin-top: 16px;
}

.ballot-actions {
  margin-top: 24px;
}

.ranking-hint {
  font-size: 12px;
  color: #999;
  margin-bottom: 8px;
}

.ranking-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  border-bottom: 1px solid #f0f0f0;
}

.ranking-pos {
  font-weight: bold;
  min-width: 24px;
}

.ranking-label {
  flex: 1;
}
</style>
