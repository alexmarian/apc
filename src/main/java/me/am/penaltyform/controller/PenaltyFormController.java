package me.am.penaltyform.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import me.am.penaltyform.model.PenaltyForm;
import me.am.penaltyform.service.PdfGenerationService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.*;

@Controller
public class PenaltyFormController {

  private final PdfGenerationService pdfGenerationService;
  private final ObjectMapper objectMapper;

  @Autowired
  public PenaltyFormController(PdfGenerationService pdfGenerationService,
      ObjectMapper objectMapper) {
    this.pdfGenerationService = pdfGenerationService;
    this.objectMapper = objectMapper;
  }

  @GetMapping("/")
  public String showForm() {
    return "index";
  }

  @PostMapping("/generate-pdf")
  public ResponseEntity<byte[]> generatePdf(@RequestBody String formData) {
    try {
      // Convert JSON to PenaltyForm object
      PenaltyForm penaltyForm = objectMapper.readValue(formData, PenaltyForm.class);

      // Generate PDF
      byte[] pdfContent = pdfGenerationService.generatePenaltyPdf(penaltyForm);

      // Set HTTP headers
      HttpHeaders headers = new HttpHeaders();
      headers.setContentType(MediaType.APPLICATION_PDF);
      headers.setContentDispositionFormData("attachment", "penalty_form.pdf");

      return new ResponseEntity<>(pdfContent, headers, HttpStatus.OK);
    } catch (Exception e) {
      e.printStackTrace();
      return new ResponseEntity<>(HttpStatus.INTERNAL_SERVER_ERROR);
    }
  }
}