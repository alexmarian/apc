package me.am.penaltyform.model;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import java.time.LocalDate;
import lombok.Data;

import java.util.List;

@Data
public class PenaltyForm {

  @NotBlank(message = "Name is required")
  private String firstName;

  @NotBlank(message = "Surname is required")
  private String lastName;

  @NotBlank(message = "Unit is required")
  private String unit;

  @NotNull(message = "At least one breach must be selected")
  private Breach selectedBreach;

  @NotNull(message = "Breach date is required")
  private LocalDate breachDate;

  private Integer occurrenceCount;

  @NotBlank(message = "Context information is required")
  private String contextInformation;

  private List<String> evidenceMaterials;

  private Double calculatedPenalty;
}
