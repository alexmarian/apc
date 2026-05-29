<script setup lang="ts">
import { computed } from 'vue'
import { Pie } from 'vue-chartjs'
import { Chart as ChartJS, ArcElement, Title, Tooltip, Legend } from 'chart.js'
import type { ChartData, ChartOptions } from 'chart.js'
import BaseChart, { type ChartDataItem } from './BaseChart.vue'
import { formatCurrency } from '@/utils/formatters'

// Register Chart.js components
ChartJS.register(ArcElement, Title, Tooltip, Legend)

// Props
const props = defineProps<{
  data: ChartDataItem[];
  title?: string;
  height?: number;
  showLegend?: boolean;
  showPercentage?: boolean;
  showCount?: boolean;
}>()

// Generate chart data for Chart.js
const pieChartData = computed<ChartData<'pie'>>(() => {
  return {
    labels: props.data.map(item => item.name),
    datasets: [
      {
        data: props.data.map(item => item.value),
        backgroundColor: props.data.map(item => item.color),
        hoverBackgroundColor: props.data.map(item => item.color),
        borderWidth: 1
      }
    ]
  }
})

// Chart options
const chartOptions = computed<ChartOptions<'pie'>>(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false // We're using our custom legend
    },
    tooltip: {
      callbacks: {
        label: (context) => {
          const label = context.label || '';
          const value = context.raw as number;

          // Calculate percentage from dataset
          const dataset = context.chart.data.datasets[0];
          const total = (dataset.data as number[]).reduce((sum, val) => sum + (val || 0), 0);
          const percentage = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0';

          return `${label}: ${formatCurrency(value)} (${percentage}%)`;
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
      <Pie
        :data="pieChartData"
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
