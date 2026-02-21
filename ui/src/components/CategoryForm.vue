<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { NForm, NFormItem, NButton, NSpace, NSpin, NAlert, NAutoComplete, useMessage } from 'naive-ui'
import { categoryApi } from '@/services/api'
import type { Category, CategoryCreateRequest, CategoryUpdateRequest } from '@/types/api'
import type { FormRules } from 'naive-ui'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  associationId: number
  categoryId?: number
}>()

const emit = defineEmits<{
  (e: 'saved', category: Category): void
  (e: 'cancelled'): void
}>()

const { t } = useI18n()
const message = useMessage()

const formData = reactive<CategoryCreateRequest>({
  type: '',
  family: '',
  name: ''
})

const rules: FormRules = {
  type: [
    { required: true, message: t('categories.validation.typeRequired'), trigger: 'blur' },
    { max: 100, message: t('categories.validation.maxLength', { max: 100 }), trigger: 'blur' }
  ],
  family: [
    { required: true, message: t('categories.validation.familyRequired'), trigger: 'blur' },
    { max: 100, message: t('categories.validation.maxLength', { max: 100 }), trigger: 'blur' }
  ],
  name: [
    { required: true, message: t('categories.validation.nameRequired'), trigger: 'blur' },
    { max: 100, message: t('categories.validation.maxLength', { max: 100 }), trigger: 'blur' }
  ]
}

const loading = ref(false)
const submitting = ref(false)
const error = ref<string | null>(null)
const formRef = ref(null)
const originalCategory = ref<Category | null>(null)
const allCategories = ref<Category[]>([])

const isEditMode = computed(() => !!props.categoryId)

// Build unique string list preserving insertion order
function uniqueStrings(values: string[]): string[] {
  const seen = new Set<string>()
  return values.filter(v => v && !seen.has(v) && seen.add(v))
}

// Suggestions derived from existing DB categories, with localized labels
const typeOptions = computed(() =>
  uniqueStrings(allCategories.value.map(c => c.type)).map(v => ({
    value: v,
    label: t(`categories.types.${v}`, v)
  }))
)

const familyOptions = computed(() => {
  const source = formData.type
    ? allCategories.value.filter(c => c.type === formData.type)
    : allCategories.value
  return uniqueStrings(source.map(c => c.family)).map(v => ({
    value: v,
    label: t(`categories.families.${v}`, v)
  }))
})

const nameOptions = computed(() => {
  let source = allCategories.value
  if (formData.type) source = source.filter(c => c.type === formData.type)
  if (formData.family) source = source.filter(c => c.family === formData.family)
  return uniqueStrings(source.map(c => c.name)).map(v => ({
    value: v,
    label: t(`categories.names.${v}`, v)
  }))
})

// Filter against localized label so user can type in their current language
function filterOptions(input: string, options: { value: string; label: string }[]) {
  if (!input) return options
  const lower = input.toLowerCase()
  return options.filter(o =>
    o.label.toLowerCase().includes(lower) || o.value.toLowerCase().includes(lower)
  )
}

const filteredTypeOptions = computed(() => filterOptions(formData.type, typeOptions.value))
const filteredFamilyOptions = computed(() => filterOptions(formData.family, familyOptions.value))
const filteredNameOptions = computed(() => filterOptions(formData.name, nameOptions.value))

const fetchAllCategories = async () => {
  try {
    const response = await categoryApi.getAllCategories(props.associationId, true)
    allCategories.value = response.data
  } catch (err) {
    console.warn('Could not fetch categories for suggestions:', err)
  }
}

const fetchCategoryDetails = async () => {
  if (!props.categoryId) return
  try {
    loading.value = true
    error.value = null
    const response = await categoryApi.getCategory(props.associationId, props.categoryId)
    const categoryData = response.data
    originalCategory.value = categoryData
    formData.type = categoryData.type
    formData.family = categoryData.family
    formData.name = categoryData.name
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('common.error')
    console.error('Error fetching category details:', err)
  } finally {
    loading.value = false
  }
}

const submitForm = async (e: MouseEvent) => {
  e.preventDefault()
  if (!formRef.value) return

  try {
    // @ts-ignore
    await formRef.value.validate()
    submitting.value = true
    error.value = null

    let response: { data: Category }

    if (isEditMode.value && props.categoryId) {
      const updateData: CategoryUpdateRequest = {
        type: formData.type,
        family: formData.family,
        name: formData.name
      }
      response = await categoryApi.updateCategory(props.associationId, props.categoryId, updateData)
      message.success(t('categories.categoryUpdated'))
    } else {
      response = await categoryApi.createCategory(props.associationId, formData)
      message.success(t('categories.categoryCreated'))
    }

    emit('saved', response.data)
  } catch (err) {
    let errorMessage: string
    if (err instanceof Error) {
      errorMessage = err.message
    } else if (typeof err === 'object' && err !== null && 'response' in err) {
      const axiosError = err as any
      errorMessage = axiosError.response?.data?.msg || t('common.error')
    } else {
      errorMessage = t('common.error')
    }
    error.value = errorMessage
    message.error(errorMessage)
    console.error('Error submitting form:', err)
  } finally {
    submitting.value = false
  }
}

const cancelForm = () => emit('cancelled')

onMounted(async () => {
  await fetchAllCategories()
  if (props.categoryId) {
    fetchCategoryDetails()
  }
})
</script>

<template>
  <div class="category-form">
    <NSpin :show="loading">
      <NAlert v-if="error" type="error" :title="t('common.error')" style="margin-bottom: 16px;" closable @close="error = null">
        {{ error }}
      </NAlert>

      <!-- Live Preview -->
      <NAlert v-if="formData.type && formData.family && formData.name" type="info" style="margin-bottom: 16px;">
        <template #header>{{ t('categories.preview') }}</template>
        {{ t(`categories.types.${formData.type}`, formData.type) }} → {{ t(`categories.families.${formData.family}`, formData.family) }} → {{ t(`categories.names.${formData.name}`, formData.name) }}
      </NAlert>

      <NForm
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-placement="left"
        label-width="120px"
        require-mark-placement="right-hanging"
      >
        <NFormItem v-if="isEditMode && originalCategory" label="Category ID">
          <NAutoComplete
            :value="originalCategory.id.toString()"
            :options="[]"
            :disabled="true"
            class="category-form__readonly-input"
          />
        </NFormItem>

        <NFormItem :label="t('categories.type')" path="type">
          <NAutoComplete
            v-model:value="formData.type"
            :options="filteredTypeOptions"
            :placeholder="t('categories.typePlaceholder')"
            clearable
          />
        </NFormItem>

        <NFormItem :label="t('categories.family')" path="family">
          <NAutoComplete
            v-model:value="formData.family"
            :options="filteredFamilyOptions"
            :placeholder="t('categories.familyPlaceholder')"
            clearable
          />
        </NFormItem>

        <NFormItem :label="t('categories.name')" path="name">
          <NAutoComplete
            v-model:value="formData.name"
            :options="filteredNameOptions"
            :placeholder="t('categories.namePlaceholder')"
            clearable
          />
        </NFormItem>

        <div style="margin-top: 24px;">
          <NSpace justify="end">
            <NButton @click="cancelForm" :disabled="submitting">
              {{ t('common.cancel') }}
            </NButton>
            <NButton type="primary" @click="submitForm" :loading="submitting">
              {{ isEditMode ? t('common.update') : t('common.create') }}
            </NButton>
          </NSpace>
        </div>
      </NForm>
    </NSpin>
  </div>
</template>

<style scoped>
.category-form {
  width: 100%;
}

.category-form__readonly-input {
  background-color: #f5f5f5;
}
</style>
