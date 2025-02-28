package me.am.penaltyform.controller;

import java.util.Arrays;
import java.util.List;
import me.am.penaltyform.model.Breach;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api")
public class ApiController {

  @GetMapping("/breaches")
  public List<Breach> getBreaches() {
    return Arrays.asList(
        // 5.1. Utilizarea spațiilor comune și siguranța
        new Breach(1L, "B5.1.1", "Utilizarea neautorizată a înregistrărilor video", 1000.0,
            "5.1.1"),
        new Breach(2L, "B5.1.2", "Depășirea vitezei sau claxonatul nejustificat", 500.0, "5.1.2"),
        new Breach(3L, "B5.1.3", "Depozitarea substanțelor periculoase", 1500.0, "5.1.3"),
        new Breach(4L, "B5.1.4", "Lansarea focului de artificii", 600.0, "5.1.4"),
        new Breach(5L, "B5.1.5", "Intervenția neautorizată asupra instalațiilor inginerești",
            1000.0, "5.1.5"),
        new Breach(6L, "B5.1.6", "Organizarea comerțului stradal neautorizat", 1000.0, "5.1.6"),

        // 5.2. Reglementări privind animalele de companie
        new Breach(7L, "B5.2.1", "Omiterea colectării dejecțiilor animale", 500.0, "5.2.1"),

        // 5.3. Menținerea curățeniei și colectarea deșeurilor
        new Breach(8L, "B5.3.1", "Aruncarea gunoiului menajer în locuri neamenajate", 1000.0,
            "5.3.1"),
        new Breach(9L, "B5.3.2", "Aruncarea gunoaielor în spațiile comune", 1500.0, "5.3.2"),
        new Breach(10L, "B5.3.3", "Lăsarea pungilor cu gunoi menajer pe coridor", 1500.0, "5.3.3"),
        new Breach(11L, "B5.3.4", "Aruncarea țigărilor în afara urnelor de gunoi", 1000.0, "5.3.4"),
        new Breach(12L, "B5.3.5", "Evacuarea deșeurilor nepermise în sistemul de canalizare",
            1500.0, "5.3.5"),
        new Breach(13L, "B5.3.6a", "Aruncarea gunoiului nemenajer în tomberoane", 1500.0, "5.3.6"),
        new Breach(14L, "B5.3.6b", "Aruncarea gunoiului nemenajer (volum mare)", 2500.0, "5.3.6"),
        new Breach(15L, "B5.3.7", "Exercitarea necesităților fiziologice în zone interzise", 2000.0,
            "5.3.7"),

        // 5.4. Restricții privind fumatul
        new Breach(16L, "B5.4.1", "Fumatul în zonele interzise", 2000.0, "5.4.1"),

        // 5.5. Liniștea și controlul zgomotelor
        new Breach(17L, "B5.5.1", "Tulburarea liniștii în timpul nopții", 1000.0, "5.5.1"),
        new Breach(18L, "B5.5.2", "Efectuarea lucrărilor cu zgomot în afara orelor permise", 1000.0,
            "5.5.2"),

        // 5.6. Utilizarea lifturilor
        new Breach(19L, "B5.6.1", "Folosirea lifturilor mici pentru materiale de construcție",
            2500.0, "5.6.1"),

        // 5.7. Reglementări privind fațada
        new Breach(20L, "B5.7.1", "Instalarea tijelor pentru uscarea rufelor", 2500.0, "5.7.1"),
        new Breach(21L, "B5.7.2", "Instalarea neautorizată a antenelor", 2500.0, "5.7.2"),
        new Breach(22L, "B5.7.3", "Montarea aparatelor de climatizare cu traseu vizibil", 2500.0,
            "5.7.3"),
        new Breach(23L, "B5.7.4", "Montarea plaselor pentru insecte de culoare nepermisă", 1000.0,
            "5.7.4"),
        new Breach(24L, "B5.7.5", "Instalarea neautorizată de obiecte pe fațadă", 2500.0, "5.7.5"),

        // 5.8. Reglementări privind parcarea
        new Breach(25L, "B5.8.1", "Parcarea în afara locurilor amenajate", 500.0, "5.8.1"),
        new Breach(26L, "B5.8.2", "Parcarea peste marcajul rutier", 500.0, "5.8.2"),
        new Breach(27L, "B5.8.3", "Parcarea îndelungată (peste 3 săptămâni)", 750.0, "5.8.3"),
        new Breach(28L, "B5.8.4", "Parcarea pe capacele de canalizare/apeduct", 500.0, "5.8.4"),
        new Breach(29L, "B5.8.5", "Parcarea blocând accesul", 1000.0, "5.8.5"),
        new Breach(30L, "B5.8.6", "Scurgeri de uleiuri/lichide din automobil", 1000.0, "5.8.6"),

        // 5.9. Reguli speciale în perioada lucrărilor de reparație
        new Breach(31L, "B5.9.3", "Utilizarea liftului de pasageri pentru materiale de construcție",
            2500.0, "5.9.3"),
        new Breach(32L, "B5.9.4", "Depozitarea materialelor/deșeurilor în LUC", 1500.0, "5.9.4"),
        new Breach(33L, "B5.9.5", "Nesalubrizarea în urma transportării materialelor", 1500.0,
            "5.9.5"),
        new Breach(34L, "B5.9.6", "Afectarea integrității/rezistenței blocului", 7500.0, "5.9.6"),
        new Breach(35L, "B5.9.8", "Conectarea hotelor în canalul de ventilare comun", 2500.0,
            "5.9.8"));
  }

  @GetMapping("/occurrences")
  public List<Integer> getOccurrences() {
    return Arrays.asList(1, 2, 3, 4, 5);
  }

  @GetMapping("/calculate-penalty")
  public double calculatePenalty(double baseAmount, int occurrenceCount) {
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