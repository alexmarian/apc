<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

// Props
const props = defineProps<{
  type: string
  family: string
  name: string
  showType?: boolean
  showFamily?: boolean
  showName?: boolean
  delimiter?: string
}>()

// Default values
const showType = props.showType !== undefined ? props.showType : true
const showFamily = props.showFamily !== undefined ? props.showFamily : true
const showName = props.showName !== undefined ? props.showName : true
const delimiter = props.delimiter || ' - '

// I18n
const { t } = useI18n()

// Compute the full localized category string
const localizedCategory = computed(() => {
  const parts = []

  if (showType && props.type) {
    parts.push(t(`categories.types.${props.type}`))
  }

  if (showFamily && props.family) {
    parts.push(t(`categories.families.${props.family}`))
  }

  if (showName && props.name) {
    parts.push(t(`categories.names.${props.name}`))
  }

  return parts.join(delimiter)
})
</script>

<template>
  <span class="localized-category">{{ localizedCategory }}</span>
</template>
