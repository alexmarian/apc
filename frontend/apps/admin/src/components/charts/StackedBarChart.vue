<script setup lang="ts">
import { computed } from 'vue'
import { Bar } from 'vue-chartjs'
import { Chart as ChartJS, Title, Tooltip, Legend, CategoryScale, LinearScale, BarElement } from 'chart.js'
import type { ChartData, ChartOptions } from 'chart.js'
import { NSpace, NTooltip } from 'naive-ui'
import { formatCurrency } from '@/utils/formatters'

// Register Chart.js components
ChartJS.register(Title, Tooltip, Legend, CategoryScale, LinearScale, BarElement)

// Type for stacked data series
export interface StackedChartSeries {
  name: string;
  color: string;
}

export interface StackedChartItem {
  label: string;
  values: Record<string, number>;
  total: number;
}

// Props
const props = defineProps<{
  data: StackedChartItem[];
  series: StackedChartSeries[];
  title?: string;
  height?: number;
}>()

// Generate chart data for Chart.js
const stackedBarChartData = computed<ChartData<'bar'>>(() => {
  const labels = props.data.map(item => item.label);

  // Create a dataset for each series
  const datasets = props.series.map((serie) => {
    return {
      label: serie.name,
      data: props.data.map(item => item.values[serie.name] || 0),
      backgroundColor: serie.color,
      borderWidth: 1
    }
  });

  return {
    labels,
    datasets
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
</script>

<template>
  <div class="stacked-bar-chart">
    <h4 v-if="title" class="chart-title">{{ title }}</h4>

    <div class="chart-container">
      <Bar
        :data="stackedBarChartData"
        :options="chartOptions"
        :height="height || 300"
      />
    </div>

    <!-- Legend for series -->
    <NSpace wrap justify="center" align="center" class="chart-legend">
      <NTooltip v-for="(serie, index) in series" :key="'legend-item-' + index" placement="top">
        <template #trigger>
          <div class="legend-item">
            <div class="legend-color" :style="{ backgroundColor: serie.color }"></div>
            <div class="legend-name">{{ serie.name }}</div>
          </div>
        </template>
        {{ serie.name }}
      </NTooltip>
    </NSpace>
  </div>
</template>

<style scoped>
.stacked-bar-chart {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.chart-title {
  margin-bottom: 1rem;
  text-align: center;
}

.chart-container {
  width: 100%;
  max-width: 500px;
  height: v-bind('`${height || 300}px`');
}

.chart-legend {
  margin-top: 1rem;
  padding: 0.5rem;
  max-width: 100%;
}

.legend-item {
  display: flex;
  align-items: center;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  transition: background-color 0.2s;
  cursor: pointer;
}

.legend-item:hover {
  background-color: rgba(0, 0, 0, 0.05);
}

.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 2px;
  margin-right: 8px;
  flex-shrink: 0;
}

.legend-name {
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}
</style>
