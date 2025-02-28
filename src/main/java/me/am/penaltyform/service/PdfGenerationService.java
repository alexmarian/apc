package me.am.penaltyform.service;

import me.am.penaltyform.model.PenaltyForm;
import org.springframework.stereotype.Service;
import org.xhtmlrenderer.pdf.ITextRenderer;
import org.thymeleaf.TemplateEngine;
import org.thymeleaf.context.Context;
import org.springframework.beans.factory.annotation.Autowired;
import com.lowagie.text.pdf.BaseFont;

import java.io.ByteArrayOutputStream;
import java.io.File;
import java.time.LocalDate;
import java.time.format.DateTimeFormatter;

@Service
public class PdfGenerationService {

  private final TemplateEngine templateEngine;

  @Autowired
  public PdfGenerationService(TemplateEngine templateEngine) {
    this.templateEngine = templateEngine;
  }

  private void setupFonts(ITextRenderer renderer) throws Exception {
    // Option 1: Use a font file with good Unicode support
    String fontPath = "./fonts/LiberationSans-Regular.ttf"; // Adjust path as needed
    renderer.getFontResolver().addFont(
        fontPath,
        BaseFont.IDENTITY_H,
        BaseFont.EMBEDDED
    );

    // Option 2: Add multiple font files for different styles if needed
    // renderer.getFontResolver().addFont("/path/to/LiberationSans-Bold.ttf", BaseFont.IDENTITY_H, BaseFont.EMBEDDED);
  }

  public byte[] generatePenaltyPdf(PenaltyForm penaltyForm) {
    try {
      // Penalty calculation logic (unchanged)
      if (penaltyForm.getCalculatedPenalty() == null && penaltyForm.getSelectedBreach() != null && penaltyForm.getOccurrenceCount() != null) {
        Double baseAmount = penaltyForm.getSelectedBreach().getBaseAmount();
        Integer occurrenceCount = penaltyForm.getOccurrenceCount();

        if (occurrenceCount == 1) {
          penaltyForm.setCalculatedPenalty(baseAmount);
        } else if (occurrenceCount == 2) {
          penaltyForm.setCalculatedPenalty(baseAmount);
        } else if (occurrenceCount == 3) {
          penaltyForm.setCalculatedPenalty(baseAmount * 1.5);
        } else {
          penaltyForm.setCalculatedPenalty(baseAmount * 2.0);
        }
      }

      // Prepare context
      Context context = new Context();
      context.setVariable("penaltyForm", penaltyForm);
      context.setVariable("today", LocalDate.now().format(DateTimeFormatter.ofPattern("dd/MM/yyyy")));

      // Process template
      String htmlContent = templateEngine.process("pdf-template", context);

      // Ensure proper XML declaration with UTF-8 encoding
      htmlContent = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" + htmlContent;

      // Generate PDF with proper font encoding
      ByteArrayOutputStream baos = new ByteArrayOutputStream();
      ITextRenderer renderer = new ITextRenderer();

      // Setup fonts before rendering
      setupFonts(renderer);

      renderer.setDocumentFromString(htmlContent);
      renderer.layout();
      renderer.createPDF(baos);

      return baos.toByteArray();
    } catch (Exception e) {
      e.printStackTrace();
      return null;
    }
  }
}