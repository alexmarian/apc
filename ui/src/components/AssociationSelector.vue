<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { NSelect, NSpin } from 'naive-ui'
import config from '@/config'
import type { Association } from '@/types/api'
import { associationApi } from '@/services/api.ts'

// Props
const props = defineProps<{
  associationId: number | null
}>()

// Emits
const emit = defineEmits<{
  (e: 'update:associationId', id: number): void
}>()

// State
const associations = ref<Association[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

// Fetch associations
const fetchAssociations = async () => {
  try {
    loading.value = true
    error.value = null

    const token = localStorage.getItem(config.authTokenKey)

    const response = await associationApi.getAssociations();

    associations.value = response.data

    // If we have associations but no selection, select the first one
    if (associations.value.length > 0 && !props.associationId) {
      emit('update:associationId', associations.value[0].id)
    }
  } catch (err) {
    console.error('Error fetching associations:', err)
    error.value = 'Failed to load associations'
  } finally {
    loading.value = false
  }
}

// Format association options for NSelect
const options = computed(() => {
  return associations.value.map(assoc => ({
    label: assoc.name,
    value: assoc.id
  }))
})

// Handle selection change
const handleChange = (value: number) => {
  emit('update:associationId', value)
}

// Load associations on mount
onMounted(() => {
  fetchAssociations()
})
</script>

<template>
  <div class="association-selector">
    <NSpin :show="loading">
      <NSelect
        :value="props.associationId"
        :options="options"
        placeholder="Select an association"
        @update:value="handleChange"
        :disabled="loading || associations.length === 0"
      />
    </NSpin>
    <p v-if="error" class="error">{{ error }}</p>
  </div>
</template>

<style scoped>
.association-selector {
  width: 100%;
  max-width: 300px;
}
</style>
