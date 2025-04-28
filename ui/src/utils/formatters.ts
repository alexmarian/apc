/**
 * Format a number as currency
 * @param value The number to format
 * @param locale The locale to use (defaults to 'en-US')
 * @param currency The currency code (defaults to 'USD')
 * @returns Formatted currency string
 */
export const formatCurrency = (
  value: number,
  locale: string = 'en-US',
  currency: string = 'MDL'
): string => {
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency: currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(value);
};

/**
 * Format a number as a percentage
 * @param value The number to format (e.g., 0.25 for 25%)
 * @param decimals Number of decimal places (defaults to 2)
 * @returns Formatted percentage string
 */
export const formatPercentage = (
  value: number,
  decimals: number = 2
): string => {
  return `${(value * 100).toFixed(decimals)}%`;
};

/**
 * Format a date to a localized string
 * @param date The date to format
 * @param format The format to use (defaults to 'short')
 * @param locale The locale to use (defaults to undefined, which uses the browser's locale)
 * @returns Formatted date string
 */
export const formatDate = (
  date: Date | string | number,
  format: 'short' | 'medium' | 'long' | 'full' = 'short',
  locale?: string
): string => {
  const dateObj = date instanceof Date ? date : new Date(date);
  return dateObj.toLocaleDateString(locale, { dateStyle: format });
};

/**
 * Format a file size in bytes to a human-readable string
 * @param bytes Number of bytes
 * @param decimals Number of decimal places (defaults to 2)
 * @returns Formatted file size string
 */
export const formatFileSize = (
  bytes: number,
  decimals: number = 2
): string => {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) + ' ' + sizes[i];
};
