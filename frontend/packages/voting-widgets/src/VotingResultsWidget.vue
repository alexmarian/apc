<template>
  <NSpin :show="loading">
    <NAlert v-if="fetchError" type="error" style="margin-bottom: 16px">
      {{ fetchError }}
    </NAlert>

    <template v-if="context">
      <!-- Summary header -->
      <NCard style="margin-bottom: 16px">
        <NDescriptions :column="2" label-placement="top" size="small">
          <NDescriptionsItem :label="t('gathering')">{{ context.gathering.title }}</NDescriptionsItem>
          <NDescriptionsItem :label="t('status')">
            <NTag :type="statusTagType" size="small">{{ context.gathering.status.toUpperCase() }}</NTag>
          </NDescriptionsItem>
        </NDescriptions>
      </NCard>

      <!-- Results not yet available -->
      <NAlert v-if="!results" type="info">
        {{ t('notAvailable') }} <strong>{{ context.gathering.status }}</strong>.
        <template v-if="context.gathering.status !== 'tallied'">
          {{ t('willBePublished') }}
        </template>
      </NAlert>

      <!-- Results available -->
      <template v-else>
        <!-- Summary statistics -->
        <NCard size="small" style="margin-bottom: 16px" :title="t('participationSummary')">
          <NDescriptions :column="3" label-placement="top" size="small">
            <NDescriptionsItem :label="t('participated')">
              {{ results.statistics.participating_units }} {{ t('units') }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('voted')">
              {{ results.statistics.voted_units }} {{ t('units') }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('participationRate')">
              {{ results.statistics.participation_rate.toFixed(1) }}%
            </NDescriptionsItem>
          </NDescriptions>
        </NCard>

        <!-- Per-matter breakdown -->
        <div v-for="matter in results.results" :key="matter.matter_id" style="margin-bottom: 16px">
          <NCard size="small">
            <template #header>
              <NSpace align="center" justify="space-between" style="flex-wrap: wrap; gap: 4px">
                <NText strong>{{ matter.matter_title }}</NText>
                <NTag :type="matter.is_passed ? 'success' : 'error'" size="small">
                  {{ matter.is_passed ? t('passed') : t('failed') }}
                </NTag>
              </NSpace>
            </template>

            <!-- Member's own vote notice -->
            <NAlert
              v-if="memberVotedOnMatter(matter.matter_id)"
              type="success"
              size="small"
              style="margin-bottom: 12px"
            >
              {{ t('yourVoteCounted') }}
            </NAlert>
            <NAlert
              v-else-if="context.ballot"
              type="default"
              size="small"
              style="margin-bottom: 12px"
            >
              {{ t('didNotVote') }}
            </NAlert>

            <!-- Vote breakdown -->
            <div v-for="vote in sortedVotes(matter)" :key="vote.choice" style="margin-bottom: 10px">
              <NSpace align="center" justify="space-between" style="margin-bottom: 4px">
                <NSpace align="center" size="small">
                  <NText :style="isMyChoice(matter.matter_id, vote.choice) ? 'font-weight:600;color:#18a058' : ''">
                    {{ formatChoice(vote.choice, matter.voting_config) }}
                  </NText>
                  <NTag v-if="isMyChoice(matter.matter_id, vote.choice)" type="success" size="tiny">
                    {{ t('yourVote') }}
                  </NTag>
                </NSpace>
                <NText :depth="2" style="font-size: 12px">
                  {{ vote.vote_count }} {{ vote.vote_count !== 1 ? t('votes') : t('vote') }}
                  ({{ vote.percentage.toFixed(1) }}%)
                </NText>
              </NSpace>
              <NProgress
                type="line"
                :percentage="vote.percentage"
                :status="progressStatus(vote.choice, matter)"
                :show-indicator="false"
                :height="8"
                :border-radius="4"
              />
            </div>

            <!-- Quorum info -->
            <NDivider style="margin: 8px 0" />
            <NText :depth="3" style="font-size: 12px">
              {{ t('quorum') }}: {{ matter.quorum_info?.met ? t('quorumMet') : t('quorumNotMet') }}
              <template v-if="matter.quorum_info">
                — {{ matter.quorum_info.achieved_percentage.toFixed(1) }}% {{ t('of') }}
                {{ matter.quorum_info.required_percentage }}% {{ t('required') }}
              </template>
            </NText>
          </NCard>
        </div>
      </template>
    </template>
  </NSpin>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NAlert,
  NCard,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NProgress,
  NSpace,
  NSpin,
  NTag,
  NText
} from 'naive-ui'
import type { VotingService, MemberContext, VoteResults, MatterResult, VoteResult, VotingConfig } from './types'

const { t } = useI18n({
  useScope: 'local',
  messages: {
    en: {
      gathering: 'Gathering', status: 'Status',
      participationSummary: 'Participation Summary',
      participated: 'Participated', voted: 'Voted', units: 'units',
      participationRate: 'Participation rate',
      passed: 'PASSED', failed: 'FAILED',
      yourVoteCounted: 'Your vote has been counted for this matter.',
      didNotVote: 'You did not vote on this matter.',
      yourVote: 'Your vote', vote: 'vote', votes: 'votes',
      abstain: 'Abstain',
      quorum: 'Quorum', quorumMet: 'Met', quorumNotMet: 'Not met',
      of: 'of', required: 'required',
      notAvailable: 'Results are not yet available. Current status:',
      willBePublished: 'Results will be published after the gathering is tallied.',
    },
    ro: {
      gathering: 'Adunare', status: 'Status',
      participationSummary: 'Rezumat participare',
      participated: 'Participat', voted: 'Votat', units: 'unități',
      participationRate: 'Rata de participare',
      passed: 'ADOPTAT', failed: 'RESPINS',
      yourVoteCounted: 'Votul dvs. a fost înregistrat pentru acest punct.',
      didNotVote: 'Nu ați votat pentru acest punct.',
      yourVote: 'Votul dvs.', vote: 'vot', votes: 'voturi',
      abstain: 'Abținere',
      quorum: 'Cvorum', quorumMet: 'Întrunit', quorumNotMet: 'Neîntrunit',
      of: 'din', required: 'necesar',
      notAvailable: 'Rezultatele nu sunt disponibile încă. Stare curentă:',
      willBePublished: 'Rezultatele vor fi publicate după numărarea voturilor.',
    },
    ru: {
      gathering: 'Собрание', status: 'Статус',
      participationSummary: 'Сводка участия',
      participated: 'Участвовало', voted: 'Проголосовало', units: 'ед.',
      participationRate: 'Явка',
      passed: 'ПРИНЯТО', failed: 'ОТКЛОНЕНО',
      yourVoteCounted: 'Ваш голос учтён по данному вопросу.',
      didNotVote: 'Вы не голосовали по данному вопросу.',
      yourVote: 'Ваш голос', vote: 'голос', votes: 'голосов',
      abstain: 'Воздержаться',
      quorum: 'Кворум', quorumMet: 'Достигнут', quorumNotMet: 'Не достигнут',
      of: 'из', required: 'требуется',
      notAvailable: 'Результаты пока недоступны. Текущий статус:',
      willBePublished: 'Результаты будут опубликованы после подсчёта голосов.',
    },
  }
})

const props = defineProps<{
  service: VotingService
}>()

const loading = ref(false)
const fetchError = ref<string | null>(null)
const context = ref<MemberContext | null>(null)
const results = ref<VoteResults | null>(null)

const statusTagType = computed((): 'success' | 'warning' | 'error' | 'info' | 'default' => {
  switch (context.value?.gathering.status) {
    case 'active': return 'success'
    case 'scheduled': return 'info'
    case 'tallied': return 'info'
    case 'closed': return 'error'
    default: return 'default'
  }
})

function memberVotedOnMatter(matterId: number): boolean {
  if (!context.value?.ballot) return false
  const vote = context.value.ballot.ballot_content[String(matterId)]
  return !!vote && vote.values.length > 0
}

function isMyChoice(matterId: number, choice: string): boolean {
  if (!context.value?.ballot) return false
  const vote = context.value.ballot.ballot_content[String(matterId)]
  return !!vote && vote.values.includes(choice)
}

function formatChoice(choice: string, config: VotingConfig): string {
  if (choice === 'abstain') return t('abstain')
  if (config.type === 'yes_no') return choice === 'yes' ? 'Yes' : 'No'
  const opt = config.options?.find(o => o.id === choice)
  return opt ? opt.text : choice
}

function sortedVotes(matter: MatterResult): VoteResult[] {
  return [...matter.votes].sort((a, b) => b.vote_count - a.vote_count)
}

function progressStatus(choice: string, matter: MatterResult): 'success' | 'error' | 'default' | 'warning' {
  if (choice === 'abstain') return 'warning'
  if (matter.voting_config.type === 'yes_no') {
    if (choice === 'yes') return matter.is_passed ? 'success' : 'default'
    if (choice === 'no') return matter.is_passed ? 'default' : 'error'
  }
  return 'default'
}

async function fetchContext() {
  loading.value = true
  fetchError.value = null
  try {
    const data = await props.service.getContext()
    context.value = data
    results.value = data.results ?? null
  } catch (err) {
    fetchError.value = err instanceof Error ? err.message : 'Network error'
  } finally {
    loading.value = false
  }
}

onMounted(fetchContext)
</script>

<style scoped>
.n-card {
  border-radius: 8px;
}
</style>
