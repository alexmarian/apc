<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { NForm, NFormItem, NInput, NButton, NSpace, NSpin, NAlert, useMessage } from 'naive-ui'
import { categoryApi } from '@/services/api'
import type { Category, CategoryCreateRequest, CategoryUpdateRequest } from '@/types/api'
import type { FormRules } from 'naive-ui'
import { useI18n } from 'vue-i18n'

// Props
const props = defineProps<{
  associationId: number
  categoryId?: number // If provided, we're editing an existing category
}>()

// Emits
const emit = defineEmits<{
  (e: 'saved', category: Category): void  // Pass the updated/created category data
  (e: 'cancelled'): void
}>()

// I18n
const { t } = useI18n()
const message = useMessage()

// Form data
const formData = reactive<CategoryCreateRequest>({
  type: '',
  family: '',
  name: ''
})

// Form validation rules
const rules: FormRules = {
  type: [
    {
      required: true,
      message: t('categories.validation.typeRequired'),
      trigger: 'blur'
    },
    {
      max: 100,
      message: t('categories.validation.maxLength', { max: 100 }),
      trigger: 'blur'
    }
  ],
  family: [
    {
      required: true,
      message: t('categories.validation.familyRequired'),
      trigger: 'blur'
    },
    {
      max: 100,
      message: t('categories.validation.maxLength', { max: 100 }),
      trigger: 'blur'
    }
  ],
  name: [
    {
      required: true,
      message: t('categories.validation.nameRequired'),
      trigger: 'blur'
    },
    {
      max: 100,
      message: t('categories.validation.maxLength', { max: 100 }),
      trigger: 'blur'
    }
  ]
}

// State
const loading = ref<boolean>(false)
const submitting = ref<boolean>(false)
const error = ref<string | null>(null)
const formRef = ref(null)
const originalCategory = ref<Category | null>(null)

const isEditMode = computed(() => !!props.categoryId)

// Fetch category details if editing
const fetchCategoryDetails = async () => {
  if (!props.categoryId) return

  try {
    loading.value = true
    error.value = null

    const response = await categoryApi.getCategory(props.associationId, props.categoryId)
    const categoryData = response.data
    originalCategory.value = categoryData

    // Update form data
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

// Submit form
const submitForm = async (e: MouseEvent) => {
  e.preventDefault()

  if (!formRef.value) return

  try {
    // @ts-ignore - Naive UI types issue with form ref
    await formRef.value.validate()

    submitting.value = true
    error.value = null

    let response: { data: Category }

    // Determine if creating or updating
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

    // Emit the updated/created category data so parent can update the list without reloading
    emit('saved', response.data)
  } catch (err) {
    let errorMessage: string
    if (err instanceof Error) {
      errorMessage = err.message
    } else if (typeof err === 'object' && err !== null && 'response' in err) {
      // Axios error
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

// Cancel form
const cancelForm = () => {
  emit('cancelled')
}

// On component mount
onMounted(() => {
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
        {{ formData.type }} → {{ formData.family }} → {{ formData.name }}
      </NAlert>

      <NForm
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-placement="left"
        label-width="120px"
        require-mark-placement="right-hanging"
      >
        <!-- Readonly Category ID for edit mode -->
        <NFormItem v-if="isEditMode && originalCategory" label="Category ID">
          <NInput
            :value="originalCategory.id.toString()"
            readonly
            class="category-form__readonly-input"
          />
        </NFormItem>

        <NFormItem :label="t('categories.type')" path="type">
          <NInput
            v-model:value="formData.type"
            :placeholder="t('categories.typePlaceholder')"
            maxlength="100"
            show-count
          />
        </NFormItem>

        <NFormItem :label="t('categories.family')" path="family">
          <NInput
            v-model:value="formData.family"
            :placeholder="t('categories.familyPlaceholder')"
            maxlength="100"
            show-count
          />
        </NFormItem>

        <NFormItem :label="t('categories.name')" path="name">
          <NInput
            v-model:value="formData.name"
            :placeholder="t('categories.namePlaceholder')"
            maxlength="100"
            show-count
          />
        </NFormItem>

        <div style="margin-top: 24px;">
          <NSpace justify="end">
            <NButton
              @click="cancelForm"
              :disabled="submitting"
            >
              {{ t('common.cancel') }}
            </NButton>

            <NButton
              type="primary"
              @click="submitForm"
              :loading="submitting"
            >
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
