<script setup lang="ts">
import { NPageHeader } from 'naive-ui'

/**
 * The frame every page renders into, so titles, back navigation, primary
 * actions and content spacing land in the same place across the app.
 *
 * Slot contract:
 * - `actions` — page-level actions (create, export). Never row-level or
 *   selection-level actions; those belong next to the table.
 * - `context` — scope selectors that qualify the whole page (building, period).
 * - default — the page body.
 */
const props = defineProps<{
  title: string
  subtitle?: string
  /** Shows the back arrow. Pair with @back. */
  back?: boolean
}>()

const emit = defineEmits<{
  back: []
}>()
</script>

<template>
  <div class="page-shell">
    <NPageHeader :on-back="props.back ? () => emit('back') : undefined">
      <template #title>{{ props.title }}</template>

      <template v-if="props.subtitle" #subtitle>{{ props.subtitle }}</template>

      <template v-if="$slots.context" #header>
        <div class="page-shell__context">
          <slot name="context" />
        </div>
      </template>

      <template v-if="$slots.actions" #extra>
        <slot name="actions" />
      </template>
    </NPageHeader>

    <div class="page-shell__content">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.page-shell {
  width: 100%;
}

.page-shell__context {
  margin-bottom: 12px;
}

/* The single source of the gap between a page header and its body. Page
 * bodies must not add their own top margin. */
.page-shell__content {
  margin-top: 16px;
}
</style>
