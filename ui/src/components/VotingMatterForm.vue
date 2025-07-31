<template>
  <div class="voting-matter-form">
    <NCard style="width: 700px">
      <template #header>
        <h2>{{ isEditMode ? $t('gatherings.matters.edit') : $t('gatherings.matters.create') }}</h2>
      </template>

      <NSpin :show="loading">
        <NAlert v-if="error" type="error" closable @close="error = null">
          {{ error }}
        </NAlert>

        <NForm ref="formRef" :model="formData" :rules="rules" label-placement="top">
          <NFormItem :label="$t('gatherings.matters.title')" path="title">
            <NInput 
              v-model:value="formData.title" 
              :placeholder="$t('gatherings.matters.titlePlaceholder')" 
            />
          </NFormItem>

          <NFormItem :label="$t('gatherings.matters.description')" path="description">
            <NInput
              v-model:value="formData.description"
              type="textarea"
              :placeholder="$t('gatherings.matters.descriptionPlaceholder')"
              :rows="3"
            />
          </NFormItem>

          <NGrid :cols="2" :x-gap="16">
            <NGridItem>
              <NFormItem :label="$t('gatherings.matters.type')" path="matter_type">
                <NSelect v-model:value="formData.matter_type" :options="typeOptions" />
              </NFormItem>
            </NGridItem>
            
            <NGridItem>
              <NFormItem :label="$t('gatherings.matters.order')" path="order_index">
                <NInputNumber
                  v-model:value="formData.order_index"
                  :min="1"
                  :max="100"
                  style="width: 100%"
                />
              </NFormItem>
            </NGridItem>
          </NGrid>

          <NDivider>{{ $t('gatherings.matters.votingConfig') }}</NDivider>

          <NFormItem :label="$t('gatherings.matters.votingType')" path="voting_config.type">
            <NSelect 
              v-model:value="formData.voting_config.type" 
              :options="votingTypeOptions"
              @update:value="handleVotingTypeChange"
            />
          </NFormItem>

          <NFormItem 
            v-if="needsOptions"
            :label="$t('gatherings.matters.options')" 
            path="voting_config.options"
          >
            <div class="options-section">
              <div 
                v-for="(option, index) in formData.voting_config.options" 
                :key="index"
                class="option-item"
              >
                <NInput
                  v-model:value="formData.voting_config.options[index].text"
                  :placeholder="$t('gatherings.matters.optionPlaceholder', { index: index + 1 })"
                />
                <NButton 
                  type="error" 
                  size="small" 
                  @click="removeOption(index)"
                  :disabled="formData.voting_config.options.length <= 2"
                >
                  {{ $t('gatherings.matters.removeOption') }}
                </NButton>
              </div>
              <NButton @click="addOption" :type="'dashed' as any" style="width: 100%">
                {{ $t('gatherings.matters.addOption') }}
              </NButton>
            </div>
          </NFormItem>

          <NGrid :cols="2" :x-gap="16">
            <NGridItem>
              <NFormItem :label="$t('gatherings.matters.majorityType')" path="voting_config.required_majority">
                <NSelect 
                  v-model:value="formData.voting_config.required_majority" 
                  :options="majorityTypeOptions"
                  @update:value="handleMajorityTypeChange"
                />
              </NFormItem>
            </NGridItem>
            
            <NGridItem>
              <NFormItem 
                v-if="needsMajorityThreshold"
                :label="$t('gatherings.matters.majorityThreshold')" 
                path="voting_config.required_majority_value"
              >
                <NInputNumber
                  v-model:value="formData.voting_config.required_majority_value"
                  :min="1"
                  :max="100"
                  :precision="1"
                  style="width: 100%"
                >
                  <template #suffix>%</template>
                </NInputNumber>
              </NFormItem>
            </NGridItem>
          </NGrid>

          <NFormItem :label="$t('gatherings.matters.quorumThreshold')" path="voting_config.quorum">
            <NInputNumber
              v-model:value="formData.voting_config.quorum"
              :min="1"
              :max="100"
              :precision="1"
              style="width: 100%"
            >
              <template #suffix>%</template>
            </NInputNumber>
          </NFormItem>

          <NGrid :cols="2" :x-gap="16">
            <NGridItem>
              <NFormItem :label="$t('gatherings.matters.isAnonymous')" path="voting_config.is_anonymous">
                <NSwitch v-model:value="formData.voting_config.is_anonymous" />
              </NFormItem>
            </NGridItem>
            
            <NGridItem>
              <NFormItem :label="$t('gatherings.matters.allowAbstention')" path="voting_config.allow_abstention">
                <NSwitch v-model:value="formData.voting_config.allow_abstention" />
              </NFormItem>
            </NGridItem>
          </NGrid>
        </NForm>

        <div class="form-actions">
          <NSpace justify="end">
            <NButton @click="handleCancel">
              {{ $t('common.cancel') }}
            </NButton>
            <NButton type="primary" @click="handleSubmit" :loading="submitting">
              {{ isEditMode ? $t('common.update') : $t('common.create') }}
            </NButton>
          </NSpace>
        </div>
      </NSpin>
    </NCard>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NCard,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSelect,
  NButton,
  NSpace,
  NAlert,
  NSpin,
  NGrid,
  NGridItem,
  NDivider,
  NSwitch,
  type FormInst,
  type FormRules
} from 'naive-ui'
import { votingMatterApi } from '@/services/api'
import type { 
  Gathering, 
  VotingMatter, 
  VotingMatterCreateRequest,
  VotingMatterUpdateRequest,
  VotingMatterType, 
  VotingType, 
  VotingOption,
  MajorityType 
} from '@/types/api'

interface Props {
  associationId: number
  gathering: Gathering
  matter?: VotingMatter
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

const isEditMode = computed(() => !!props.matter)

const formData = reactive<{
  title: string
  description: string
  matter_type: VotingMatterType
  order_index: number
  voting_config: {
    type: VotingType
    options: VotingOption[]
    required_majority: MajorityType
    required_majority_value: number | null
    quorum: number | null
    is_anonymous: boolean
    allow_abstention: boolean
  }
}>({
  title: '',
  description: '',
  matter_type: 'policy' as VotingMatterType,
  order_index: 1,
  voting_config: {
    type: 'yes_no' as VotingType,
    options: [],
    required_majority: 'simple' as MajorityType,
    required_majority_value: null,
    quorum: 50,
    is_anonymous: false,
    allow_abstention: true
  }
})

const typeOptions = computed(() => [
  { label: t('gatherings.matters.types.budget'), value: 'budget' },
  { label: t('gatherings.matters.types.election'), value: 'election' },
  { label: t('gatherings.matters.types.policy'), value: 'policy' },
  { label: t('gatherings.matters.types.poll'), value: 'poll' },
  { label: t('gatherings.matters.types.extraordinary'), value: 'extraordinary' }
])

const votingTypeOptions = computed(() => [
  { label: t('gatherings.matters.votingTypes.yes_no'), value: 'yes_no' },
  { label: t('gatherings.matters.votingTypes.multiple_choice'), value: 'multiple_choice' },
  { label: t('gatherings.matters.votingTypes.single_choice'), value: 'single_choice' },
  { label: t('gatherings.matters.votingTypes.ranking'), value: 'ranking' }
])

const majorityTypeOptions = computed(() => [
  { label: t('gatherings.matters.majorityTypes.simple'), value: 'simple' },
  { label: t('gatherings.matters.majorityTypes.absolute'), value: 'absolute' },
  { label: t('gatherings.matters.majorityTypes.qualified'), value: 'qualified' },
  { label: t('gatherings.matters.majorityTypes.unanimous'), value: 'unanimous' }
])

const needsOptions = computed(() => {
  return ['multiple_choice', 'single_choice', 'ranking'].includes(formData.voting_config.type)
})

const needsMajorityThreshold = computed(() => {
  return formData.voting_config.required_majority === 'qualified'
})

const rules: FormRules = {
  title: [
    { required: true, message: t('gatherings.matters.titleRequired') }
  ],
  description: [
    { required: true, message: t('gatherings.matters.descriptionRequired') }
  ],
  matter_type: [
    { required: true, message: t('gatherings.matters.typeRequired') }
  ],
  order_index: [
    { required: true, message: t('gatherings.matters.orderRequired'), type: 'number' }
  ],
  'voting_config.type': [
    { required: true, message: t('gatherings.matters.votingTypeRequired') }
  ],
  'voting_config.required_majority': [
    { required: true, message: t('gatherings.matters.majorityTypeRequired') }
  ]
}

const handleVotingTypeChange = (value: VotingType) => {
  if (needsOptions.value && formData.voting_config.options.length === 0) {
    formData.voting_config.options = [{ id: '1', text: '' }, { id: '2', text: '' }]
  } else if (!needsOptions.value) {
    formData.voting_config.options = []
  }
}

const handleMajorityTypeChange = (value: MajorityType) => {
  if (value === 'qualified' && !formData.voting_config.required_majority_value) {
    formData.voting_config.required_majority_value = 66.7
  } else if (value !== 'qualified') {
    formData.voting_config.required_majority_value = null
  }
}

const addOption = () => {
  const nextId = (formData.voting_config.options.length + 1).toString()
  formData.voting_config.options.push({ id: nextId, text: '' })
}

const removeOption = (index: number) => {
  if (formData.voting_config.options.length > 2) {
    formData.voting_config.options.splice(index, 1)
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    submitting.value = true
    error.value = null

    const requestData = {
      ...formData,
      voting_config: {
        ...formData.voting_config,
        options: needsOptions.value ? formData.voting_config.options.filter(opt => opt.text.trim()) : undefined
      }
    }

    if (isEditMode.value) {
      await votingMatterApi.updateVotingMatter(
        props.associationId, 
        props.gathering.id, 
        props.matter!.id, 
        requestData as VotingMatterUpdateRequest
      )
    } else {
      await votingMatterApi.createVotingMatter(
        props.associationId, 
        props.gathering.id, 
        requestData as VotingMatterCreateRequest
      )
    }

    emit('saved')
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || t('common.error')
  } finally {
    submitting.value = false
  }
}

const handleCancel = () => {
  emit('cancelled')
}

watch(() => props.matter, (newMatter) => {
  if (newMatter) {
    Object.assign(formData, {
      title: newMatter.title,
      description: newMatter.description,
      matter_type: newMatter.matter_type,
      order_index: newMatter.order_index,
      voting_config: {
        type: newMatter.voting_config.type,
        options: newMatter.voting_config.options || [],
        required_majority: newMatter.voting_config.required_majority,
        required_majority_value: newMatter.voting_config.required_majority_value,
        quorum: newMatter.voting_config.quorum,
        is_anonymous: newMatter.voting_config.is_anonymous,
        allow_abstention: newMatter.voting_config.allow_abstention
      }
    })
  }
}, { immediate: true })
</script>

<style scoped>
.voting-matter-form {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 16px;
}

.options-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.option-item {
  display: flex;
  gap: 8px;
  align-items: center;
}

.form-actions {
  margin-top: 24px;
}
</style>