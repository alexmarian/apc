<template>
  <div>
    <div
      v-for="matter in matters"
      :key="matter.id"
      style="margin-bottom: 20px"
    >
      <NCard size="small">
        <template #header>{{ matterTitle(matter) }}</template>

        <NText v-if="matter.description || matter.description_ru" :depth="2" style="display: block; margin-bottom: 12px; font-size: 13px">
          {{ matterDescription(matter) }}
        </NText>

        <!-- yes_no -->
        <NRadioGroup
          v-if="matter.voting_config.type === 'yes_no'"
          :value="singleValue(matter.id)"
          @update:value="setSingle(matter.id, $event)"
        >
          <NSpace>
            <NRadio value="yes">{{ t('yes') }}</NRadio>
            <NRadio value="no">{{ t('no') }}</NRadio>
            <NRadio v-if="matter.voting_config.allow_abstention" value="abstain">
              {{ t('abstain') }}
            </NRadio>
          </NSpace>
        </NRadioGroup>

        <!-- single_choice -->
        <NRadioGroup
          v-else-if="matter.voting_config.type === 'single_choice'"
          :value="singleValue(matter.id)"
          @update:value="setSingle(matter.id, $event)"
        >
          <NSpace vertical>
            <NRadio
              v-for="option in matter.voting_config.options ?? []"
              :key="option.id"
              :value="option.id"
            >
              {{ option.text }}
            </NRadio>
            <NRadio v-if="matter.voting_config.allow_abstention" value="abstain">
              {{ t('abstain') }}
            </NRadio>
          </NSpace>
        </NRadioGroup>

        <!-- multiple_choice -->
        <NCheckboxGroup
          v-else-if="matter.voting_config.type === 'multiple_choice'"
          :value="multiValue(matter.id)"
          @update:value="setMulti(matter.id, $event)"
        >
          <NSpace vertical>
            <NCheckbox
              v-for="option in matter.voting_config.options ?? []"
              :key="option.id"
              :value="option.id"
            >
              {{ option.text }}
            </NCheckbox>
          </NSpace>
        </NCheckboxGroup>

        <!-- ranking -->
        <div v-else-if="matter.voting_config.type === 'ranking'">
          <NText :depth="3" style="font-size: 12px; margin-bottom: 10px; display: block">
            {{ t('rankingHint') }}
          </NText>
          <div
            v-for="(optId, idx) in rankValue(matter.id)"
            :key="optId"
            style="display: flex; align-items: center; gap: 8px; margin-bottom: 6px; padding: 6px 10px; border: 1px solid var(--n-border-color); border-radius: 4px"
          >
            <NText :depth="3" style="min-width: 20px; text-align: center; font-weight: 600">{{ idx + 1 }}</NText>
            <NText style="flex: 1">{{ optionText(matter, optId) }}</NText>
            <NButton size="tiny" :disabled="idx === 0" @click="moveRank(matter.id, idx, -1)">↑</NButton>
            <NButton
              size="tiny"
              :disabled="idx === rankValue(matter.id).length - 1"
              @click="moveRank(matter.id, idx, 1)"
            >↓</NButton>
          </div>
        </div>
      </NCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NButton,
  NCard,
  NCheckbox,
  NCheckboxGroup,
  NRadio,
  NRadioGroup,
  NSpace,
  NText
} from 'naive-ui'
import type { MatterInfo } from './types'

const { t, locale } = useI18n({
  useScope: 'local',
  messages: {
    en: {
      yes: 'Yes',
      no: 'No',
      abstain: 'Abstain',
      rankingHint: 'Rank options from most preferred (top) to least preferred (bottom).',
    },
    ro: {
      yes: 'Da',
      no: 'Nu',
      abstain: 'Abținere',
      rankingHint: 'Ordonați opțiunile de la cea mai preferată (sus) la cea mai puțin preferată (jos).',
    },
    ru: {
      yes: 'Да',
      no: 'Нет',
      abstain: 'Воздержаться',
      rankingHint: 'Упорядочьте варианты от наиболее предпочтительного (вверху) до наименее предпочтительного (внизу).',
    },
  }
})

const props = defineProps<{
  matters: MatterInfo[]
  modelValue: Record<string, string[]>
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, string[]>]
}>()

// Initialise empty entries for matters not yet in modelValue
watch(
  () => props.matters,
  (matters) => {
    const current = { ...props.modelValue }
    let changed = false
    for (const m of matters) {
      const key = String(m.id)
      if (current[key] === undefined) {
        current[key] = m.voting_config.type === 'ranking'
          ? (m.voting_config.options ?? []).map(o => o.id)
          : []
        changed = true
      }
    }
    if (changed) emit('update:modelValue', current)
  },
  { immediate: true }
)

function singleValue(matterId: number): string {
  return props.modelValue[String(matterId)]?.[0] ?? ''
}

function setSingle(matterId: number, val: string) {
  emit('update:modelValue', { ...props.modelValue, [String(matterId)]: val ? [val] : [] })
}

function multiValue(matterId: number): string[] {
  return props.modelValue[String(matterId)] ?? []
}

function setMulti(matterId: number, vals: (string | number)[]) {
  emit('update:modelValue', { ...props.modelValue, [String(matterId)]: vals.map(String) })
}

function rankValue(matterId: number): string[] {
  return props.modelValue[String(matterId)] ?? []
}

function moveRank(matterId: number, idx: number, dir: -1 | 1) {
  const arr = [...rankValue(matterId)]
  const newIdx = idx + dir
  if (newIdx < 0 || newIdx >= arr.length) return
  ;[arr[idx], arr[newIdx]] = [arr[newIdx], arr[idx]]
  emit('update:modelValue', { ...props.modelValue, [String(matterId)]: arr })
}

function matterTitle(matter: MatterInfo): string {
  const lang = locale.value?.slice(0, 2)
  if (lang === 'ru' && matter.title_ru) return matter.title_ru
  return matter.title
}

function matterDescription(matter: MatterInfo): string {
  const lang = locale.value?.slice(0, 2)
  if (lang === 'ru' && matter.description_ru) return matter.description_ru
  return matter.description
}

function optionText(matter: MatterInfo, optId: string): string {
  return matter.voting_config.options?.find(o => o.id === optId)?.text ?? optId
}
</script>
