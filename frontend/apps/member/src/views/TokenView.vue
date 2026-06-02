<template>
  <NSpin v-if="loading" size="large" style="display: flex; justify-content: center; padding: 64px 0" />

  <NResult
    v-else-if="unauthorized"
    status="403"
    :title="t('linkInvalid')"
    :description="t('linkInvalidDesc')"
  />

  <NAlert v-else-if="fetchError" type="error" style="margin: 32px auto; max-width: 600px">
    {{ fetchError }}
  </NAlert>

  <template v-else-if="context">
    <VotingWidget v-if="context.gathering.status === 'active'" :service="service" :initial-context="context" />
    <VotingResultsWidget v-else-if="context.gathering.status === 'tallied'" :service="service" :initial-context="context" />

    <NResult
      v-else-if="context.gathering.status === 'draft' || context.gathering.status === 'published' || context.gathering.status === 'scheduled'"
      status="info"
      :title="t('notStartedTitle')"
      :description="t('notStartedDesc', { status: context.gathering.status })"
    >
      <template #footer>
        <NTag type="default" size="large">{{ context.gathering.status.toUpperCase() }}</NTag>
      </template>
    </NResult>

    <NResult
      v-else-if="context.gathering.status === 'closed'"
      status="info"
      :title="t('endedTitle')"
      :description="t('endedDesc')"
    >
      <template #footer>
        <NTag type="warning" size="large">CLOSED</NTag>
      </template>
    </NResult>

    <NResult
      v-else
      status="info"
      :title="t('unavailableTitle')"
      :description="t('unavailableDesc', { status: context.gathering.status })"
    />
  </template>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NAlert, NResult, NSpin, NTag } from 'naive-ui'
import { VotingWidget, VotingResultsWidget, createMemberVotingService } from '@apc/voting-widgets'
import type { VotingService, MemberContext } from '@apc/voting-widgets'
import { HttpError } from '@apc/voting-widgets'

const { t } = useI18n()

const route = useRoute()
const token = route.params.token as string
const service: VotingService = createMemberVotingService(token)

const loading = ref(true)
const unauthorized = ref(false)
const fetchError = ref<string | null>(null)
const context = ref<MemberContext | null>(null)

onMounted(async () => {
  try {
    context.value = await service.getContext()
  } catch (err) {
    if (err instanceof HttpError && err.status === 401) {
      unauthorized.value = true
    } else {
      fetchError.value = err instanceof Error ? err.message : t('networkError')
    }
  } finally {
    loading.value = false
  }
})
</script>
