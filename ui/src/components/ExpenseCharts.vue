<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  NCard,
  NDivider,
  NSelect,
  NEmpty,
  NRadioGroup,
  NRadio,
  NButton,
  NTooltip,
  NSpace,
  NCollapse,
  NCollapseItem
} from 'naive-ui'
import { Pie, Bar, Doughnut } from 'vue-chartjs'
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  ArcElement,
  CategoryScale,
  LinearScale,
  BarElement
} from 'chart.js'
import type { ChartData, ChartOptions } from 'chart.js'
import type { Expense } from '@/types/api'
import { formatCurrency } from '@/utils/formatters'

// Register Chart.js components
ChartJS.register(
  Title,
  Tooltip,
  Legend,
  ArcElement,
  CategoryScale,
  LinearScale,
  BarElement
)

// Props
const props = defineProps<{
  expenses: Expense[]
}>()

// Chart display mode for each section
const typeChartMode = ref<'pie' | 'bar'>('pie')
const categoryChartMode = ref<'pie' | 'bar'>('pie')
const monthlyChartMode = ref<'bar'>('bar')

// Selected type for category breakdown
const selectedType = ref<string | null>(null)

// Refs for chart elements to export
const typeChartRef = ref<InstanceType<typeof Pie | typeof Bar> | null>(null)
const monthlyChartRef = ref<InstanceType<typeof Bar> | null>(null)
const categoryChartRefs = ref<Record<string, any>>({})

// Generate a color palette for the charts
const COLORS = [
  '#3366FF', '#FF6633', '#33CC99', '#FFCC33', '#FF33CC',
  '#33CCFF', '#CC99FF', '#99CC33', '#FF9966', '#6699FF'
]

// Interface for chart data items
interface ChartDataItem {
  name: string;
  value: number;
  count: number;
  color?: string;
  percentage?: number;
}

// Computed data for the charts
const expensesByType = computed<ChartDataItem[]>(() => {
  if (!props.expenses || props.expenses.length === 0) return []

  const grouped = props.expenses.reduce((acc, expense) => {
    const type = expense.category_type || 'Uncategorized'

    if (!acc[type]) {
      acc[type] = {
        name: type,
        value: 0,
        count: 0
      }
    }

    acc[type].value += expense.amount
    acc[type].count += 1

    return acc
  }, {} as Record<string, ChartDataItem>)

  return Object.values(grouped).map((item, index) => ({
    ...item,
    color: COLORS[index % COLORS.length],
    percentage: (item.value / props.expenses.reduce((sum, exp) => sum + exp.amount, 0)) * 100
  }))
})

// Family categorization - group expenses within each type by family
const expensesByTypeAndFamily = computed<Record<string, Record<string, ChartDataItem>>>(() => {
  if (!props.expenses || props.expenses.length === 0) return {}

  const typeAndFamily: Record<string, Record<string, ChartDataItem>> = {}

  // First pass - group by type and family
  props.expenses.forEach(expense => {
    const type = expense.category_type || 'Uncategorized'
    const family = expense.category_family || 'General'

    if (!typeAndFamily[type]) {
      typeAndFamily[type] = {}
    }

    if (!typeAndFamily[type][family]) {
      typeAndFamily[type][family] = {
        name: family,
        value: 0,
        count: 0
      }
    }

    typeAndFamily[type][family].value += expense.amount
    typeAndFamily[type][family].count += 1
  })

  // Second pass - calculate percentages and add colors
  Object.keys(typeAndFamily).forEach(type => {
    const totalForType = Object.values(typeAndFamily[type])
    .reduce((sum, family) => sum + family.value, 0)

    Object.keys(typeAndFamily[type]).forEach((family, index) => {
      typeAndFamily[type][family].color = COLORS[index % COLORS.length]
      typeAndFamily[type][family].percentage =
        (typeAndFamily[type][family].value / totalForType) * 100
    })
  })

  return typeAndFamily
})

// Get families for a given type as array
const getFamiliesForType = (type: string): ChartDataItem[] => {
  if (!expensesByTypeAndFamily.value[type]) return []

  return Object.values(expensesByTypeAndFamily.value[type])
  .sort((a, b) => b.value - a.value)
  .map((family, index) => ({
    ...family,
    color: COLORS[index % COLORS.length]
  }))
}

// Get categories for a specific type and family
const getCategoriesForTypeAndFamily = (type: string, family: string): ChartDataItem[] => {
  if (!props.expenses || props.expenses.length === 0) return []

  const filteredExpenses = props.expenses.filter(
    expense => expense.category_type === type &&
      (expense.category_family === family ||
        (!expense.category_family && family === 'General'))
  )

  const grouped = filteredExpenses.reduce((acc, expense) => {
    const category = expense.category_name || 'Uncategorized'

    if (!acc[category]) {
      acc[category] = {
        name: category,
        value: 0,
        count: 0
      }
    }

    acc[category].value += expense.amount
    acc[category].count += 1

    return acc
  }, {} as Record<string, ChartDataItem>)

  const totalForFamily = filteredExpenses.reduce((sum, exp) => sum + exp.amount, 0)

  return Object.values(grouped)
  .map((item, index) => ({
    ...item,
    color: COLORS[index % COLORS.length],
    percentage: (item.value / totalForFamily) * 100
  }))
  .sort((a, b) => b.value - a.value)
}

// Interface for monthly expense data
interface MonthlyExpenseData {
  month: string;
  total: number;
  [key: string]: number | string;
}

// Monthly expenses data
const expensesByMonth = computed<MonthlyExpenseData[]>(() => {
  if (!props.expenses || props.expenses.length === 0) return []

  // Create a map to store expenses by month
  const monthlyData: Record<string, Record<string, number>> = {}

  // Process each expense
  props.expenses.forEach(expense => {
    const date = new Date(expense.date)
    const monthYear = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
    const category = expense.category_type || 'Uncategorized'

    if (!monthlyData[monthYear]) {
      monthlyData[monthYear] = {}
    }

    if (!monthlyData[monthYear][category]) {
      monthlyData[monthYear][category] = 0
    }

    monthlyData[monthYear][category] += expense.amount
  })

  // Convert to array format for the chart
  const months = Object.keys(monthlyData).sort()

  return months.map(month => {
    const result: MonthlyExpenseData = { month, total: 0 }

    // Get all unique categories
    const categories = [...new Set(props.expenses.map(e => e.category_type).filter(Boolean))]

    // Add data for each category
    categories.forEach(category => {
      if (category) {
        result[category] = monthlyData[month][category] || 0
      }
    })

    // Calculate total for this month
    result.total = Object.values(monthlyData[month]).reduce((sum: number, val: number) => sum + val, 0)

    return result
  })
})

// Get all unique expense types
const expenseTypes = computed<string[]>(() => {
  return [...new Set(props.expenses
  .map(e => e.category_type || 'Uncategorized')
  .filter(Boolean) as string[])]
})

// When expenses change, set the default selected type
watch(() => props.expenses, () => {
  if (expensesByType.value.length > 0 && (!selectedType.value || !expensesByType.value.find(t => t.name === selectedType.value))) {
    selectedType.value = expensesByType.value[0].name
  }
}, { immediate: true })

// Chart.js chart options
const pieChartOptions = computed<ChartOptions<'pie'>>(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'bottom',
      labels: {
        boxWidth: 12,
        font: {
          size: 12
        }
      }
    },
    tooltip: {
      callbacks: {
        label: (context) => {
          const label = context.label || '';
          const value = context.raw as number;

          // Calculate percentage manually using dataset values instead of relying on chart meta
          const dataset = context.chart.data.datasets[0];
          const total = (dataset.data as number[]).reduce((sum, val) => sum + (val || 0), 0);
          const percentage = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0';

          return `${label}: ${formatCurrency(value)} (${percentage}%)`;
        }
      }
    }
  }
}))

const barChartOptions = computed<ChartOptions<'bar'>>(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'bottom',
      labels: {
        boxWidth: 12,
        font: {
          size: 12
        }
      }
    },
    tooltip: {
      callbacks: {
        label: (context) => {
          const label = context.dataset.label || '';
          const value = context.raw as number;
          return `${label}: ${formatCurrency(value)}`;
        }
      }
    }
  },
  scales: {
    x: {
      ticks: {
        maxRotation: 45,
        minRotation: 45
      }
    },
    y: {
      beginAtZero: true,
      ticks: {
        callback: (value) => {
          return formatCurrency(value as number);
        }
      }
    }
  }
}))

const stackedBarChartOptions = computed<ChartOptions<'bar'>>(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'bottom',
      labels: {
        boxWidth: 12,
        font: {
          size: 12
        }
      }
    },
    tooltip: {
      callbacks: {
        label: (context) => {
          const label = context.dataset.label || '';
          const value = context.raw as number;
          return `${label}: ${formatCurrency(value)}`;
        }
      }
    }
  },
  scales: {
    x: {
      stacked: true,
      ticks: {
        maxRotation: 45,
        minRotation: 45
      }
    },
    y: {
      stacked: true,
      beginAtZero: true,
      ticks: {
        callback: (value) => {
          return formatCurrency(value as number);
        }
      }
    }
  }
}))

// Chart data for Expenses by Type (Pie Chart)
const typesPieChartData = computed<ChartData<'pie'>>(() => {
  return {
    labels: expensesByType.value.map(item => item.name),
    datasets: [
      {
        data: expensesByType.value.map(item => item.value),
        backgroundColor: expensesByType.value.map(item => item.color),
        hoverBackgroundColor: expensesByType.value.map(item => item.color),
        borderWidth: 1
      }
    ]
  }
})

// Chart data for Expenses by Type (Bar Chart)
const typesBarChartData = computed<ChartData<'bar'>>(() => {
  return {
    labels: expensesByType.value.map(item => item.name),
    datasets: [
      {
        label: 'Amount',
        data: expensesByType.value.map(item => item.value),
        backgroundColor: expensesByType.value.map(item => item.color),
        borderWidth: 1
      }
    ]
  }
})

// Function to generate pie chart data for a specific type and family
const getFamilyPieChartData = (type: string): ChartData<'pie'> => {
  const families = getFamiliesForType(type)
  return {
    labels: families.map(item => item.name),
    datasets: [
      {
        data: families.map(item => item.value),
        backgroundColor: families.map(item => item.color),
        hoverBackgroundColor: families.map(item => item.color),
        borderWidth: 1
      }
    ]
  }
}

// Function to generate bar chart data for a specific type and family
const getFamilyBarChartData = (type: string): ChartData<'bar'> => {
  const families = getFamiliesForType(type)
  return {
    labels: families.map(item => item.name),
    datasets: [
      {
        label: 'Amount',
        data: families.map(item => item.value),
        backgroundColor: families.map(item => item.color),
        borderWidth: 1
      }
    ]
  }
}

// Function to generate pie chart data for categories
const getCategoryPieChartData = (type: string, family: string): ChartData<'pie'> => {
  const categories = getCategoriesForTypeAndFamily(type, family)
  return {
    labels: categories.map(item => item.name),
    datasets: [
      {
        data: categories.map(item => item.value),
        backgroundColor: categories.map(item => item.color),
        hoverBackgroundColor: categories.map(item => item.color),
        borderWidth: 1
      }
    ]
  }
}

// Function to generate bar chart data for categories
const getCategoryBarChartData = (type: string, family: string): ChartData<'bar'> => {
  const categories = getCategoriesForTypeAndFamily(type, family)
  return {
    labels: categories.map(item => item.name),
    datasets: [
      {
        label: 'Amount',
        data: categories.map(item => item.value),
        backgroundColor: categories.map(item => item.color),
        borderWidth: 1
      }
    ]
  }
}

// Generate monthly bar chart data (stacked)
const monthlyBarChartData = computed<ChartData<'bar'>>(() => {
  if (expensesByMonth.value.length === 0) return { labels: [], datasets: [] }

  const months = expensesByMonth.value
  const labels = months.map(month => month.month)

  // Get all category types used in any month
  const allTypes = expenseTypes.value

  // Create a dataset for each type
  const datasets = allTypes.map((type, index) => {
    return {
      label: type,
      data: months.map(month => month[type] as number || 0),
      backgroundColor: COLORS[index % COLORS.length],
      borderWidth: 1
    }
  })

  return {
    labels,
    datasets
  }
})

// Helper function to get chart ID for refs
const getChartRefId = (type: string, family?: string): string => {
  if (family) {
    return `${type.replace(/\s+/g, '-')}-${family.replace(/\s+/g, '-')}`
  }
  return `${type.replace(/\s+/g, '-')}`
}

// Helper function to capture chart canvas
const captureChart = (chartRef: any): HTMLCanvasElement | null => {
  if (!chartRef?.$el) return null

  try {
    // Get the chart canvas element
    const chartCanvas = chartRef.$el.querySelector('canvas')
    if (!chartCanvas) return null

    return chartCanvas
  } catch (error) {
    console.error('Error capturing chart', error)
    return null
  }
}

// References for the print sections
const printSectionRef = ref<HTMLElement | null>(null)

/**
 * Print the currently visible charts
 * Uses the browser's built-in print functionality
 */
const printVisibleCharts = () => {
  // Create a new window for printing
  const printWindow = window.open('', '_blank')
  if (!printWindow) {
    alert('Please allow pop-ups to print charts')
    return
  }

  // Write the print document HTML
  printWindow.document.write(`
    <html>
      <head>
        <title>Expense Charts</title>
        <style>
          body { font-family: Arial, sans-serif; padding: 20px; }
          .chart-container { margin-bottom: 40px; page-break-after: auto; }
          .chart-title { font-size: 18px; font-weight: bold; margin-bottom: 10px; }
          table { width: 100%; border-collapse: collapse; margin-top: 20px; }
          th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
          th { background-color: #f2f2f2; }
          .chart-image { max-width: 100%; height: auto; }
          @media print {
            .page-break { page-break-before: always; }
          }
        </style>
      </head>
      <body>
        <h1>Expense Charts</h1>
  `)

  // Capture the main type chart
  if (typeChartRef.value?.$el) {
    const typeChart = captureChart(typeChartRef.value)
    if (typeChart) {
      printWindow.document.write(`
        <div class="chart-container">
          <div class="chart-title">Expenses by Type</div>
          <img class="chart-image" src="${typeChart.toDataURL('image/png')}" />

          <table>
            <tr>
              <th>Category</th>
              <th>Amount</th>
              <th>Percentage</th>
            </tr>
      `)

      expensesByType.value.forEach(item => {
        printWindow.document.write(`
          <tr>
            <td>${item.name}</td>
            <td>${formatCurrency(item.value)}</td>
            <td>${Math.round(item.percentage || 0)}%</td>
          </tr>
        `)
      })

      printWindow.document.write('</table></div>')
    }
  }

  // Capture the monthly chart
  if (monthlyChartRef.value?.$el) {
    const monthlyChart = captureChart(monthlyChartRef.value)
    if (monthlyChart) {
      printWindow.document.write(`
        <div class="chart-container page-break">
          <div class="chart-title">Monthly Expense Trends</div>
          <img class="chart-image" src="${monthlyChart.toDataURL('image/png')}" />

          <table>
            <tr>
              <th>Month</th>
              <th>Total Amount</th>
            </tr>
      `)

      expensesByMonth.value.forEach(month => {
        printWindow.document.write(`
          <tr>
            <td>${month.month}</td>
            <td>${formatCurrency(month.total)}</td>
          </tr>
        `)
      })

      printWindow.document.write('</table></div>')
    }
  }

  // Capture expanded family charts (for the first type only, to keep it manageable)
  if (expensesByType.value.length > 0) {
    const firstType = expensesByType.value[0]
    const typeId = getChartRefId(firstType.name)

    if (categoryChartRefs.value[typeId]?.$el) {
      const familyChart = captureChart(categoryChartRefs.value[typeId])
      if (familyChart) {
        printWindow.document.write(`
          <div class="chart-container page-break">
            <div class="chart-title">Families in ${firstType.name}</div>
            <img class="chart-image" src="${familyChart.toDataURL('image/png')}" />

            <table>
              <tr>
                <th>Family</th>
                <th>Amount</th>
                <th>Percentage</th>
              </tr>
        `)

        getFamiliesForType(firstType.name).forEach(family => {
          printWindow.document.write(`
            <tr>
              <td>${family.name}</td>
              <td>${formatCurrency(family.value)}</td>
              <td>${Math.round(family.percentage || 0)}%</td>
            </tr>
          `)
        })

        printWindow.document.write('</table></div>')
      }
    }
  }

  // Finish the HTML document
  printWindow.document.write(`
      </body>
    </html>
  `)

  // Wait for images to load before printing
  printWindow.document.close()
  printWindow.onload = () => {
    setTimeout(() => {
      printWindow.print()
      // Close the window after printing (or if printing is canceled)
      printWindow.onafterprint = () => {
        printWindow.close()
      }
    }, 500)
  }
}

/**
 * Print a comprehensive expense report
 * Including summary, types, monthly trends, and category breakdowns
 */
const printFullReport = () => {
  // Create a new window for printing
  const printWindow = window.open('', '_blank')
  if (!printWindow) {
    alert('Please allow pop-ups to print expense report')
    return
  }

  // Calculate totals
  const totalAmount = props.expenses.reduce((sum, expense) => sum + expense.amount, 0)
  const averageAmount = totalAmount / (props.expenses.length || 1)

  // Write the print document HTML
  printWindow.document.write(`
    <html>
      <head>
        <title>Expense Analysis Report</title>
        <style>
          body { font-family: Arial, sans-serif; padding: 20px; }
          .section { margin-bottom: 40px; page-break-after: auto; }
          h1 { margin-bottom: 5px; }
          h2 { margin-top: 30px; margin-bottom: 15px; color: #444; }
          .subtitle { font-size: 14px; color: #777; margin-bottom: 30px; }
          table { width: 100%; border-collapse: collapse; margin-top: 20px; }
          th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
          th { background-color: #f2f2f2; }
          .summary-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; margin-top: 20px; }
          .summary-card { background-color: #f9f9f9; padding: 15px; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
          .summary-card-title { font-size: 14px; color: #777; margin-bottom: 5px; }
          .summary-card-value { font-size: 22px; font-weight: bold; }
          .chart-image { max-width: 100%; height: auto; }
          @media print {
            .page-break { page-break-before: always; }
          }
        </style>
      </head>
      <body>
        <h1>Expense Analysis Report</h1>
        <div class="subtitle">All time</div>

        <section class="section">
          <h2>Summary</h2>
          <div class="summary-grid">
            <div class="summary-card">
              <div class="summary-card-title">Total Expenses</div>
              <div class="summary-card-value">${formatCurrency(totalAmount)}</div>
            </div>
            <div class="summary-card">
              <div class="summary-card-title">Number of Expenses</div>
              <div class="summary-card-value">${props.expenses.length}</div>
            </div>
            <div class="summary-card">
              <div class="summary-card-title">Average Expense</div>
              <div class="summary-card-value">${formatCurrency(averageAmount)}</div>
            </div>
          </div>
        </section>
  `)

  // Expenses by Type section
  if (typeChartRef.value?.$el) {
    const typeChart = captureChart(typeChartRef.value)
    if (typeChart) {
      printWindow.document.write(`
        <section class="section">
          <h2>Expenses by Type</h2>
          <img class="chart-image" src="${typeChart.toDataURL('image/png')}" />

          <table>
            <tr>
              <th>Type</th>
              <th>Amount</th>
              <th>Percentage</th>
              <th>Count</th>
            </tr>
      `)

      expensesByType.value.forEach(item => {
        printWindow.document.write(`
          <tr>
            <td>${item.name}</td>
            <td>${formatCurrency(item.value)}</td>
            <td>${Math.round(item.percentage || 0)}%</td>
            <td>${item.count}</td>
          </tr>
        `)
      })

      printWindow.document.write('</table></section>')
    }
  }

  // Monthly Trends section
  if (monthlyChartRef.value?.$el) {
    const monthlyChart = captureChart(monthlyChartRef.value)
    if (monthlyChart) {
      printWindow.document.write(`
        <section class="section page-break">
          <h2>Monthly Expense Trends</h2>
          <img class="chart-image" src="${monthlyChart.toDataURL('image/png')}" />

          <table>
            <tr>
              <th>Month</th>
              <th>Total Amount</th>
            </tr>
      `)

      expensesByMonth.value.forEach(month => {
        printWindow.document.write(`
          <tr>
            <td>${month.month}</td>
            <td>${formatCurrency(month.total)}</td>
          </tr>
        `)
      })

      printWindow.document.write('</table></section>')
    }
  }

  // Detailed Month Breakdown with Categories
  printWindow.document.write(`
    <section class="section page-break">
      <h2>Detailed Monthly Breakdown by Category</h2>
      <table>
        <tr>
          <th>Month</th>
  `)

  // Add headers for each expense type
  expenseTypes.value.forEach(type => {
    printWindow.document.write(`<th>${type}</th>`)
  })
  printWindow.document.write(`<th>Total</th></tr>`)

  // Add data for each month
  expensesByMonth.value.forEach(month => {
    printWindow.document.write(`<tr><td>${month.month}</td>`)

    expenseTypes.value.forEach(type => {
      const value = month[type] as number || 0
      printWindow.document.write(`<td>${formatCurrency(value)}</td>`)
    })

    printWindow.document.write(`<td>${formatCurrency(month.total)}</td></tr>`)
  })

  printWindow.document.write('</table></section>')

  // Type Breakdown section (for each type)
  printWindow.document.write('<section class="section page-break"><h2>Category Type Breakdown</h2>')

  expensesByType.value.forEach((type, index) => {
    // Add page break between types except for the first one
    if (index > 0) {
      printWindow.document.write('<div class="page-break"></div>')
    }

    printWindow.document.write(`
      <div style="margin-top: 30px; margin-bottom: 20px;">
        <h3>${type.name} - ${formatCurrency(type.value)}</h3>
        <table>
          <tr>
            <th>Family</th>
            <th>Amount</th>
            <th>Percentage</th>
          </tr>
    `)

    const families = getFamiliesForType(type.name)
    families.forEach(family => {
      printWindow.document.write(`
        <tr>
          <td>${family.name}</td>
          <td>${formatCurrency(family.value)}</td>
          <td>${Math.round(family.percentage || 0)}%</td>
        </tr>
      `)
    })

    printWindow.document.write('</table></div>')

    // If there are families, add a breakdown of the first family's categories
    if (families.length > 0) {
      const firstFamily = families[0]
      const categories = getCategoriesForTypeAndFamily(type.name, firstFamily.name)

      if (categories.length > 0) {
        printWindow.document.write(`
          <div style="margin-left: 20px; margin-top: 15px;">
            <h4>Categories in ${firstFamily.name}</h4>
            <table>
              <tr>
                <th>Category</th>
                <th>Amount</th>
                <th>Percentage</th>
              </tr>
        `)

        categories.forEach(category => {
          printWindow.document.write(`
            <tr>
              <td>${category.name}</td>
              <td>${formatCurrency(category.value)}</td>
              <td>${Math.round(category.percentage || 0)}%</td>
            </tr>
          `)
        })

        printWindow.document.write('</table></div>')
      }
    }
  })

  printWindow.document.write('</section>')

  // Finish the HTML document
  printWindow.document.write(`
      </body>
    </html>
  `)

  // Wait for images to load before printing
  printWindow.document.close()
  printWindow.onload = () => {
    setTimeout(() => {
      printWindow.print()
      // Close the window after printing (or if printing is canceled)
      printWindow.onafterprint = () => {
        printWindow.close()
      }
    }, 500)
  }
}
</script>

<template>
  <div class="expense-charts">
    <template v-if="props.expenses.length === 0">
      <NEmpty description="No expenses found for the selected filters" />
    </template>

    <template v-else>
      <div class="export-buttons">
        <NSpace>
          <NTooltip>
            <template #trigger>
              <NButton type="primary" ghost size="small" @click="printVisibleCharts">
                Print Visible Charts
              </NButton>
            </template>
            Print all currently visible charts
          </NTooltip>

          <NTooltip>
            <template #trigger>
              <NButton type="primary" size="small" @click="printFullReport">
                Print Full Report
              </NButton>
            </template>
            Print a comprehensive report with summary data and all charts
          </NTooltip>
        </NSpace>
      </div>

      <!-- Section 1: Expenses by Type -->
      <NCard title="Expenses by Type" style="margin-bottom: 24px;">
        <NRadioGroup v-model:value="typeChartMode" class="mode-selector">
          <NRadio value="pie">Pie Chart</NRadio>
          <NRadio value="bar">Bar Chart</NRadio>
        </NRadioGroup>

        <div class="chart-container">
          <!-- Pie Chart View -->
          <template v-if="typeChartMode === 'pie' && expensesByType.length > 0">
            <div class="chart">
              <Pie
                ref="typeChartRef"
                :data="typesPieChartData"
                :options="pieChartOptions"
                :height="300"
              />
            </div>

            <!-- Custom Legend -->
            <div class="chart-legend">
              <div v-for="(item, index) in expensesByType" :key="'type-legend-' + index"
                   class="legend-item">
                <div class="legend-color" :style="{ backgroundColor: item.color }"></div>
                <div class="legend-text">
                  <div class="legend-name">{{ item.name }}</div>
                  <div class="legend-value">{{ formatCurrency(item.value) }}
                    ({{ Math.round(item.percentage || 0) }}%)
                  </div>
                </div>
              </div>
            </div>
          </template>

          <!-- Bar Chart View -->
          <template v-else-if="typeChartMode === 'bar' && expensesByType.length > 0">
            <div class="chart">
              <Bar
                ref="typeChartRef"
                :data="typesBarChartData"
                :options="barChartOptions"
                :height="300"
              />
            </div>

            <!-- Custom Legend -->
            <div class="chart-legend">
              <div v-for="(item, index) in expensesByType" :key="'type-bar-legend-' + index"
                   class="legend-item">
                <div class="legend-color" :style="{ backgroundColor: item.color }"></div>
                <div class="legend-text">
                  <div class="legend-name">{{ item.name }}</div>
                  <div class="legend-value">{{ formatCurrency(item.value) }}</div>
                </div>
              </div>
            </div>
          </template>

          <template v-else>
            <NEmpty description="Not enough data to display chart" />
          </template>
        </div>
      </NCard>

      <!-- Section 2: Monthly Trends -->
      <NCard title="Monthly Expense Trends" style="margin-bottom: 24px;">
        <NRadioGroup v-model:value="monthlyChartMode" class="mode-selector">
          <NRadio value="bar">Stacked Bar Chart</NRadio>
        </NRadioGroup>

        <div class="chart-container">
          <template v-if="expensesByMonth.length > 0">
            <div class="chart">
              <Bar
                ref="monthlyChartRef"
                :data="monthlyBarChartData"
                :options="stackedBarChartOptions"
                :height="300"
              />
            </div>

            <!-- Custom Legend -->
            <div class="chart-legend">
              <div v-for="(type, index) in expenseTypes" :key="'month-legend-' + index"
                   class="legend-item">
                <div class="legend-color"
                     :style="{ backgroundColor: COLORS[index % COLORS.length] }"></div>
                <div class="legend-text">
                  <div class="legend-name">{{ type }}</div>
                </div>
              </div>
            </div>
          </template>

          <template v-else>
            <NEmpty description="No monthly data available" />
          </template>
        </div>
      </NCard>

      <!-- Section 3: Expense Type Breakdown -->
      <NCard title="Expense Type Breakdown" style="margin-bottom: 24px;">
        <NRadioGroup v-model:value="categoryChartMode" class="mode-selector">
          <NRadio value="pie">Pie Chart</NRadio>
          <NRadio value="bar">Bar Chart</NRadio>
        </NRadioGroup>

        <NCollapse>
          <NCollapseItem
            v-for="type in expensesByType"
            :key="type.name"
            :title="type.name + ' - ' + formatCurrency(type.value)"
          >
            <!-- Families within this type section -->
            <div class="chart-container">
              <NCard title="Family Breakdown" size="small" style="margin-bottom: 16px;">
                <template v-if="getFamiliesForType(type.name).length > 0">
                  <!-- Pie Chart for Families -->
                  <div v-if="categoryChartMode === 'pie'" class="chart">
                    <Pie
                      :ref="el => categoryChartRefs[getChartRefId(type.name)] = el"
                      :data="getFamilyPieChartData(type.name)"
                      :options="pieChartOptions"
                      :height="300"
                    />

                    <!-- Custom Legend -->
                    <div class="chart-legend">
                      <div v-for="(item, index) in getFamiliesForType(type.name)"
                           :key="'family-legend-' + index" class="legend-item">
                        <div class="legend-color" :style="{ backgroundColor: item.color }"></div>
                        <div class="legend-text">
                          <div class="legend-name">{{ item.name }}</div>
                          <div class="legend-value">{{ formatCurrency(item.value) }}
                            ({{ Math.round(item.percentage || 0) }}%)
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>

                  <!-- Bar Chart for Families -->
                  <div v-else class="chart">
                    <Bar
                      :ref="el => categoryChartRefs[getChartRefId(type.name)] = el"
                      :data="getFamilyBarChartData(type.name)"
                      :options="barChartOptions"
                      :height="300"
                    />

                    <!-- Custom Legend -->
                    <div class="chart-legend">
                      <div v-for="(item, index) in getFamiliesForType(type.name)"
                           :key="'family-bar-legend-' + index" class="legend-item">
                        <div class="legend-color" :style="{ backgroundColor: item.color }"></div>
                        <div class="legend-text">
                          <div class="legend-name">{{ item.name }}</div>
                          <div class="legend-value">{{ formatCurrency(item.value) }}</div>
                        </div>
                      </div>
                    </div>
                  </div>
                </template>
                <template v-else>
                  <NEmpty description="No families found for this type" />
                </template>
              </NCard>

              <!-- Categories within each family collapsible sections -->
              <NCollapse>
                <NCollapseItem
                  v-for="family in getFamiliesForType(type.name)"
                  :key="family.name"
                  :title="family.name + ' - ' + formatCurrency(family.value)"
                >
                  <div class="chart-container">
                    <template
                      v-if="getCategoriesForTypeAndFamily(type.name, family.name).length > 0">
                      <!-- Pie Chart for Categories -->
                      <div v-if="categoryChartMode === 'pie'" class="chart">
                        <Pie
                          :ref="el => categoryChartRefs[getChartRefId(type.name, family.name)] = el"
                          :data="getCategoryPieChartData(type.name, family.name)"
                          :options="pieChartOptions"
                          :height="300"
                        />

                        <!-- Custom Legend -->
                        <div class="chart-legend">
                          <div
                            v-for="(item, index) in getCategoriesForTypeAndFamily(type.name, family.name)"
                            :key="'category-legend-' + index" class="legend-item">
                            <div class="legend-color"
                                 :style="{ backgroundColor: item.color }"></div>
                            <div class="legend-text">
                              <div class="legend-name">{{ item.name }}</div>
                              <div class="legend-value">{{ formatCurrency(item.value) }}
                                ({{ Math.round(item.percentage || 0) }}%)
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>

                      <!-- Bar Chart for Categories -->
                      <div v-else class="chart">
                        <Bar
                          :ref="el => categoryChartRefs[getChartRefId(type.name, family.name)] = el"
                          :data="getCategoryBarChartData(type.name, family.name)"
                          :options="barChartOptions"
                          :height="300"
                        />

                        <!-- Custom Legend -->
                        <div class="chart-legend">
                          <div
                            v-for="(item, index) in getCategoriesForTypeAndFamily(type.name, family.name)"
                            :key="'category-bar-legend-' + index" class="legend-item">
                            <div class="legend-color"
                                 :style="{ backgroundColor: item.color }"></div>
                            <div class="legend-text">
                              <div class="legend-name">{{ item.name }}</div>
                              <div class="legend-value">{{ formatCurrency(item.value) }}</div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </template>
                    <template v-else>
                      <NEmpty description="No categories found for this family" />
                    </template>
                  </div>
                </NCollapseItem>
              </NCollapse>
            </div>
          </NCollapseItem>
        </NCollapse>
      </NCard>
    </template>
  </div>
</template>
<style scoped>
.expense-charts {
  margin-bottom: 20px;
}

.export-buttons {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}

.chart-container {
  min-height: 300px;
  margin: 10px 0;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.mode-selector {
  margin-bottom: 16px;
  display: flex;
  justify-content: center;
}

.chart {
  width: 100%;
  max-width: 600px;
  height: 300px;
  margin-bottom: 20px;
}

.chart-legend {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 12px;
  max-width: 100%;
  padding: 10px;
}

.legend-item {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.legend-color {
  width: 16px;
  height: 16px;
  border-radius: 3px;
  margin-right: 8px;
}

.legend-text {
  display: flex;
  flex-direction: column;
}

.legend-name {
  font-size: 14px;
  font-weight: 500;
}

.legend-value {
  font-size: 12px;
  opacity: 0.8;
}

@media (max-width: 768px) {
  .chart-legend {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
