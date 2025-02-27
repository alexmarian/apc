package me.am.penaltyform.controller;

import me.am.penaltyform.model.Breach;
import me.am.penaltyform.model.Penalty;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.math.BigDecimal;
import java.util.Arrays;
import java.util.List;

@RestController
@RequestMapping("/api")
public class ApiController {

  @GetMapping("/breaches")
  public List<Breach> getBreaches() {
    return Arrays.asList(new Breach(1L, "B001", "Speeding (below 10 mph over limit)"),
        new Breach(2L, "B002", "Speeding (10-20 mph over limit)"),
        new Breach(3L, "B003", "Speeding (above 20 mph over limit)"),
        new Breach(4L, "B004", "Running a red light"), new Breach(5L, "B005", "Illegal parking"),
        new Breach(6L, "B006", "Driving without a valid license"),
        new Breach(7L, "B007", "Failure to yield right of way"));
  }

  @GetMapping("/penalties")
  public List<Penalty> getPenalties() {
    return Arrays.asList(new Penalty(1L, "P001", "Warning - No Fine", BigDecimal.ZERO),
        new Penalty(2L, "P002", "Minor Penalty - $50 Fine", new BigDecimal("50.00")),
        new Penalty(3L, "P003", "Standard Penalty - $100 Fine", new BigDecimal("100.00")),
        new Penalty(4L, "P004", "Major Penalty - $200 Fine", new BigDecimal("200.00")),
        new Penalty(5L, "P005", "Severe Penalty - $500 Fine", new BigDecimal("500.00")));
  }
}
