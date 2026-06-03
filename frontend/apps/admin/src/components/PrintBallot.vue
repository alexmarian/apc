<script setup lang="ts">
import { ref } from 'vue'
import { NButton, NSpin } from 'naive-ui'
import { votingMatterApi, associationApi } from '@/services/api'
import type { Gathering, VotingMatter } from '@/types/api'
import { VotingType } from '@/types/api'

const props = defineProps<{
  associationId: number
  gathering: Gathering
  lang: 'ro' | 'ru'
}>()

const printing = ref(false)

const LABELS = {
  ro: {
    ballot: 'BULETIN DE VOT',
    ownerName: 'Nume proprietar',
    signature: 'Semnătură',
    yes: 'DA',
    no: 'NU',
    abstain: 'ABȚINERE',
    rankPrefix: 'Locul',
    informative: 'Punct informativ (fără vot)',
    page: 'Pagina',
  },
  ru: {
    ballot: 'БЮЛЛЕТЕНЬ ДЛЯ ГОЛОСОВАНИЯ',
    ownerName: 'Имя владельца',
    signature: 'Подпись',
    yes: 'ДА',
    no: 'НЕТ',
    abstain: 'ВОЗДЕРЖАТЬСЯ',
    rankPrefix: 'Место',
    informative: 'Информационный пункт (без голосования)',
    page: 'Страница',
  },
} as const

function matterTitle(m: VotingMatter): string {
  return props.lang === 'ru' ? (m.title_ru || m.title) : m.title
}

function matterDescription(m: VotingMatter): string {
  return props.lang === 'ru' ? (m.description_ru || m.description) : m.description
}

function renderMatterInputs(m: VotingMatter): string {
  const L = LABELS[props.lang]
  const cfg = m.voting_config

  if (m.is_informative) {
    return `<p class="informative">${L.informative}</p>`
  }

  if (cfg.type === VotingType.YesNo) {
    const abstainBox = cfg.allow_abstention
      ? `<label class="option"><span class="box"></span>${L.abstain}</label>`
      : ''
    return `
      <div class="options-row">
        <label class="option"><span class="box"></span>${L.yes}</label>
        <label class="option"><span class="box"></span>${L.no}</label>
        ${abstainBox}
      </div>`
  }

  if (cfg.type === VotingType.SingleChoice || cfg.type === VotingType.MultipleChoice) {
    const options = (cfg.options ?? []).map(o =>
      `<label class="option block"><span class="box"></span>${o.text}</label>`
    ).join('')
    const abstainBox = cfg.allow_abstention
      ? `<label class="option block"><span class="box"></span>${L.abstain}</label>`
      : ''
    return `<div class="options-list">${options}${abstainBox}</div>`
  }

  if (cfg.type === VotingType.Ranking) {
    const options = (cfg.options ?? []).map((o, i) =>
      `<div class="rank-row"><span class="rank-label">${L.rankPrefix} ${i + 1}</span><span class="rank-line"></span><span class="rank-option">${o.text}</span></div>`
    ).join('')
    return `<div class="ranking">${options}</div>`
  }

  return ''
}

function buildHtml(matters: VotingMatter[], associationName: string): string {
  const L = LABELS[props.lang]
  const date = props.gathering.scheduled_date
    ? new Date(props.gathering.scheduled_date).toLocaleDateString(props.lang === 'ru' ? 'ru-RU' : 'ro-RO', { day: '2-digit', month: 'long', year: 'numeric' })
    : ''

  const mattersHtml = matters
    .sort((a, b) => a.order_index - b.order_index)
    .map((m, i) => `
      <div class="matter">
        <div class="matter-header">${i + 1}. ${matterTitle(m)}</div>
        ${matterDescription(m) ? `<div class="matter-desc">${matterDescription(m)}</div>` : ''}
        ${renderMatterInputs(m)}
      </div>
    `).join('')

  return `<!DOCTYPE html>
<html lang="${props.lang}">
<head>
  <meta charset="UTF-8">
  <title>${L.ballot}</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body { font-family: Arial, sans-serif; font-size: 12px; color: #000; padding: 20mm; }
    h1 { font-size: 18px; text-align: center; text-transform: uppercase; letter-spacing: 2px; margin-bottom: 4px; }
    .meta { text-align: center; font-size: 11px; color: #444; margin-bottom: 16px; }
    .gathering-title { font-size: 14px; font-weight: bold; text-align: center; margin-bottom: 16px; }
    hr { border: none; border-top: 2px solid #000; margin: 12px 0; }
    .owner-block { display: flex; gap: 32px; margin-bottom: 16px; }
    .owner-field { flex: 1; }
    .field-label { font-size: 10px; text-transform: uppercase; letter-spacing: 1px; color: #666; margin-bottom: 2px; }
    .field-line { border-bottom: 1px solid #000; height: 24px; }
    .matter { margin-bottom: 20px; page-break-inside: avoid; }
    .matter-header { font-weight: bold; font-size: 12px; margin-bottom: 4px; }
    .matter-desc { font-size: 11px; color: #555; margin-bottom: 8px; }
    .informative { font-size: 11px; font-style: italic; color: #777; }
    .options-row { display: flex; gap: 24px; }
    .options-list { display: flex; flex-direction: column; gap: 6px; }
    .option { display: flex; align-items: center; gap: 8px; font-size: 12px; }
    .option.block { display: flex; }
    .box { display: inline-block; width: 14px; height: 14px; border: 1.5px solid #000; flex-shrink: 0; }
    .rank-row { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
    .rank-label { font-size: 11px; color: #666; width: 50px; flex-shrink: 0; }
    .rank-line { flex: 1; border-bottom: 1px solid #999; }
    .rank-option { font-size: 11px; width: 160px; flex-shrink: 0; }
    @media print {
      body { padding: 15mm; }
      .matter { page-break-inside: avoid; }
    }
  </style>
</head>
<body>
  <h1>${L.ballot}</h1>
  <div class="meta">${associationName} &mdash; ${date}</div>
  <div class="gathering-title">${props.gathering.title}</div>
  <hr>
  <div class="owner-block">
    <div class="owner-field">
      <div class="field-label">${L.ownerName}</div>
      <div class="field-line"></div>
    </div>
    <div class="owner-field">
      <div class="field-label">${L.signature}</div>
      <div class="field-line"></div>
    </div>
  </div>
  <hr>
  ${mattersHtml}
</body>
</html>`
}

async function print() {
  printing.value = true
  try {
    const [mattersRes, assocRes] = await Promise.all([
      votingMatterApi.getVotingMatters(props.associationId, props.gathering.id),
      associationApi.getAssociation(props.associationId),
    ])
    const matters: VotingMatter[] = mattersRes.data
    const associationName: string = assocRes.data?.name ?? ''
    const html = buildHtml(matters, associationName)
    const win = window.open('', '_blank')
    if (!win) return
    win.document.write(html)
    win.document.close()
    win.focus()
    win.print()
  } finally {
    printing.value = false
  }
}
</script>

<template>
  <NButton :loading="printing" size="small" @click="print">
    <slot />
  </NButton>
</template>
