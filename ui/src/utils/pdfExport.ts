import { formatCurrency } from './formatters'

/**
 * Utility for exporting charts and reports to PDF
 *
 * This implementation uses browser's built-in functionality to print to PDF
 * For a more advanced solution, you could integrate a library like jsPDF or html2pdf
 */

/**
 * Create a printable document with the chart SVG and data
 * @param chartTitle Title of the chart
 * @param svgElement The SVG element to export
 * @param data Optional data to include in the PDF
 * @param dateRange Optional date range string to include in the report
 */
export const exportChartToPdf = (
  chartTitle: string,
  svgElement: SVGElement | null,
  data: any[] = [],
  dateRange: string = ''
) => {
  if (!svgElement) {
    console.error('No SVG element found for export')
    return
  }

  // Create a new window for printing
  const printWindow = window.open('', '_blank')
  if (!printWindow) {
    alert('Please allow pop-ups to export charts as PDF')
    return
  }

  // Clone the SVG to avoid modifying the original
  const svgClone = svgElement.cloneNode(true) as SVGElement

  // Set width and height attributes explicitly for better rendering
  svgClone.setAttribute('width', '600')
  svgClone.setAttribute('height', '400')

  // Convert SVG to a data URL
  const svgData = new XMLSerializer().serializeToString(svgClone)
  const svgBlob = new Blob([svgData], { type: 'image/svg+xml' })
  const svgUrl = URL.createObjectURL(svgBlob)

  // Generate the HTML content for the print window
  const htmlContent = `
    <!DOCTYPE html>
    <html>
    <head>
      <title>${chartTitle} - Export</title>
      <style>
        body {
          font-family: Arial, sans-serif;
          margin: 20px;
          color: #333;
        }
        h1 {
          font-size: 20px;
          margin-bottom: 10px;
        }
        h2 {
          font-size: 16px;
          margin-top: 20px;
          margin-bottom: 10px;
        }
        .chart-container {
          margin: 20px 0;
          text-align: center;
        }
        .chart-image {
          max-width: 100%;
          height: auto;
        }
        table {
          width: 100%;
          border-collapse: collapse;
          margin-top: 20px;
        }
        th, td {
          border: 1px solid #ddd;
          padding: 8px;
          text-align: left;
        }
        th {
          background-color: #f2f2f2;
        }
        .report-header {
          margin-bottom: 20px;
        }
        .report-footer {
          margin-top: 30px;
          font-size: 12px;
          color: #666;
          text-align: center;
        }
        @media print {
          body {
            padding: 0;
            margin: 0;
          }
        }
      </style>
    </head>
    <body>
      <div class="report-header">
        <h1>${chartTitle}</h1>
        ${dateRange ? `<p>Period: ${dateRange}</p>` : ''}
        <p>Generated on: ${new Date().toLocaleString()}</p>
      </div>

      <div class="chart-container">
        <img src="${svgUrl}" class="chart-image" alt="${chartTitle}" />
      </div>

      ${data.length > 0 ? `
        <h2>Data Summary</h2>
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Amount</th>
              <th>Percentage</th>
            </tr>
          </thead>
          <tbody>
            ${data.map(item => `
              <tr>
                <td>${item.name}</td>
                <td>${formatCurrency(item.value)}</td>
                <td>${item.percentage ? item.percentage.toFixed(1) + '%' : '-'}</td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      ` : ''}

      <div class="report-footer">
        <p>APC Management Portal - Expense Report</p>
      </div>

      <script>
        // Auto-print when loaded
        window.onload = function() {
          // Short delay to ensure everything is rendered
          setTimeout(() => {
            window.print();
            // Close the window after printing (or after print is cancelled)
            window.addEventListener('afterprint', function() {
              window.close();
            });
          }, 500);
        };
      </script>
    </body>
    </html>
  `

  // Write HTML to the new window and trigger printing
  printWindow.document.open()
  printWindow.document.write(htmlContent)
  printWindow.document.close()
}

/**
 * Export full expense report to PDF
 * @param title Report title
 * @param summaryData Summary statistics
 * @param chartsSvg SVG elements of charts
 * @param breakdownData Data tables for the report
 * @param dateRange Date range covered in the report
 */
export const exportFullReportToPdf = (
  title: string,
  summaryData: { label: string, value: string }[],
  chartsSvg: SVGElement | null,
  breakdownData: {
    types: { name: string, value: number, percentage: number }[],
    months: { month: string, value: number }[]
  },
  dateRange: string
) => {
  // Create a new window for printing
  const printWindow = window.open('', '_blank')
  if (!printWindow) {
    alert('Please allow pop-ups to export reports as PDF')
    return
  }

  // Convert chart SVG to a data URL if available
  let chartUrl = ''
  if (chartsSvg) {
    const svgClone = chartsSvg.cloneNode(true) as SVGElement
    svgClone.setAttribute('width', '600')
    svgClone.setAttribute('height', '400')

    const svgData = new XMLSerializer().serializeToString(svgClone)
    const svgBlob = new Blob([svgData], { type: 'image/svg+xml' })
    chartUrl = URL.createObjectURL(svgBlob)
  }

  // Generate the HTML content for the print window
  const htmlContent = `
    <!DOCTYPE html>
    <html>
    <head>
      <title>${title}</title>
      <style>
        body {
          font-family: Arial, sans-serif;
          margin: 20px;
          color: #333;
        }
        h1 {
          font-size: 22px;
          margin-bottom: 10px;
        }
        h2 {
          font-size: 18px;
          margin-top: 25px;
          margin-bottom: 10px;
        }
        h3 {
          font-size: 16px;
          margin-top: 20px;
          margin-bottom: 10px;
        }
        .report-header {
          margin-bottom: 20px;
        }
        .summary-section {
          display: flex;
          justify-content: space-between;
          margin-bottom: 30px;
          flex-wrap: wrap;
        }
        .summary-box {
          border: 1px solid #ddd;
          padding: 15px;
          border-radius: 5px;
          width: calc(33% - 20px);
          min-width: 200px;
          margin-bottom: 10px;
          box-sizing: border-box;
        }
        .summary-label {
          font-size: 14px;
          color: #666;
        }
        .summary-value {
          font-size: 20px;
          font-weight: bold;
          margin-top: 5px;
        }
        .chart-container {
          margin: 20px 0;
          text-align: center;
          page-break-inside: avoid;
        }
        .chart-image {
          max-width: 100%;
          height: auto;
        }
        table {
          width: 100%;
          border-collapse: collapse;
          margin-top: 15px;
          page-break-inside: avoid;
        }
        th, td {
          border: 1px solid #ddd;
          padding: 8px;
          text-align: left;
        }
        th {
          background-color: #f2f2f2;
        }
        .breakdown-section {
          margin-top: 30px;
          page-break-before: always;
        }
        .report-footer {
          margin-top: 30px;
          font-size: 12px;
          color: #666;
          text-align: center;
          page-break-before: always;
        }
        @media print {
          body {
            padding: 0;
            margin: 0;
          }
          .page-break {
            page-break-before: always;
          }
        }
      </style>
    </head>
    <body>
      <div class="report-header">
        <h1>${title}</h1>
        <p>Period: ${dateRange}</p>
        <p>Generated on: ${new Date().toLocaleString()}</p>
      </div>

      <!-- Summary Section -->
      <h2>Summary Statistics</h2>
      <div class="summary-section">
        ${summaryData.map(item => `
          <div class="summary-box">
            <div class="summary-label">${item.label}</div>
            <div class="summary-value">${item.value}</div>
          </div>
        `).join('')}
      </div>

      <!-- Chart Section -->
      ${chartUrl ? `
        <div class="chart-container">
          <h2>Expense Distribution</h2>
          <img src="${chartUrl}" class="chart-image" alt="Expense Chart" />
        </div>
      ` : ''}

      <!-- Breakdown Section -->
      <div class="breakdown-section">
        <h2>Expense Breakdown</h2>

        <h3>By Type</h3>
        <table>
          <thead>
            <tr>
              <th>Type</th>
              <th>Amount</th>
              <th>Percentage</th>
            </tr>
          </thead>
          <tbody>
            ${breakdownData.types.map(item => `
              <tr>
                <td>${item.name}</td>
                <td>${formatCurrency(item.value)}</td>
                <td>${item.percentage.toFixed(1)}%</td>
              </tr>
            `).join('')}
          </tbody>
        </table>

        <h3>By Month</h3>
        <table>
          <thead>
            <tr>
              <th>Month</th>
              <th>Amount</th>
            </tr>
          </thead>
          <tbody>
            ${breakdownData.months.map(item => `
              <tr>
                <td>${item.month}</td>
                <td>${formatCurrency(item.value)}</td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      </div>

      <div class="report-footer">
        <p>APC Management Portal - Expense Report</p>
        <p>This report is automatically generated and does not require a signature.</p>
      </div>

      <script>
        // Auto-print when loaded
        window.onload = function() {
          // Short delay to ensure everything is rendered
          setTimeout(() => {
            window.print();
            // Close the window after printing (or after print is cancelled)
            window.addEventListener('afterprint', function() {
              window.close();
            });
          }, 500);
        };
      </script>
    </body>
    </html>
  `

  // Write HTML to the new window and trigger printing
  printWindow.document.open()
  printWindow.document.write(htmlContent)
  printWindow.document.close()
}
