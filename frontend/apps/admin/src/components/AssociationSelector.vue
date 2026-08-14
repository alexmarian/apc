<script setup lang="ts">
import { computed } from 'vue'
import { NSelect } from 'naive-ui'
import { storeToRefs } from 'pinia'
import { useAssociationStore } from '@/stores/association'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const store = useAssociationStore()
const { associationId, associations, loading } = storeToRefs(store)

const options = computed(() =>
  associations.value.map((assoc) => ({
    label: assoc.name,
    value: assoc.id,
  })),
)
</script>

<template>
  <NSelect
    :value="associationId"
    :options="options"
    :loading="loading"
    :placeholder="t('associations.select')"
    :disabled="loading || associations.length === 0"
    :consistent-menu-width="false"
    @update:value="store.select"
  />
</template>
