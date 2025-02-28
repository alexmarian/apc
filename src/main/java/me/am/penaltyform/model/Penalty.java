package me.am.penaltyform.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class Penalty {
  private Long id;
  private String code;
  private String description;
  private BigDecimal amount;
}