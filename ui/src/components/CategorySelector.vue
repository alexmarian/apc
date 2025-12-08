<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { NSelect, NSpin } from 'naive-ui'
import { categoryApi } from '@/services/api'
import type { Category } from '@/types/api'
import { useI18n } from 'vue-i18n'


const props = defineProps<{
  modelValue: number | null
  associationId: number
  placeholder?: string
  includeAllOption?: boolean
  disabled?: boolean
}>()
const { t } = useI18n()

const emit = defineEmits<{
  (e: 'update:modelValue', id: number | null): void
}>()


const categories = ref<Category[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

const options = computed(() => {
  const categoryOptions = categories.value
  .filter(category => !category.is_deleted)
  .map(category => ({
    label: `${category.type} → ${category.family} → ${category.name}`,
    value: category.id
  }))


  if (props.includeAllOption) {
    categoryOptions.unshift({
      label: t('categories.allCategories','All Categories'),
      value: null as any // Need to cast this because TypeScript doesn't like null values
    })
  }

  return categoryOptions
})

// Fetch categories
const fetchCategories = async () => {
  if (!props.associationId) return

  try {
    loading.value = true
    error.value = null

    const response = await categoryApi.getCategories(props.associationId)
    categories.value = response.data

    // If no category is selected and we have categories, select the first one
    if (!props.includeAllOption && !props.modelValue && categories.value.length > 0 && !props.disabled) {
      emit('update:modelValue', categories.value[0].id)
    }
  } catch (err) {
    console.error('Error fetching categories:', err)
    error.value = 'Failed to load categories'
  } finally {
    loading.value = false
  }
}

// Handle selection change
const handleChange = (value: number | null) => {
  emit('update:modelValue', value)
}
watch(
  () => props.associationId,
  (newValue) => {
    fetchCategories()
  },
  { immediate: true }
)
</script>

<template>
  <div class="category-selector">
    <NSpin :show="loading">
      <NSelect
        filterable
        :value="props.modelValue"
        :options="options"
        :placeholder="placeholder || t('common.select','Select a category')"
        @update:value="handleChange"
        :disabled="loading || categories.length === 0 || props.disabled"
      />
    </NSpin>
    <p v-if="error" class="error">{{ error }}</p>
  </div>
</template>

<style scoped>
.category-selector {
  width: 100%;
}
.error {
  color: #d03050;
  font-size: 0.8rem;
  margin-top: 4px;
}
</style>
