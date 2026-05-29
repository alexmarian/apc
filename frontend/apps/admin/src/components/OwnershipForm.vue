<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import {
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NSpin,
  NAlert,
  NDatePicker,
  NSelect,
  NCheckbox,
  NDivider,
  useMessage
} from 'naive-ui'
import { ownerApi, unitApi, ownershipApi } from '@/services/api'
import type { Owner } from '@/types/api'
import type { FormRules } from 'naive-ui'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  associationId: number
  buildingId: number
  unitId: number
  mode: 'create' | 'select' // create new owner or select existing
}>()

const emit = defineEmits<{
  (e: 'saved'): void
  (e: 'cancelled'): void
}>()

// Form data
const formData = reactive({
  // Owner data (for create mode)
  owner: {
    name: '',
    identification_number: '',
    contact_phone: '',
    contact_email: ''
  },
  // Owner selection (for select mode)
  owner_id: null as number | null,
  // Ownership data (for both modes)
  start_date: Date.now(),
  end_date: null as number | null,
  registration_document: '',
  registration_date: Date.now(),
  // Is this an exclusive ownership (replaces all others)?
  is_exclusive: false
})

// State
const loading = ref(false)
const submitting = ref(false)
const error = ref<string | null>(null)
const owners = ref<Owner[]>([])
const message = useMessage()

// Form validation rules
const rules: FormRules = {
  'owner.name': [
    { required: props.mode === 'create', message: t('validation.required', { field: t('owners.name', 'Owner name') }, '{field} is required'), trigger: 'blur' }
  ],
  'owner.identification_number': [
    { required: props.mode === 'create', message: t('validation.required', { field: t('owners.identification', 'Identification number') }, '{field} is required'), trigger: 'blur' }
  ],
  owner_id: [
    {
      required: props.mode === 'select',
      type: 'number',
      message: t('units.ownership.selectOwnerRequired', 'Please select an owner'),
      trigger: 'change'
    }
  ],
  start_date: [
    { required: true, message: t('validation.required', { field: t('units.ownership.startDate', 'Start date') }, '{field} is required'), trigger: 'blur' }
  ],
  registration_document: [
    { required: true, message: t('validation.required', { field: t('units.ownership.registrationDoc', 'Registration document') }, '{field} is required'), trigger: 'blur' }
  ],
  registration_date: [
    { required: true, message: t('validation.required', { field: t('units.ownership.registrationDate', 'Registration date') }, '{field} is required'), trigger: 'blur' }
  ]
}

// Fetch existing owners for selection mode
const fetchOwners = async () => {
  if (props.mode !== 'select') return

  try {
    loading.value = true
    error.value = null

    const response = await ownerApi.getOwners(props.associationId)
    owners.value = response.data
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('common.error', 'Failed to load owners')
    console.error('Error fetching owners:', err)
  } finally {
    loading.value = false
  }
}

// Owner options for select dropdown
const ownerOptions = computed(() => {
  return owners.value.map(owner => ({
    label: `${owner.name} (${owner.identification_number})`,
    value: owner.id
  }))
})

// Submit form handler
const handleSubmit = async () => {
  try {
    submitting.value = true
    error.value = null

    let ownerId: number;

    // If in create mode, create the owner first
    if (props.mode === 'create') {
      const createOwnerResponse = await ownerApi.createOwner(
        props.associationId,
        {
          name: formData.owner.name,
          identification_number: formData.owner.identification_number,
          contact_phone: formData.owner.contact_phone,
          contact_email: formData.owner.contact_email
        }
      );
      ownerId = createOwnerResponse.data.id;
    } else {
      // In select mode, use the selected owner ID
      if (!formData.owner_id) {
        error.value = t('units.ownership.selectOwnerRequired', 'Please select an owner');
        return;
      }
      ownerId = formData.owner_id;
    }

    // Create the ownership
    await ownershipApi.createUnitOwnership(
      props.associationId,
      props.buildingId,
      props.unitId,
      {
        owner_id: ownerId,
        start_date: new Date(formData.start_date).toISOString(),
        end_date: formData.end_date ? new Date(formData.end_date).toISOString() : null,
        registration_document: formData.registration_document,
        registration_date: new Date(formData.registration_date).toISOString(),
        is_exclusive: formData.is_exclusive
      }
    );

    message.success(t('units.ownership.savedSuccess', 'Ownership saved successfully'));
    emit('saved');
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('common.error', 'An error occurred while saving');
    console.error('Error submitting form:', err);
  } finally {
    submitting.value = false;
  }
}

// Load owners when component mounts
onMounted(() => {
  fetchOwners()
})
</script>

<template>
  <div class="ownership-form">
    <h2>{{ props.mode === 'create'
      ? t('units.ownership.addNewOwner', 'Add New Owner')
      : t('units.ownership.selectExistingOwner', 'Select Existing Owner') }}
    </h2>

    <NSpin :show="loading">
      <NAlert v-if="error" type="error" style="margin-bottom: 16px;">
        {{ error }}
      </NAlert>

      <NForm :rules="rules" :model="formData">
        <!-- Owner creation form fields -->
        <template v-if="props.mode === 'create'">
          <NFormItem :label="t('owners.name', 'Owner Name')" path="owner.name">
            <NInput
              v-model:value="formData.owner.name"
              :placeholder="t('owners.enterName', 'Enter owner name')"
            />
          </NFormItem>

          <NFormItem :label="t('owners.identification', 'Identification Number')" path="owner.identification_number">
            <NInput
              v-model:value="formData.owner.identification_number"
              :placeholder="t('owners.enterIdentification', 'Enter ID number')"
            />
          </NFormItem>

          <NFormItem :label="t('owners.contactPhone', 'Phone Number')" path="owner.contact_phone">
            <NInput
              v-model:value="formData.owner.contact_phone"
              :placeholder="t('owners.enterPhone', 'Enter phone number')"
            />
          </NFormItem>

          <NFormItem :label="t('owners.contactEmail', 'Email')" path="owner.contact_email">
            <NInput
              v-model:value="formData.owner.contact_email"
              :placeholder="t('owners.enterEmail', 'Enter email address')"
            />
          </NFormItem>
        </template>

        <!-- Owner selection dropdown -->
        <template v-else>
          <NFormItem :label="t('units.ownership.selectOwner', 'Select Owner')" path="owner_id">
            <NSelect
              v-model:value="formData.owner_id"
              :options="ownerOptions"
              :placeholder="t('units.ownership.selectOwnerPlaceholder', 'Select an owner')"
              filterable
            />
          </NFormItem>
        </template>

        <!-- Ownership details (common for both modes) -->
        <NDivider>{{ t('units.ownership.details', 'Ownership Details') }}</NDivider>

        <NFormItem :label="t('units.ownership.startDate', 'Start Date')" path="start_date">
          <NDatePicker
            v-model:value="formData.start_date"
            type="date"
            clearable
            style="width: 100%"
          />
        </NFormItem>

        <NFormItem :label="t('units.ownership.endDate', 'End Date')" path="end_date">
          <NDatePicker
            v-model:value="formData.end_date"
            type="date"
            clearable
            style="width: 100%"
          />
        </NFormItem>

        <NFormItem :label="t('units.ownership.registrationDoc', 'Registration Document')" path="registration_document">
          <NInput
            v-model:value="formData.registration_document"
            :placeholder="t('units.ownership.enterRegistrationDoc', 'Enter registration document number')"
          />
        </NFormItem>

        <NFormItem :label="t('units.ownership.registrationDate', 'Registration Date')" path="registration_date">
          <NDatePicker
            v-model:value="formData.registration_date"
            type="date"
            clearable
            style="width: 100%"
          />
        </NFormItem>

        <NFormItem path="is_exclusive">
          <NCheckbox v-model:checked="formData.is_exclusive">
            {{ t('units.ownership.exclusiveOwnership', 'Exclusive Ownership (deactivates all other current ownerships)') }}
          </NCheckbox>
        </NFormItem>

        <!-- Form actions -->
        <div style="margin-top: 24px;">
          <NSpace justify="end">
            <NButton @click="emit('cancelled')" :disabled="submitting">
              {{ t('common.cancel', 'Cancel') }}
            </NButton>

            <NButton type="primary" @click="handleSubmit" :loading="submitting">
              {{ t('units.ownership.saveOwnership', 'Save Ownership') }}
            </NButton>
          </NSpace>
        </div>
      </NForm>
    </NSpin>
  </div>
</template>

<style scoped>
.ownership-form {
  max-width: 600px;
  margin: 0 auto;
}
</style>
