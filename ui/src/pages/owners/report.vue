<script setup lang="ts">
import { ref } from 'vue'
import { NPageHeader, NCard } from 'naive-ui'
import OwnersReport from '@/components/OwnersReport.vue'
import AssociationSelector from '@/components/AssociationSelector.vue'
import { useI18n } from 'vue-i18n'

// State
const associationId = ref<number | null>(null)
const { t } = useI18n()
</script>

<template>
  <div class="owners-report-page">
    <NPageHeader>
      <template #title>
        {{ t('owners.report', 'Owners Report') }}
      </template>

      <template #header>
        <div style="margin-bottom: 12px;">
          <AssociationSelector v-model:associationId="associationId" />
        </div>
      </template>
    </NPageHeader>

    <div class="content">
      <div v-if="!associationId">
        <NCard style="margin-top: 16px;">
          <div style="text-align: center; padding: 32px;">
            <p>{{ t('owners.selectAssociation', 'Please select an association to view the owners report') }}</p>
          </div>
        </NCard>
      </div>

      <div v-else>
        <OwnersReport :association-id="associationId" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.owners-report-page {
  width: 100%;
}

.content {
  margin-top: 16px;
}
</style>
