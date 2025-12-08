<script setup lang="ts">
import { ref, computed } from 'vue'
import { NCard, NPageHeader, NModal } from 'naive-ui'
import CategoriesList from '@/components/CategoriesList.vue'
import CategoryForm from '@/components/CategoryForm.vue'
import AssociationSelector from '@/components/AssociationSelector.vue'
import type { Category } from '@/types/api'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

// State
const associationId = ref<number | null>(null)
const showCategoryModal = ref(false)
const editingCategoryId = ref<number | undefined>(undefined)

// Reference to the CategoriesList component
const categoriesListRef = ref<InstanceType<typeof CategoriesList> | null>(null)

// Computed properties
const modalTitle = computed(() => {
  return editingCategoryId.value ? t('categories.editCategory') : t('categories.createNew')
})

const canShowCategories = computed(() => {
  return associationId.value !== null
})

// Methods
const handleEditCategory = (categoryId: number) => {
  editingCategoryId.value = categoryId
  showCategoryModal.value = true
}

const handleCreateCategory = () => {
  editingCategoryId.value = undefined
  showCategoryModal.value = true
}

const handleCategorySaved = (savedCategory: Category) => {
  console.log('Category saved:', savedCategory)

  // Update or add the category in the list without reloading
  if (categoriesListRef.value) {
    if (editingCategoryId.value) {
      // Update existing category
      categoriesListRef.value.updateCategory(savedCategory)
    } else {
      // Add new category
      categoriesListRef.value.addCategory(savedCategory)
    }
  }

  // Close the modal
  showCategoryModal.value = false
  editingCategoryId.value = undefined
}

const handleCategoryFormCancelled = () => {
  showCategoryModal.value = false
  editingCategoryId.value = undefined
}

const handleAssociationChanged = (newAssociationId: number | null) => {
  associationId.value = newAssociationId
  // Close any open modals when association changes
  showCategoryModal.value = false
  editingCategoryId.value = undefined
}
</script>

<template>
  <div class="categories-page">
    <NPageHeader>
      <template #title>
        {{ t('categories.title') }}
      </template>

      <template #header>
        <div style="margin-bottom: 12px;">
          <AssociationSelector
            v-model:associationId="associationId"
            @update:associationId="handleAssociationChanged"
          />
        </div>
      </template>
    </NPageHeader>

    <div v-if="!associationId">
      <NCard style="margin-top: 16px;">
        <div style="text-align: center; padding: 32px;">
          <p>{{ t('categories.selectAssociation') }}</p>
        </div>
      </NCard>
    </div>

    <div v-else-if="canShowCategories">
      <!-- Categories List -->
      <CategoriesList
        ref="categoriesListRef"
        :association-id="associationId"
        @edit="handleEditCategory"
        @create="handleCreateCategory"
      />
    </div>

    <!-- Category Edit/Create Modal -->
    <NModal
      v-model:show="showCategoryModal"
      style="width: 650px"
      preset="card"
      :title="modalTitle"
      :mask-closable="false"
      :close-on-esc="true"
    >
      <CategoryForm
        v-if="showCategoryModal && associationId"
        :association-id="associationId"
        :category-id="editingCategoryId"
        @saved="handleCategorySaved"
        @cancelled="handleCategoryFormCancelled"
      />
    </NModal>
  </div>
</template>

<style scoped>
.categories-page {
  width: 100%;
}
</style>
