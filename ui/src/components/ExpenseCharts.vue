<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  NCard,
  NEmpty,
  NRadioGroup,
  NRadio,
  NCollapse,
  NCollapseItem,
  NButton,
  NSpace,
  NTooltip,
  NIcon
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { OpenInNewRound } from '@vicons/material'
import type { Expense } from '@/types/api'
import { formatCurrency } from '@/utils/formatters'
import PieChart from '@/components/charts/PieChart.vue'
import BarChart from '@/components/charts/BarChart.vue'
import StackedBarChart from '@/components/charts/StackedBarChart.vue'
import type { ChartDataItem } from '@/components/charts/BaseChart.vue'
import type { StackedChartItem, StackedChartSeries } from '@/components/charts/StackedBarChart.vue'
import LocalizedCategoryDisplay from './LocalizedCategoryDisplay.vue'

// Props
const props = defineProps<{
  expenses: Expense[],
  dateRange?: [number, number] | null
}>()

// I18n
const { t, locale } = useI18n()

// Chart display mode for each section
const typeChartMode = ref<'pie' | 'bar'>('pie')
const categoryChartMode = ref<'pie' | 'bar'>('pie')
const monthlyChartMode = ref<'bar'>('bar')

// Selected type for category breakdown
const selectedType = ref<string | null>(null)

// Generate a color palette for the charts
const COLORS = [
  '#3366FF', '#FF6633', '#33CC99', '#FFCC33', '#FF33CC',
  '#33CCFF', '#CC99FF', '#99CC33', '#FF9966', '#6699FF'
]

// Computed data for the charts
const expensesByType = computed<ChartDataItem[]>(() => {
  if (!props.expenses || props.expenses.length === 0) return []

  const grouped = props.expenses.reduce((acc, expense) => {
    const type = expense.category_type || 'Uncategorized'

    if (!acc[type]) {
      acc[type] = {
        name: t(`categories.types.${type}`),
        rawName: type, // Store original value for lookup
        value: 0,
        count: 0
      }
    }

    acc[type].value += expense.amount
    acc[type].count += 1

    return acc
  }, {} as Record<string, ChartDataItem & { rawName: string }>)

  return Object.values(grouped).map((item, index) => ({
    ...item,
    color: COLORS[index % COLORS.length],
    percentage: (item.value / props.expenses.reduce((sum, exp) => sum + exp.amount, 0)) * 100
  }))
})

// Family categorization - group expenses within each type by family
const expensesByTypeAndFamily = computed<Record<string, Record<string, ChartDataItem & { rawName: string }>>>(() => {
  if (!props.expenses || props.expenses.length === 0) return {}

  const typeAndFamily: Record<string, Record<string, ChartDataItem & { rawName: string }>> = {}

  // First pass - group by type and family
  props.expenses.forEach(expense => {
    const type = expense.category_type || 'Uncategorized'
    const family = expense.category_family || 'General'

    if (!typeAndFamily[type]) {
      typeAndFamily[type] = {}
    }

    if (!typeAndFamily[type][family]) {
      typeAndFamily[type][family] = {
        name: t(`categories.families.${family}`),
        rawName: family, // Store original value for lookup
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
        name: t(`categories.names.${category}`),
        rawName: category,
        value: 0,
        count: 0
      }
    }

    acc[category].value += expense.amount
    acc[category].count += 1

    return acc
  }, {} as Record<string, ChartDataItem & { rawName: string }>)

  const totalForFamily = filteredExpenses.reduce((sum, exp) => sum + exp.amount, 0)

  return Object.values(grouped)
    .map((item, index) => ({
      ...item,
      color: COLORS[index % COLORS.length],
      percentage: (item.value / totalForFamily) * 100
    }))
    .sort((a, b) => b.value - a.value)
}

// Get all unique expense types
const expenseTypes = computed<string[]>(() => {
  return [...new Set(props.expenses
    .map(e => e.category_type || 'Uncategorized')
    .filter(Boolean) as string[])]
})

// Monthly expenses data for stacked chart
const monthlyExpensesData = computed<{items: StackedChartItem[], series: StackedChartSeries[]}>(() => {
  if (!props.expenses || props.expenses.length === 0) {
    return { items: [], series: [] }
  }

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
  const uniqueCategories = expenseTypes.value

  // Create series data
  const series: StackedChartSeries[] = uniqueCategories.map((category, index) => ({
    name: category,
    color: COLORS[index % COLORS.length]
  }))

  // Create items data
  const items: StackedChartItem[] = months.map(month => {
    const values: Record<string, number> = {}
    let total = 0

    // Add data for each category
    uniqueCategories.forEach(category => {
      const value = monthlyData[month][category] || 0
      values[category] = value
      total += value
    })

    return {
      label: month,
      values,
      total
    }
  })

  return { items, series }
})

// Format date range as string for display
const formattedDateRange = computed(() => {
  if (!props.dateRange) return t('expenses.allTime', 'All time')

  const start = new Date(props.dateRange[0]).toLocaleDateString()
  const end = new Date(props.dateRange[1]).toLocaleDateString()

  return `${start} - ${end}`
})

// When expenses change, set the default selected type
watch(() => props.expenses, () => {
  if (expensesByType.value.length > 0 && (!selectedType.value || !expensesByType.value.find(t => t.name === selectedType.value))) {
    selectedType.value = expensesByType.value[0].name
  }
}, { immediate: true })

const openStandaloneChartsPage = () => {
  try {
    // Prepare data for standalone view
    const typeDetails = expensesByType.value.map(type => {
      // Get the families for this type
      const families = getFamiliesForType(type.rawName);

      return {
        type: type.name,
        rawName: type.rawName, // Important: Include the raw name for translation
        value: type.value,
        families: families.map(family => ({
          ...family,
          // Ensure rawName is preserved for each family
          rawName: family.rawName
        }))
      };
    });

    // Create the data object to pass to the standalone page with rawNames
    const chartData = {
      expenses: props.expenses,
      expensesByType: expensesByType.value.map(item => ({
        ...item,
        // Ensure rawName is preserved for each type
        rawName: item.rawName
      })),
      expensesByMonth: monthlyExpensesData.value,
      typeDetails
    };

    // Store data in localStorage so the new window can access it
    localStorage.setItem('standalone_chart_data', JSON.stringify(chartData));
    localStorage.setItem('standalone_chart_title', t('expenses.expenseAnalysis', 'Expense Analysis'));
    localStorage.setItem('standalone_chart_date_range', formattedDateRange.value);
    localStorage.setItem('standalone_chart_language', locale.value);

    // Open a new window with the standalone HTML page (not a route)
    const standaloneWindow = window.open('/standalone-charts.html', '_blank');

    if (!standaloneWindow) {
      alert(t('expenses.allowPopups', 'Please allow pop-ups to open the charts in a new window'));
    }
  } catch (error) {
    console.error('Error opening standalone charts page:', error);
    alert(t('expenses.chartError', 'There was an error opening the charts page. Please try again.'));
  }
}

// Format a category and value for display
const formatCategoryAndValue = (categoryName: string, value: number) => {
  return `${categoryName} - ${formatCurrency(value)}`
}
</script>

<template>
  <div class="expense-charts">
    <template v-if="props.expenses.length === 0">
      <NEmpty :description="t('expenses.noExpensesFilters', 'No expenses found for the selected filters')" />
    </template>

    <template v-else>
      <!-- Export Button -->
      <div class="charts-actions">
        <NSpace justify="end">
          <NTooltip>
            <template #trigger>
              <NButton type="primary" @click="openStandaloneChartsPage">
                <template #icon>
                  <NIcon>
                    <OpenInNewRound />
                  </NIcon>
                </template>
                {{ t('expenses.openChartsView', 'Open Charts View') }}
              </NButton>
            </template>
            {{ t('expenses.printablePage', 'Open all charts in a printable page') }}
          </NTooltip>
        </NSpace>
      </div>

      <!-- Section 1: Expenses by Type -->
      <NCard :title="t('expenses.expensesByType', 'Expenses by Type')" style="margin-bottom: 24px;">
        <NRadioGroup v-model:value="typeChartMode" class="mode-selector">
          <NRadio value="pie">{{ t('charts.pieChart', 'Pie Chart') }}</NRadio>
          <NRadio value="bar">{{ t('charts.barChart', 'Bar Chart') }}</NRadio>
        </NRadioGroup>

        <div class="chart-container">
          <PieChart
            v-if="typeChartMode === 'pie' && expensesByType.length > 0"
            :data="expensesByType"
            :showPercentage="true"
            :height="300"
          />

          <BarChart
            v-else-if="typeChartMode === 'bar' && expensesByType.length > 0"
            :data="expensesByType"
            :height="300"
          />

          <NEmpty v-else :description="t('charts.noData', 'No data available')" />
        </div>
      </NCard>

      <!-- Section 2: Monthly Trends -->
      <NCard :title="t('expenses.monthlyTrends', 'Monthly Expense Trends')" style="margin-bottom: 24px;">
        <div class="chart-container">
          <StackedBarChart
            v-if="monthlyExpensesData.items.length > 0"
            :data="monthlyExpensesData.items"
            :series="monthlyExpensesData.series"
            :height="300"
          />

          <NEmpty v-else :description="t('expenses.noMonthlyData', 'No monthly data available')" />
        </div>
      </NCard>

      <!-- Section 3: Expense Type Breakdown -->
      <NCard :title="t('expenses.expenseTypeBreakdown', 'Expense Type Breakdown')" style="margin-bottom: 24px;">
        <NRadioGroup v-model:value="categoryChartMode" class="mode-selector">
          <NRadio value="pie">{{ t('charts.pieChart', 'Pie Chart') }}</NRadio>
          <NRadio value="bar">{{ t('charts.barChart', 'Bar Chart') }}</NRadio>
        </NRadioGroup>

        <NCollapse>
          <NCollapseItem
            v-for="type in expensesByType"
            :key="type.rawName"
            :title="formatCategoryAndValue(type.name, type.value)"
          >
            <!-- Families within this type section -->
            <div class="chart-container">
              <NCard :title="t('expenses.familyBreakdown', 'Family Breakdown')" size="small" style="margin-bottom: 16px;">
                <template v-if="getFamiliesForType(type.rawName).length > 0">
                  <PieChart
                    v-if="categoryChartMode === 'pie'"
                    :data="getFamiliesForType(type.rawName)"
                    :showPercentage="true"
                    :height="300"
                  />

                  <BarChart
                    v-else
                    :data="getFamiliesForType(type.rawName)"
                    :height="300"
                  />
                </template>
                <template v-else>
                  <NEmpty :description="t('expenses.noFamiliesFound', 'No families found for this type')" />
                </template>
              </NCard>

              <!-- Categories within each family collapsible sections -->
              <NCollapse>
                <NCollapseItem
                  v-for="family in getFamiliesForType(type.rawName)"
                  :key="family.rawName"
                  :title="formatCategoryAndValue(family.name, family.value)"
                >
                  <div class="chart-container">
                    <template v-if="getCategoriesForTypeAndFamily(type.rawName, family.rawName).length > 0">
                      <PieChart
                        v-if="categoryChartMode === 'pie'"
                        :data="getCategoriesForTypeAndFamily(type.rawName, family.rawName)"
                        :showPercentage="true"
                        :height="300"
                      />

                      <BarChart
                        v-else
                        :data="getCategoriesForTypeAndFamily(type.rawName, family.rawName)"
                        :height="300"
                      />
                    </template>
                    <template v-else>
                      <NEmpty :description="t('expenses.noCategoriesFound', 'No categories found for this family')" />
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

.charts-actions {
  margin-bottom: 16px;
  text-align: right;
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
</style>
