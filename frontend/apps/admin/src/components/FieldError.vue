<!-- src/components/FieldError.vue -->
<script setup lang="ts">
import { NText } from 'naive-ui'
</script>

<template>
  <!-- type="error" takes its colour from the active naive theme. -->
  <NText type="error" tag="p" class="field-error">
    <slot></slot>
  </NText>
</template>

<style scoped>
.field-error {
  margin-top: 4px;
}
</style>
