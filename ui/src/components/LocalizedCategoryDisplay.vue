<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  type: string
  family: string
  name: string
  showType?: boolean
  showFamily?: boolean
  showName?: boolean
  delimiter?: string
}>()

const { t } = useI18n()

const showType = props.showType !== undefined ? props.showType : true
const showFamily = props.showFamily !== undefined ? props.showFamily : true
const showName = props.showName !== undefined ? props.showName : true
const delimiter = props.delimiter || ' - '

const displayCategory = computed(() => {
  const parts = []
  if (showType && props.type) parts.push(t(`categories.types.${props.type}`, props.type))
  if (showFamily && props.family) parts.push(t(`categories.families.${props.family}`, props.family))
  if (showName && props.name) parts.push(t(`categories.names.${props.name}`, props.name))
  return parts.join(delimiter)
})
</script>

<template>
  <span class="localized-category">{{ displayCategory }}</span>
</template>
