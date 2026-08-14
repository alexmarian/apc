import { defineStore } from 'pinia'
import { ref } from 'vue'
import { associationApi } from '@/services/api'
import { syncCategoryLabels } from '@/utils/categoryLabels'
import type { Association } from '@/types/api'

export const useAssociationStore = defineStore('association', () => {
  const associationId = ref<number | null>(null)
  const associations = ref<Association[]>([])
  const loading = ref(false)

  function select(id: number) {
    associationId.value = id
    syncCategoryLabels(id)
  }

  async function init() {
    // Guard on the list, not on associationId: the switcher needs the full list
    // even once a selection exists.
    if (associations.value.length > 0) return
    loading.value = true
    try {
      const response = await associationApi.getAssociations()
      associations.value = response.data
      if (associationId.value === null && response.data.length > 0) {
        select(response.data[0].id)
      }
    } catch (err) {
      console.error('Failed to load association:', err)
    } finally {
      loading.value = false
    }
  }

  function reset() {
    associationId.value = null
    associations.value = []
    loading.value = false
  }

  return { associationId, associations, loading, select, init, reset }
})
