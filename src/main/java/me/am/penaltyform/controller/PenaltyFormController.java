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
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.ResponseBody;

@Controller
public class PenaltyFormController {

  private final PdfGenerationService pdfGenerationService;
  private final ObjectMapper objectMapper;

  @Autowired
  public PenaltyFormController(PdfGenerationService pdfGenerationService, ObjectMapper objectMapper) {
    this.pdfGenerationService = pdfGenerationService;
    this.objectMapper = objectMapper;
  }

  @GetMapping("/")
  public String showForm(Model model) {
    // Create an empty form object to avoid null reference in the template
    model.addAttribute("penaltyForm", new PenaltyForm());
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

  @GetMapping("/api/calculate")
  @ResponseBody
  public Double calculatePenalty(@RequestParam("baseAmount") Double baseAmount,
      @RequestParam("occurrenceCount") Integer occurrenceCount) {
    // Conform articolului 3.5 din Regulament
    if (occurrenceCount == 1) {
      return baseAmount;
    } else if (occurrenceCount == 2) {
      // Al doilea incident în 12 luni - doar avertisment, dar vom păstra amenda inițială
      return baseAmount;
    } else if (occurrenceCount == 3) {
      // Al treilea incident în 12 luni - penalitatea inițială majorată cu 50%
      return baseAmount * 1.5;
    } else {
      // Al patrulea incident și ulterioarele - penalitatea inițială dublată
      return baseAmount * 2.0;
    }
  }
}