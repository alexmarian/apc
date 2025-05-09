<script setup lang="ts">
import { computed } from 'vue'
import { Bar } from 'vue-chartjs'
import { Chart as ChartJS, Title, Tooltip, Legend, CategoryScale, LinearScale, BarElement } from 'chart.js'
import type { ChartData, ChartOptions } from 'chart.js'
import BaseChart, { type ChartDataItem } from './BaseChart.vue'
import { formatCurrency } from '@/utils/formatters'

// Register Chart.js components
ChartJS.register(Title, Tooltip, Legend, CategoryScale, LinearScale, BarElement)

// Props
const props = defineProps<{
  data: ChartDataItem[];
  title?: string;
  height?: number;
  showLegend?: boolean;
  showPercentage?: boolean;
  showCount?: boolean;
  stacked?: boolean;
}>()

// Generate chart data for Chart.js
const barChartData = computed<ChartData<'bar'>>(() => {
  return {
    labels: props.data.map(item => item.name),
    datasets: [
      {
        label: 'Amount',
        data: props.data.map(item => item.value),
        backgroundColor: props.data.map(item => item.color),
        borderWidth: 1
      }
    ]
  }
})

// Chart options
const chartOptions = computed<ChartOptions<'bar'>>(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false // We're using our custom legend
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
      stacked: !!props.stacked,
      ticks: {
        maxRotation: 45,
        minRotation: 45
      }
    },
    y: {
      stacked: !!props.stacked,
      beginAtZero: true,
      ticks: {
        callback: (value) => {
          return formatCurrency(value as number);
        }
      }
    }
  }
}))
</script>

<template>
  <BaseChart
    :data="data"
    :title="title"
    :showPercentage="showPercentage"
    :showCount="showCount"
  >
    <div class="chart-container">
      <Bar
        :data="barChartData"
        :options="chartOptions"
        :height="height || 300"
      />
    </div>
  </BaseChart>
</template>

<style scoped>
.chart-container {
  width: 100%;
  max-width: 500px;
  height: v-bind('`${height || 300}px`');
}
</style>
