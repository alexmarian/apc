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
import { exportChartToPdf, exportFullReportToPdf } from '@/utils/pdfExport'
import html2canvas from 'html2canvas'
import jsPDF from 'jspdf'

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
  return [...new Set(props.expenses.map(e => e.category_type).filter(Boolean))]
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
          const label = context.label || ''
          const value = context.raw as number
          const percentage = ((value / context.chart.getDatasetMeta(0).total) * 100).toFixed(1)
          return `${label}: ${formatCurrency(value)} (${percentage}%)`
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
          const label = context.dataset.label || ''
          const value = context.raw as number
          return `${label}: ${formatCurrency(value)}`
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
          return formatCurrency(value as number)
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
          const label = context.dataset.label || ''
          const value = context.raw as number
          return `${label}: ${formatCurrency(value)}`
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
          return formatCurrency(value as number)
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

// Export helpers
const captureChart = async (chartRef: any): Promise<HTMLCanvasElement | null> => {
  if (!chartRef) return null

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

// Export the current chart to PDF
const exportCurrentChart = async () => {
  // Collect all visible charts
  const charts = []

  // Main type chart
  let typeChart
  if (typeChartMode.value === 'pie') {
    typeChart = await captureChart(typeChartRef.value)
    if (typeChart) {
      charts.push({
        title: 'Expenses by Type',
        element: typeChart,
        data: expensesByType.value
      })
    }
  } else {
    typeChart = await captureChart(typeChartRef.value)
    if (typeChart) {
      charts.push({
        title: 'Expenses by Type',
        element: typeChart,
        data: expensesByType.value
      })
    }
  }

  // Monthly chart
  const monthlyChart = await captureChart(monthlyChartRef.value)
  if (monthlyChart) {
    charts.push({
      title: 'Monthly Expenses',
      element: monthlyChart,
      data: expensesByMonth.value.map(month => ({
        name: month.month,
        value: month.total
      }))
    })
  }

  // Find all expanded type and family charts
  const expandedTypes = expensesByType.value.slice(0, 3) // Assume first 3 types are expanded for demo

  for (const type of expandedTypes) {
    const typeId = getChartRefId(type.name)
    if (categoryChartRefs.value[typeId]) {
      const familyChart = await captureChart(categoryChartRefs.value[typeId])
      if (familyChart) {
        charts.push({
          title: `Families in ${type.name}`,
          element: familyChart,
          data: getFamiliesForType(type.name)
        })
      }

      // Add some expanded families for this type (for demo)
      const families = getFamiliesForType(type.name).slice(0, 2)
      for (const family of families) {
        const familyId = getChartRefId(type.name, family.name)
        if (categoryChartRefs.value[familyId]) {
          const categoryChart = await captureChart(categoryChartRefs.value[familyId])
          if (categoryChart) {
            charts.push({
              title: `Categories in ${family.name}`,
              element: categoryChart,
              data: getCategoriesForTypeAndFamily(type.name, family.name)
            })
          }
        }
      }
    }
  }

  // Export each chart
  charts.forEach((chart, index) => {
    if (chart.element) {
      // Small delay between exports to prevent browser issues
      setTimeout(() => {
        exportChartToPdf(chart.title, chart.element, chart.data)
      }, index * 500)
    }
  })
}

// Export a full report with all charts and data
const exportFullReport = async () => {
  // Use the first available chart
  let chartElement = null

  if (typeChartRef.value) {
    chartElement = await captureChart(typeChartRef.value)
  }

  // Prepare summary data
  const totalAmount = props.expenses.reduce((sum, expense) => sum + expense.amount, 0)
  const summaryData = [
    {
      label: 'Total Expenses',
      value: formatCurrency(totalAmount)
    },
    {
      label: 'Number of Expenses',
      value: props.expenses.length.toString()
    },
    {
      label: 'Average Expense',
      value: formatCurrency(totalAmount / (props.expenses.length || 1))
    }
  ]

  // Prepare breakdown data
  const breakdownData = {
    types: expensesByType.value,
    months: expensesByMonth.value.map(month => ({
      month: month.month,
      value: month.total
    }))
  }

  // Export the full report
  if (chartElement) {
    exportFullReportToPdf(
      'Expense Analysis Report',
      summaryData,
      chartElement,
      breakdownData,
      'All time'
    )
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
              <NButton type="primary" ghost size="small" @click="exportCurrentChart">
                Export Visible Charts
              </NButton>
            </template>
            Export all currently visible charts as PDF
          </NTooltip>

          <NTooltip>
            <template #trigger>
              <NButton type="primary" size="small" @click="exportFullReport">
                Export Full Report
              </NButton>
            </template>
            Export a comprehensive report with summary data and all charts
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
