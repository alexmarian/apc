<script setup lang="ts">
import { NCard, NEmpty } from 'naive-ui'
import PageShell from '@/components/PageShell.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <PageShell :title="t('home.title', 'Home Page')">
    <NCard>
      <NEmpty :description="t('home.title', 'Home Page')" />
    </NCard>
  </PageShell>
</template>
