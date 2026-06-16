import { defineStore } from 'pinia'
import { ref } from 'vue'
import { associationApi } from '@/services/api'
import { syncCategoryLabels } from '@/utils/categoryLabels'

export const useAssociationStore = defineStore('association', () => {
  const associationId = ref<number | null>(null)
  const loading = ref(false)

  async function init() {
    if (associationId.value !== null) return
    loading.value = true
    try {
      const response = await associationApi.getAssociations()
      if (response.data.length > 0) {
        associationId.value = response.data[0].id
        syncCategoryLabels(response.data[0].id)
      }
    } catch (err) {
      console.error('Failed to load association:', err)
    } finally {
      loading.value = false
    }
  }

  function reset() {
    associationId.value = null
    loading.value = false
  }

  return { associationId, loading, init, reset }
})
