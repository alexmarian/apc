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
        <!-- Overall Statistics -->
        <NCard>
          <template #header>
            <h3>{{ $t('gatherings.results.statistics') }}</h3>
          </template>
          
          <NGrid :cols="4" :x-gap="16">
            <NGridItem>
              <NStatistic
                :label="$t('gatherings.results.qualifiedUnits')"
                :value="results.statistics.qualified_units"
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
                :label="$t('gatherings.results.checkedInParticipants')"
                :value="results.statistics.checked_in_participants"
              />
            </NGridItem>
            <NGridItem>
              <NStatistic
                :label="$t('gatherings.results.participationRate')"
                :value="results.statistics.participation_rate"
                suffix="%"
              />
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

            <NGrid :cols="2" :x-gap="16">
              <!-- Vote Results -->
              <NGridItem>
                <h5>{{ $t('gatherings.results.votes') }}</h5>
                <NDataTable
                  :columns="voteColumns"
                  :data="result.votes"
                  :pagination="false"
                  size="small"
                />
              </NGridItem>

              <!-- Statistics -->
              <NGridItem>
                <h5>{{ $t('gatherings.results.statistics') }}</h5>
                <NDescriptions :column="1" size="small">
                  <NDescriptionsItem :label="$t('gatherings.results.totalParticipants')">
                    {{ result.statistics.total_participants }}
                  </NDescriptionsItem>
                  <NDescriptionsItem :label="$t('gatherings.results.totalVotes')">
                    {{ result.statistics.total_votes }}
                  </NDescriptionsItem>
                  <NDescriptionsItem :label="$t('gatherings.results.totalWeight')">
                    {{ result.statistics.total_weight }}
                  </NDescriptionsItem>
                  <NDescriptionsItem :label="$t('gatherings.results.abstentions')">
                    {{ result.statistics.abstentions }}
                  </NDescriptionsItem>
                  <NDescriptionsItem :label="$t('gatherings.results.participationRate')">
                    {{ result.statistics.participation_rate }}%
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
import { ref, computed, onMounted, watch, h } from 'vue'
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
  type DataTableColumns
} from 'naive-ui'
import { votingApi } from '@/services/api'
import type { Gathering, VotingResults, VoteResult } from '@/types/api'

interface Props {
  associationId: number
  gathering: Gathering
}

const props = defineProps<Props>()

const { t } = useI18n()

const loading = ref(false)
const error = ref<string | null>(null)
const results = ref<VotingResults | null>(null)

const voteColumns: DataTableColumns<VoteResult> = [
  {
    title: t('gatherings.results.choice'),
    key: 'choice',
    width: 120
  },
  {
    title: t('gatherings.results.votes'),
    key: 'vote_count',
    width: 80
  },
  {
    title: t('gatherings.results.percentage'),
    key: 'percentage',
    width: 100,
    render: (row) => h(NProgress, {
      type: 'line',
      percentage: row.percentage,
      height: 8,
      showIndicator: false
    })
  },
  {
    title: t('gatherings.results.weight'),
    key: 'weight_sum',
    width: 80
  },
  {
    title: t('gatherings.results.weightPercentage'),
    key: 'weight_percentage',
    width: 100,
    render: (row) => h(NProgress, {
      type: 'line',
      percentage: row.weight_percentage,
      height: 8,
      showIndicator: false
    })
  }
]

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
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

// Watch for gathering status changes and reload results
watch(() => props.gathering.status, (newStatus, oldStatus) => {
  // Reload results when status changes to closed or tallied
  if ((newStatus === 'closed' || newStatus === 'tallied') && newStatus !== oldStatus) {
    loadResults()
  }
})

// Watch for gathering ID changes (in case user navigates to different gathering)
watch(() => props.gathering.id, () => {
  // Clear current results and reload
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