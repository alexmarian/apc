<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import { NDataTable, NButton, NSpace, NTag, NEmpty, NSpin, NAlert, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { accountApi } from '@/services/api'
import type { Account } from '@/types/api'
import { useI18n } from 'vue-i18n'

// Props
const props = defineProps<{
  associationId: number
}>()

// Emits
const emit = defineEmits<{
  (e: 'edit', accountId: number): void
}>()

// I18n
const { t } = useI18n()

// Data
const accounts = ref<Account[]>([])
const loading = ref<boolean>(true)
const error = ref<string | null>(null)
const message = useMessage()

// Table columns definition
const columns = ref<DataTableColumns<Account>>([
  {
    title: t('accounts.accountNumber'),
    key: 'number',
    sorter: 'default'
  },
  {
    title: t('accounts.description'),
    key: 'description'
  },
  {
    title: t('accounts.destination'),
    key: 'destination'
  },
  {
    title: t('accounts.status'),
    key: 'is_active',
    render(row) {
      return h(
        NTag,
        {
          type: row.is_active ? 'success' : 'error',
          bordered: false
        },
        { default: () => row.is_active ? t('accounts.active') : t('accounts.inactive') }
      )
    }
  },
  {
    title: t('common.actions'),
    key: 'actions',
    render(row) {
      return h(
        NSpace,
        {
          justify: 'center',
          align: 'center'
        },
        {
          default: () => [
            h(
              NButton,
              {
                strong: true,
                secondary: true,
                type: 'info',
                size: 'small',
                disabled: !row.is_active,
                onClick: () => emit('edit', row.id)
              },
              { default: () => t('common.edit') }
            ),
            h(
              NButton,
              {
                strong: true,
                secondary: true,
                type: 'error',
                size: 'small',
                disabled: !row.is_active,
                onClick: () => disableAccount(row.id)
              },
              { default: () => t('common.disable') }
            )
          ]
        }
      )
    }
  }
])

// Fetch accounts
const fetchAccounts = async () => {
  try {
    loading.value = true
    error.value = null

    const response = await accountApi.getAccounts(props.associationId)
    accounts.value = response.data
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('common.error')
    console.error('Error fetching accounts:', err)
  } finally {
    loading.value = false
  }
}

// Disable account
const disableAccount = async (accountId: number) => {
  try {
    const confirmDisable = window.confirm(t('accounts.confirmDisable'))
    if (!confirmDisable) return

    await accountApi.disableAccount(props.associationId, accountId)

    // Update the local state
    const index = accounts.value.findIndex(acc => acc.id === accountId)
    if (index !== -1) {
      accounts.value[index].is_active = false
    }

    message.success(t('accounts.accountDisabled'))
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : t('common.error')
    error.value = errorMessage
    console.error('Error disabling account:', err)
    message.error(t('common.error') + ': ' + errorMessage)
  }
}

// Load accounts on component mount
onMounted(() => {
  fetchAccounts()
})
</script>

<template>
  <div class="accounts-list">
    <h2>{{ t('accounts.list') }}</h2>

    <NSpin :show="loading">
      <NAlert v-if="error" type="error" :title="t('common.error')" closable>
        {{ error }}
        <NButton @click="fetchAccounts">{{ t('common.retry') }}</NButton>
      </NAlert>

      <NDataTable
        v-else
        :columns="columns"
        :data="accounts"
        :bordered="false"
        :single-line="false"
        :pagination="{
          pageSize: 10
        }"
        :row-props="row => ({
          style: !row.is_active ? 'opacity: 0.6' : ''
        })"
      >
        <template #empty>
          <NEmpty :description="t('accounts.noAccounts')">
            <template #extra>
              <p>{{ t('accounts.createToStart') }}</p>
            </template>
          </NEmpty>
        </template>
      </NDataTable>
    </NSpin>
  </div>
</template>

<style scoped>
.accounts-list {
  margin: 2rem 0;
}
</style>
