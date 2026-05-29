<template>
  <div class="participant-form">
    <NCard style="width: 600px">
      <template #header>
        <h2>{{ $t('gatherings.participants.add') }}</h2>
      </template>

      <NSpin :show="loading">
        <NAlert v-if="error" type="error" closable @close="error = null">
          {{ error }}
        </NAlert>

        <!-- Step 1: Participant Information -->
        <div v-if="formData.step === 'participant'">
          <NForm ref="formRef" :model="formData" :rules="rules" label-placement="top">
            <NFormItem :label="$t('gatherings.participants.participantType')" path="type">
              <NSelect v-model:value="formData.type" :options="typeOptions" />
            </NFormItem>

            <NFormItem :label="$t('gatherings.participants.owner')" path="owner_id">
              <NSelect
                v-model:value="formData.owner_id"
                :options="ownerOptions"
                :loading="ownersLoading"
                filterable
                :placeholder="$t('gatherings.participants.selectOwner')"
              />
            </NFormItem>

            <NFormItem :label="$t('gatherings.participants.units')" path="unit_ids">
              <NSelect
                v-model:value="formData.unit_ids"
                :options="unitOptions"
                :loading="unitsLoading"
                multiple
                filterable
                :placeholder="$t('gatherings.participants.selectUnits')"
              />
            </NFormItem>

            <div v-if="formData.type === 'delegate'" class="delegate-fields">
              <NFormItem :label="$t('gatherings.participants.delegateName')" path="delegate_name">
                <NInput 
                  v-model:value="formData.delegate_name" 
                  :placeholder="$t('gatherings.participants.delegateNamePlaceholder')"
                />
              </NFormItem>

              <NFormItem :label="$t('gatherings.participants.delegateContact')" path="delegate_contact">
                <NInput 
                  v-model:value="formData.delegate_contact" 
                  :placeholder="$t('gatherings.participants.delegateContactPlaceholder')"
                />
              </NFormItem>

              <NFormItem :label="$t('gatherings.participants.delegationDocument')" path="delegation_document">
                <NInput 
                  v-model:value="formData.delegation_document" 
                  :placeholder="$t('gatherings.participants.delegationDocumentPlaceholder')"
                />
              </NFormItem>
            </div>

            <NFormItem :label="$t('gatherings.participants.addBallotInfo')" path="add_ballot_info">
              <NSwitch v-model:value="formData.add_ballot_info" />
              <NText depth="3" style="margin-left: 8px">
                {{ $t('gatherings.participants.addBallotInfoDescription') }}
              </NText>
            </NFormItem>
          </NForm>
        </div>

        <!-- Step 2: Ballot Information -->
        <div v-if="formData.step === 'ballot' && formData.created_participant">
          <BallotForm
            :association-id="associationId"
            :gathering="gathering"
            :participant="formData.created_participant"
            :is-offline-mode="true"
            @saved="handleBallotSaved"
            @cancelled="handleBallotCancelled"
          />
        </div>

        <div v-if="formData.step === 'participant'" class="form-actions">
          <NSpace justify="end">
            <NButton @click="handleCancel">
              {{ $t('common.cancel') }}
            </NButton>
            <NButton type="primary" @click="handleSubmit" :loading="submitting">
              {{ formData.add_ballot_info ? $t('common.next') : $t('common.create') }}
            </NButton>
          </NSpace>
        </div>
      </NSpin>
    </NCard>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NCard,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NButton,
  NSpace,
  NAlert,
  NSpin,
  NSwitch,
  NText,
  type FormInst,
  type FormRules
} from 'naive-ui'
import { participantApi, ownerApi, gatheringApi } from '@/services/api'
import type { 
  Gathering, 
  ParticipantCreateRequest, 
  ParticipantType,
  Owner,
  QualifiedUnit,
  GatheringParticipant
} from '@/types/api'
import BallotForm from './BallotForm.vue'

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
const ownersLoading = ref(false)
const unitsLoading = ref(false)
const owners = ref<Owner[]>([])
const qualifiedUnits = ref<QualifiedUnit[]>([])

const formData = reactive<{
  type: ParticipantType
  owner_id: number | null
  unit_ids: number[]
  delegate_name: string
  delegate_contact: string
  delegation_document: string
  step: 'participant' | 'ballot'
  add_ballot_info: boolean
  created_participant: GatheringParticipant | null
}>({
  type: 'owner' as ParticipantType,
  owner_id: null,
  unit_ids: [],
  delegate_name: '',
  delegate_contact: '',
  delegation_document: '',
  step: 'participant',
  add_ballot_info: false,
  created_participant: null
})

const typeOptions = computed(() => [
  { label: t('gatherings.participants.types.owner'), value: 'owner' },
  { label: t('gatherings.participants.types.delegate'), value: 'delegate' }
])

// Only show owners who have qualified units
const ownerOptions = computed(() => {
  const ownersWithUnits = new Set<number>()
  qualifiedUnits.value.forEach(unit => {
    if (unit.owner_id) {
      ownersWithUnits.add(unit.owner_id)
    }
  })
  
  return owners.value
    .filter(owner => ownersWithUnits.has(owner.id))
    .map(owner => ({
      label: `${owner.name} (${owner.identification_number})`,
      value: owner.id
    }))
})

// Only show units that belong to the selected owner
const unitOptions = computed(() => {
  if (!formData.owner_id) {
    return []
  }
  
  return qualifiedUnits.value
    .filter(unit => unit.owner_id === formData.owner_id)
    .map(unit => ({
      label: `${unit.unit_number} - ${unit.building_name}`,
      value: unit.id,
      disabled: unit.is_participating
    }))
})

const rules: FormRules = {
  type: [
    { required: true, message: t('gatherings.participants.typeRequired') }
  ],
  owner_id: [
    { required: true, message: t('gatherings.participants.ownerRequired'), type: 'number' }
  ],
  unit_ids: [
    { required: true, message: t('gatherings.participants.unitsRequired'), type: 'array', min: 1 }
  ],
  delegate_name: [
    { required: true, message: t('gatherings.participants.delegateNameRequired'), trigger: 'blur' }
  ],
  delegate_contact: [
    { required: true, message: t('gatherings.participants.delegateContactRequired'), trigger: 'blur' }
  ]
}

const loadOwners = async () => {
  ownersLoading.value = true
  try {
    const response = await ownerApi.getOwners(props.associationId)
    owners.value = response.data
  } catch (err: unknown) {
    console.error('Failed to load owners:', err)
  } finally {
    ownersLoading.value = false
  }
}

const loadQualifiedUnits = async () => {
  unitsLoading.value = true
  try {
    const response = await gatheringApi.getQualifiedUnits(props.associationId, props.gathering.id)
    qualifiedUnits.value = response.data
  } catch (err: unknown) {
    console.error('Failed to load qualified units:', err)
  } finally {
    unitsLoading.value = false
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    submitting.value = true
    error.value = null

    const requestData: ParticipantCreateRequest = {
      participant_type: formData.type,
      unit_ids: formData.unit_ids,
      ...(formData.type === 'owner' && {
        owner_id: formData.owner_id!
      }),
      ...(formData.type === 'delegate' && {
        delegating_owner_id: formData.owner_id!,
        delegation_document_ref: formData.delegation_document,
        delegate_name: formData.delegate_name,
        delegate_contact: formData.delegate_contact
      })
    }

    const response = await participantApi.addParticipant(props.associationId, props.gathering.id, requestData)
    formData.created_participant = response.data

    if (formData.add_ballot_info) {
      formData.step = 'ballot'
    } else {
      emit('saved')
    }
  } catch (err: unknown) {
    error.value = (err as Error).message || t('common.error')
  } finally {
    submitting.value = false
  }
}

const handleBallotSaved = () => {
  emit('saved')
}

const handleBallotCancelled = () => {
  formData.step = 'participant'
}

const handleCancel = () => {
  emit('cancelled')
}

// Clear unit selection when owner changes
watch(() => formData.owner_id, () => {
  formData.unit_ids = []
})

onMounted(() => {
  loadOwners()
  loadQualifiedUnits()
})
</script>

<style scoped>
.participant-form {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 16px;
}

.delegate-fields {
  margin-top: 16px;
}

.form-actions {
  margin-top: 24px;
}
</style>