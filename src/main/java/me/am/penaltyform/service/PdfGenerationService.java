package me.am.penaltyform.service;

import me.am.penaltyform.model.PenaltyForm;
import org.springframework.stereotype.Service;
import org.xhtmlrenderer.pdf.ITextRenderer;
import org.thymeleaf.TemplateEngine;
import org.thymeleaf.context.Context;
import org.springframework.beans.factory.annotation.Autowired;

import java.io.ByteArrayOutputStream;
import java.time.LocalDate;
import java.time.format.DateTimeFormatter;

@Service
public class PdfGenerationService {

  private final TemplateEngine templateEngine;

  @Autowired
  public PdfGenerationService(TemplateEngine templateEngine) {
    this.templateEngine = templateEngine;
  }

  public byte[] generatePenaltyPdf(PenaltyForm penaltyForm) {
    try {
      // Prepare context
      Context context = new Context();
      context.setVariable("penaltyForm", penaltyForm);
      context.setVariable("today", LocalDate.now().format(DateTimeFormatter.ofPattern("dd/MM/yyyy")));

      // Process template
      String htmlContent = templateEngine.process("pdf-template", context);

      // Ensure proper XML declaration for XHTML
      htmlContent = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" + htmlContent;

      // Generate PDF
      ByteArrayOutputStream baos = new ByteArrayOutputStream();
      ITextRenderer renderer = new ITextRenderer();
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
