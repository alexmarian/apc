package me.am.penaltyform.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class Breach {
  private Long id;
  private String code;
  private String description;
  private Double baseAmount;
  private String regulationReference;
}