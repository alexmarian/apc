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
                  <NRadioGroup v-model:value="formData.votes[matter.id] as string">
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
                  <NRadioGroup v-model:value="formData.votes[matter.id] as string">
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
                  <NCheckboxGroup v-model:value="formData.votes[matter.id] as string[]">
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

                <div v-else class="vote-options">
                  <NSelect
                    v-model:value="formData.votes[matter.id]"
                    :options="getMatterOptions(matter)"
                    :placeholder="$t('gatherings.voting.selectChoice')"
                  />
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
  NSelect,
  NDatePicker,
  type FormInst,
  type FormRules
} from 'naive-ui'
import { votingMatterApi, votingApi } from '@/services/api'
import type {
  Gathering,
  VotingMatter,
  GatheringParticipant,
  Vote
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
  votes: Record<number, string | string[] | undefined>
  ballot_number: string
  submitted_at: number | null
  notes: string
}>({
  votes: {},
  ballot_number: '',
  submitted_at: null,
  notes: ''
})

const totalWeight = computed(() => {
  return props.participant.unit_ids.length || 0
})

const canSubmit = computed(() => {
  const allMattersAnswered = matters.value.every(matter =>
    formData.votes[matter.id] !== undefined
  )

  if (props.isOfflineMode) {
    return allMattersAnswered && formData.ballot_number.trim() !== ''
  }

  return allMattersAnswered
})

const rules: FormRules = {
  ballot_number: [
    { required: true, message: t('gatherings.voting.ballotNumberRequired'), trigger: 'blur' }
  ]
}

const getMatterOptions = (matter: VotingMatter) => {
  const options = matter.voting_config.options?.map(option => ({
    label: option.text,
    value: option.id
  })) || []

  if (matter.voting_config.allow_abstention) {
    options.push({
      label: t('gatherings.voting.abstain'),
      value: 'abstain'
    })
  }

  return options
}

const loadMatters = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await votingMatterApi.getVotingMatters(props.associationId, props.gathering.id)
    matters.value = response.data.sort((a, b) => a.order_index - b.order_index)

    // Initialize form data
    matters.value.forEach(matter => {
      formData.votes[matter.id] = matter.voting_config.type === 'multiple_choice' ? [] : undefined
    })

  } catch (err: unknown) {
    error.value = (err as Error).message || t('common.error')
  } finally {
    loading.value = false
  }
}

const handleReset = () => {
  matters.value.forEach(matter => {
    formData.votes[matter.id] = matter.voting_config.type === 'multiple_choice' ? [] : undefined
  })
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

    const votes: Vote[] = matters.value
      .filter(matter => formData.votes[matter.id] !== undefined)
      .map(matter => ({
        matter_id: matter.id,
        choice: formData.votes[matter.id]!,
        weight: totalWeight.value
      }))

    const ballotData = {
      votes,
      ...(props.isOfflineMode && {
        ballot_number: formData.ballot_number,
        submitted_at: formData.submitted_at ? new Date(formData.submitted_at).toISOString() : new Date().toISOString(),
        notes: formData.notes,
        is_offline: true
      })
    }

    await votingApi.submitBallot(props.associationId, props.gathering.id, { participantId: props.participant.id, ...ballotData })

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
</style>
