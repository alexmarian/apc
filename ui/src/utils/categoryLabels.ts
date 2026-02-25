import { categoryApi } from '@/services/api'
import i18n from '@/i18n'

/**
 * Fetches categories and merges original_labels into i18n messages so that
 * t('categories.types.some_new_key') resolves to the friendly label even
 * before it is added to locale files.
 *
 * Also stores the overrides in localStorage so the standalone-charts page
 * (which has its own translation system including ru) can pick them up.
 */
export async function syncCategoryLabels(associationId: number) {
  try {
    const response = await categoryApi.getAllCategories(associationId, true)
    const categories = response.data

    const overrides: Record<string, Record<string, string>> = {
      types: {},
      families: {},
      names: {}
    }

    for (const cat of categories) {
      if (!cat.original_labels) continue
      if (cat.original_labels.type) overrides.types[cat.type] = cat.original_labels.type
      if (cat.original_labels.family) overrides.families[cat.family] = cat.original_labels.family
      if (cat.original_labels.name) overrides.names[cat.name] = cat.original_labels.name
    }

    const hasOverrides = Object.values(overrides).some(o => Object.keys(o).length > 0)
    if (!hasOverrides) {
      localStorage.removeItem('standalone_category_overrides')
      return
    }

    for (const locale of ['en', 'ro']) {
      i18n.global.mergeLocaleMessage(locale, {
        categories: overrides
      })
    }

    localStorage.setItem('standalone_category_overrides', JSON.stringify(overrides))
  } catch (err) {
    console.warn('Could not sync category labels:', err)
  }
}
