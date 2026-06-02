<script setup lang="ts">
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import { NCard, NPageHeader, NSpin } from 'naive-ui'
import UnitExpenseDistributionReport from '@/components/UnitExpenseDistributionReport.vue'
import BuildingSelector from '@/components/BuildingSelector.vue'
import { useAssociationStore } from '@/stores/association'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const { associationId } = storeToRefs(useAssociationStore())
const buildingId = ref<number | null>(null)
</script>

<template>
  <div class="expense-distribution-page">
    <NPageHeader>
      <template #title>
        {{ t('expenses.distribution', 'Unit Expense Distribution Report') }}
      </template>

      <template #header>
        <div style="display: flex; gap: 16px; margin-bottom: 12px;">
          <BuildingSelector
            v-model:building-id="buildingId"
            v-model:association-id="associationId"
          />
        </div>
      </template>
    </NPageHeader>

    <div class="content">
      <UnitExpenseDistributionReport
        v-if="associationId"
        :association-id="associationId"
        :building-id="buildingId"
      />
    </div>
  </div>
</template>

<style scoped>
.expense-distribution-page {
  width: 100%;
}

.content {
  margin-top: 16px;
}
</style>
