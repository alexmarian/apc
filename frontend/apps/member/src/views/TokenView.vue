<template>
  <NSpin v-if="loading" size="large" style="display: flex; justify-content: center; padding: 64px 0" />

  <NResult
    v-else-if="unauthorized"
    status="403"
    title="Link Invalid or Expired"
    description="This voting link is no longer valid. Please contact your association administrator for a new invitation."
  />

  <NAlert v-else-if="fetchError" type="error" style="margin: 32px auto; max-width: 600px">
    {{ fetchError }}
  </NAlert>

  <template v-else-if="status">
    <VotingWidget v-if="status === 'active'" :service="service" />
    <VotingResultsWidget v-else-if="status === 'tallied'" :service="service" />

    <NResult
      v-else-if="status === 'draft' || status === 'scheduled'"
      status="info"
      title="Voting Has Not Started Yet"
      :description="`This gathering is currently ${status}. Please check back when voting opens.`"
    >
      <template #footer>
        <NTag type="default" size="large">{{ status.toUpperCase() }}</NTag>
      </template>
    </NResult>

    <NResult
      v-else-if="status === 'closed'"
      status="info"
      title="Voting Has Ended"
      description="Voting for this gathering has closed. Results are being tallied and will be available shortly."
    >
      <template #footer>
        <NTag type="warning" size="large">CLOSED</NTag>
      </template>
    </NResult>

    <NResult
      v-else
      status="info"
      title="Gathering Unavailable"
      :description="`This gathering is currently in an unexpected state: ${status}.`"
    />
  </template>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { NAlert, NResult, NSpin, NTag } from 'naive-ui'
import { VotingWidget, VotingResultsWidget, createMemberVotingService } from '@apc/voting-widgets'
import type { VotingService } from '@apc/voting-widgets'
import { HttpError } from '@apc/voting-widgets'

const route = useRoute()
const token = route.params.token as string
const service: VotingService = createMemberVotingService(token)

const loading = ref(true)
const unauthorized = ref(false)
const fetchError = ref<string | null>(null)
const status = ref<string | null>(null)

onMounted(async () => {
  try {
    const data = await service.getContext()
    status.value = data.gathering?.status ?? null
  } catch (err) {
    if (err instanceof HttpError && err.status === 401) {
      unauthorized.value = true
    } else {
      fetchError.value = err instanceof Error ? err.message : 'Network error'
    }
  } finally {
    loading.value = false
  }
})
</script>
