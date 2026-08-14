<script setup lang="ts">
import { NText } from 'naive-ui'

/**
 * A labelled control inside a FilterBar. Filters carry visible labels rather
 * than relying on placeholders, which disappear once a value is selected.
 */
const props = defineProps<{
  label: string
  /** Flex grow factor, for filters that deserve more room than their siblings. */
  grow?: number
}>()
</script>

<template>
  <div class="filter-field" :style="{ flexGrow: props.grow ?? 0 }">
    <NText class="filter-field__label">{{ props.label }}:</NText>
    <div class="filter-field__control">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.filter-field {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.filter-field__label {
  white-space: nowrap;
}

.filter-field__control {
  flex: 1;
  min-width: 0;
}
</style>
