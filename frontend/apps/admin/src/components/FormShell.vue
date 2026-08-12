<!-- src/components/FormShell.vue -->
<template>
  <div class="form-shell">
    <slot></slot>
  </div>
</template>

<style scoped>
/*
 * Padding, background and radius are deliberately absent: every call site
 * renders these forms inside an NCard (directly, or via NModal preset="card"),
 * which already owns that chrome. Only the width constraint is ours.
 */
.form-shell {
  max-width: 600px;
  margin: 0 auto;
}

/*
 * naive-ui styles `disabled` inputs but not `readonly` ones, so readonly
 * fields are indistinguishable from editable ones without this. :deep()
 * resolves against .form-shell, which contains all slotted content.
 * !important is carried over from the per-form copies this replaces — it
 * overrides naive's own .n-input rules for the inner element.
 */
.form-shell :deep(.readonly-input) {
  background-color: var(--input-color-disabled);
  color: var(--text-color-disabled);
}

.form-shell :deep(.readonly-input .n-input__input-el) {
  background-color: var(--input-color-disabled) !important;
  color: var(--text-color-disabled) !important;
  cursor: not-allowed;
}
</style>
