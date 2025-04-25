// Export all utility functions

// Theme utilities
export * from './theme'

// Formatter utilities
export * from './formatters'

// Date utilities
export const getCurrentMonthRange = (): [string, string] => {
  const now = new Date()
  const firstDay = new Date(now.getFullYear(), now.getMonth(), 1)
  const lastDay = new Date(now.getFullYear(), now.getMonth() + 1, 0)

  return [
    firstDay.toISOString().split('T')[0],
    lastDay.toISOString().split('T')[0]
  ]
}

export const getPreviousMonthRange = (): [string, string] => {
  const now = new Date()
  const firstDay = new Date(now.getFullYear(), now.getMonth() - 1, 1)
  const lastDay = new Date(now.getFullYear(), now.getMonth(), 0)

  return [
    firstDay.toISOString().split('T')[0],
    lastDay.toISOString().split('T')[0]
  ]
}

export const getCurrentYearRange = (): [string, string] => {
  const now = new Date()
  const firstDay = new Date(now.getFullYear(), 0, 1)
  const lastDay = new Date(now.getFullYear(), 11, 31)

  return [
    firstDay.toISOString().split('T')[0],
    lastDay.toISOString().split('T')[0]
  ]
}

// String utilities
export const truncateString = (str: string, maxLength: number = 50): string => {
  if (!str) return ''
  return str.length <= maxLength ? str : `${str.slice(0, maxLength)}...`
}
