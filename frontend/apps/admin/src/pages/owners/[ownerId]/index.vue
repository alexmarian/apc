<script setup lang="ts">
import { ref, computed, onMounted, h, watch } from 'vue'
import {
  NPageHeader,
  NCard,
  NDescriptions,
  NDescriptionsItem,
  NSpin,
  NAlert,
  NButton,
  NSpace,
  NDataTable,
  NTag,
  NStatistic,
  NGrid,
  NGridItem,
  NDivider,
  NBreadcrumb,
  NBreadcrumbItem
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useRoute, useRouter } from 'vue-router'
import { ownerApi } from '@/services/api'
import { formatPercentage } from '@/utils/formatters'
import type { OwnerReportItem, OwnerUnit, OwnerCoOwner } from '@/types/api'
import { useAssociationStore } from '@/stores/association'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { arrayToCsv, downloadCsv } from '@/utils/csvUtils'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { associationId } = storeToRefs(useAssociationStore())

const ownerId = computed(() => parseInt(route.params.ownerId as string))

const loading = ref(true)
const error = ref<string | null>(null)
const ownerData = ref<OwnerReportItem | null>(null)
const selectedUnitTypes = ref<string[]>([])

// Breadcrumb chain: array of {id, name} read from query param, plus the current owner appended once loaded
interface BreadcrumbEntry { id: number; name: string }
const breadcrumbChain = computed<BreadcrumbEntry[]>(() => {
  const raw = route.query.chain
  if (!raw) return []
  try {
    return JSON.parse(raw as string) as BreadcrumbEntry[]
  } catch {
    return []
  }
})

const fetchOwnerDetail = async () => {
  if (!associationId.value) {
    error.value = t('common.selectAssociation', 'Please select an association first')
    loading.value = false
    return
  }

  try {
    loading.value = true
    error.value = null
    const response = await ownerApi.getOwnerReport(associationId.value, true, true, ownerId.value)
    const items: OwnerReportItem[] = response.data
    ownerData.value = items.length > 0 ? items[0] : null
    if (!ownerData.value) {
      error.value = t('owners.detail.notFound', 'Owner not found')
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('owners.loadError', 'Failed to load owner data')
  } finally {
    loading.value = false
  }
}

// Refetch when navigating between co-owner pages (ownerId param changes)
watch(ownerId, () => {
  ownerData.value = null
  selectedUnitTypes.value = []
  fetchOwnerDetail()
})

const unitTypeSummary = computed(() => {
  const units = ownerData.value?.units
  if (!units) return []
  const byType = new Map<string, { unit_type: string; count: number; area: number; part: number }>()
  for (const unit of units) {
    const entry = byType.get(unit.unit_type) ?? { unit_type: unit.unit_type, count: 0, area: 0, part: 0 }
    entry.count += 1
    entry.area += unit.area
    entry.part += unit.part
    byType.set(unit.unit_type, entry)
  }
  return Array.from(byType.values())
})

const filteredUnits = computed<OwnerUnit[]>(() => {
  const units = ownerData.value?.units ?? []
  if (selectedUnitTypes.value.length === 0) return units
  return units.filter(u => selectedUnitTypes.value.includes(u.unit_type))
})

const handleSelectUnitType = (unitType: string) => {
  const idx = selectedUnitTypes.value.indexOf(unitType)
  if (idx === -1) {
    selectedUnitTypes.value = [...selectedUnitTypes.value, unitType]
  } else {
    selectedUnitTypes.value = selectedUnitTypes.value.filter(t => t !== unitType)
  }
}

const unitsColumns = computed<DataTableColumns<OwnerUnit>>(() => [
  {
    title: t('units.building', 'Building'),
    key: 'building_name',
    sorter: (a, b) => a.building_name.localeCompare(b.building_name)
  },
  { title: t('units.cadastralNumber', 'Cadastral Number'), key: 'unit_cadastral_number' },
  { title: t('units.address', 'Address'), key: 'unit_address' },
  {
    title: t('units.unit', 'Unit'),
    key: 'unit_number',
    sorter: (a, b) => a.unit_number.localeCompare(b.unit_number, undefined, { numeric: true })
  },
  {
    title: t('units.area', 'Area'),
    key: 'area',
    sorter: (a, b) => a.area - b.area,
    render: (row) => `${row.area.toFixed(2)} m²`
  },
  {
    title: t('units.part', 'Part'),
    key: 'part',
    sorter: (a, b) => a.part - b.part,
    render: (row) => formatPercentage(row.part, 4)
  },
  {
    title: t('units.type', 'Type'),
    key: 'unit_type',
    render: (row) => t(`unitTypes.${row.unit_type}`, row.unit_type)
  },
  {
    title: t('common.actions', 'Actions'),
    key: 'actions',
    render: (row) => h(
      NButton,
      {
        size: 'small',
        onClick: () => router.push({
          path: `/units/${row.unit_id}`,
          query: {
            associationId: associationId.value?.toString(),
            buildingId: row.building_id.toString(),
            from: route.fullPath
          }
        })
      },
      { default: () => t('common.view', 'View') }
    )
  }
])

const downloadFilteredUnits = () => {
  const units = filteredUnits.value
  if (!units.length) return
  const headers = [
    t('units.building', 'Building'),
    t('units.cadastralNumber', 'Cadastral Number'),
    t('units.address', 'Address'),
    t('units.unit', 'Unit'),
    t('units.area', 'Area (m²)'),
    t('units.part', 'Part (%)'),
    t('units.type', 'Type')
  ]
  const rows = units.map(u => [
    u.building_name,
    u.unit_cadastral_number,
    u.unit_address,
    u.unit_number,
    u.area.toFixed(2),
    (u.part * 100).toFixed(4),
    t(`unitTypes.${u.unit_type}`, u.unit_type)
  ])
  const ownerName = ownerData.value?.owner.name ?? 'owner'
  const filename = `${ownerName.replace(/\s+/g, '_')}_units.csv`
  downloadCsv(arrayToCsv([headers, ...rows]), filename)
}

const navigateToCoOwner = (coOwner: OwnerCoOwner) => {
  // Build the new breadcrumb chain: existing chain + current owner
  const currentEntry: BreadcrumbEntry = { id: ownerId.value, name: ownerData.value?.owner.name ?? String(ownerId.value) }
  const newChain = [...breadcrumbChain.value, currentEntry].slice(-3) // cap at 3 ancestors
  router.push({
    path: `/owners/${coOwner.id}`,
    query: { chain: JSON.stringify(newChain) }
  })
}

const coOwnersColumns = computed<DataTableColumns<OwnerCoOwner>>(() => [
  { title: t('owners.name', 'Name'), key: 'name' },
  { title: t('owners.identification', 'Identification'), key: 'identification_number' },
  { title: t('owners.contactPhone', 'Contact Phone'), key: 'contact_phone' },
  { title: t('owners.contactEmail', 'Contact Email'), key: 'contact_email' },
  {
    title: t('owners.sharedUnits', 'Shared Units'),
    key: 'shared_unit_nums',
    render: (row: OwnerCoOwner) => row.shared_unit_nums.join(', ')
  },
  {
    title: t('common.actions', 'Actions'),
    key: 'actions',
    render: (row: OwnerCoOwner) => h(
      NButton,
      { size: 'small', onClick: () => navigateToCoOwner(row) },
      { default: () => t('common.details', 'Details') }
    )
  }
])

const handleBack = () => {
  if (breadcrumbChain.value.length > 0) {
    const prev = breadcrumbChain.value[breadcrumbChain.value.length - 1]
    const parentChain = breadcrumbChain.value.slice(0, -1)
    router.push({
      path: `/owners/${prev.id}`,
      query: parentChain.length > 0 ? { chain: JSON.stringify(parentChain) } : {}
    })
  } else {
    router.push('/owners/report')
  }
}

const navigateToBreadcrumb = (entry: BreadcrumbEntry, index: number) => {
  const parentChain = breadcrumbChain.value.slice(0, index)
  router.push({
    path: `/owners/${entry.id}`,
    query: parentChain.length > 0 ? { chain: JSON.stringify(parentChain) } : {}
  })
}

onMounted(fetchOwnerDetail)
</script>

<template>
  <div class="owner-detail-page">
    <NPageHeader @back="handleBack">
      <template #title>
        {{ ownerData?.owner.name ?? t('owners.detail.title', 'Owner Detail') }}
      </template>
      <template #subtitle>
        {{ ownerData?.owner.identification_number }}
      </template>
      <template #header>
        <NBreadcrumb>
          <NBreadcrumbItem @click="router.push('/owners/report')" style="cursor:pointer">
            {{ t('owners.report', 'Owners Report') }}
          </NBreadcrumbItem>
          <NBreadcrumbItem
            v-for="(entry, idx) in breadcrumbChain"
            :key="entry.id"
            @click="navigateToBreadcrumb(entry, idx)"
            style="cursor:pointer"
          >
            {{ entry.name }}
          </NBreadcrumbItem>
          <NBreadcrumbItem>
            {{ ownerData?.owner.name ?? t('owners.detail.title', 'Owner Detail') }}
          </NBreadcrumbItem>
        </NBreadcrumb>
      </template>
    </NPageHeader>

    <NSpin :show="loading">
      <NAlert v-if="error" type="error" style="margin-top: 16px;">
        {{ error }}
        <NButton style="margin-left: 12px;" @click="fetchOwnerDetail">{{ t('common.retry', 'Retry') }}</NButton>
      </NAlert>

      <div v-else-if="ownerData" class="detail-content">
        <!-- Owner Info -->
        <NCard style="margin-top: 16px;">
          <NDescriptions bordered :column="2">
            <NDescriptionsItem :label="t('owners.name', 'Name')">
              {{ ownerData.owner.name }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('owners.identification', 'Identification')">
              {{ ownerData.owner.identification_number }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('owners.contactPhone', 'Contact Phone')">
              {{ ownerData.owner.contact_phone || '—' }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('owners.contactEmail', 'Contact Email')">
              {{ ownerData.owner.contact_email || '—' }}
            </NDescriptionsItem>
          </NDescriptions>
        </NCard>

        <!-- Summary Stats -->
        <NCard style="margin-top: 16px;">
          <NGrid :cols="3" :x-gap="24">
            <NGridItem>
              <NStatistic
                :label="t('owners.totalArea', 'Total Area')"
                :value="ownerData.statistics.total_area.toFixed(2)"
                suffix="m²"
              />
            </NGridItem>
            <NGridItem>
              <NStatistic
                :label="t('owners.totalPart', 'Condo Part')"
                :value="formatPercentage(ownerData.statistics.total_condo_part, 4)"
              />
            </NGridItem>
            <NGridItem>
              <NStatistic
                :label="t('units.title', 'Units')"
                :value="ownerData.statistics.total_units"
              />
            </NGridItem>
          </NGrid>
        </NCard>

        <!-- Holdings by type -->
        <NCard v-if="unitTypeSummary.length > 0" style="margin-top: 16px;">
          <h3 class="section-title">{{ t('owners.holdingsByType', 'Holdings by unit type') }}</h3>
          <div class="type-cards-row">
            <div
              v-for="entry in unitTypeSummary"
              :key="entry.unit_type"
              class="type-card"
              :class="{ 'type-card--active': selectedUnitTypes.includes(entry.unit_type) }"
              @click="handleSelectUnitType(entry.unit_type)"
            >
              <div class="type-card__label">{{ t(`unitTypes.${entry.unit_type}`, entry.unit_type) }}</div>
              <div class="type-card__row">
                <span class="type-card__key">{{ t('owners.unitCount', 'Number') }}</span>
                <span class="type-card__val">{{ entry.count }}</span>
              </div>
              <div class="type-card__row">
                <span class="type-card__key">{{ t('units.area', 'Area') }}</span>
                <span class="type-card__val">{{ entry.area.toFixed(2) }} m²</span>
              </div>
              <div class="type-card__row">
                <span class="type-card__key">{{ t('units.part', 'Part') }}</span>
                <span class="type-card__val">{{ formatPercentage(entry.part, 4) }}</span>
              </div>
            </div>
          </div>
        </NCard>

        <!-- Units table -->
        <NCard style="margin-top: 16px;">
          <div class="section-header">
            <h3 class="section-title">{{ t('owners.unitsDetails', "Owner's Units") }}</h3>
            <NTag
              v-for="type in selectedUnitTypes"
              :key="type"
              type="info"
              closable
              @close="handleSelectUnitType(type)"
            >
              {{ t(`unitTypes.${type}`, type) }}
            </NTag>
            <NButton size="small" @click="downloadFilteredUnits" style="margin-left: auto;">
              {{ t('owners.exportToCsv', 'Export to CSV') }}
            </NButton>
          </div>
          <NDataTable
            :columns="unitsColumns"
            :data="filteredUnits"
            :pagination="{ pageSize: 10 }"
            :row-key="(row: OwnerUnit) => row.unit_id"
            :bordered="false"
          />
        </NCard>

        <!-- Co-owners table -->
        <NCard v-if="ownerData.co_owners && ownerData.co_owners.length > 0" style="margin-top: 16px;">
          <h3 class="section-title">{{ t('owners.coOwners', 'Co-Owners') }}</h3>
          <NDataTable
            :columns="coOwnersColumns"
            :data="ownerData.co_owners"
            :pagination="{ pageSize: 10 }"
            :row-key="(row: OwnerCoOwner) => row.id"
            :bordered="false"
          />
        </NCard>
        <NDivider v-else />
      </div>
    </NSpin>
  </div>
</template>

<style scoped>
.owner-detail-page {
  width: 100%;
}

.detail-content {
  padding-bottom: 32px;
}

.section-title {
  font-size: 1.05rem;
  font-weight: 600;
  margin: 0 0 12px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.type-cards-row {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  overflow-x: auto;
  gap: 10px;
  padding-bottom: 4px;
}

.type-card {
  flex: 0 0 auto;
  min-width: 130px;
  padding: 10px 14px;
  border-radius: 6px;
  border: 1px solid var(--border-color, #e0e0e6);
  cursor: pointer;
  transition: border-color 0.15s, background-color 0.15s;
}

.type-card:hover {
  border-color: #2080f0;
}

.type-card--active {
  border-color: #2080f0;
  background-color: rgba(32, 128, 240, 0.08);
}

.type-card__label {
  font-weight: 600;
  font-size: 0.85rem;
  margin-bottom: 8px;
  text-transform: capitalize;
}

.type-card__row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 8px;
  font-size: 0.82rem;
  line-height: 1.7;
}

.type-card__key {
  opacity: 0.6;
  white-space: nowrap;
}

.type-card__val {
  font-weight: 500;
  white-space: nowrap;
}

</style>
