<template>
  <NSpin :show="loading">
    <NAlert v-if="fetchError" type="error" style="margin-bottom: 16px">
      {{ fetchError }}
    </NAlert>

    <template v-if="context">
      <!-- Summary header -->
      <NCard style="margin-bottom: 16px">
        <NDescriptions :column="3" label-placement="top" size="small">
          <NDescriptionsItem :label="t('owner')">{{ context.owner.name }}</NDescriptionsItem>
          <NDescriptionsItem :label="t('units')">{{ context.units.length }}</NDescriptionsItem>
          <NDescriptionsItem :label="t('votingWeight')">{{ totalWeight.toFixed(4) }}</NDescriptionsItem>
        </NDescriptions>
        <div style="margin-top: 8px">
          <NTag :type="statusTagType" size="small">{{ context.gathering.status.toUpperCase() }}</NTag>
        </div>
      </NCard>

      <!-- Gathering not active -->
      <NAlert
        v-if="context.gathering.status !== 'active'"
        :type="context.gathering.status === 'tallied' ? 'info' : 'warning'"
      >
        <template v-if="['draft', 'scheduled'].includes(context.gathering.status)">
          {{ t('statusNotStarted') }}
        </template>
        <template v-else-if="context.gathering.status === 'tallied'">
          {{ t('statusTallied') }}
        </template>
        <template v-else>
          {{ t('statusClosed') }}
        </template>
      </NAlert>

      <!-- Active gathering -->
      <template v-else>
        <!-- Read-only receipt (already voted or just submitted) -->
        <NCard v-if="receipt">
          <template #header>
            <span style="color: #18a058">✓ {{ t('ballotSubmitted') }}</span>
          </template>

          <NDescriptions :column="1" label-placement="left" size="small" style="margin-bottom: 16px">
            <NDescriptionsItem :label="t('ballotId')">{{ receipt.ballot_id }}</NDescriptionsItem>
            <NDescriptionsItem :label="t('verificationHash')">
              <NText code style="font-size: 11px; word-break: break-all">
                {{ receipt.ballot_hash }}
              </NText>
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('submittedAt')">
              {{ receipt.submitted_at ? new Date(receipt.submitted_at).toLocaleString() : '—' }}
            </NDescriptionsItem>
          </NDescriptions>

          <NDivider title-placement="left">{{ t('yourVotes') }}</NDivider>

          <div
            v-for="matter in context.matters.filter(m => !m.is_informative)"
            :key="matter.id"
            style="margin-bottom: 12px; padding-left: 4px"
          >
            <NText strong style="display: block">{{ matter.title }}</NText>
            <NText :depth="2" style="margin-top: 4px; display: block; padding-left: 12px">
              {{ formatVotedValues(matter) }}
            </NText>
          </div>
        </NCard>

        <!-- Interactive ballot form -->
        <NCard v-else>
          <template #header>
            <span>{{ t('yourBallot') }}</span>
            <NText :depth="3" style="font-size: 13px; margin-left: 8px">
              — {{ context.gathering.title }}
            </NText>
          </template>

          <NAlert v-if="submitError" type="error" closable style="margin-bottom: 16px" @close="submitError = null">
            {{ submitError }}
          </NAlert>

          <!-- Informative matters -->
          <div
            v-for="matter in informativeMatters"
            :key="matter.id"
            style="margin-bottom: 16px"
          >
            <NCard size="small" embedded>
              <template #header>
                <NText style="font-size: 14px">{{ matter.title }}</NText>
                <NTag size="tiny" style="margin-left: 8px">{{ t('informational') }}</NTag>
              </template>
              <NText :depth="2" style="font-size: 13px">{{ matter.description }}</NText>
            </NCard>
          </div>

          <!-- Votable matters via shared BallotForm -->
          <BallotForm :matters="votableMatters" v-model="ballotVotes" />

          <NSpace justify="end" style="margin-top: 8px">
            <NButton
              type="primary"
              :loading="submitting"
              :disabled="!canSubmit"
              @click="handleSubmit"
            >
              {{ t('submitBallot') }}
            </NButton>
          </NSpace>
        </NCard>
      </template>
    </template>
  </NSpin>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NAlert,
  NButton,
  NCard,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NSpace,
  NSpin,
  NTag,
  NText
} from 'naive-ui'
import BallotForm from './BallotForm.vue'
import type { VotingService, MemberContext, MatterInfo, BallotVote } from './types'
import { HttpError } from './types'

const { t } = useI18n({
  useScope: 'local',
  messages: {
    en: {
      owner: 'Owner',
      units: 'Units',
      votingWeight: 'Voting weight',
      yourBallot: 'Your Ballot',
      submitBallot: 'Submit Ballot',
      ballotSubmitted: 'Ballot Submitted',
      ballotId: 'Ballot ID',
      verificationHash: 'Verification hash',
      submittedAt: 'Submitted at',
      yourVotes: 'Your votes',
      informational: 'Informational',
      yes: 'Yes', no: 'No', abstain: 'Abstain',
      statusNotStarted: 'Voting has not started yet. Please check back later.',
      statusTallied: 'Voting has closed. Results are being tallied.',
      statusClosed: 'Voting is closed.',
      errAlreadySubmitted: 'A ballot has already been submitted for this gathering.',
      errInvalidBallot: 'Invalid ballot.',
    },
    ro: {
      owner: 'Proprietar',
      units: 'Unități',
      votingWeight: 'Pondere de vot',
      yourBallot: 'Buletinul dvs. de vot',
      submitBallot: 'Trimite buletinul',
      ballotSubmitted: 'Buletin trimis',
      ballotId: 'ID buletin',
      verificationHash: 'Hash de verificare',
      submittedAt: 'Trimis la',
      yourVotes: 'Voturile dvs.',
      informational: 'Informativ',
      yes: 'Da', no: 'Nu', abstain: 'Abținere',
      statusNotStarted: 'Votul nu a început încă. Verificați mai târziu.',
      statusTallied: 'Votul s-a încheiat. Rezultatele sunt în curs de numărare.',
      statusClosed: 'Votul este închis.',
      errAlreadySubmitted: 'Un buletin de vot a fost deja trimis pentru această adunare.',
      errInvalidBallot: 'Buletin de vot invalid.',
    },
    ru: {
      owner: 'Владелец',
      units: 'Единицы',
      votingWeight: 'Вес голоса',
      yourBallot: 'Ваш бюллетень',
      submitBallot: 'Подать бюллетень',
      ballotSubmitted: 'Бюллетень подан',
      ballotId: 'ID бюллетеня',
      verificationHash: 'Хэш верификации',
      submittedAt: 'Подан в',
      yourVotes: 'Ваши голоса',
      informational: 'Информационный',
      yes: 'Да', no: 'Нет', abstain: 'Воздержаться',
      statusNotStarted: 'Голосование ещё не началось. Загляните позже.',
      statusTallied: 'Голосование завершено. Результаты подсчитываются.',
      statusClosed: 'Голосование закрыто.',
      errAlreadySubmitted: 'Бюллетень для этого собрания уже был подан.',
      errInvalidBallot: 'Недействительный бюллетень.',
    },
  }
})

const props = defineProps<{
  service: VotingService
}>()

const loading = ref(false)
const submitting = ref(false)
const fetchError = ref<string | null>(null)
const submitError = ref<string | null>(null)
const context = ref<MemberContext | null>(null)
const receipt = ref<{ ballot_id: number; ballot_hash: string; submitted_at: string | null; ballot_content: Record<string, BallotVote> } | null>(null)

const ballotVotes = reactive<Record<string, string[]>>({})

const totalWeight = computed(() =>
  context.value?.units.reduce((sum, u) => sum + u.voting_weight, 0) ?? 0
)

const votableMatters = computed(() =>
  (context.value?.matters ?? []).filter(m => !m.is_informative)
)

const informativeMatters = computed(() =>
  (context.value?.matters ?? []).filter(m => m.is_informative)
)

const canSubmit = computed(() =>
  votableMatters.value.length > 0 &&
  votableMatters.value.every(m => (ballotVotes[String(m.id)]?.length ?? 0) > 0)
)

const statusTagType = computed((): 'success' | 'warning' | 'error' | 'info' | 'default' => {
  switch (context.value?.gathering.status) {
    case 'active': return 'success'
    case 'scheduled': return 'info'
    case 'tallied': return 'info'
    case 'closed': return 'error'
    default: return 'default'
  }
})

function buildBallotContent(): Record<string, BallotVote> {
  const content: Record<string, BallotVote> = {}
  for (const m of votableMatters.value) {
    const key = String(m.id)
    content[key] = { matter_id: m.id, values: ballotVotes[key] ?? [] }
  }
  return content
}

function formatVotedValues(matter: MatterInfo): string {
  if (!receipt.value) return '—'
  const vote = receipt.value.ballot_content[String(matter.id)]
  if (!vote || vote.values.length === 0) return '—'

  if (matter.voting_config.type === 'ranking') {
    return vote.values
      .map((v, i) => {
        const opt = matter.voting_config.options?.find(o => o.id === v)
        return `${i + 1}. ${opt ? opt.text : v}`
      })
      .join(', ')
  }

  return vote.values
    .map(v => {
      if (v === 'abstain') return t('abstain')
      if (matter.voting_config.type === 'yes_no') return v === 'yes' ? t('yes') : t('no')
      const opt = matter.voting_config.options?.find(o => o.id === v)
      return opt ? opt.text : v
    })
    .join(', ')
}

async function fetchContext() {
  loading.value = true
  fetchError.value = null
  try {
    const data = await props.service.getContext()
    context.value = data
    if (data.ballot) {
      receipt.value = {
        ballot_id: data.ballot.ballot_id,
        ballot_hash: data.ballot.ballot_hash,
        submitted_at: data.ballot.submitted_at,
        ballot_content: data.ballot.ballot_content
      }
    }
  } catch (err) {
    fetchError.value = err instanceof Error ? err.message : 'Network error'
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (!canSubmit.value) return
  submitting.value = true
  submitError.value = null
  const ballotContent = buildBallotContent()
  try {
    const data = await props.service.submitBallot(ballotContent)
    receipt.value = {
      ballot_id: data.ballot_id,
      ballot_hash: data.ballot_hash,
      submitted_at: data.submitted_at,
      ballot_content: data.ballot_content ?? ballotContent
    }
  } catch (err) {
    if (err instanceof HttpError) {
      if (err.status === 409) submitError.value = t('errAlreadySubmitted')
      else if (err.status === 400) submitError.value = err.message || t('errInvalidBallot')
      else submitError.value = err.message
    } else {
      submitError.value = err instanceof Error ? err.message : 'Network error'
    }
  } finally {
    submitting.value = false
  }
}

onMounted(fetchContext)
</script>

<style scoped>
.n-card {
  border-radius: 8px;
}
</style>
