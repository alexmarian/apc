<template>
  <NSpin :show="loading">
    <NAlert v-if="error" type="error" closable style="margin-bottom: 16px" @close="error = null">
      {{ error }}
    </NAlert>

    <NSpace justify="space-between" align="center" style="margin-bottom: 16px">
      <NText strong>{{ filteredRows.length }} qualified owner{{ filteredRows.length !== 1 ? 's' : '' }}</NText>
      <NSpace>
        <NButton @click="handleGenerateAll" :loading="bulkGenerating" :disabled="ownersWithoutActive.length === 0">
          Generate for visible ({{ ownersWithoutActive.length }})
        </NButton>
        <NButton @click="exportCsv" :disabled="activeRows.length === 0">
          Export CSV
        </NButton>
      </NSpace>
    </NSpace>

    <NInput
      v-model:value="searchQuery"
      placeholder="Search owners..."
      clearable
      style="margin-bottom: 12px"
    />

    <NDataTable :columns="columns" :data="filteredRows" :bordered="false" size="small" />
  </NSpin>

  <!-- Generate link modal -->
  <NModal v-model:show="generateModalVisible" :mask-closable="false" style="width: 480px">
    <NCard :title="generateModalOwner ? `Generate link for ${generateModalOwner.ownerName}` : 'Generate all links'">
      <template v-if="!generatedTokens.length">
        <NFormItem label="Expiry date" :feedback="expiryError || undefined" :validation-status="expiryError ? 'error' : undefined">
          <NDatePicker
            v-model:value="expiryTimestamp"
            type="date"
            style="width: 100%"
            :is-date-disabled="isPastDate"
          />
        </NFormItem>
        <NSpace justify="end" style="margin-top: 8px">
          <NButton @click="generateModalVisible = false">Cancel</NButton>
          <NButton type="primary" :loading="generating" @click="confirmGenerate">
            Generate
          </NButton>
        </NSpace>
      </template>

      <template v-else>
        <NAlert type="warning" style="margin-bottom: 16px">
          These links cannot be recovered after closing this dialog. Copy them now.
        </NAlert>

        <div v-for="item in generatedTokens" :key="item.ownerId" style="margin-bottom: 12px">
          <NText strong style="display: block; margin-bottom: 4px">{{ item.ownerName }}</NText>
          <NInputGroup>
            <NInput :value="item.url" readonly style="font-size: 12px; font-family: monospace" />
            <NButton @click="copyUrl(item.url)" style="min-width: 70px">
              {{ copiedUrl === item.url ? '✓ Copied' : 'Copy' }}
            </NButton>
          </NInputGroup>
        </div>

        <NSpace justify="end" style="margin-top: 16px">
          <NButton v-if="generatedTokens.length > 1" @click="exportGeneratedCsv">Export CSV</NButton>
          <NButton type="primary" @click="closeGenerateModal">Done</NButton>
        </NSpace>
      </template>
    </NCard>
  </NModal>

  <!-- Revoke confirm modal -->
  <NModal v-model:show="revokeModalVisible" style="width: 400px">
    <NCard title="Revoke invitation">
      <NText>
        Revoking this link for <strong>{{ revokeTarget?.ownerName }}</strong> will immediately
        invalidate their access. You can generate a new link afterward.
      </NText>
      <NSpace justify="end" style="margin-top: 16px">
        <NButton @click="revokeModalVisible = false">Cancel</NButton>
        <NButton type="error" :loading="revoking" @click="confirmRevoke">Revoke</NButton>
      </NSpace>
    </NCard>
  </NModal>
</template>

<script setup lang="ts">
import { ref, computed, watch, h } from 'vue'
import {
  NAlert,
  NButton,
  NCard,
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

const revokeModalVisible = ref(false)
const revokeTarget = ref<OwnerRow | null>(null)
const revoking = ref(false)

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

const columns: DataTableColumns<OwnerRow> = [
  { title: 'Owner', key: 'ownerName' },
  {
    title: 'Status',
    key: 'status',
    render(row) {
      const types: Record<string, 'success' | 'warning' | 'error' | 'default'> = {
        active: 'success',
        expired: 'warning',
        revoked: 'error',
        none: 'default'
      }
      const labels: Record<string, string> = {
        active: 'Active',
        expired: 'Expired',
        revoked: 'Revoked',
        none: 'Not generated'
      }
      return h(NTag, { type: types[row.status], size: 'small' }, () => labels[row.status])
    }
  },
  {
    title: 'Expires',
    key: 'expires',
    render(row) {
      if (!row.invitation || row.status === 'revoked') return '—'
      return new Date(row.invitation.expires_at).toLocaleDateString()
    }
  },
  {
    title: 'Actions',
    key: 'actions',
    render(row) {
      const actions: ReturnType<typeof h>[] = []
      if (row.status !== 'active') {
        actions.push(
          h(NButton, {
            size: 'small',
            onClick: () => openGenerateSingle(row)
          }, () => 'Generate link')
        )
      }
      if (row.status === 'active') {
        actions.push(
          h(NButton, {
            size: 'small',
            type: 'error',
            onClick: () => openRevoke(row)
          }, () => 'Revoke')
        )
      }
      return h(NSpace, { size: 'small' }, () => actions)
    }
  }
]

function openGenerateSingle(row: OwnerRow) {
  generateModalOwner.value = row
  generatedTokens.value = []
  expiryTimestamp.value = defaultExpiry()
  expiryError.value = null
  generateModalVisible.value = true
}

function openRevoke(row: OwnerRow) {
  revokeTarget.value = row
  revokeModalVisible.value = true
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
    expiryError.value = 'Select an expiry date'
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
      error.value = err.response?.data?.error ?? err.message ?? 'Failed to generate invitation'
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

async function confirmRevoke() {
  if (!revokeTarget.value?.invitation) return
  revoking.value = true
  try {
    await invitationApi.revoke(props.associationId, props.gathering.id, revokeTarget.value.invitation.id)
    const row = rows.value.find(r => r.ownerId === revokeTarget.value!.ownerId)
    if (row) {
      row.status = 'revoked'
    }
    revokeModalVisible.value = false
  } catch (err: any) {
    error.value = err.response?.data?.error ?? err.message ?? 'Failed to revoke invitation'
  } finally {
    revoking.value = false
  }
}

function exportGeneratedCsv() {
  const lines = ['Owner,Link']
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
  const lines = ['Owner,Link,Expires']
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
    error.value = err.response?.data?.error ?? err.message ?? 'Failed to load invitations'
  } finally {
    loading.value = false
  }
}

watch(() => [props.associationId, props.gathering.id], loadData, { immediate: true })
</script>
