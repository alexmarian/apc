package me.am.penaltyform.model;

import java.math.BigDecimal;
import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class Penalty {

  private Long id;
  private String code;
  private String description;
  private BigDecimal amount;

}
