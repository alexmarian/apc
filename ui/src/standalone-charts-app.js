// src/standalone-charts-app.js
import { h, ref, onMounted, computed } from 'vue'
import {
  NSpace, NButton, NCard, NPageHeader, NEmpty, NDivider, NIcon
} from 'naive-ui'
import { PrintRound } from '@vicons/material'
import PieChart from './components/charts/PieChart.vue'
import BarChart from './components/charts/BarChart.vue'
import StackedBarChart from './components/charts/StackedBarChart.vue'
import { formatCurrency } from './utils/formatters'

// Create a simplified app with just the charts
export default {
  setup() {
    // State
    const chartData = ref(null)
    const title = ref('Expense Analysis')
    const dateRange = ref('')
    const totalAmount = ref(0)

    // Print function
    const printCharts = () => {
      window.print()
    }

    // Initialize on mount - get data from localStorage
    onMounted(() => {
      const storedData = localStorage.getItem('standalone_chart_data')
      if (storedData) {
        try {
          chartData.value = JSON.parse(storedData)

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

    return {
      chartData, title, dateRange, totalAmount, printCharts, formatCurrency
    }
  },

  render() {
    return h('div', { class: 'standalone-container' }, [// Header
      h(NPageHeader, {}, {
        title: () => h('div', {}, this.title),
        subtitle: () => this.dateRange ? h('div', {}, `Period: ${this.dateRange}`) : null,
        extra: () => h(NButton, {
          type: 'primary', onClick: this.printCharts
        }, {
          default: () => 'Print Charts', icon: () => h(NIcon, {}, {
            default: () => h(PrintRound)
          })
        })
      }),

      // Content
      this.chartData ? [// Summary Section
        h(NCard, { title: 'Summary', class: 'summary-card' }, {
          default: () => h('div', { class: 'summary-content' }, [h('div', { class: 'summary-item' }, [h('div', { class: 'summary-label' }, 'Total Expenses'), h('div', { class: 'summary-value' }, this.formatCurrency(this.totalAmount))]), h('div', { class: 'summary-item' }, [h('div', { class: 'summary-label' }, 'Number of Expenses'), h('div', { class: 'summary-value' }, this.chartData.expenses.length)]), h('div', { class: 'summary-item' }, [h('div', { class: 'summary-label' }, 'Average Expense'), h('div', { class: 'summary-value' }, this.formatCurrency(this.totalAmount / this.chartData.expenses.length))])])
        }),

        // Page break
        h('div', { class: 'page-break' }),

        // Expenses by Type
        h(NCard, { title: 'Expenses by Type', class: 'chart-card' }, {
          default: () => [h('div', { class: 'chart-container' }, [h(PieChart, {
            data: this.chartData.expensesByType, showPercentage: true, height: 400
          })])
            // Type data table would be added here
          ]
        })

        // Additional chart sections would follow the same pattern
      ] : h(NEmpty, { description: 'No chart data available' })])
  }
}
