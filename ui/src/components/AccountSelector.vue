<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { NSelect, NSpin } from 'naive-ui'
import { accountApi } from '@/services/api'
import type { Account } from '@/types/api'
import { useI18n } from 'vue-i18n'

// Props
const props = defineProps<{
  modelValue: number | null
  associationId: number
  activeOnly?: boolean
  placeholder?: string
  disabled?: boolean
}>()

// Emits
const emit = defineEmits<{
  (e: 'update:modelValue', id: number | null): void
}>()

// I18n
const { t } = useI18n()

// State
const accounts = ref<Account[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

// Computed options for NSelect
const options = computed(() => {
  let filteredAccounts = accounts.value

  // Filter to only active accounts if activeOnly is true
  if (props.activeOnly) {
    filteredAccounts = filteredAccounts.filter(account => account.is_active)
  }

  return filteredAccounts.map(account => ({
    label: `${account.number} - ${account.description}`,
    value: account.id
  }))
})

// Fetch accounts
const fetchAccounts = async () => {
  if (!props.associationId) return

  try {
    loading.value = true
    error.value = null

    const response = await accountApi.getAccounts(props.associationId)
    accounts.value = response.data

    // If no account is selected and we have accounts, select the first one
    if (!props.modelValue && accounts.value.length > 0 && !props.disabled) {
      // If activeOnly is true, find the first active account
      if (props.activeOnly) {
        const firstActiveAccount = accounts.value.find(account => account.is_active)
        if (firstActiveAccount) {
          emit('update:modelValue', firstActiveAccount.id)
        }
      } else {
        emit('update:modelValue', accounts.value[0].id)
      }
    }
  } catch (err) {
    console.error('Error fetching accounts:', err)
    error.value = t('common.error')
  } finally {
    loading.value = false
  }
}

// Handle selection change
const handleChange = (value: number | null) => {
  emit('update:modelValue', value)
}

// Load accounts on mount
onMounted(() => {
  fetchAccounts()
})
</script>

<template>
  <div class="account-selector">
    <NSpin :show="loading">
      <NSelect
        :value="props.modelValue"
        :options="options"
        :placeholder="placeholder || t('accounts.select', 'Select an account')"
        @update:value="handleChange"
        :disabled="loading || accounts.length === 0 || props.disabled"
      />
    </NSpin>
    <p v-if="error" class="error">{{ error }}</p>
  </div>
</template>

<style scoped>
.account-selector {
  width: 100%;
}
.error {
  color: #d03050;
  font-size: 0.8rem;
  margin-top: 4px;
}
</style>
