<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import {
  NSpace,
  NButton,
  NCard,
  NPageHeader,
  NEmpty,
  NDivider,
  NIcon
} from 'naive-ui'
import { PrintRound } from '@vicons/material'
import PieChart from '@/components/charts/PieChart.vue'
import BarChart from '@/components/charts/BarChart.vue'
import StackedBarChart from '@/components/charts/StackedBarChart.vue'
import type { ChartDataItem } from '@/components/charts/BaseChart.vue'
import type { StackedChartItem, StackedChartSeries } from '@/components/charts/StackedBarChart.vue'
import type { Expense } from '@/types/api'
import { formatCurrency } from '@/utils/formatters'

// Store chart data
const chartData = ref<{
  expenses: Expense[],
  expensesByType: ChartDataItem[],
  expensesByMonth: {
    items: StackedChartItem[],
    series: StackedChartSeries[]
  },
  typeDetails: Array<{
    type: string,
    value: number,
    families: ChartDataItem[]
  }>
} | null>(null)

// Chart display preferences
const typeChartMode = ref<'pie' | 'bar'>('pie')
const categoryChartMode = ref<'pie' | 'bar'>('pie')

// Set title and date information
const title = ref<string>('Expense Analysis')
const dateRange = ref<string>('')
const totalAmount = ref<number>(0)

// Print function
const printCharts = () => {
  window.print()
}

// Try to retrieve data from localStorage on mount
onMounted(() => {
  // Get data from localStorage (passed from parent window)
  const storedData = localStorage.getItem('standalone_chart_data')
  if (storedData) {
    try {
      chartData.value = JSON.parse(storedData)

      // Clean up - remove the data from localStorage after retrieving it
      localStorage.removeItem('standalone_chart_data')

      // Set summary data
      if (chartData.value?.expenses) {
        title.value = localStorage.getItem('standalone_chart_title') || 'Expense Analysis'
        dateRange.value = localStorage.getItem('standalone_chart_date_range') || ''
        totalAmount.value = chartData.value.expenses.reduce((sum, exp) => sum + exp.amount, 0)
      }
    } catch (error) {
      console.error('Error parsing chart data:', error)
    }
  }
})
</script>

<template>
  <div class="standalone-charts-page">
    <div class="print-header">
      <!-- Header for both screen and print -->
      <NPageHeader>
        <template #title>
          {{ title }}
        </template>
        <template #subtitle v-if="dateRange">
          Period: {{ dateRange }}
        </template>
        <template #extra>
          <NButton type="primary" @click="printCharts">
            <template #icon>
              <NIcon>
                <PrintRound />
              </NIcon>
            </template>
            Print
          </NButton>
        </template>
      </NPageHeader>
    </div>

    <div class="page-content" v-if="chartData">
      <!-- Summary Section -->
      <NCard title="Summary" class="summary-card">
        <div class="summary-content">
          <div class="summary-item">
            <div class="summary-label">Total Expenses</div>
            <div class="summary-value">{{ formatCurrency(totalAmount) }}</div>
          </div>
          <div class="summary-item">
            <div class="summary-label">Number of Expenses</div>
            <div class="summary-value">{{ chartData.expenses.length }}</div>
          </div>
          <div class="summary-item">
            <div class="summary-label">Average Expense</div>
            <div class="summary-value">{{ formatCurrency(totalAmount / chartData.expenses.length) }}</div>
          </div>
        </div>
      </NCard>

      <!-- Section 1: Expenses by Type -->
      <div class="page-break"></div>
      <NCard title="Expenses by Type" class="chart-card">
        <NSpace justify="center" class="chart-section">
          <div class="chart-container">
            <PieChart
              :data="chartData.expensesByType"
              :showPercentage="true"
              :height="400"
            />
          </div>

          <div class="chart-table">
            <table>
              <thead>
              <tr>
                <th>Type</th>
                <th>Amount</th>
                <th>Percentage</th>
                <th>Count</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="item in chartData.expensesByType" :key="item.name">
                <td>{{ item.name }}</td>
                <td>{{ formatCurrency(item.value) }}</td>
                <td>{{ item.percentage ? item.percentage.toFixed(1) + '%' : '0%' }}</td>
                <td>{{ item.count }}</td>
              </tr>
              </tbody>
            </table>
          </div>
        </NSpace>
      </NCard>

      <!-- Section 2: Monthly Trends -->
      <div class="page-break"></div>
      <NCard title="Monthly Expense Trends" class="chart-card">
        <div class="chart-container wide">
          <StackedBarChart
            v-if="chartData.expensesByMonth.items.length > 0"
            :data="chartData.expensesByMonth.items"
            :series="chartData.expensesByMonth.series"
            :height="400"
          />
          <NEmpty v-else description="No monthly data available" />
        </div>

        <!-- Monthly data table -->
        <div class="chart-table" v-if="chartData.expensesByMonth.items.length > 0">
          <table>
            <thead>
            <tr>
              <th>Month</th>
              <th>Total Amount</th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="item in chartData.expensesByMonth.items" :key="item.label">
              <td>{{ item.label }}</td>
              <td>{{ formatCurrency(item.total) }}</td>
            </tr>
            </tbody>
          </table>
        </div>
      </NCard>

      <!-- Section 3: Type Breakdown -->
      <template v-if="chartData.typeDetails && chartData.typeDetails.length > 0">
        <div v-for="(typeDetail, index) in chartData.typeDetails" :key="typeDetail.type" class="type-detail-section">
          <div class="page-break"></div>
          <NCard :title="typeDetail.type + ' - ' + formatCurrency(typeDetail.value)" class="chart-card">
            <!-- Family Breakdown -->
            <div class="chart-container">
              <PieChart
                :data="typeDetail.families"
                :title="`Family Breakdown for ${typeDetail.type}`"
                :showPercentage="true"
                :height="400"
              />
            </div>

            <!-- Families Table -->
            <div class="chart-table">
              <table>
                <thead>
                <tr>
                  <th>Family</th>
                  <th>Amount</th>
                  <th>Percentage</th>
                  <th>Count</th>
                </tr>
                </thead>
                <tbody>
                <tr v-for="family in typeDetail.families" :key="family.name">
                  <td>{{ family.name }}</td>
                  <td>{{ formatCurrency(family.value) }}</td>
                  <td>{{ family.percentage ? family.percentage.toFixed(1) + '%' : '0%' }}</td>
                  <td>{{ family.count }}</td>
                </tr>
                </tbody>
              </table>
            </div>
          </NCard>
        </div>
      </template>
    </div>

    <NEmpty v-else description="No chart data available" />
  </div>
</template>

<style>
/* Global styles that apply to both screen and print */
@media screen {
  .standalone-charts-page {
    padding: 20px;
    max-width: 1200px;
    margin: 0 auto;
  }
}

/* Print-specific styles */
@media print {
  /* Reset page margins */
  @page {
    margin: 0.5cm;
  }

  body {
    font-size: 12pt;
  }

  .standalone-charts-page {
    padding: 0;
  }

  .print-header {
    text-align: center;
    margin-bottom: 20px;
  }

  /* Hide print button when printing */
  button {
    display: none !important;
  }

  /* Force page breaks */
  .page-break {
    page-break-after: always;
    break-after: page;
  }

  /* Ensure each chart card starts on a new page */
  .chart-card {
    page-break-inside: avoid;
  }

  /* Make sure text is black for better printing */
  * {
    color: black !important;
    text-shadow: none !important;
    box-shadow: none !important;
  }
}

/* Common styles */
.summary-card {
  margin-bottom: 30px;
}

.summary-content {
  display: flex;
  justify-content: space-around;
  flex-wrap: wrap;
  gap: 20px;
}

.summary-item {
  text-align: center;
  min-width: 150px;
}

.summary-label {
  font-size: 14px;
  opacity: 0.8;
  margin-bottom: 8px;
}

.summary-value {
  font-size: 24px;
  font-weight: bold;
}

.chart-card {
  margin-bottom: 30px;
}

.chart-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.chart-container {
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
}

.chart-container.wide {
  max-width: 800px;
}

.chart-table {
  width: 100%;
  margin-top: 20px;
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 10px;
}

th, td {
  padding: 8px;
  text-align: left;
}

th {
  font-weight: bold;
}

tr:nth-child(even) {
}

.type-detail-section {
  margin-top: 30px;
}

@media (max-width: 768px) {
  .summary-content {
    flex-direction: column;
    align-items: center;
  }
}
</style>
