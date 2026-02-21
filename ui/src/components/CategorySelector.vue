<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NSelect, NSpin } from 'naive-ui'
import type { SelectInst } from 'naive-ui'
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
    label: `${t(`categories.types.${category.type}`, category.type)} → ${t(`categories.families.${category.family}`, category.family)} → ${t(`categories.names.${category.name}`, category.name)}`,
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

  } catch (err) {
    console.error('Error fetching categories:', err)
    error.value = 'Failed to load categories'
  } finally {
    loading.value = false
  }
}

const normalize = (s: string) =>
  s.normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase()

const filterOption = (pattern: string, option: { label: string; value: number | null }) => {
  return normalize(option.label).includes(normalize(pattern))
}

const selectRef = ref<SelectInst | null>(null)
const currentFilter = ref('')

const handleFocus = () => {
  // When tabbing into the select, immediately move focus to the search input
  // so the user can start typing without an extra click
  selectRef.value?.focusInput()
}

const handleSearch = (value: string) => {
  currentFilter.value = value
}

const handleTabKey = (e: KeyboardEvent) => {
  if (e.key !== 'Tab') return
  const query = currentFilter.value.trim()
  if (!query) return
  const matched = options.value.find(o =>
    o.value !== null && normalize(o.label).includes(normalize(query))
  )
  if (matched) {
    emit('update:modelValue', matched.value)
    currentFilter.value = ''
  }
}

// Handle selection change
const handleChange = (value: number | null) => {
  currentFilter.value = ''
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
        ref="selectRef"
        filterable
        :filter="filterOption"
        :value="props.modelValue"
        :options="options"
        :placeholder="placeholder || t('common.select','Select a category')"
        @update:value="handleChange"
        @search="handleSearch"
        @focus="handleFocus"
        :disabled="loading || categories.length === 0 || props.disabled"
        :input-props="{ onKeydown: handleTabKey }"
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
