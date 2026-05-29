<script setup lang="ts">
import { ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { NCard, NModal } from 'naive-ui'
import CategoriesList from '@/components/CategoriesList.vue'
import CategoryForm from '@/components/CategoryForm.vue'
import { useAssociationStore } from '@/stores/association'
import type { Category } from '@/types/api'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const { associationId } = storeToRefs(useAssociationStore())
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

</script>

<template>
  <div class="categories-page">
    <div v-if="canShowCategories">
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
