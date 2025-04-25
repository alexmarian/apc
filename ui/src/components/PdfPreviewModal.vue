<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { NModal, NSpin, NButton, NSpace, NProgress, useMessage } from 'naive-ui'
import jsPDF from 'jspdf'
import html2canvas from 'html2canvas'

// Message system
const message = useMessage()

// Props
const props = defineProps<{
  show: boolean
  title: string
  contentElement?: HTMLElement | null
  onGenerate?: () => void
}>()

// Emits
const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'generated'): void
}>()

// State
const previewSrc = ref<string>('')
const loading = ref(false)
const progress = ref(0)
const error = ref<string | null>(null)

// Generate PDF preview
const generatePreview = async () => {
  if (!props.contentElement) {
    error.value = 'No content to preview'
    return
  }

  try {
    loading.value = true
    error.value = null
    progress.value = 10

    // Create PDF using jsPDF
    const pdf = new jsPDF({
      orientation: 'portrait',
      unit: 'mm',
      format: 'a4'
    })

    // Start progress
    progress.value = 20

    // Capture content as image
    const canvas = await html2canvas(props.contentElement, {
      scale: 2,
      useCORS: true,
      logging: false,
      allowTaint: true,
      backgroundColor: '#ffffff'
    })

    progress.value = 70

    // Calculate dimensions
    const pageWidth = pdf.internal.pageSize.getWidth() - 20
    const imgWidth = pageWidth
    const ratio = canvas.height / canvas.width
    const imgHeight = pageWidth * ratio

    // Add title
    pdf.setFontSize(16)
    pdf.text(props.title, 10, 10)

    // Add date
    pdf.setFontSize(10)
    pdf.text(`Generated preview on: ${new Date().toLocaleString()}`, 10, 18)

    // Add image
    const imgData = canvas.toDataURL('image/png')
    pdf.addImage(imgData, 'PNG', 10, 25, imgWidth, imgHeight)

    // Add footer
    pdf.setFontSize(9)
    pdf.setTextColor(100, 100, 100)
    pdf.text(
      `APC Management Portal - Preview`,
      pdf.internal.pageSize.getWidth() / 2,
      pdf.internal.pageSize.getHeight() - 10,
      { align: 'center' }
    )

    progress.value = 90

    // Convert to data URL for preview
    const pdfDataUrl = pdf.output('datauristring')
    previewSrc.value = pdfDataUrl

    progress.value = 100
  } catch (err) {
    console.error('Error generating preview:', err)
    error.value = 'Failed to generate preview. Please try again.'
  } finally {
    loading.value = false
  }
}

// Handle download
const handleDownload = () => {
  if (props.onGenerate) {
    props.onGenerate()
  }
  emit('generated')
  handleClose()
  message.success('PDF export successfully completed')
}

// Close modal
const handleClose = () => {
  emit('update:show', false)
}

// Watch for show changes
watch(() => props.show, async (show) => {
  if (show) {
    await generatePreview()
  }
})

// Generate preview on mount if modal is open
onMounted(async () => {
  if (props.show) {
    await generatePreview()
  }
})
</script>

<template>
  <NModal
    v-model:show="props.show"
    style="width: 800px; max-width: 90vw"
    preset="card"
    :title="'Preview: ' + props.title"
    :closable="true"
    @close="handleClose"
  >
    <NSpin :show="loading">
      <NProgress
        v-if="loading"
        type="line"
        :percentage="progress"
        :show-indicator="true"
        processing
      />

      <div v-if="error" class="error">{{ error }}</div>

      <div v-else-if="previewSrc" class="preview-container">
        <iframe
          :src="previewSrc"
          class="pdf-preview"
          frameborder="0"
        ></iframe>
      </div>

      <div v-else-if="!loading" class="error">
        No preview available. Please try again.
      </div>
    </NSpin>

    <template #footer>
      <NSpace justify="end">
        <NButton @click="handleClose">Cancel</NButton>
        <NButton type="primary" @click="handleDownload" :loading="loading">
          Download PDF
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped>
.preview-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-height: 60vh;
}

.pdf-preview {
  width: 100%;
  height: 60vh;
  border: 1px solid #eee;
  margin-top: 16px;
  margin-bottom: 16px;
}

.error {
  color: #e03;
  text-align: center;
  padding: 2rem;
}
</style>
