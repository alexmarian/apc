import { formatCurrency } from './formatters'
import jsPDF from 'jspdf'
import html2canvas from 'html2canvas'

/**
 * Improved utility for exporting charts and reports to PDF
 * Uses jsPDF and html2canvas for better quality PDFs
 */

/**
 * Convert an HTML element to a canvas for PDF generation
 */
const elementToCanvas = async (element: HTMLElement): Promise<HTMLCanvasElement> => {
  return await html2canvas(element, {
    scale: 2, // Higher scale for better quality
    useCORS: true,
    logging: false,
    allowTaint: true,
    backgroundColor: '#ffffff'
  })
}

/**
 * Create a PDF from a chart element
 * @param chartTitle Title of the chart
 * @param chartElement The DOM element containing the chart
 * @param data Optional data to include in the PDF
 * @param dateRange Optional date range string to include in the report
 */
export const exportChartToPdf = async (
  chartTitle: string,
  chartElement: HTMLElement | SVGElement | null,
  data: any[] = [],
  dateRange: string = ''
) => {
  if (!chartElement) {
    console.error('No chart element found for export')
    return
  }

  // Create PDF document (A4 size)
  const pdf = new jsPDF({
    orientation: 'portrait',
    unit: 'mm',
    format: 'a4'
  })

  // Add title
  pdf.setFontSize(18)
  pdf.text(chartTitle, 15, 15)

  // Add date information
  pdf.setFontSize(10)
  pdf.text(`Generated on: ${new Date().toLocaleString()}`, 15, 22)
  if (dateRange) {
    pdf.text(`Period: ${dateRange}`, 15, 28)
  }

  try {
    // Convert chart element to canvas
    const canvas = await elementToCanvas(chartElement as HTMLElement)

    // Calculate dimensions to fit PDF while maintaining aspect ratio
    const pageWidth = pdf.internal.pageSize.getWidth() - 30 // margins
    const imgWidth = pageWidth
    const imgHeight = (canvas.height * imgWidth) / canvas.width

    // Add chart image to PDF
    const imgData = canvas.toDataURL('image/png')
    pdf.addImage(imgData, 'PNG', 15, 35, imgWidth, imgHeight)

    // Add data table if provided
    if (data.length > 0) {
      // Move to a new page if the chart takes up most of the current page
      if (imgHeight > 180) {
        pdf.addPage()
        pdf.setFontSize(14)
        pdf.text('Data Summary', 15, 15)
        pdf.setFontSize(10)

        // Set up table
        const startY = 25
        const cellPadding = 5
        const colWidths = [70, 50, 40]
        const rowHeight = 10

        // Header row
        pdf.setFillColor(240, 240, 240)
        pdf.rect(15, startY, colWidths[0] + colWidths[1] + colWidths[2], rowHeight, 'F')
        pdf.setFont('helvetica', 'bold')
        pdf.text('Name', 15 + cellPadding, startY + rowHeight - cellPadding)
        pdf.text('Amount', 15 + colWidths[0] + cellPadding, startY + rowHeight - cellPadding)
        pdf.text('Percentage', 15 + colWidths[0] + colWidths[1] + cellPadding, startY + rowHeight - cellPadding)

        // Data rows
        pdf.setFont('helvetica', 'normal')
        let currentY = startY + rowHeight

        data.forEach((item, index) => {
          // Add new page if needed
          if (currentY > pdf.internal.pageSize.getHeight() - 20) {
            pdf.addPage()
            currentY = 15
          }

          // Alternate row background
          if (index % 2 === 0) {
            pdf.setFillColor(248, 248, 248)
            pdf.rect(15, currentY, colWidths[0] + colWidths[1] + colWidths[2], rowHeight, 'F')
          }

          // Add row data
          pdf.text(item.name.toString(), 15 + cellPadding, currentY + rowHeight - cellPadding)
          pdf.text(formatCurrency(item.value), 15 + colWidths[0] + cellPadding, currentY + rowHeight - cellPadding)
          pdf.text(item.percentage ? `${item.percentage.toFixed(1)}%` : '-',
            15 + colWidths[0] + colWidths[1] + cellPadding,
            currentY + rowHeight - cellPadding)

          currentY += rowHeight
        })
      } else {
        // Add data table below chart
        pdf.setFontSize(14)
        pdf.text('Data Summary', 15, imgHeight + 45)

        // Set up table
        const startY = imgHeight + 55
        const cellPadding = 5
        const colWidths = [70, 50, 40]
        const rowHeight = 10

        // Header row
        pdf.setFillColor(240, 240, 240)
        pdf.rect(15, startY, colWidths[0] + colWidths[1] + colWidths[2], rowHeight, 'F')
        pdf.setFont('helvetica', 'bold')
        pdf.text('Name', 15 + cellPadding, startY + rowHeight - cellPadding)
        pdf.text('Amount', 15 + colWidths[0] + cellPadding, startY + rowHeight - cellPadding)
        pdf.text('Percentage', 15 + colWidths[0] + colWidths[1] + cellPadding, startY + rowHeight - cellPadding)

        // Data rows
        pdf.setFont('helvetica', 'normal')
        let currentY = startY + rowHeight

        data.forEach((item, index) => {
          // Add new page if needed
          if (currentY > pdf.internal.pageSize.getHeight() - 20) {
            pdf.addPage()
            currentY = 15
          }

          // Alternate row background
          if (index % 2 === 0) {
            pdf.setFillColor(248, 248, 248)
            pdf.rect(15, currentY, colWidths[0] + colWidths[1] + colWidths[2], rowHeight, 'F')
          }

          // Add row data
          pdf.text(item.name.toString(), 15 + cellPadding, currentY + rowHeight - cellPadding)
          pdf.text(formatCurrency(item.value), 15 + colWidths[0] + cellPadding, currentY + rowHeight - cellPadding)
          pdf.text(item.percentage ? `${item.percentage.toFixed(1)}%` : '-',
            15 + colWidths[0] + colWidths[1] + cellPadding,
            currentY + rowHeight - cellPadding)

          currentY += rowHeight
        })
      }
    }

    // Add footer
    const pageCount = pdf.internal.getNumberOfPages()
    for (let i = 1; i <= pageCount; i++) {
      pdf.setPage(i)
      pdf.setFontSize(8)
      pdf.setTextColor(100, 100, 100)
      pdf.text(`APC Management Portal - ${chartTitle} - Page ${i} of ${pageCount}`,
        pdf.internal.pageSize.getWidth() / 2,
        pdf.internal.pageSize.getHeight() - 10,
        { align: 'center' })
    }

    // Download the PDF
    pdf.save(`${chartTitle.replace(/\s+/g, '_')}_${new Date().toISOString().split('T')[0]}.pdf`)
  } catch (error) {
    console.error('Error generating PDF:', error)
  }
}

/**
 * Export full expense report to PDF with multiple sections and charts
 * @param title Report title
 * @param summaryData Summary statistics
 * @param chartElement DOM element containing the chart
 * @param breakdownData Data tables for the report
 * @param dateRange Date range covered in the report
 */
export const exportFullReportToPdf = async (
  title: string,
  summaryData: { label: string, value: string }[],
  chartElement: HTMLElement | SVGElement | null,
  breakdownData: {
    types: { name: string, value: number, percentage: number }[],
    months: { month: string, value: number }[]
  },
  dateRange: string
) => {
  // Create PDF document (A4 size)
  const pdf = new jsPDF({
    orientation: 'portrait',
    unit: 'mm',
    format: 'a4'
  })

  // Document title
  pdf.setFontSize(20)
  pdf.setTextColor(0, 0, 0)
  pdf.text(title, 15, 15)

  // Report metadata
  pdf.setFontSize(10)
  pdf.text(`Generated on: ${new Date().toLocaleString()}`, 15, 22)
  pdf.text(`Period: ${dateRange}`, 15, 28)

  // Summary section
  pdf.setFontSize(16)
  pdf.text('Summary Statistics', 15, 38)

  // Create summary boxes
  const boxWidth = 55
  const boxHeight = 20
  const boxSpacing = 10
  let currentX = 15
  const summaryY = 45

  summaryData.forEach((item, index) => {
    // Box background
    pdf.setFillColor(248, 248, 252)
    pdf.setDrawColor(200, 200, 220)
    pdf.roundedRect(currentX, summaryY, boxWidth, boxHeight, 2, 2, 'FD')

    // Box content
    pdf.setFontSize(9)
    pdf.setTextColor(100, 100, 130)
    pdf.text(item.label, currentX + 5, summaryY + 7)

    pdf.setFontSize(12)
    pdf.setTextColor(0, 0, 0)
    pdf.text(item.value, currentX + 5, summaryY + 16)

    currentX += boxWidth + boxSpacing
  })

  let yPosition = summaryY + boxHeight + 15

  // Add chart if available
  if (chartElement) {
    try {
      pdf.setFontSize(16)
      pdf.text('Expense Distribution', 15, yPosition)
      yPosition += 10

      const canvas = await elementToCanvas(chartElement as HTMLElement)
      const pageWidth = pdf.internal.pageSize.getWidth() - 30
      const imgWidth = pageWidth
      const imgHeight = (canvas.height * imgWidth) / canvas.width

      const imgData = canvas.toDataURL('image/png')
      pdf.addImage(imgData, 'PNG', 15, yPosition, imgWidth, imgHeight)

      yPosition += imgHeight + 15
    } catch (error) {
      console.error('Error converting chart to image:', error)
    }
  }

  // Check if we need a new page
  if (yPosition > pdf.internal.pageSize.getHeight() - 60) {
    pdf.addPage()
    yPosition = 15
  }

  // Breakdown by type
  pdf.setFontSize(16)
  pdf.text('Expense Breakdown', 15, yPosition)
  yPosition += 10

  pdf.setFontSize(12)
  pdf.text('By Type', 15, yPosition)
  yPosition += 8

  // Type table
  const colWidths = [70, 50, 40]
  const rowHeight = 10
  const cellPadding = 3

  // Header row
  pdf.setFillColor(240, 240, 240)
  pdf.rect(15, yPosition, colWidths[0] + colWidths[1] + colWidths[2], rowHeight, 'F')
  pdf.setFont('helvetica', 'bold')
  pdf.setFontSize(10)
  pdf.text('Type', 15 + cellPadding, yPosition + rowHeight - cellPadding)
  pdf.text('Amount', 15 + colWidths[0] + cellPadding, yPosition + rowHeight - cellPadding)
  pdf.text('Percentage', 15 + colWidths[0] + colWidths[1] + cellPadding, yPosition + rowHeight - cellPadding)

  // Type data rows
  pdf.setFont('helvetica', 'normal')
  yPosition += rowHeight

  breakdownData.types.forEach((item, index) => {
    // Check if we need a new page
    if (yPosition > pdf.internal.pageSize.getHeight() - 20) {
      pdf.addPage()
      yPosition = 15
    }

    // Alternate row background
    if (index % 2 === 0) {
      pdf.setFillColor(248, 248, 248)
      pdf.rect(15, yPosition, colWidths[0] + colWidths[1] + colWidths[2], rowHeight, 'F')
    }

    // Add row data
    pdf.text(item.name.toString(), 15 + cellPadding, yPosition + rowHeight - cellPadding)
    pdf.text(formatCurrency(item.value), 15 + colWidths[0] + cellPadding, yPosition + rowHeight - cellPadding)
    pdf.text(`${item.percentage.toFixed(1)}%`, 15 + colWidths[0] + colWidths[1] + cellPadding, yPosition + rowHeight - cellPadding)

    yPosition += rowHeight
  })

  // Add some spacing
  yPosition += 10

  // Check if we need a new page
  if (yPosition > pdf.internal.pageSize.getHeight() - 60) {
    pdf.addPage()
    yPosition = 15
  }

  // Monthly breakdown
  pdf.setFontSize(12)
  pdf.text('By Month', 15, yPosition)
  yPosition += 8

  // Month table
  const monthColWidths = [70, 70]

  // Header row
  pdf.setFillColor(240, 240, 240)
  pdf.rect(15, yPosition, monthColWidths[0] + monthColWidths[1], rowHeight, 'F')
  pdf.setFont('helvetica', 'bold')
  pdf.text('Month', 15 + cellPadding, yPosition + rowHeight - cellPadding)
  pdf.text('Amount', 15 + monthColWidths[0] + cellPadding, yPosition + rowHeight - cellPadding)

  // Month data rows
  pdf.setFont('helvetica', 'normal')
  yPosition += rowHeight

  breakdownData.months.forEach((item, index) => {
    // Check if we need a new page
    if (yPosition > pdf.internal.pageSize.getHeight() - 20) {
      pdf.addPage()
      yPosition = 15
    }

    // Alternate row background
    if (index % 2 === 0) {
      pdf.setFillColor(248, 248, 248)
      pdf.rect(15, yPosition, monthColWidths[0] + monthColWidths[1], rowHeight, 'F')
    }

    // Add row data
    pdf.text(item.month, 15 + cellPadding, yPosition + rowHeight - cellPadding)
    pdf.text(formatCurrency(item.value), 15 + monthColWidths[0] + cellPadding, yPosition + rowHeight - cellPadding)

    yPosition += rowHeight
  })

  // Add footer
  const pageCount = pdf.internal.getNumberOfPages()
  for (let i = 1; i <= pageCount; i++) {
    pdf.setPage(i)
    pdf.setFontSize(8)
    pdf.setTextColor(100, 100, 100)
    pdf.text(`APC Management Portal - Expense Report - Page ${i} of ${pageCount}`,
      pdf.internal.pageSize.getWidth() / 2,
      pdf.internal.pageSize.getHeight() - 10,
      { align: 'center' })
  }

  // Download the PDF
  pdf.save(`${title.replace(/\s+/g, '_')}_${new Date().toISOString().split('T')[0]}.pdf`)
}
