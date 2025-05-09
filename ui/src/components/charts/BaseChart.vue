<script setup lang="ts">
import { computed } from 'vue'
import { NTooltip, NSpace } from 'naive-ui'
import { formatCurrency } from '@/utils/formatters'
import { useI18n } from 'vue-i18n'

// Common data item interface used across chart types
export interface ChartDataItem {
  name: string;
  value: number;
  color?: string;
  percentage?: number;
  count: number;
}

// Props
const props = defineProps<{
  data: ChartDataItem[];
  title?: string;
  showPercentage?: boolean;
  showCount?: boolean;
}>()

// I18n
const { t } = useI18n()

// Calculate total value for percentages if not already calculated
const totalValue = computed(() => {
  return props.data.reduce((sum, item) => sum + item.value, 0)
})

// Calculate percentages if not already calculated
const dataWithPercentages = computed(() => {
  return props.data.map(item => ({
    ...item,
    percentage: item.percentage ?? (totalValue.value > 0 ? (item.value / totalValue.value) * 100 : 0)
  }))
})
</script>

<template>
  <div class="base-chart">
    <h4 v-if="title" class="chart-title">{{ title }}</h4>
    <slot></slot>

    <!-- Standard Legend using Naive UI components -->
    <NSpace wrap justify="center" align="center" class="chart-legend">
      <NTooltip v-for="(item, index) in dataWithPercentages" :key="'legend-item-' + index" placement="top">
        <template #trigger>
          <div class="legend-item">
            <div class="legend-color" :style="{ backgroundColor: item.color }"></div>
            <div class="legend-text">
              <div class="legend-name" :title="item.name">{{ item.name }}</div>
              <div class="legend-value">
                {{ formatCurrency(item.value) }}
                <span v-if="showPercentage">({{ Math.round(item.percentage!) }}%)</span>
                <span v-if="showCount && item.count">[{{ item.count }}]</span>
              </div>
            </div>
          </div>
        </template>
        <div>
          <div>{{ item.name }}</div>
          <div>{{ formatCurrency(item.value) }}</div>
          <div v-if="showPercentage">{{ item.percentage!.toFixed(1) }}%</div>
          <div v-if="showCount && item.count">{{ t('common.total') }}: {{ item.count }}</div>
        </div>
      </NTooltip>
    </NSpace>
  </div>
</template>

<style scoped>
.base-chart {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.chart-title {
  margin-bottom: 1rem;
  text-align: center;
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

.legend-text {
  display: flex;
  flex-direction: column;
  max-width: 150px;
}

.legend-name {
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.legend-value {
  font-size: 11px;
  opacity: 0.8;
  white-space: nowrap;
}
</style>
