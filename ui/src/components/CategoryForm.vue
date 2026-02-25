<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from 'vue'
import {
  NForm, NFormItem, NButton, NSpace, NSpin, NAlert,
  NSelect, NInput, NRadioGroup, NRadioButton, useMessage
} from 'naive-ui'
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

type FieldMode = 'existing' | 'new'

const typeMode = ref<FieldMode>('existing')
const familyMode = ref<FieldMode>('existing')

const typeExisting = ref<string | null>(null)
const typeLabel = ref('')
const typeKey = ref('')
const typeKeyManuallyEdited = ref(false)

const familyExisting = ref<string | null>(null)
const familyLabel = ref('')
const familyKey = ref('')
const familyKeyManuallyEdited = ref(false)

const nameLabel = ref('')
const nameKey = ref('')
const nameKeyManuallyEdited = ref(false)

const loading = ref(false)
const submitting = ref(false)
const error = ref<string | null>(null)
const formRef = ref(null)
const originalCategory = ref<Category | null>(null)
const allCategories = ref<Category[]>([])

const isEditMode = computed(() => !!props.categoryId)

function generateI18nKey(label: string): string {
  return label
    .toLowerCase()
    .trim()
    .replace(/\s+/g, '_')
    .replace(/[^a-z0-9_]/g, '')
}

function ensureUniqueKey(baseKey: string, field: 'type' | 'family' | 'name'): string {
  const existingKeys = new Set(allCategories.value.map(c => c[field]))
  if (!existingKeys.has(baseKey)) return baseKey
  let i = 1
  while (existingKeys.has(`${baseKey}_${i}`)) i++
  return `${baseKey}_${i}`
}

// Auto-generate keys from labels
watch(typeLabel, (val) => {
  if (!typeKeyManuallyEdited.value) {
    typeKey.value = val ? ensureUniqueKey(generateI18nKey(val), 'type') : ''
  }
})
watch(familyLabel, (val) => {
  if (!familyKeyManuallyEdited.value) {
    familyKey.value = val ? ensureUniqueKey(generateI18nKey(val), 'family') : ''
  }
})
watch(nameLabel, (val) => {
  if (!nameKeyManuallyEdited.value) {
    nameKey.value = val ? ensureUniqueKey(generateI18nKey(val), 'name') : ''
  }
})

function uniqueStrings(values: string[]): string[] {
  const seen = new Set<string>()
  return values.filter(v => v && !seen.has(v) && seen.add(v))
}

const typeSelectOptions = computed(() =>
  uniqueStrings(allCategories.value.map(c => c.type)).map(v => ({
    value: v,
    label: t(`categories.types.${v}`, v)
  }))
)

const familySelectOptions = computed(() => {
  const resolvedType = typeMode.value === 'existing' ? typeExisting.value : typeKey.value
  const source = resolvedType
    ? allCategories.value.filter(c => c.type === resolvedType)
    : allCategories.value
  return uniqueStrings(source.map(c => c.family)).map(v => ({
    value: v,
    label: t(`categories.families.${v}`, v)
  }))
})

const resolvedType = computed(() =>
  typeMode.value === 'existing' ? (typeExisting.value || '') : typeKey.value
)
const resolvedFamily = computed(() =>
  familyMode.value === 'existing' ? (familyExisting.value || '') : familyKey.value
)
const resolvedName = computed(() => nameKey.value)

const rules: FormRules = {
  type: [{ required: true, message: t('categories.validation.typeRequired'), trigger: 'blur' }],
  family: [{ required: true, message: t('categories.validation.familyRequired'), trigger: 'blur' }],
  name: [{ required: true, message: t('categories.validation.nameRequired'), trigger: 'blur' }]
}

const formModel = computed(() => ({
  type: resolvedType.value,
  family: resolvedFamily.value,
  name: resolvedName.value
}))

const previewType = computed(() => {
  if (typeMode.value === 'new' && typeLabel.value) return typeLabel.value
  if (typeMode.value === 'existing' && typeExisting.value) return t(`categories.types.${typeExisting.value}`, typeExisting.value)
  return ''
})
const previewFamily = computed(() => {
  if (familyMode.value === 'new' && familyLabel.value) return familyLabel.value
  if (familyMode.value === 'existing' && familyExisting.value) return t(`categories.families.${familyExisting.value}`, familyExisting.value)
  return ''
})
const previewName = computed(() => {
  if (nameLabel.value) return nameLabel.value
  return ''
})

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
    const cat = response.data
    originalCategory.value = cat

    if (cat.original_labels?.type) {
      typeMode.value = 'new'
      typeLabel.value = cat.original_labels.type
      typeKey.value = cat.type
      typeKeyManuallyEdited.value = true
    } else {
      typeMode.value = 'existing'
      typeExisting.value = cat.type
    }

    if (cat.original_labels?.family) {
      familyMode.value = 'new'
      familyLabel.value = cat.original_labels.family
      familyKey.value = cat.family
      familyKeyManuallyEdited.value = true
    } else {
      familyMode.value = 'existing'
      familyExisting.value = cat.family
    }

    if (cat.original_labels?.name) {
      nameLabel.value = cat.original_labels.name
      nameKey.value = cat.name
      nameKeyManuallyEdited.value = true
    } else {
      nameLabel.value = ''
      nameKey.value = cat.name
      nameKeyManuallyEdited.value = true
    }
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

    const originalLabels: Record<string, string> = {}
    if (typeMode.value === 'new' && typeLabel.value) originalLabels.type = typeLabel.value
    if (familyMode.value === 'new' && familyLabel.value) originalLabels.family = familyLabel.value
    if (nameLabel.value) originalLabels.name = nameLabel.value

    let response: { data: Category }

    if (isEditMode.value && props.categoryId) {
      const updateData: CategoryUpdateRequest = {
        type: resolvedType.value,
        family: resolvedFamily.value,
        name: resolvedName.value,
        original_labels: Object.keys(originalLabels).length > 0 ? originalLabels : undefined
      }
      response = await categoryApi.updateCategory(props.associationId, props.categoryId, updateData)
      message.success(t('categories.categoryUpdated'))
    } else {
      const createData: CategoryCreateRequest = {
        type: resolvedType.value,
        family: resolvedFamily.value,
        name: resolvedName.value,
        original_labels: Object.keys(originalLabels).length > 0 ? originalLabels : undefined
      }
      response = await categoryApi.createCategory(props.associationId, createData)
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

      <NAlert v-if="previewType && previewFamily && previewName" type="info" style="margin-bottom: 16px;">
        <template #header>{{ t('categories.preview') }}</template>
        {{ previewType }} → {{ previewFamily }} → {{ previewName }}
      </NAlert>

      <NForm
        ref="formRef"
        :model="formModel"
        :rules="rules"
        label-placement="left"
        label-width="120px"
        require-mark-placement="right-hanging"
      >
        <NFormItem v-if="isEditMode && originalCategory" label="Category ID">
          <NInput :value="originalCategory.id.toString()" disabled />
        </NFormItem>

        <!-- Type -->
        <NFormItem :label="t('categories.type')" path="type">
          <div style="width: 100%">
            <NRadioGroup v-model:value="typeMode" size="small" style="margin-bottom: 8px;">
              <NRadioButton value="existing">{{ t('categories.selectExisting') }}</NRadioButton>
              <NRadioButton value="new">{{ t('categories.createNewOption') }}</NRadioButton>
            </NRadioGroup>
            <NSelect
              v-if="typeMode === 'existing'"
              v-model:value="typeExisting"
              :options="typeSelectOptions"
              :placeholder="t('categories.typePlaceholder')"
              filterable
              clearable
            />
            <div v-else>
              <NInput
                v-model:value="typeLabel"
                :placeholder="t('categories.labelPlaceholder')"
                style="margin-bottom: 4px;"
              />
              <NInput
                v-model:value="typeKey"
                :placeholder="t('categories.keyPlaceholder')"
                :readonly="!typeKeyManuallyEdited"
                size="small"
                @click="typeKeyManuallyEdited = true"
              >
                <template #prefix>
                  <span style="color: #999; font-size: 12px;">{{ t('categories.i18nKey') }}:</span>
                </template>
              </NInput>
            </div>
          </div>
        </NFormItem>

        <!-- Family -->
        <NFormItem :label="t('categories.family')" path="family">
          <div style="width: 100%">
            <NRadioGroup v-model:value="familyMode" size="small" style="margin-bottom: 8px;">
              <NRadioButton value="existing">{{ t('categories.selectExisting') }}</NRadioButton>
              <NRadioButton value="new">{{ t('categories.createNewOption') }}</NRadioButton>
            </NRadioGroup>
            <NSelect
              v-if="familyMode === 'existing'"
              v-model:value="familyExisting"
              :options="familySelectOptions"
              :placeholder="t('categories.familyPlaceholder')"
              filterable
              clearable
            />
            <div v-else>
              <NInput
                v-model:value="familyLabel"
                :placeholder="t('categories.labelPlaceholder')"
                style="margin-bottom: 4px;"
              />
              <NInput
                v-model:value="familyKey"
                :placeholder="t('categories.keyPlaceholder')"
                :readonly="!familyKeyManuallyEdited"
                size="small"
                @click="familyKeyManuallyEdited = true"
              >
                <template #prefix>
                  <span style="color: #999; font-size: 12px;">{{ t('categories.i18nKey') }}:</span>
                </template>
              </NInput>
            </div>
          </div>
        </NFormItem>

        <!-- Name (always create new) -->
        <NFormItem :label="t('categories.name')" path="name">
          <div style="width: 100%">
            <NInput
              v-model:value="nameLabel"
              :placeholder="t('categories.labelPlaceholder')"
              style="margin-bottom: 4px;"
            />
            <NInput
              v-model:value="nameKey"
              :placeholder="t('categories.keyPlaceholder')"
              :readonly="!nameKeyManuallyEdited"
              size="small"
              @click="nameKeyManuallyEdited = true"
            >
              <template #prefix>
                <span style="color: #999; font-size: 12px;">{{ t('categories.i18nKey') }}:</span>
              </template>
            </NInput>
          </div>
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
</style>
