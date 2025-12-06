/**
 * Escape a CSV field value
 * Handles commas, quotes, and newlines
 */
export function escapeCsvField(value: string | number): string {
  const stringValue = String(value)

  // If the value contains comma, quote, or newline, wrap in quotes and escape quotes
  if (stringValue.includes(',') || stringValue.includes('"') || stringValue.includes('\n')) {
    return `"${stringValue.replace(/"/g, '""')}"`
  }

  return stringValue
}

/**
 * Convert array of rows to CSV string with BOM for Excel compatibility
 */
export function arrayToCsv(rows: (string | number)[][]): string {
  // Add BOM for UTF-8 Excel compatibility
  const BOM = '\uFEFF'

  const csvContent = rows
    .map(row => row.map(escapeCsvField).join(','))
    .join('\n')

  return BOM + csvContent
}

/**
 * Download CSV file
 */
export function downloadCsv(csvContent: string, filename: string): void {
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.setAttribute('href', url)
  link.setAttribute('download', filename)
  link.style.visibility = 'hidden'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}
