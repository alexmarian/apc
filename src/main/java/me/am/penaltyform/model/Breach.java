package me.am.penaltyform.model;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class Breach {

  private Long id;
  private String code;
  private String description;
}