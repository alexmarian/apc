<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { NSelect, NSpin } from 'naive-ui'
import axios from 'axios'
import config from '@/config'
import type { Association } from '@/types/api'

// Props
const props = defineProps<{
  modelValue: number | null
}>()

// Emits
const emit = defineEmits<{
  (e: 'update:modelValue', id: number): void
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

    const response = await axios.get<Association[]>(`${config.apiBaseUrl}/associations`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })

    associations.value = response.data

    // If we have associations but no selection, select the first one
    if (associations.value.length > 0 && !props.modelValue) {
      emit('update:modelValue', associations.value[0].id)
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
  emit('update:modelValue', value)
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
        :value="props.modelValue"
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
