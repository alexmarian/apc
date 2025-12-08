<template>
  <div class="voting-interface">
    <NSpin :show="loading">
      <NAlert v-if="error" type="error" closable @close="error = null">
        {{ error }}
      </NAlert>

      <div v-if="!participant" class="no-participant">
        <NCard>
          <NAlert type="warning">
            {{ $t('gatherings.voting.noParticipant') }}
          </NAlert>
        </NCard>
      </div>

      <div v-else-if="!participant.checked_in_at" class="not-checked-in">
        <NCard>
          <NAlert type="warning">
            {{ $t('gatherings.voting.mustCheckIn') }}
          </NAlert>
        </NCard>
      </div>

      <div v-else class="voting-form">
        <NCard>
          <template #header>
            <h3>{{ $t('gatherings.voting.ballot') }}</h3>
            <p>{{ $t('gatherings.voting.instructions') }}</p>
          </template>

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
                    <NRadioGroup v-model:value="formData.votes[matter.id]">
                      <NSpace>
                        <NRadio value="yes">{{ $t('common.yes') }}</NRadio>
                        <NRadio value="no">{{ $t('common.no') }}</NRadio>
                        <NRadio v-if="matter.voting_config.allow_abstention" value="abstain">
                          {{ $t('gatherings.voting.abstain') }}
                        </NRadio>
                      </NSpace>
                    </NRadioGroup>
                  </div>

                  <div v-else-if="matter.voting_config.type === 'single_choice'"
                       class="vote-options">
                    <NRadioGroup v-model:value="formData.votes[matter.id]">
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

                  <div v-else-if="matter.voting_config.type === 'multiple_choice'"
                       class="vote-options">
                    <NCheckboxGroup v-model:value="formData.votes[matter.id]">
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

            <div class="voting-summary">
              <NCard>
                <template #header>
                  <h4>{{ $t('gatherings.voting.summary') }}</h4>
                </template>

                <NDescriptions :column="1">
                  <NDescriptionsItem :label="$t('gatherings.participants.owner')">
                    {{ participant.owner_name }}
                  </NDescriptionsItem>
                  <NDescriptionsItem :label="$t('gatherings.participants.units')">
                    {{ participant.unit_ids.length }}
                  </NDescriptionsItem>
                  <NDescriptionsItem :label="$t('gatherings.voting.totalWeight')">
                    {{ totalWeight }}
                  </NDescriptionsItem>
                </NDescriptions>
              </NCard>
            </div>
          </NForm>

          <div class="voting-actions">
            <NSpace justify="end">
              <NButton @click="resetForm">
                {{ $t('common.reset') }}
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
        </NCard>
      </div>
    </NSpin>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NCard,
  NForm,
  NFormItem,
  NButton,
  NSpace,
  NAlert,
  NSpin,
  NRadioGroup,
  NRadio,
  NCheckboxGroup,
  NCheckbox,
  NSelect,
  NDescriptions,
  NDescriptionsItem,
  type FormInst,
  type FormRules
} from 'naive-ui'
import { votingMatterApi, participantApi, votingApi } from '@/services/api'
import type {
  Gathering,
  VotingMatter,
  GatheringParticipant,
  Vote
} from '@/types/api'

interface Props {
  associationId: number
  gathering: Gathering
}

const props = defineProps<Props>()

const { t } = useI18n()

const formRef = ref<FormInst | null>(null)
const loading = ref(false)
const submitting = ref(false)
const error = ref<string | null>(null)
const matters = ref<VotingMatter[]>([])
const participant = ref<GatheringParticipant | null>(null)

const formData = reactive<{
  votes: Record<number, any>
}>({
  votes: {}
})

const totalWeight = computed(() => {
  return participant.value?.unit_ids.length || 0
})

const canSubmit = computed(() => {
  return matters.value.every(matter =>
    formData.votes[matter.id] !== undefined && formData.votes[matter.id] !== null
  )
})

const rules: FormRules = {
  // Dynamic rules based on matters
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

const loadData = async () => {
  loading.value = true
  error.value = null

  try {
    // Load matters
    const mattersResponse = await votingMatterApi.getVotingMatters(props.associationId, props.gathering.id)
    matters.value = mattersResponse.data.sort((a, b) => a.order_index - b.order_index)

    // Load participants to find current user's participant record
    const participantsResponse = await participantApi.getParticipants(props.associationId, props.gathering.id)
    // In a real app, you would filter by current user
    participant.value = participantsResponse.data[0] || null

    // Initialize form data
    matters.value.forEach(matter => {
      formData.votes[matter.id] = matter.voting_config.type === 'multiple_choice' ? [] : null
    })

  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  matters.value.forEach(matter => {
    formData.votes[matter.id] = matter.voting_config.type === 'multiple_choice' ? [] : null
  })
}

const handleSubmit = async () => {
  if (!formRef.value || !participant.value) return

  try {
    submitting.value = true
    error.value = null

    const votes: Vote[] = matters.value.map(matter => ({
      matter_id: matter.id,
      choice: formData.votes[matter.id],
      weight: totalWeight.value
    }))

    await votingApi.submitBallot(props.associationId, props.gathering.id, {
      participantId: participant.value.id,
      votes
    })

    // Show success message
    error.value = null
    // In a real app, you might redirect or show a success modal

  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.voting-interface {
  margin-top: 16px;
}

.no-participant,
.not-checked-in {
  margin: 16px 0;
}

.voting-form {
  margin-top: 16px;
}

.matter-section {
  margin-bottom: 16px;
}

.vote-options {
  margin-top: 8px;
}

.voting-summary {
  margin-top: 16px;
}

.voting-actions {
  margin-top: 24px;
}
</style>
