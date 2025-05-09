<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <div>
    <h1>{{ t('common.notFound', '404 Not Found') }}</h1>
  </div>
</template>
