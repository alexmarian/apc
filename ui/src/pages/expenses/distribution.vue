<script setup lang="ts">
import { ref } from 'vue'
import { NCard, NPageHeader } from 'naive-ui'
import UnitExpenseDistributionReport from '@/components/UnitExpenseDistributionReport.vue'
import AssociationSelector from '@/components/AssociationSelector.vue'
import BuildingSelector from '@/components/BuildingSelector.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const associationId = ref<number | null>(null)
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
          <AssociationSelector
            v-model:associationId="associationId"
          />
          <BuildingSelector
            v-if="associationId"
            v-model:building-id="buildingId"
            v-model:association-id="associationId"
          />
        </div>
      </template>
    </NPageHeader>

    <div v-if="!associationId">
      <NCard style="margin-top: 16px;">
        <div style="text-align: center; padding: 32px;">
          <p>{{ t('expenses.selectAssociation', 'Please select an association to manage expenses') }}</p>
        </div>
      </NCard>
    </div>

    <div v-else class="content">
      <UnitExpenseDistributionReport
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
