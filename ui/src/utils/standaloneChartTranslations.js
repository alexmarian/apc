// ui/src/utils/standaloneChartTranslations.js
import i18n from '@/i18n'
import en from '@/i18n/locales/en.json'
import ro from '@/i18n/locales/ro.json'

// Import only the ru translations for the standalone charts
const ruStandaloneCharts = {
  title: 'Анализ Расходов',
  period: 'Период',
  printCharts: 'Печать Графиков',
  summary: 'Сводка',
  totalExpenses: 'Всего Расходов',
  numberOfExpenses: 'Количество Расходов',
  averageExpense: 'Средний Расход',
  expensesByType: 'Расходы по Типу',
  pieChart: 'Круговая Диаграмма',
  barChart: 'Столбчатая Диаграмма',
  type: 'Тип',
  amount: 'Сумма',
  percentage: 'Процент',
  count: 'Количество',
  monthlyExpenseTrends: 'Ежемесячные Тенденции Расходов',
  month: 'Месяц',
  totalAmount: 'Общая Сумма',
  noMonthlyData: 'Нет доступных ежемесячных данных',
  expenseTypeBreakdown: 'Разбивка по Типу Расходов',
  family: 'Семейство',
  noChartData: 'Нет доступных данных. Пожалуйста, вернитесь в основное приложение и снова откройте просмотр графиков.'
}

// Convert keys from i18n structure to standalone chart structure
function mapI18nToStandaloneFormat(i18nSource) {
  return {
    title: i18nSource.expenses.expenseAnalysis || i18nSource.expenses.title,
    period: i18nSource.expenses.period,
    printCharts: i18nSource.expenses.printCharts || 'Print Charts',
    summary: i18nSource.expenses.summary,
    totalExpenses: i18nSource.expenses.totalExpenses,
    numberOfExpenses: i18nSource.expenses.numberOfExpenses,
    averageExpense: i18nSource.expenses.averageExpense,
    expensesByType: i18nSource.expenses.expensesByType,
    pieChart: i18nSource.charts.pieChart,
    barChart: i18nSource.charts.barChart,
    type: i18nSource.categories.types.title,
    amount: i18nSource.expenses.amount,
    percentage: i18nSource.charts.percentage || 'Percentage',
    count: i18nSource.charts.count,
    monthlyExpenseTrends: i18nSource.expenses.monthlyTrends,
    month: i18nSource.expenses.month || 'Month',
    totalAmount: i18nSource.expenses.totalAmount,
    noMonthlyData: i18nSource.expenses.noMonthlyData,
    expenseTypeBreakdown: i18nSource.expenses.expenseTypeBreakdown,
    family: i18nSource.categories.families.title,
    noChartData: i18nSource.expenses.noChartData || 'No chart data available. Please go back to the main application and open charts view again.'
  }
}

// Create the combined translations object
const standaloneTranslations = {
  en: mapI18nToStandaloneFormat(en),
  ro: mapI18nToStandaloneFormat(ro),
  ru: ruStandaloneCharts
}

// Locale information for formatting
export const localeInfo = {
  'en': {
    locale: 'en-US',
    currency: 'MDL',
    rtl: false
  },
  'ro': {
    locale: 'ro-RO',
    currency: 'MDL',
    rtl: false
  },
  'ru': {
    locale: 'ru-RU',
    currency: 'MDL',
    rtl: false
  }
}

// Function to get standalone chart translations
export function getStandaloneTranslations() {
  return standaloneTranslations
}

// Function to prepare data for standalone charts
export function prepareDataForStandaloneCharts(expenses, expensesByType, monthlyData, typeDetails, dateRange) {
  // Ensure category translations are applied
  const translatedExpensesByType = expensesByType.map(item => ({
    ...item,
    name: i18n.global.t(`categories.types.${item.rawName || item.name}`)
  }))

  // Process monthly data with translations
  const translatedMonthlyData = {
    items: monthlyData.items,
    series: monthlyData.series.map(series => ({
      ...series,
      name: i18n.global.t(`categories.types.${series.name}`)
    }))
  }

  // Process type details with translations
  const translatedTypeDetails = typeDetails.map(type => ({
    ...type,
    type: i18n.global.t(`categories.types.${type.type}`),
    families: type.families.map(family => ({
      ...family,
      name: i18n.global.t(`categories.families.${family.rawName || family.name}`)
    }))
  }))

  // Create the data object
  const chartData = {
    expenses,
    expensesByType: translatedExpensesByType,
    expensesByMonth: translatedMonthlyData,
    typeDetails: translatedTypeDetails
  }

  // Store data in localStorage
  localStorage.setItem('standalone_chart_data', JSON.stringify(chartData))
  localStorage.setItem('standalone_chart_title', i18n.global.t('expenses.expenseAnalysis', 'Expense Analysis'))
  localStorage.setItem('standalone_chart_date_range', dateRange)
  localStorage.setItem('standalone_chart_language', i18n.global.locale.value)

  return chartData
}
