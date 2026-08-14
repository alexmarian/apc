<script setup lang="ts">
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import PageShell from '@/components/PageShell.vue'
import UnitExpenseDistributionReport from '@/components/UnitExpenseDistributionReport.vue'
import BuildingSelector from '@/components/BuildingSelector.vue'
import { useAssociationStore } from '@/stores/association'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const { associationId } = storeToRefs(useAssociationStore())
const buildingId = ref<number | null>(null)
</script>

<template>
  <PageShell :title="t('expenses.distribution', 'Unit Expense Distribution Report')">
    <template #context>
      <BuildingSelector v-model:building-id="buildingId" v-model:association-id="associationId" />
    </template>

    <UnitExpenseDistributionReport
      v-if="associationId"
      :association-id="associationId"
      :building-id="buildingId"
    />
  </PageShell>
</template>
