![Go](https://img.shields.io/badge/Go-1.24+-blue)
![Docker](https://img.shields.io/badge/Docker-ready-blue)
![License](https://img.shields.io/badge/License-MIT-green)

# Mortgage Calculator

A simple web-based mortgage calculator built with **Go**, **HTML templates**, and **Docker**.  
It calculates a fixed-period mortgage amortization schedule, displays the breakdown in the browser, and allows downloading a **PDF report** with the results.

---

## 📌 Features

- Enter:
  - Mortgage principal
  - Annual interest rate (percentage)
  - Fixed period in months
  - Monthly payment amount
- Displays a full amortization table (interest, principal & balance)
- Shows mortgage summary totals
- Generates a downloadable PDF report
- Works in any modern browser
- Docker ready

---

## 🧠 How It Works

### Main Components

1. **`browser.go`**
   - HTTP web server handling:
     - GET `/` → serve form
     - POST `/` → compute mortgage and render results
     - POST `/download-pdf` → generate and serve PDF
   - Defines types:
     - `Mortgage` — inputs
     - `MonthlyData` — monthly amortization
     - `PageData` — composite DTO passed to templates

2. **`GenerateMonthlyBreakdown`**
   - Given a mortgage with principal, annual rate (%) and fixed months
   - Converts annual rate to **monthly rate**
   - Computes interest and principal for each month
   - Produces slice of amortization rows

3. **HTML Template (`index.html`)**
   - Form for inputs
   - Summary and results table rendered with `html/template`
   - Hidden form to send same data to PDF endpoint

4. **PDF Generation**
   - Uses the `gofpdf` library to generate an in-memory PDF
   - Writes the mortgage summary and full monthly table

---

## 🚀 Running Locally (without Docker)

### Prerequisites

Make sure Go is installed:

```bash
go version
# Example: go version go1.24.3
```

### Install Dependencies

```bash
go mod download
```

### Run the Server
```bash
go run browser.go
```

### Open the browser
```bash
http://localhost:8080
```

## 🐳 Running With Docker (Recommended)

### Build the Docker image

From the project root:
```bash
docker build -t mortgage_calculator .
```

### Run the Container (Expose Port 8080)
```bash
docker run --rm -p 8080:8080 mortgage_calculator
```

### Access the application
```bash
http://localhost:8080
```
This allows the containerized Go application to be accessed directly from your local browser.

## 🧪 How to Use the Application

1. Open `http://localhost:8080`
2. Enter:
    - Principal (e.g. 360000)
    - Annual interest rate (e.g. 5.25)
    - Fixed rate period in months (e.g. 24)
    - Monthly payment (e.g. 3200.99)
3. Click **Calculate**
4. Review:
    - Mortgage summary
    - Monthly breakdown table
5. Click **Download PDF** to export the report

## 📄 Notes About PDF Output

- Floating-point values are formatted to two decimal places
- Currency symbols require UTF-8 compatible fonts
- PDF values exactly match the HTML calculations

## 🔧 Potential Improvements

- Input validation and error handling
- Auto-calculate monthly payment
- Thousands separators (e.g. £333,802.37)
- Improved mobile responsiveness
- Paginated PDF tables for long mortgage terms
