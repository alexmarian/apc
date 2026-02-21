<script setup lang="ts">
import { ref, onMounted, h, watch, computed } from 'vue'
import {
  NDataTable, NButton, NSpace, NTag, NEmpty, NSpin, NAlert, NInput,
  NCheckbox, NCard, NSelect, useMessage, useDialog
} from 'naive-ui'
import type { DataTableColumns, DataTableRowKey } from 'naive-ui'
import { categoryApi } from '@/services/api'
import type { Category } from '@/types/api'
import { useI18n } from 'vue-i18n'

// Props
const props = defineProps<{
  associationId: number
}>()

// Emits
const emit = defineEmits<{
  (e: 'edit', categoryId: number): void
  (e: 'create'): void
}>()

// I18n
const { t } = useI18n()

// Data
const categories = ref<Category[]>([])
const loading = ref<boolean>(false)
const error = ref<string | null>(null)
const message = useMessage()
const dialog = useDialog()
const hasInitialized = ref(false)

// Filter state
const searchQuery = ref('')
const includeInactive = ref(false)
const selectedRowKeys = ref<DataTableRowKey[]>([])
const typeFilter = ref<string | null>(null)
const familyFilter = ref<string | null>(null)

// Get unique types for filter dropdown
const uniqueTypes = computed(() => {
  const types = [...new Set(categories.value.map(c => c.type))]
  return types.sort()
})

// Get unique families for filter dropdown (filtered by selected type if applicable)
const uniqueFamilies = computed(() => {
  const cats = typeFilter.value
    ? categories.value.filter(c => c.type === typeFilter.value)
    : categories.value
  const families = [...new Set(cats.map(c => c.family))]
  return families.sort()
})

// Filtered categories based on search and filters
const filteredCategories = computed(() => {
  let result = categories.value

  // Apply type filter
  if (typeFilter.value) {
    result = result.filter(category => category.type === typeFilter.value)
  }

  // Apply family filter
  if (familyFilter.value) {
    result = result.filter(category => category.family === familyFilter.value)
  }

  // Apply search query
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(category =>
      category.type.toLowerCase().includes(query) ||
      category.family.toLowerCase().includes(query) ||
      category.name.toLowerCase().includes(query)
    )
  }

  return result
})

// Check if any selected categories are active or inactive
const hasSelectedActive = computed(() => {
  return selectedRowKeys.value.some(key => {
    const category = categories.value.find(c => c.id === key)
    return category && !category.is_deleted
  })
})

const hasSelectedInactive = computed(() => {
  return selectedRowKeys.value.some(key => {
    const category = categories.value.find(c => c.id === key)
    return category && category.is_deleted
  })
})

// Table columns definition
const columns = computed<DataTableColumns<Category>>(() => [
  {
    type: 'selection'
  },
  {
    title: 'ID',
    key: 'id',
    width: 80,
    sorter: (a, b) => a.id - b.id
  },
  {
    title: t('categories.fullPath'),
    key: 'fullPath',
    sorter: (a, b) => {
      const pathA = `${t(`categories.types.${a.type}`, a.type)} → ${t(`categories.families.${a.family}`, a.family)} → ${t(`categories.names.${a.name}`, a.name)}`
      const pathB = `${t(`categories.types.${b.type}`, b.type)} → ${t(`categories.families.${b.family}`, b.family)} → ${t(`categories.names.${b.name}`, b.name)}`
      return pathA.localeCompare(pathB)
    },
    render(row) {
      return `${t(`categories.types.${row.type}`, row.type)} → ${t(`categories.families.${row.family}`, row.family)} → ${t(`categories.names.${row.name}`, row.name)}`
    }
  },
  {
    title: t('categories.type'),
    key: 'type',
    width: 150,
    sorter: 'default',
    render(row) {
      return t(`categories.types.${row.type}`, row.type)
    }
  },
  {
    title: t('categories.family'),
    key: 'family',
    width: 150,
    sorter: 'default',
    render(row) {
      return t(`categories.families.${row.family}`, row.family)
    }
  },
  {
    title: t('categories.name'),
    key: 'name',
    width: 150,
    sorter: 'default',
    render(row) {
      return t(`categories.names.${row.name}`, row.name)
    }
  },
  {
    title: t('categories.status'),
    key: 'is_deleted',
    width: 120,
    render(row) {
      return h(
        NTag,
        {
          type: !row.is_deleted ? 'success' : 'error',
          bordered: false
        },
        { default: () => !row.is_deleted ? t('categories.active') : t('categories.inactive') }
      )
    }
  },
  {
    title: t('common.actions'),
    key: 'actions',
    width: 200,
    render(row) {
      return h(
        NSpace,
        {
          justify: 'center',
          align: 'center'
        },
        {
          default: () => [
            h(
              NButton,
              {
                strong: true,
                secondary: true,
                type: 'info',
                size: 'small',
                disabled: row.is_deleted,
                onClick: () => emit('edit', row.id)
              },
              { default: () => t('common.edit') }
            ),
            !row.is_deleted ? h(
              NButton,
              {
                strong: true,
                secondary: true,
                type: 'error',
                size: 'small',
                onClick: () => deactivateCategory(row.id)
              },
              { default: () => t('categories.deactivate') }
            ) : h(
              NButton,
              {
                strong: true,
                secondary: true,
                type: 'success',
                size: 'small',
                onClick: () => reactivateCategory(row.id)
              },
              { default: () => t('categories.reactivate') }
            )
          ]
        }
      )
    }
  }
])

// Fetch categories
const fetchCategories = async () => {
  if (loading.value) {
    console.log('Already loading categories, skipping fetch')
    return
  }

  if (!props.associationId) {
    console.log('Missing associationId, skipping fetch')
    return
  }

  try {
    loading.value = true
    error.value = null
    console.log('Fetching categories for association:', props.associationId)

    const response = await categoryApi.getAllCategories(props.associationId, includeInactive.value)
    categories.value = response.data
    hasInitialized.value = true

    console.log('Categories fetched successfully:', response.data.length, 'categories')
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('common.error')
    console.error('Error fetching categories:', err)
  } finally {
    loading.value = false
  }
}

// Method to update a single category in the list
const updateCategory = (updatedCategory: Category) => {
  const index = categories.value.findIndex(category => category.id === updatedCategory.id)
  if (index !== -1) {
    categories.value.splice(index, 1, updatedCategory)
    console.log('Category updated in list:', updatedCategory)
  }
}

// Method to add a new category to the list
const addCategory = (newCategory: Category) => {
  categories.value.push(newCategory)
  console.log('Category added to list:', newCategory)
}

// Expose methods for parent components
defineExpose({
  updateCategory,
  addCategory,
  refreshData: fetchCategories
})

// Deactivate single category
const deactivateCategory = async (categoryId: number) => {
  try {
    dialog.warning({
      title: t('categories.confirmDeactivate'),
      content: t('categories.deactivateWarning'),
      positiveText: t('common.confirm'),
      negativeText: t('common.cancel'),
      onPositiveClick: async () => {
        try {
          await categoryApi.deactivateCategory(props.associationId, categoryId)

          // Update the local state
          const index = categories.value.findIndex(cat => cat.id === categoryId)
          if (index !== -1) {
            categories.value[index].is_deleted = true
          }

          message.success(t('categories.categoryDeactivated'))
        } catch (err) {
          const errorMessage = err instanceof Error ? err.message : t('common.error')
          error.value = errorMessage
          console.error('Error deactivating category:', err)
          message.error(t('common.error') + ': ' + errorMessage)
        }
      }
    })
  } catch (err) {
    console.error('Error showing dialog:', err)
  }
}

// Reactivate single category
const reactivateCategory = async (categoryId: number) => {
  try {
    await categoryApi.reactivateCategory(props.associationId, categoryId)

    // Update the local state
    const index = categories.value.findIndex(cat => cat.id === categoryId)
    if (index !== -1) {
      categories.value[index].is_deleted = false
    }

    message.success(t('categories.categoryReactivated'))
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : t('common.error')
    error.value = errorMessage
    console.error('Error reactivating category:', err)
    message.error(t('common.error') + ': ' + errorMessage)
  }
}

// Bulk deactivate
const bulkDeactivate = async () => {
  if (selectedRowKeys.value.length === 0) return

  dialog.warning({
    title: t('categories.confirmBulkDeactivate', { count: selectedRowKeys.value.length }),
    content: t('categories.bulkDeactivateWarning'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      try {
        const categoryIds = selectedRowKeys.value.map(key => Number(key))
        await categoryApi.bulkDeactivate(props.associationId, categoryIds)

        // Update local state
        categoryIds.forEach(id => {
          const index = categories.value.findIndex(cat => cat.id === id)
          if (index !== -1) {
            categories.value[index].is_deleted = true
          }
        })

        selectedRowKeys.value = []
        message.success(t('categories.bulkDeactivateSuccess'))
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : t('common.error')
        error.value = errorMessage
        console.error('Error bulk deactivating categories:', err)
        message.error(t('common.error') + ': ' + errorMessage)
      }
    }
  })
}

// Bulk reactivate
const bulkReactivate = async () => {
  if (selectedRowKeys.value.length === 0) return

  try {
    const categoryIds = selectedRowKeys.value.map(key => Number(key))
    await categoryApi.bulkReactivate(props.associationId, categoryIds)

    // Update local state
    categoryIds.forEach(id => {
      const index = categories.value.findIndex(cat => cat.id === id)
      if (index !== -1) {
        categories.value[index].is_deleted = false
      }
    })

    selectedRowKeys.value = []
    message.success(t('categories.bulkReactivateSuccess'))
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : t('common.error')
    error.value = errorMessage
    console.error('Error bulk reactivating categories:', err)
    message.error(t('common.error') + ': ' + errorMessage)
  }
}

// Watch for associationId changes
watch(() => props.associationId,
  (newAssocId, oldAssocId) => {
    console.log('Association ID changed:', { newAssocId, oldAssocId })

    if (newAssocId && newAssocId !== oldAssocId) {
      hasInitialized.value = false
      categories.value = []
      selectedRowKeys.value = []
      fetchCategories()
    }
  },
  { immediate: false }
)

// Watch for includeInactive changes
watch(includeInactive, () => {
  fetchCategories()
})

// Load categories on component mount
onMounted(() => {
  if (props.associationId && !hasInitialized.value) {
    fetchCategories()
  }
})
</script>

<template>
  <div class="categories-list">
    <NCard style="margin-top: 16px;">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>{{ t('categories.list') }}</span>
          <NSpace>
            <NButton
              v-if="hasSelectedActive"
              type="error"
              secondary
              size="small"
              :disabled="selectedRowKeys.length === 0"
              @click="bulkDeactivate"
            >
              {{ t('categories.bulkDeactivate') }} ({{ selectedRowKeys.length }})
            </NButton>
            <NButton
              v-if="hasSelectedInactive"
              type="success"
              secondary
              size="small"
              :disabled="selectedRowKeys.length === 0"
              @click="bulkReactivate"
            >
              {{ t('categories.bulkReactivate') }} ({{ selectedRowKeys.length }})
            </NButton>
            <NButton
              type="primary"
              @click="emit('create')"
            >
              {{ t('categories.createNew') }}
            </NButton>
          </NSpace>
        </div>
      </template>

      <NSpace vertical :size="16">
        <!-- Search and Filter Controls -->
        <NSpace>
          <NInput
            v-model:value="searchQuery"
            :placeholder="t('categories.searchPlaceholder')"
            clearable
            style="width: 300px;"
          />
          <NSelect
            v-model:value="typeFilter"
            :options="[
              { label: t('categories.allTypes'), value: undefined, type:'ignored' },
              ...uniqueTypes.map(type => ({ label: type, value: type }))
            ]"
            :placeholder="t('categories.filterByType')"
            clearable
            style="width: 200px;"
          />
          <NSelect
            v-model:value="familyFilter"
            :options="[
              { label: t('categories.allFamilies'), value: undefined, type:'ignored' },
              ...uniqueFamilies.map(family => ({ label: family, value: family }))
            ]"
            :placeholder="t('categories.filterByFamily')"
            clearable
            style="width: 200px;"
            :disabled="!typeFilter && uniqueFamilies.length === 0"
          />
          <NCheckbox v-model:checked="includeInactive">
            {{ t('categories.includeInactive') }}
          </NCheckbox>
        </NSpace>

        <NAlert v-if="error" type="error" :title="t('common.error')" closable @close="error = null">
          {{ error }}
        </NAlert>

        <NSpin :show="loading">
          <NDataTable
            v-if="filteredCategories.length > 0"
            :columns="columns"
            :data="filteredCategories"
            :row-key="(row: Category) => row.id"
            v-model:checked-row-keys="selectedRowKeys"
            :pagination="{
              pageSize: 50
            }"
          />
          <NEmpty
            v-else
            :description="t('categories.noCategories')"
            style="padding: 40px 0;"
          >
            <template #extra>
              <NButton type="primary" @click="emit('create')">
                {{ t('categories.createNew') }}
              </NButton>
            </template>
          </NEmpty>
        </NSpin>
      </NSpace>
    </NCard>
  </div>
</template>

<style scoped>
.categories-list {
  width: 100%;
}
</style>
