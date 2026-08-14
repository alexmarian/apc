<script setup lang="ts">
import { ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { NButton, NModal } from 'naive-ui'
import PageShell from '@/components/PageShell.vue'
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
  // Update or add the category in the list without reloading
  if (categoriesListRef.value) {
    if (editingCategoryId.value) {
      categoriesListRef.value.updateCategory(savedCategory)
    } else {
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
  <PageShell :title="t('categories.title')">
    <template #actions>
      <NButton type="primary" :disabled="!associationId" @click="handleCreateCategory">
        {{ t('categories.createNew') }}
      </NButton>
    </template>

    <CategoriesList
      v-if="canShowCategories"
      ref="categoriesListRef"
      :association-id="associationId"
      @edit="handleEditCategory"
      @create="handleCreateCategory"
    />

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
  </PageShell>
</template>
