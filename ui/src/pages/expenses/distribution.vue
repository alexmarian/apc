<script setup lang="ts">
import { ref } from 'vue'
import { NPageHeader } from 'naive-ui'
import UnitExpenseDistributionReport from '@/components/UnitExpenseDistributionReport.vue'

// State
const associationId = ref<number | null>(null)
const buildingId = ref<number | null>(null)

// Methods
const handleUpdateAssociationId = (id: number) => {
  associationId.value = id
}

const handleUpdateBuildingId = (id: number) => {
  buildingId.value = id
}
</script>

<template>
  <div class="expense-distribution-page">
    <NPageHeader>
      <template #title>
        Unit Expense Distribution Report
      </template>
    </NPageHeader>

    <div class="content">
      <UnitExpenseDistributionReport
        v-model:associationId="associationId"
        v-model:buildingId="buildingId"
        @update:associationId="handleUpdateAssociationId"
        @update:buildingId="handleUpdateBuildingId"
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
