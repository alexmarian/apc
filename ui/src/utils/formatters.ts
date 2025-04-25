/**
 * Format number as currency
 * @param value Number to format as currency
 * @param currency Currency code (default: RON)
 * @returns Formatted currency string
 */
export const formatCurrency = (value: number, currency: string = 'MDL'): string => {
  return new Intl.NumberFormat('ro-RO', {
    style: 'currency',
    currency,
    minimumFractionDigits: 2
  }).format(value)
}

/**
 * Format date to locale string
 * @param dateStr Date string in ISO format
 * @param format Format option: 'short', 'medium', 'long'
 * @returns Formatted date string
 */
export const formatDate = (dateStr: string, format: 'short' | 'medium' | 'long' = 'medium'): string => {
  const date = new Date(dateStr)
  let options: Intl.DateTimeFormatOptions

  switch (format) {
    case 'short':
      options = { day: '2-digit', month: '2-digit', year: 'numeric' }
      break
    case 'long':
      options = { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' }
      break
    case 'medium':
    default:
      options = { day: 'numeric', month: 'long', year: 'numeric' }
  }

  return date.toLocaleDateString('ro-RO', options)
}
