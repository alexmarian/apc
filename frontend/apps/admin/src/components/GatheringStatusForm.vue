<template>
  <div class="gathering-status-form">
    <NCard style="width: 500px">
      <template #header>
        <h2>{{ $t('gatherings.status.title') }}</h2>
      </template>

      <NSpin :show="loading">
        <NAlert v-if="error" type="error" closable @close="error = null">
          {{ error }}
        </NAlert>

        <div class="current-status">
          <p><strong>{{ $t('gatherings.status.current') }}:</strong></p>
          <NTag :type="getStatusType(gathering.status)" size="large">
            {{ $t(`gatherings.status.${gathering.status}`) }}
          </NTag>
        </div>

        <NForm ref="formRef" :model="formData" :rules="rules" label-placement="top">
          <NFormItem :label="$t('gatherings.status.new')" path="status">
            <NSelect 
              v-model:value="formData.status" 
              :options="availableStatuses"
              :placeholder="$t('gatherings.status.selectNew')"
            />
          </NFormItem>

          <div v-if="formData.status" class="status-info">
            <NAlert type="info">
              {{ getStatusDescription(formData.status) }}
            </NAlert>
          </div>

          <div v-if="formData.status && isDestructiveChange" class="warning">
            <NAlert type="warning">
              {{ $t('gatherings.status.destructiveWarning') }}
            </NAlert>
          </div>
        </NForm>

        <div class="form-actions">
          <NSpace justify="end">
            <NButton @click="handleCancel">
              {{ $t('common.cancel') }}
            </NButton>
            <NButton 
              type="primary" 
              @click="handleSubmit" 
              :loading="submitting"
              :disabled="!formData.status || formData.status === gathering.status"
            >
              {{ $t('gatherings.status.update') }}
            </NButton>
          </NSpace>
        </div>
      </NSpin>
    </NCard>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NCard,
  NForm,
  NFormItem,
  NSelect,
  NButton,
  NSpace,
  NAlert,
  NSpin,
  NTag,
  type FormInst,
  type FormRules
} from 'naive-ui'
import { gatheringApi } from '@/services/api'
import type { Gathering, GatheringStatus } from '@/types/api'

interface Props {
  associationId: number
  gathering: Gathering
}

const props = defineProps<Props>()

const emit = defineEmits<{
  saved: []
  cancelled: []
}>()

const { t } = useI18n()

const formRef = ref<FormInst | null>(null)
const loading = ref(false)
const submitting = ref(false)
const error = ref<string | null>(null)

const formData = reactive<{
  status: GatheringStatus | null
}>({
  status: null
})

const statusTransitions: Record<GatheringStatus, GatheringStatus[]> = {
  draft: ['published' as GatheringStatus],
  published: ['active' as GatheringStatus, 'draft' as GatheringStatus],
  active: ['closed' as GatheringStatus],
  closed: ['tallied' as GatheringStatus, 'active' as GatheringStatus],
  tallied: []
}

const availableStatuses = computed(() => {
  const available = statusTransitions[props.gathering.status] || []
  return available.map(status => ({
    label: t(`gatherings.status.${status}`),
    value: status
  }))
})

const isDestructiveChange = computed(() => {
  if (!formData.status) return false
  const destructiveTransitions = ['draft', 'active']
  return destructiveTransitions.includes(formData.status)
})

const getStatusType = (status: GatheringStatus) => {
  switch (status) {
    case 'draft':
      return 'default'
    case 'published':
      return 'info'
    case 'active':
      return 'success'
    case 'closed':
      return 'warning'
    case 'tallied':
      return 'error'
    default:
      return 'default'
  }
}

const getStatusDescription = (status: GatheringStatus) => {
  switch (status) {
    case 'draft':
      return t('gatherings.status.descriptions.draft')
    case 'published':
      return t('gatherings.status.descriptions.published')
    case 'active':
      return t('gatherings.status.descriptions.active')
    case 'closed':
      return t('gatherings.status.descriptions.closed')
    case 'tallied':
      return t('gatherings.status.descriptions.tallied')
    default:
      return ''
  }
}

const rules: FormRules = {
  status: [
    { required: true, message: t('gatherings.status.required') }
  ]
}

const handleSubmit = async () => {
  if (!formRef.value || !formData.status) return

  try {
    await formRef.value.validate()
    submitting.value = true
    error.value = null

    await gatheringApi.updateGatheringStatus(props.associationId, props.gathering.id, {
      status: formData.status
    })

    emit('saved')
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : t('common.error')
    error.value = errorMessage
  } finally {
    submitting.value = false
  }
}

const handleCancel = () => {
  emit('cancelled')
}
</script>

<style scoped>
.gathering-status-form {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 16px;
}

.current-status {
  margin-bottom: 20px;
}

.status-info {
  margin-top: 12px;
}

.warning {
  margin-top: 12px;
}

.form-actions {
  margin-top: 24px;
}
</style>