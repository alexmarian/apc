<template>
  <NSpin :show="loading">
    <NAlert v-if="error" type="error" closable style="margin-bottom: 16px" @close="error = null">
      {{ error }}
    </NAlert>

    <NSpace justify="space-between" align="center" style="margin-bottom: 16px">
      <NText strong>{{ t('gatherings.invitations.qualifiedOwners', filteredRows.length) }}</NText>
      <NSpace>
        <NButton @click="handleGenerateAll" :loading="bulkGenerating" :disabled="ownersWithoutActive.length === 0">
          {{ t('gatherings.invitations.generateForVisible', { count: ownersWithoutActive.length }) }}
        </NButton>
        <NButton @click="exportCsv" :disabled="activeRows.length === 0">
          {{ t('gatherings.invitations.exportCsv') }}
        </NButton>
        <PrintBallot :association-id="associationId" :gathering="gathering" lang="ro">
          {{ $t('gatherings.invitations.printBallotRo') }}
        </PrintBallot>
        <PrintBallot :association-id="associationId" :gathering="gathering" lang="ru">
          {{ $t('gatherings.invitations.printBallotRu') }}
        </PrintBallot>
      </NSpace>
    </NSpace>

    <NInput
      v-model:value="searchQuery"
      :placeholder="t('gatherings.invitations.searchOwners')"
      clearable
      style="margin-bottom: 12px"
    />

    <NDataTable :columns="columns" :data="filteredRows" :bordered="false" size="small" />
  </NSpin>

  <!-- Generate link modal -->
  <NModal
    v-model:show="generateModalVisible"
    style="width: 480px"
    preset="card"
    :title="generateModalOwner
      ? t('gatherings.invitations.generateLinkFor', { name: generateModalOwner.ownerName })
      : t('gatherings.invitations.generateAllLinks')"
    :mask-closable="false"
    :close-on-esc="true"
  >
    <template v-if="!generatedTokens.length">
        <NFormItem :label="t('gatherings.invitations.expiryDate')" :feedback="expiryError || undefined" :validation-status="expiryError ? 'error' : undefined">
          <NDatePicker
            v-model:value="expiryTimestamp"
            type="date"
            style="width: 100%"
            :is-date-disabled="isPastDate"
          />
        </NFormItem>
        <NSpace justify="end" style="margin-top: 8px">
          <NButton @click="generateModalVisible = false">{{ t('common.cancel') }}</NButton>
          <NButton type="primary" :loading="generating" @click="confirmGenerate">
            {{ t('gatherings.invitations.generate') }}
          </NButton>
        </NSpace>
    </template>

    <template v-else>
        <NAlert type="warning" style="margin-bottom: 16px">
          {{ t('gatherings.invitations.linksNotRecoverable') }}
        </NAlert>

        <div v-for="item in generatedTokens" :key="item.ownerId" style="margin-bottom: 12px">
          <NText strong style="display: block; margin-bottom: 4px">{{ item.ownerName }}</NText>
          <NInputGroup>
            <NInput :value="item.url" readonly style="font-size: var(--text-sm); font-family: monospace" />
            <NButton @click="copyUrl(item.url)" style="min-width: 70px">
              {{ copiedUrl === item.url ? `✓ ${t('gatherings.invitations.copied')}` : t('gatherings.invitations.copy') }}
            </NButton>
          </NInputGroup>
        </div>

        <NSpace justify="end" style="margin-top: 16px">
          <NButton v-if="generatedTokens.length > 1" @click="exportGeneratedCsv">
            {{ t('gatherings.invitations.exportCsv') }}
          </NButton>
          <NButton type="primary" @click="closeGenerateModal">
            {{ t('gatherings.invitations.done') }}
          </NButton>
        </NSpace>
    </template>
  </NModal>

</template>

<script setup lang="ts">
import { ref, computed, watch, h } from 'vue'
import {
  NAlert,
  NButton,
  NDataTable,
  NDatePicker,
  NFormItem,
  NInput,
  NInputGroup,
  NModal,
  NSpace,
  NSpin,
  NTag,
  NText
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { gatheringApi, invitationApi } from '@/services/api'
import { useConfirm } from '@/composables/useConfirm'
import { useI18n } from 'vue-i18n'
import PrintBallot from './PrintBallot.vue'
import type { Gathering, MemberInvitation, QualifiedUnit } from '@/types/api'

const MEMBER_BASE_URL = import.meta.env.VITE_MEMBER_BASE_URL || 'https://member.blocul-nostru.online'

const props = defineProps<{
  associationId: number
  gathering: Gathering
}>()

interface OwnerRow {
  ownerId: number
  ownerName: string
  invitation: MemberInvitation | null
  status: 'none' | 'active' | 'expired' | 'revoked'
}

const { t } = useI18n()
const confirm = useConfirm()

const loading = ref(false)
const error = ref<string | null>(null)
const rows = ref<OwnerRow[]>([])
const generatedUrlsByOwner = ref<Map<number, string>>(new Map())

const generateModalVisible = ref(false)
const generateModalOwner = ref<OwnerRow | null>(null)
const generatedTokens = ref<{ ownerId: number; ownerName: string; url: string }[]>([])
const expiryTimestamp = ref<number | null>(null)
const expiryError = ref<string | null>(null)
const generating = ref(false)
const bulkGenerating = ref(false)
const copiedUrl = ref<string | null>(null)


const searchQuery = ref('')

function normalize(s: string): string {
  return s.normalize('NFD').replace(/\p{M}/gu, '').toLowerCase()
}

const filteredRows = computed(() => {
  const q = normalize(searchQuery.value.trim())
  if (!q) return rows.value
  return rows.value.filter(r => normalize(r.ownerName).includes(q))
})

const ownersWithoutActive = computed(() => filteredRows.value.filter(r => r.status !== 'active'))
const activeRows = computed(() => filteredRows.value.filter(r => r.status === 'active'))

function defaultExpiry(): number {
  const base = props.gathering.scheduled_date
    ? new Date(props.gathering.scheduled_date)
    : new Date()
  base.setMonth(base.getMonth() + 3)
  if (base.getTime() <= Date.now()) {
    const fallback = new Date()
    fallback.setMonth(fallback.getMonth() + 1)
    return fallback.getTime()
  }
  return base.getTime()
}

function isPastDate(ts: number): boolean {
  return ts < Date.now() - 86400_000
}

const columns = computed<DataTableColumns<OwnerRow>>(() => [
  { title: t('gatherings.invitations.owner'), key: 'ownerName' },
  {
    title: t('common.status'),
    key: 'status',
    render(row) {
      const types: Record<string, 'success' | 'warning' | 'error' | 'default'> = {
        active: 'success',
        expired: 'warning',
        revoked: 'error',
        none: 'default'
      }
      const labels: Record<string, string> = {
        active: t('gatherings.invitations.statusActive'),
        expired: t('gatherings.invitations.statusExpired'),
        revoked: t('gatherings.invitations.statusRevoked'),
        none: t('gatherings.invitations.statusNone')
      }
      return h(NTag, { type: types[row.status], size: 'small' }, () => labels[row.status])
    }
  },
  {
    title: t('gatherings.invitations.expires'),
    key: 'expires',
    render(row) {
      if (!row.invitation || row.status === 'revoked') return '—'
      return new Date(row.invitation.expires_at).toLocaleDateString()
    }
  },
  {
    title: t('common.actions'),
    key: 'actions',
    render(row) {
      const actions: ReturnType<typeof h>[] = []
      if (row.status !== 'active') {
        actions.push(
          h(NButton, {
            size: 'small',
            onClick: () => openGenerateSingle(row)
          }, () => t('gatherings.invitations.generateLink'))
        )
      }
      if (row.status === 'active') {
        actions.push(
          h(NButton, {
            size: 'small',
            type: 'error',
            onClick: () => openRevoke(row)
          }, () => t('gatherings.invitations.revoke'))
        )
      }
      return h(NSpace, { size: 'small' }, () => actions)
    }
  }
])

function openGenerateSingle(row: OwnerRow) {
  generateModalOwner.value = row
  generatedTokens.value = []
  expiryTimestamp.value = defaultExpiry()
  expiryError.value = null
  generateModalVisible.value = true
}

function openRevoke(row: OwnerRow) {
  confirm({
    title: t('gatherings.invitations.revokeTitle'),
    content: t('gatherings.invitations.revokeWarning', { name: row.ownerName }),
    positiveText: t('gatherings.invitations.revoke'),
    onConfirm: () => revokeInvitation(row)
  })
}

function handleGenerateAll() {
  generateModalOwner.value = null
  generatedTokens.value = []
  expiryTimestamp.value = defaultExpiry()
  expiryError.value = null
  generateModalVisible.value = true
}

async function confirmGenerate() {
  if (!expiryTimestamp.value) {
    expiryError.value = t('gatherings.invitations.selectExpiry')
    return
  }
  expiryError.value = null
  const expiresAt = new Date(expiryTimestamp.value).toISOString()

  if (generateModalOwner.value) {
    generating.value = true
    try {
      const res = await invitationApi.create(props.associationId, props.gathering.id, {
        owner_id: generateModalOwner.value.ownerId,
        expires_at: expiresAt
      })
      const url = `${MEMBER_BASE_URL}/${res.data.token}`
      generatedUrlsByOwner.value.set(generateModalOwner.value.ownerId, url)
      generatedTokens.value = [{
        ownerId: generateModalOwner.value.ownerId,
        ownerName: generateModalOwner.value.ownerName,
        url
      }]
    } catch (err: any) {
      error.value = err.response?.data?.error ?? err.message ?? t('gatherings.invitations.generateError')
      generateModalVisible.value = false
    } finally {
      generating.value = false
    }
  } else {
    bulkGenerating.value = true
    const results: typeof generatedTokens.value = []
    for (const row of ownersWithoutActive.value) {
      try {
        const res = await invitationApi.create(props.associationId, props.gathering.id, {
          owner_id: row.ownerId,
          expires_at: expiresAt
        })
        const url = `${MEMBER_BASE_URL}/${res.data.token}`
        generatedUrlsByOwner.value.set(row.ownerId, url)
        results.push({ ownerId: row.ownerId, ownerName: row.ownerName, url })
      } catch {
        // skip already-active or errored
      }
    }
    generatedTokens.value = results
    bulkGenerating.value = false
  }
}

async function copyUrl(url: string) {
  await navigator.clipboard.writeText(url)
  copiedUrl.value = url
  setTimeout(() => { copiedUrl.value = null }, 2000)
}

function closeGenerateModal() {
  generateModalVisible.value = false
  generatedTokens.value = []
  loadData()
}

async function revokeInvitation(target: OwnerRow) {
  if (!target.invitation) return
  try {
    await invitationApi.revoke(props.associationId, props.gathering.id, target.invitation.id)
    const row = rows.value.find(r => r.ownerId === target.ownerId)
    if (row) {
      row.status = 'revoked'
    }
  } catch (err: any) {
    error.value = err.response?.data?.error ?? err.message ?? t('gatherings.invitations.revokeError')
  }
}

function exportGeneratedCsv() {
  const lines = [[t('gatherings.invitations.owner'), t('gatherings.invitations.link')].join(',')]
  for (const item of generatedTokens.value) {
    const ownerName = `"${item.ownerName.replace(/"/g, '""')}"`
    lines.push(`${ownerName},${item.url}`)
  }
  const blob = new Blob([lines.join('\n')], { type: 'text/csv' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `invitations-gathering-${props.gathering.id}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

function exportCsv() {
  const lines = [[
    t('gatherings.invitations.owner'),
    t('gatherings.invitations.link'),
    t('gatherings.invitations.expires')
  ].join(',')]
  for (const row of activeRows.value) {
    if (!row.invitation) continue
    const ownerName = `"${row.ownerName.replace(/"/g, '""')}"`
    const url = generatedUrlsByOwner.value.get(row.ownerId) ?? '—'
    const expires = new Date(row.invitation.expires_at).toLocaleDateString()
    lines.push(`${ownerName},${url},${expires}`)
  }
  const blob = new Blob([lines.join('\n')], { type: 'text/csv' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `invitations-gathering-${props.gathering.id}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

async function loadData() {
  if (!props.associationId || !props.gathering.id) return
  loading.value = true
  error.value = null
  generatedUrlsByOwner.value = new Map()
  try {
    const [unitsRes, invitationsRes] = await Promise.all([
      gatheringApi.getQualifiedUnits(props.associationId, props.gathering.id),
      invitationApi.list(props.associationId, props.gathering.id)
    ])

    const units: QualifiedUnit[] = unitsRes.data
    const invitations: MemberInvitation[] = invitationsRes.data

    // De-duplicate owners from unit list
    const ownerMap = new Map<number, string>()
    for (const u of units) {
      ownerMap.set(u.owner_id, u.owner_name)
    }

    // Build invitation lookup by owner_id (latest non-revoked preferred)
    const invByOwner = new Map<number, MemberInvitation>()
    for (const inv of invitations) {
      const existing = invByOwner.get(inv.owner_id)
      if (!existing || inv.status === 'active') {
        invByOwner.set(inv.owner_id, inv)
      }
    }

    rows.value = Array.from(ownerMap.entries()).map(([ownerId, ownerName]) => {
      const inv = invByOwner.get(ownerId) ?? null
      return {
        ownerId,
        ownerName,
        invitation: inv,
        status: inv ? (inv.status as OwnerRow['status']) : 'none'
      }
    }).sort((a, b) => a.ownerName.localeCompare(b.ownerName))
  } catch (err: any) {
    error.value = err.response?.data?.error ?? err.message ?? t('gatherings.invitations.loadError')
  } finally {
    loading.value = false
  }
}

watch(() => [props.associationId, props.gathering.id], loadData, { immediate: true })
</script>
