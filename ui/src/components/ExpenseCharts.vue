
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { NCard, NTabs, NTabPane, NDivider, NSelect, NEmpty, NRadioGroup, NRadio, NButton, NTooltip, NSpace } from 'naive-ui'
import type { Expense } from '@/types/api'
import { formatCurrency } from '@/utils/formatters'
import { exportChartToPdf, exportFullReportToPdf } from '@/utils/pdfExport'

// Props
const props = defineProps<{
  expenses: Expense[]
}>()

// Tab state
const activeTab = ref('byType')

// Selected type for category breakdown
const selectedType = ref<string | null>(null)

// Chart display mode
const chartMode = ref('pie')

// Generate a color palette for the charts
const COLORS = [
  '#3366FF', '#FF6633', '#33CC99', '#FFCC33', '#FF33CC',
  '#33CCFF', '#CC99FF', '#99CC33', '#FF9966', '#6699FF'
]

// Computed data for the charts
const expensesByType = computed(() => {
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
  }, {} as Record<string, { name: string, value: number, count: number }>)

  return Object.values(grouped).map((item, index) => ({
    ...item,
    color: COLORS[index % COLORS.length],
    percentage: (item.value / props.expenses.reduce((sum, exp) => sum + exp.amount, 0)) * 100
  }))
})

const availableTypes = computed(() => {
  return expensesByType.value.map(type => ({
    label: `${type.name} (${formatCurrency(type.value)})`,
    value: type.name
  }))
})

const expensesByCategoryForType = computed(() => {
  if (!props.expenses || props.expenses.length === 0 || !selectedType.value) return []

  const filteredExpenses = props.expenses.filter(
    expense => expense.category_type === selectedType.value
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
  }, {} as Record<string, { name: string, value: number, count: number }>)

  const totalForType = filteredExpenses.reduce((sum, exp) => sum + exp.amount, 0)

  return Object.values(grouped).map((item, index) => ({
    ...item,
    color: COLORS[index % COLORS.length],
    percentage: (item.value / totalForType) * 100
  }))
})

// Monthly expenses data
const expensesByMonth = computed(() => {
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
    const result: any = { month }

    // Get all unique categories
    const categories = [...new Set(props.expenses.map(e => e.category_type).filter(Boolean))]

    // Add data for each category
    categories.forEach(category => {
      if (category) {
        result[category] = monthlyData[month][category] || 0
      }
    })

    // Calculate total for this month
    result.total = Object.values(monthlyData[month]).reduce((sum: number, val: any) => sum + val, 0)

    return result
  })
})

// Get all unique expense types
const expenseTypes = computed(() => {
  return [...new Set(props.expenses.map(e => e.category_type).filter(Boolean))]
})

// When expenses change, set the default selected type
watch(() => props.expenses, () => {
  if (expensesByType.value.length > 0 && (!selectedType.value || !expensesByType.value.find(t => t.name === selectedType.value))) {
    selectedType.value = expensesByType.value[0].name
  }
}, { immediate: true })

// Handle type selection change
const handleTypeChange = (value: string) => {
  selectedType.value = value
}

// Refs for SVG elements to export
const typesPieChartRef = ref<SVGElement | null>(null)
const categoriesPieChartRef = ref<SVGElement | null>(null)
const typesBarChartRef = ref<SVGElement | null>(null)
const categoriesBarChartRef = ref<SVGElement | null>(null)
const monthlyChartRef = ref<SVGElement | null>(null)

// Export the current chart to PDF
const exportCurrentChart = () => {
  let chartElement: HTMLElement | null = null
  let chartData: any[] = []
  let chartTitle = ''

  // Determine which chart to export based on active tab and chart mode
  if (activeTab.value === 'byType') {
    chartTitle = 'Expenses by Type'
    chartData = expensesByType.value

    if (chartMode.value === 'pie') {
      chartElement = typesPieChartRef.value?.$el || typesPieChartRef.value
    } else {
      chartElement = typesBarChartRef.value?.$el || typesBarChartRef.value
    }
  } else if (activeTab.value === 'byCategory' && selectedType.value) {
    chartTitle = `Categories for ${selectedType.value}`
    chartData = expensesByCategoryForType.value

    if (chartMode.value === 'pie') {
      chartElement = categoriesPieChartRef.value?.$el || categoriesPieChartRef.value
    } else {
      chartElement = categoriesBarChartRef.value?.$el || categoriesBarChartRef.value
    }
  } else if (activeTab.value === 'byMonth') {
    chartTitle = 'Monthly Expenses'
    chartElement = monthlyChartRef.value?.$el || monthlyChartRef.value
    // For monthly chart, we need to format data differently
    chartData = expensesByMonth.value.map(month => ({
      name: month.month,
      value: month.total,
    }))
  }

  // If we couldn't find the chart element by ref, try to get it by selection
  if (!chartElement) {
    // First, try to find container
    const container = document.querySelector('.chart-container') as HTMLElement
    if (container) {
      // Then look for SVG within the container
      const svg = container.querySelector('svg')
      if (svg) {
        chartElement = svg.parentElement || container
      } else {
        chartElement = container
      }
    }
  }

  // Export the chart to PDF
  exportChartToPdf(
    chartTitle,
    chartElement,
    chartData
  )
}

// Export a full report with all charts and data
const exportFullReport = () => {
  // Get chart element based on active tab
  let chartElement: HTMLElement | null = null

  if (activeTab.value === 'byType') {
    chartElement = chartMode.value === 'pie'
      ? (typesPieChartRef.value?.$el || typesPieChartRef.value || document.querySelector('.pie-chart'))
      : (typesBarChartRef.value?.$el || typesBarChartRef.value || document.querySelector('.bar-chart'))
  } else if (activeTab.value === 'byCategory' && selectedType.value) {
    chartElement = chartMode.value === 'pie'
      ? (categoriesPieChartRef.value?.$el || categoriesPieChartRef.value || document.querySelector('.pie-chart'))
      : (categoriesBarChartRef.value?.$el || categoriesBarChartRef.value || document.querySelector('.bar-chart'))
  } else if (activeTab.value === 'byMonth') {
    chartElement = monthlyChartRef.value?.$el || monthlyChartRef.value || document.querySelector('.stacked-bar-chart')
  }

  // If we couldn't find the chart element by ref, try to get the whole chart container
  if (!chartElement) {
    chartElement = document.querySelector('.chart-container') as HTMLElement
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
      value: formatCurrency(totalAmount / props.expenses.length)
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
  exportFullReportToPdf(
    'Expense Analysis Report',
    summaryData,
    chartElement,
    breakdownData,
    'All time'
  )
}

// Generate data for SVG pie chart
const generatePieChartSVG = (data: { name: string; value: number; color: string; percentage: number }[]) => {
  const totalValue = data.reduce((sum, item) => sum + item.value, 0)
  let startAngle = 0

  return data.map(item => {
    const percentage = (item.value / totalValue) * 100
    const angle = (percentage / 100) * 360
    const endAngle = startAngle + angle

    // Calculate path for pie slice
    const x1 = 100 + 80 * Math.cos((startAngle - 90) * (Math.PI / 180))
    const y1 = 100 + 80 * Math.sin((startAngle - 90) * (Math.PI / 180))
    const x2 = 100 + 80 * Math.cos((endAngle - 90) * (Math.PI / 180))
    const y2 = 100 + 80 * Math.sin((endAngle - 90) * (Math.PI / 180))

    // Create SVG path for slice
    const largeArcFlag = angle > 180 ? 1 : 0
    const pathData = `M 100 100 L ${x1} ${y1} A 80 80 0 ${largeArcFlag} 1 ${x2} ${y2} Z`

    // Calculate position for label
    const labelAngle = startAngle + angle / 2
    const labelRadius = 60
    const labelX = 100 + labelRadius * Math.cos((labelAngle - 90) * (Math.PI / 180))
    const labelY = 100 + labelRadius * Math.sin((labelAngle - 90) * (Math.PI / 180))

    // Save current angle as start for next slice
    startAngle = endAngle

    return {
      pathData,
      color: item.color,
      name: item.name,
      value: item.value,
      percentage,
      labelX,
      labelY
    }
  })
}

const typesPieChartData = computed(() => generatePieChartSVG(expensesByType.value))
const categoriesPieChartData = computed(() => generatePieChartSVG(expensesByCategoryForType.value))

// Function to generate bar chart data
const generateBarChartData = (data: any[], valueKey: string = 'value', nameKey: string = 'name') => {
  // Sort data by value in descending order
  const sortedData = [...data].sort((a, b) => b[valueKey] - a[valueKey])

  // Get the maximum value for scaling
  const maxValue = sortedData.length > 0 ? sortedData[0][valueKey] : 0

  // Number of bars
  const count = sortedData.length

  // Width per bar (80% of total width divided by count)
  const barWidth = count > 0 ? (80 / count) : 0

  // Space between bars (20% of total width divided by count+1)
  const barSpacing = count > 0 ? (20 / (count + 1)) : 0

  return sortedData.map((item, index) => {
    // Calculate height (proportional to max value)
    const barHeight = (item[valueKey] / maxValue) * 80

    // X position (centered, with spacing)
    const x = 10 + barSpacing * (index + 1) + barWidth * index

    // Y position (from bottom)
    const y = 90 - barHeight

    return {
      x,
      y,
      width: barWidth,
      height: barHeight,
      color: item.color,
      name: item[nameKey],
      value: item[valueKey],
      percentage: item.percentage
    }
  })
}

const typesBarChartData = computed(() => generateBarChartData(expensesByType.value))
const categoriesBarChartData = computed(() => generateBarChartData(expensesByCategoryForType.value))

// Generate monthly bar chart data
const monthlyBarChartData = computed(() => {
  if (expensesByMonth.value.length === 0) return []

  const months = expensesByMonth.value
  const count = months.length

  // Width per month group (80% of total width divided by count)
  const groupWidth = count > 0 ? (80 / count) : 0

  // Space between groups (20% of total width divided by count+1)
  const groupSpacing = count > 0 ? (20 / (count + 1)) : 0

  // Get maximum value across all months for scaling
  const maxValue = Math.max(...months.map(month => month.total))

  return months.map((month, monthIndex) => {
    // X position for this month group
    const groupX = 10 + groupSpacing * (monthIndex + 1) + groupWidth * monthIndex

    // Get all expense types for this month
    const typesInMonth = Object.keys(month).filter(key => key !== 'month' && key !== 'total')

    // Cumulative height for stacking bars
    let cumulativeHeight = 0

    // Generate bars for each expense type
    const bars = typesInMonth.map((type) => {
      const value = month[type]
      const barHeight = (value / maxValue) * 80

      // All bars in a month have same width
      const barWidth = groupWidth

      // X position is the group X
      const x = groupX

      // Y position from bottom, accounting for already stacked bars
      const y = 90 - cumulativeHeight - barHeight

      // Add this bar's height to cumulative
      cumulativeHeight += barHeight

      // Find color for this type
      const typeIndex = expenseTypes.value.indexOf(type)
      const color = COLORS[typeIndex % COLORS.length]

      return {
        x,
        y,
        width: barWidth,
        height: barHeight,
        color,
        name: type,
        value,
        month: month.month
      }
    })

    return {
      month: month.month,
      x: groupX,
      width: groupWidth,
      total: month.total,
      bars
    }
  })
})
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
                Export Current Chart
              </NButton>
            </template>
            Export the current chart as PDF
          </NTooltip>

          <NTooltip>
            <template #trigger>
              <NButton type="primary" size="small" @click="exportFullReport">
                Export Full Report
              </NButton>
            </template>
            Export a comprehensive report with all data and charts
          </NTooltip>
        </NSpace>
      </div>

      <NTabs v-model:value="activeTab" type="line" animated>
        <!-- Expense by Type Tab -->
        <NTabPane name="byType" tab="By Expense Type">
          <NRadioGroup v-model:value="chartMode" class="mode-selector">
            <NRadio value="pie">Pie Chart</NRadio>
            <NRadio value="bar">Bar Chart</NRadio>
          </NRadioGroup>

          <div class="chart-container">
            <!-- Pie Chart View -->
            <template v-if="chartMode === 'pie' && typesPieChartData.length > 0">
              <div class="pie-chart">
                <svg ref="typesPieChartRef" viewBox="0 0 200 200" width="100%" height="100%">
                  <!-- Render pie slices -->
                  <g v-for="(slice, index) in typesPieChartData" :key="'type-pie-' + index">
                    <path :d="slice.pathData" :fill="slice.color" stroke="#fff" stroke-width="1" />
                    <!-- Add percentage labels in center of each slice -->
                    <text
                      v-if="slice.percentage > 5"
                      :x="slice.labelX"
                      :y="slice.labelY"
                      fill="#fff"
                      font-size="8"
                      text-anchor="middle"
                      alignment-baseline="middle"
                    >
                      {{ Math.round(slice.percentage) }}%
                    </text>
                  </g>
                  <!-- Inner circle for donut chart -->
                  <circle cx="100" cy="100" r="40" fill="var(--background-color)" />
                </svg>

                <!-- Chart legend -->
                <div class="chart-legend">
                  <div v-for="(item, index) in expensesByType" :key="'type-legend-' + index" class="legend-item">
                    <div class="legend-color" :style="{ backgroundColor: item.color }"></div>
                    <div class="legend-text">
                      <div class="legend-name">{{ item.name }}</div>
                      <div class="legend-value">{{ formatCurrency(item.value) }} ({{ Math.round(item.percentage) }}%)</div>
                    </div>
                  </div>
                </div>
              </div>
            </template>

            <!-- Bar Chart View -->
            <template v-else-if="chartMode === 'bar' && typesBarChartData.length > 0">
              <div class="bar-chart">
                <svg ref="typesBarChartRef" viewBox="0 0 100 100" width="100%" height="100%">
                  <!-- Y-axis and grid lines -->
                  <line x1="10" y1="10" x2="10" y2="90" stroke="var(--border-color)" stroke-width="0.5" />
                  <line x1="10" y1="90" x2="90" y2="90" stroke="var(--border-color)" stroke-width="0.5" />

                  <!-- Render bars -->
                  <g v-for="(bar, index) in typesBarChartData" :key="'type-bar-' + index">
                    <rect
                      :x="bar.x"
                      :y="bar.y"
                      :width="bar.width"
                      :height="bar.height"
                      :fill="bar.color"
                      stroke="#fff"
                      stroke-width="0.3"
                    />
                    <!-- Bar labels -->
                    <text
                      v-if="bar.height > 10"
                      :x="bar.x + bar.width/2"
                      :y="bar.y + bar.height/2"
                      fill="#fff"
                      font-size="3"
                      text-anchor="middle"
                      alignment-baseline="middle"
                    >
                      {{ Math.round(bar.percentage) }}%
                    </text>

                    <!-- X-axis labels (rotated for readability) -->
                    <text
                      :x="bar.x + bar.width/2"
                      :y="93"
                      fill="var(--text-color)"
                      font-size="3"
                      text-anchor="middle"
                      transform="rotate(45, 0, 0)"
                    >
                      {{ bar.name }}
                    </text>
                  </g>
                </svg>

                <!-- Chart legend -->
                <div class="chart-legend">
                  <div v-for="(item, index) in expensesByType" :key="'type-bar-legend-' + index" class="legend-item">
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
              <NEmpty description="Not enough data to display chart" />
            </template>
          </div>
        </NTabPane>

        <!-- Expense by Category Tab -->
        <NTabPane name="byCategory" tab="By Category">
          <div class="select-container">
            <NSelect
              v-model:value="selectedType"
              :options="availableTypes"
              placeholder="Select expense type"
              @update:value="handleTypeChange"
              :disabled="availableTypes.length === 0"
            />
          </div>

          <div v-if="selectedType" class="chart-container">
            <NDivider>Categories for {{ selectedType }}</NDivider>

            <NRadioGroup v-model:value="chartMode" class="mode-selector">
              <NRadio value="pie">Pie Chart</NRadio>
              <NRadio value="bar">Bar Chart</NRadio>
            </NRadioGroup>

            <template v-if="expensesByCategoryForType.length > 0">
              <!-- Pie Chart View -->
              <template v-if="chartMode === 'pie' && categoriesPieChartData.length > 0">
                <div class="pie-chart">
                  <svg ref="categoriesPieChartRef" viewBox="0 0 200 200" width="100%" height="100%">
                    <!-- Render pie slices -->
                    <g v-for="(slice, index) in categoriesPieChartData" :key="'cat-pie-' + index">
                      <path :d="slice.pathData" :fill="slice.color" stroke="#fff" stroke-width="1" />
                      <!-- Add percentage labels in center of each slice -->
                      <text
                        v-if="slice.percentage > 5"
                        :x="slice.labelX"
                        :y="slice.labelY"
                        fill="#fff"
                        font-size="8"
                        text-anchor="middle"
                        alignment-baseline="middle"
                      >
                        {{ Math.round(slice.percentage) }}%
                      </text>
                    </g>
                    <!-- Inner circle for donut chart -->
                    <circle cx="100" cy="100" r="40" fill="var(--background-color)" />
                  </svg>

                  <!-- Chart legend -->
                  <div class="chart-legend">
                    <div v-for="(item, index) in expensesByCategoryForType" :key="'cat-legend-' + index" class="legend-item">
                      <div class="legend-color" :style="{ backgroundColor: item.color }"></div>
                      <div class="legend-text">
                        <div class="legend-name">{{ item.name }}</div>
                        <div class="legend-value">{{ formatCurrency(item.value) }} ({{ Math.round(item.percentage) }}%)</div>
                      </div>
                    </div>
                  </div>
                </div>
              </template>

              <!-- Bar Chart View -->
              <template v-else-if="chartMode === 'bar' && categoriesBarChartData.length > 0">
                <div class="bar-chart">
                  <svg ref="categoriesBarChartRef" viewBox="0 0 100 100" width="100%" height="100%">
                    <!-- Y-axis and grid lines -->
                    <line x1="10" y1="10" x2="10" y2="90" stroke="var(--border-color)" stroke-width="0.5" />
                    <line x1="10" y1="90" x2="90" y2="90" stroke="var(--border-color)" stroke-width="0.5" />

                    <!-- Render bars -->
                    <g v-for="(bar, index) in categoriesBarChartData" :key="'cat-bar-' + index">
                      <rect
                        :x="bar.x"
                        :y="bar.y"
                        :width="bar.width"
                        :height="bar.height"
                        :fill="bar.color"
                        stroke="#fff"
                        stroke-width="0.3"
                      />
                      <!-- Bar labels -->
                      <text
                        v-if="bar.height > 10"
                        :x="bar.x + bar.width/2"
                        :y="bar.y + bar.height/2"
                        fill="#fff"
                        font-size="3"
                        text-anchor="middle"
                        alignment-baseline="middle"
                      >
                        {{ Math.round(bar.percentage) }}%
                      </text>

                      <!-- X-axis labels (rotated for readability) -->
                      <text
                        :x="bar.x + bar.width/2"
                        :y="93"
                        fill="var(--text-color)"
                        font-size="3"
                        text-anchor="middle"
                        transform="rotate(45, 0, 0)"
                      >
                        {{ bar.name }}
                      </text>
                    </g>
                  </svg>

                  <!-- Chart legend -->
                  <div class="chart-legend">
                    <div v-for="(item, index) in expensesByCategoryForType" :key="'cat-bar-legend-' + index" class="legend-item">
                      <div class="legend-color" :style="{ backgroundColor: item.color }"></div>
                      <div class="legend-text">
                        <div class="legend-name">{{ item.name }}</div>
                        <div class="legend-value">{{ formatCurrency(item.value) }}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </template>
            </template>

            <template v-else>
              <NEmpty description="No categories found for the selected type" />
            </template>
          </div>
        </NTabPane>

        <!-- Monthly Trend Tab -->
        <NTabPane name="byMonth" tab="Monthly Trends">
          <div class="chart-container">
            <template v-if="monthlyBarChartData.length > 0">
              <div class="stacked-bar-chart">
                <svg ref="monthlyChartRef" viewBox="0 0 100 100" width="100%" height="100%">
                  <!-- Y-axis and grid lines -->
                  <line x1="10" y1="10" x2="10" y2="90" stroke="var(--border-color)" stroke-width="0.5" />
                  <line x1="10" y1="90" x2="90" y2="90" stroke="var(--border-color)" stroke-width="0.5" />

                  <!-- Render month groups and bars -->
                  <g v-for="(month, monthIndex) in monthlyBarChartData" :key="'month-' + monthIndex">
                    <!-- Month label -->
                    <text
                      :x="month.x + month.width/2"
                      :y="94"
                      fill="var(--text-color)"
                      font-size="3"
                      text-anchor="middle"
                    >
                      {{ month.month }}
                    </text>

                    <!-- Bars for each expense type in this month -->
                    <g v-for="(bar, barIndex) in month.bars" :key="'month-bar-' + monthIndex + '-' + barIndex">
                      <rect
                        :x="bar.x"
                        :y="bar.y"
                        :width="bar.width"
                        :height="bar.height"
                        :fill="bar.color"
                        stroke="#fff"
                        stroke-width="0.3"
                      />

                      <!-- Bar labels (only if bar is tall enough) -->
                      <text
                        v-if="bar.height > 10"
                        :x="bar.x + bar.width/2"
                        :y="bar.y + bar.height/2"
                        fill="#fff"
                        font-size="2.5"
                        text-anchor="middle"
                        alignment-baseline="middle"
                      >
                        {{ bar.name }}
                      </text>
                    </g>

                    <!-- Total amount label above stack -->
                    <text
                      :x="month.x + month.width/2"
                      :y="8"
                      fill="var(--text-color)"
                      font-size="3"
                      text-anchor="middle"
                    >
                      {{ formatCurrency(month.total) }}
                    </text>
                  </g>
                </svg>

                <!-- Chart legend -->
                <div class="chart-legend">
                  <div v-for="(type, index) in expenseTypes" :key="'month-legend-' + index" class="legend-item">
                    <div class="legend-color" :style="{ backgroundColor: COLORS[index % COLORS.length] }"></div>
                    <div class="legend-text">
                      <div class="legend-name">{{ type }}</div>
                    </div>
                  </div>
                </div>
              </div>
            </template>

            <template v-else>
              <NEmpty description="No monthly data available" />
            </template>
          </div>
        </NTabPane>
      </NTabs>
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
  min-height: 400px;
  margin: 10px 0;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.select-container {
  margin: 20px 0;
  max-width: 400px;
}

.mode-selector {
  margin-bottom: 16px;
  display: flex;
  justify-content: center;
}

.pie-chart, .bar-chart, .stacked-bar-chart {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  max-width: 600px;
}

.pie-chart svg, .bar-chart svg, .stacked-bar-chart svg {
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
