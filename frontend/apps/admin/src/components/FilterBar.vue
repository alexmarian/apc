<script setup lang="ts">
import { NCard, NFlex, NButton } from 'naive-ui'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  /** Renders the trailing reset button. Pair with @reset. */
  resettable?: boolean
}>()

const emit = defineEmits<{
  reset: []
}>()
</script>

<template>
  <NCard size="small" class="filter-bar">
    <NFlex align="center" :size="12" wrap>
      <slot />
      <NButton v-if="props.resettable" @click="emit('reset')">
        {{ t('common.reset_filters') }}
      </NButton>
    </NFlex>
  </NCard>
</template>

<style scoped>
/* Spacing between the bar and the table below it is owned here, so callers
 * stop hand-rolling margin-top on their table cards. */
.filter-bar {
  margin-bottom: 16px;
}
</style>
