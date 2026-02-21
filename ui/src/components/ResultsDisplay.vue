<template>
  <div class="results-display">
    <NSpin :show="loading">
      <NAlert v-if="error" type="error" closable @close="error = null">
        {{ error }}
      </NAlert>

      <div v-if="!results" class="no-results">
        <NCard>
          <NAlert type="info">
            {{ $t('gatherings.results.noResults') }}
          </NAlert>
        </NCard>
      </div>

      <div v-else class="results-content">
        <div style="display: flex; justify-content: flex-end; margin-bottom: 16px;">
          <NButton
            type="primary"
            @click="handleDownloadResults"
            :loading="downloading"
          >
            <template #icon>
              <NIcon><DownloadOutlined /></NIcon>
            </template>
            {{ $t('gatherings.results.downloadResults') }}
          </NButton>
        </div>

        <!-- Overall Statistics -->
        <NCard>
          <template #header>
            <h3>{{ $t('gatherings.results.statistics') }}</h3>
          </template>

          <NGrid :cols="3" :x-gap="16" :y-gap="16">
            <NGridItem>
              <NStatistic
                :label="$t('gatherings.results.qualifiedUnits')"
                :value="results.statistics.qualified_units"
              />
            </NGridItem>
            <NGridItem>
              <NStatistic
                :label="$t('gatherings.results.qualifiedWeight')"
                :value="formatWeight(results.statistics.qualified_weight)"
              />
            </NGridItem>
            <NGridItem>
              <NStatistic
                :label="$t('gatherings.results.qualifiedArea')"
                :value="formatArea(results.statistics.qualified_area)"
              />
            </NGridItem>
            <NGridItem>
              <NStatistic
                :label="$t('gatherings.results.participatingUnits')"
                :value="results.statistics.participating_units"
              />
            </NGridItem>
            <NGridItem>
              <NStatistic
                :label="$t('gatherings.results.participatingWeight')"
                :value="formatWeight(results.statistics.participating_weight)"
              />
            </NGridItem>
            <NGridItem>
              <NStatistic
                :label="$t('gatherings.results.participationRate')"
                :value="formatPercent(results.statistics.participation_rate)"
                suffix="%"
              >
                <template #suffix>
                  <NTooltip>
                    <template #trigger>
                      <span style="cursor: help; font-size: 12px; color: #888;">
                        &nbsp;({{ $t('gatherings.results.ofQualified') }})
                      </span>
                    </template>
                    {{ $t('gatherings.results.participationRateTooltip') }}
                  </NTooltip>
                </template>
              </NStatistic>
            </NGridItem>
            <NGridItem>
              <NStatistic
                :label="$t('gatherings.results.votedUnits')"
                :value="results.statistics.voted_units"
              />
            </NGridItem>
            <NGridItem>
              <NStatistic
                :label="$t('gatherings.results.votedWeight')"
                :value="formatWeight(results.statistics.voted_weight)"
              />
            </NGridItem>
            <NGridItem>
              <NStatistic
                :label="$t('gatherings.results.votingCompletionRate')"
                :value="formatPercent(results.statistics.voting_completion_rate)"
                suffix="%"
              >
                <template #suffix>
                  <NTooltip>
                    <template #trigger>
                      <span style="cursor: help; font-size: 12px; color: #888;">
                        &nbsp;({{ $t('gatherings.results.ofParticipating') }})
                      </span>
                    </template>
                    {{ $t('gatherings.results.votingCompletionRateTooltip') }}
                  </NTooltip>
                </template>
              </NStatistic>
            </NGridItem>
          </NGrid>
        </NCard>

        <!-- Matter Results -->
        <div v-for="result in results.results" :key="result.matter_id" class="matter-result">
          <NCard>
            <template #header>
              <div class="matter-header">
                <h4>{{ result.matter_title }}</h4>
                <NTag
                  :type="result.is_passed ? 'success' : 'error'"
                  size="large"
                >
                  {{ result.is_passed ? $t('gatherings.results.passed') : $t('gatherings.results.failed') }}
                </NTag>
              </div>
            </template>

            <!-- Quorum info banner -->
            <NAlert
              v-if="result.quorum_info"
              :type="result.quorum_info.met ? 'success' : 'warning'"
              style="margin-bottom: 12px;"
            >
              {{ $t('gatherings.results.quorum') }}:
              {{ formatPercent(result.quorum_info.achieved_percentage) }}%
              {{ $t('gatherings.results.quorumRequired') }}
              {{ formatPercent(result.quorum_info.required_percentage) }}%
              ({{ result.quorum_info.met ? $t('gatherings.results.quorumMet') : $t('gatherings.results.quorumNotMet') }})
            </NAlert>

            <NGrid :cols="2" :x-gap="16">
              <!-- Vote Results -->
              <NGridItem>
                <h5>{{ $t('gatherings.results.votes') }}</h5>
                <NDataTable
                  :columns="getVoteColumns(result)"
                  :data="result.votes"
                  :pagination="false"
                  size="small"
                />
              </NGridItem>

              <!-- Statistics -->
              <NGridItem>
                <h5>{{ $t('gatherings.results.statistics') }}</h5>
                <NDescriptions :column="1" size="small">
                  <NDescriptionsItem :label="$t('gatherings.results.totalVotes')">
                    {{ result.statistics.total_votes }}
                  </NDescriptionsItem>
                  <NDescriptionsItem :label="$t('gatherings.results.totalWeight')">
                    {{ formatWeight(result.statistics.total_weight) }}
                    <NTooltip>
                      <template #trigger>
                        <span style="cursor: help; font-size: 11px; color: #888;">
                          &nbsp;({{ formatPercent(result.statistics.participation_rate) }}% {{ $t('gatherings.results.ofQualified') }})
                        </span>
                      </template>
                      {{ $t('gatherings.results.weightOfQualifiedTooltip') }}
                    </NTooltip>
                  </NDescriptionsItem>
                  <NDescriptionsItem :label="$t('gatherings.results.abstentions')">
                    {{ result.statistics.abstentions }}
                  </NDescriptionsItem>
                </NDescriptions>
              </NGridItem>
            </NGrid>

            <!-- Voting Configuration -->
            <NDivider />
            <NDescriptions :column="3" size="small">
              <NDescriptionsItem :label="$t('gatherings.matters.votingType')">
                {{ $t(`gatherings.matters.votingTypes.${result.voting_config.type}`) }}
              </NDescriptionsItem>
              <NDescriptionsItem :label="$t('gatherings.matters.majorityType')">
                {{ $t(`gatherings.matters.majorityTypes.${result.voting_config.required_majority}`) }}
              </NDescriptionsItem>
              <NDescriptionsItem :label="$t('gatherings.matters.isAnonymous')">
                {{ result.voting_config.is_anonymous ? $t('common.yes') : $t('common.no') }}
              </NDescriptionsItem>
            </NDescriptions>
          </NCard>
        </div>

        <!-- Generation Info -->
        <NCard>
          <template #header>
            <h4>{{ $t('gatherings.results.generationInfo') }}</h4>
          </template>

          <p>{{ $t('gatherings.results.generatedAt') }}: {{ formatDate(results.generated_at) }}</p>
        </NCard>
      </div>
    </NSpin>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, h } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NCard,
  NButton,
  NAlert,
  NSpin,
  NGrid,
  NGridItem,
  NStatistic,
  NTag,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NProgress,
  NIcon,
  NTooltip,
  type DataTableColumns
} from 'naive-ui'
import { DownloadOutlined } from '@vicons/antd'
import { votingApi } from '@/services/api'
import type { Gathering, VotingResults, VoteResult, MatterResult } from '@/types/api'

interface Props {
  associationId: number
  gathering: Gathering
}

const props = defineProps<Props>()

const { t } = useI18n()

const loading = ref(false)
const downloading = ref(false)
const error = ref<string | null>(null)
const results = ref<VotingResults | null>(null)

const formatWeight = (v: number | undefined) => {
  if (v == null) return '0'
  return v.toFixed(4)
}

const formatArea = (v: number | undefined) => {
  if (v == null) return '0'
  return v.toFixed(2)
}

const formatPercent = (v: number | undefined) => {
  if (v == null) return '0.00'
  return v.toFixed(2)
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

// Returns vote table columns; percentage columns show % among votes cast,
// plus a separate column for weight as % of qualified
const getVoteColumns = (result: MatterResult): DataTableColumns<VoteResult> => {
  const qualifiedWeight = results.value?.statistics?.qualified_weight ?? 0

  return [
    {
      title: t('gatherings.results.choice'),
      key: 'choice',
      width: 120
    },
    {
      title: t('gatherings.results.votes'),
      key: 'vote_count',
      width: 60
    },
    {
      title: t('gatherings.results.percentageOfCast'),
      key: 'percentage',
      width: 140,
      render: (row) => h('div', { style: 'display:flex; align-items:center; gap:6px;' }, [
        h(NProgress, {
          type: 'line',
          percentage: row.percentage,
          height: 8,
          showIndicator: false,
          style: 'flex:1; min-width:60px;'
        }),
        h('span', { style: 'font-size:12px; white-space:nowrap;' }, `${formatPercent(row.percentage)}%`)
      ])
    },
    {
      title: t('gatherings.results.weight'),
      key: 'weight_sum',
      width: 80,
      render: (row) => formatWeight(row.weight_sum)
    },
    {
      title: t('gatherings.results.weightPctOfCast'),
      key: 'weight_percentage',
      width: 150,
      render: (row) => h('div', { style: 'display:flex; align-items:center; gap:6px;' }, [
        h(NProgress, {
          type: 'line',
          percentage: row.weight_percentage,
          height: 8,
          showIndicator: false,
          style: 'flex:1; min-width:60px;'
        }),
        h('span', { style: 'font-size:12px; white-space:nowrap;' }, `${formatPercent(row.weight_percentage)}%`)
      ])
    },
    {
      title: t('gatherings.results.weightPctOfQualified'),
      key: 'weight_sum_qualified',
      width: 150,
      render: (row) => {
        const pct = qualifiedWeight > 0 ? (row.weight_sum / qualifiedWeight * 100) : 0
        return h('div', { style: 'display:flex; align-items:center; gap:6px;' }, [
          h(NProgress, {
            type: 'line',
            percentage: pct,
            height: 8,
            showIndicator: false,
            style: 'flex:1; min-width:60px;'
          }),
          h('span', { style: 'font-size:12px; white-space:nowrap;' }, `${formatPercent(pct)}%`)
        ])
      }
    }
  ]
}

const loadResults = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await votingApi.getResults(props.associationId, props.gathering.id)
    results.value = response.data
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  } finally {
    loading.value = false
  }
}

const handleDownloadResults = async () => {
  downloading.value = true
  try {
    const response = await votingApi.downloadResults(props.associationId, props.gathering.id)

    const blob = new Blob([response.data], { type: 'text/markdown' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `voting-results-${props.gathering.title}-${new Date().toISOString().split('T')[0]}.md`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  } finally {
    downloading.value = false
  }
}

// Watch for gathering status changes and reload results
watch(() => props.gathering.status, (newStatus, oldStatus) => {
  if ((newStatus === 'closed' || newStatus === 'tallied') && newStatus !== oldStatus) {
    loadResults()
  }
})

// Watch for gathering ID changes
watch(() => props.gathering.id, () => {
  results.value = null
  loadResults()
})

onMounted(() => {
  loadResults()
})
</script>

<style scoped>
.results-display {
  margin-top: 16px;
}

.no-results {
  margin: 16px 0;
}

.results-content {
  margin-top: 16px;
}

.matter-result {
  margin-top: 16px;
}

.matter-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.matter-header h4 {
  margin: 0;
}
</style>
