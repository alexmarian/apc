import type { Expense } from '@/types/api'

export interface MonthlyExpenseData {
  month: string
  value: number
}

/**
 * Group expenses by month
 * @param expenses - Array of expenses to group
 * @returns Array of monthly expense data sorted by month
 */
export function groupExpensesByMonth(expenses: Expense[]): MonthlyExpenseData[] {
  const months: Record<string, number> = {}

  expenses.forEach(expense => {
    const date = new Date(expense.date)
    const monthYear = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`

    if (!months[monthYear]) {
      months[monthYear] = 0
    }

    months[monthYear] += expense.amount
  })

  return Object.entries(months).map(([month, value]) => ({
    month,
    value
  })).sort((a, b) => a.month.localeCompare(b.month))
}

/**
 * Calculate monthly average from expenses
 * @param expenses - Array of expenses
 * @returns Average expense amount per month
 */
export function calculateMonthlyAverage(expenses: Expense[]): number {
  if (expenses.length === 0) return 0

  const monthlyData = groupExpensesByMonth(expenses)
  const totalAmount = expenses.reduce((sum, expense) => sum + expense.amount, 0)

  return monthlyData.length > 0 ? totalAmount / monthlyData.length : 0
}
