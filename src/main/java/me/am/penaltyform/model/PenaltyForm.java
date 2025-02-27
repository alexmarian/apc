package me.am.penaltyform.model;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class PenaltyForm {

  @NotBlank(message = "Name is required")
  private String firstName;

  @NotBlank(message = "Surname is required")
  private String lastName;

  @NotNull(message = "At least one breach must be selected")
  private Breach selectedBreach;

  @NotNull(message = "A penalty must be selected")
  private Penalty selectedPenalty;

}
