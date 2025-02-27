# Penalty Form Application

A Spring Boot application with Vue.js frontend for creating and downloading penalty forms as PDFs.

## Features

- Modern SPA interface built with Vue.js
- Form for collecting offender information
- Selection from predefined lists of breaches and penalties
- Real-time form validation
- Form preview before PDF generation
- PDF generation and download functionality

## Technologies Used

- **Backend**:
    - Java 17
    - Spring Boot 3.2.0
    - Thymeleaf for PDF templates
    - Flying Saucer PDF (OpenPDF) for PDF generation

- **Frontend**:
    - Vue.js 2.6 (CDN version, no build step required)
    - Axios for API calls
    - Bootstrap 5 for styling

## Project Structure

```
src/
├── main/
│   ├── java/
│   │   └── me/
│   │       └── am/
│   │           └── penaltyform/
│   │               ├── PenaltyFormApplication.java
│   │               ├── controller/
│   │               │   ├── ApiController.java
│   │               │   └── PenaltyFormController.java
│   │               ├── model/
│   │               │   ├── Breach.java
│   │               │   ├── Penalty.java
│   │               │   └── PenaltyForm.java
│   │               └── service/
│   │                   └── PdfGenerationService.java
│   └── resources/
│       ├── application.properties
│       └── templates/
│           ├── index.html
│           └── pdf-template.html
```

## Getting Started

### Prerequisites

- Java 17 or higher
- Maven 3.6 or higher

### Building the Application

1. Clone the repository:

```bash
git clone https://github.com/yourusername/penalty-form.git
cd penalty-form
```

2. Build the application:

```bash
mvn clean install
```

### Running the Application

```bash
mvn spring-boot:run
```

The application will be available at http://localhost:8080

## How It Works

1. **User Form Input**: The user fills out the form with the person's name, surname, and selects a breach and penalty from predefined lists.

2. **Form Validation**: Vue.js provides real-time validation to ensure all required fields are completed.

3. **Preview**: Before generating the PDF, users can preview the filled-out form and make adjustments if needed.

4. **PDF Generation**: When the user confirms the information, the application:
    - Sends form data to the backend
    - Processes data with the PdfGenerationService
    - Uses Thymeleaf to render an HTML template with the data
    - Converts the HTML to a PDF using Flying Saucer
    - Returns the PDF to the browser for download

## Customization

### Adding New Breaches or Penalties

To add new breach types or penalties, modify the `getBreaches()` and `getPenalties()` methods in the `ApiController` class.

### Customizing the PDF Template

The PDF appearance is controlled by the `pdf-template.html` file. You can modify the HTML and CSS in this file to change the appearance of the generated PDF.

### Extending the Form

To add more fields to the form:

1. Update the `PenaltyForm` class with new fields
2. Add corresponding form inputs in the Vue.js app in `index.html`
3. Update the `pdf-template.html` to include the new fields in the PDF output

## API Endpoints

- `GET /api/breaches` - Returns the list of possible breaches
- `GET /api/penalties` - Returns the list of possible penalties
- `POST /generate-pdf` - Accepts form data and returns a PDF document

## License

This project is licensed under the MIT License - see the LICENSE file for details.