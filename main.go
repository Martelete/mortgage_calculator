package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

type Mortgage struct {
	Principal      float64
	AnnualRate     float64
	FixedMonths    int
	MonthlyPayment float64
}

type MonthlyData struct {
	Month            int
	InterestPayment  float64
	PrincipalPayment float64
	Balance          float64
}

type PageData struct {
	Mortgage       Mortgage
	Breakdown      []MonthlyData
	TotalPaid      float64
	TotalInterest  float64
	TotalPrincipal float64
	Remaining      float64
}

func main() {
	http.HandleFunc("/", mortgageHandler)
	http.HandleFunc("/download-pdf", downloadPDFHandler)
	http.HandleFunc("/download-csv", downloadCSVHandler)

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func mortgageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(
		template.New("index.html").
			Funcs(template.FuncMap{"gbp": formatGBP}).
			ParseFiles("index.html"),
	)

	if r.Method != http.MethodPost {
		tmpl.Execute(w, PageData{})
		return
	}

	m, err := parseMortgageForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	breakdown := GenerateMonthlyBreakdown(m)

	var totalInterest, totalPrincipal float64
	for _, d := range breakdown {
		totalInterest += d.InterestPayment
		totalPrincipal += d.PrincipalPayment
	}

	data := PageData{
		Mortgage:       m,
		Breakdown:      breakdown,
		TotalPaid:      totalInterest + totalPrincipal,
		TotalInterest:  totalInterest,
		TotalPrincipal: totalPrincipal,
		Remaining:      breakdown[len(breakdown)-1].Balance,
	}

	tmpl.Execute(w, data)
}

func downloadPDFHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	m, err := parseMortgageForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	breakdown := GenerateMonthlyBreakdown(m)

	pdfBytes, err := GeneratePDFBytes(m, breakdown)
	if err != nil {
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=mortgage_breakdown.pdf")
	w.Write(pdfBytes)
}

func downloadCSVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	m, err := parseMortgageForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	breakdown := GenerateMonthlyBreakdown(m)

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=mortgage_breakdown.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write CSV headers
	writer.Write([]string{"Month", "Interest", "Principal", "Balance"})

	// Write rows
	for _, d := range breakdown {
		writer.Write([]string{
			strconv.Itoa(d.Month),
			fmt.Sprintf("%.2f", d.InterestPayment),
			fmt.Sprintf("%.2f", d.PrincipalPayment),
			fmt.Sprintf("%.2f", d.Balance),
		})
	}
}

func parseMortgageForm(r *http.Request) (Mortgage, error) {
	if err := r.ParseForm(); err != nil {
		return Mortgage{}, err
	}

	get := func(name string) (string, error) {
		v := strings.TrimSpace(r.FormValue(name))
		if v == "" {
			return "", fmt.Errorf("missing field: %s", name)
		}
		return v, nil
	}

	log.Printf("FORM months=%q", r.FormValue("months"))

	principalStr, err := get("principal")
	if err != nil {
		return Mortgage{}, err
	}
	rateStr, err := get("rate")
	if err != nil {
		return Mortgage{}, err
	}
	monthsStr, err := get("months")
	if err != nil {
		return Mortgage{}, err
	}
	monthlyStr, err := get("monthly")
	if err != nil {
		return Mortgage{}, err
	}

	principal, err := strconv.ParseFloat(principalStr, 64)
	if err != nil {
		return Mortgage{}, fmt.Errorf("invalid principal")
	}

	ratePercent, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		return Mortgage{}, fmt.Errorf("invalid rate")
	}

	fixedMonths, err := strconv.Atoi(monthsStr)
	if err != nil {
		return Mortgage{}, fmt.Errorf("invalid months")
	}

	monthlyPayment, err := strconv.ParseFloat(monthlyStr, 64)
	if err != nil {
		return Mortgage{}, fmt.Errorf("invalid monthly payment")
	}

	return Mortgage{
		Principal:      principal,
		AnnualRate:     ratePercent,
		FixedMonths:    fixedMonths,
		MonthlyPayment: monthlyPayment,
	}, nil
}

func GenerateMonthlyBreakdown(m Mortgage) []MonthlyData {
	balance := m.Principal
	monthlyRate := (m.AnnualRate / 100) / 12

	data := make([]MonthlyData, m.FixedMonths)

	for i := 0; i < m.FixedMonths; i++ {
		interest := balance * monthlyRate
		principalPayment := m.MonthlyPayment - interest
		balance -= principalPayment

		if balance < 0 {
			principalPayment += balance
			balance = 0
		}

		data[i] = MonthlyData{
			Month:            i + 1,
			InterestPayment:  interest,
			PrincipalPayment: principalPayment,
			Balance:          balance,
		}
	}

	return data
}

func formatGBP(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	parts := strings.Split(s, ".")
	intPart, decPart := parts[0], parts[1]

	for i := len(intPart) - 3; i > 0; i -= 3 {
		intPart = intPart[:i] + "," + intPart[i:]
	}

	return "£" + intPart + "." + decPart
}

func GeneratePDFBytes(m Mortgage, data []MonthlyData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.AddUTF8Font("DejaVu", "B", "fonts/DejaVuSans-Bold.ttf")
	pdf.SetFont("DejaVu", "B", 14)
	pdf.Cell(40, 10, "Mortgage Breakdown")
	pdf.Ln(12)

	var totalInterest, totalPrincipal float64
	for _, d := range data {
		totalInterest += d.InterestPayment
		totalPrincipal += d.PrincipalPayment
	}

	totalPaid := totalInterest + totalPrincipal
	remaining := data[len(data)-1].Balance

	pdf.SetFont("DejaVu", "B", 11)
	pdf.Cell(60, 8, "Principal: "+formatGBP(m.Principal))
	pdf.Ln(6)
	pdf.Cell(60, 8, fmt.Sprintf("Fixed rate period: %d months", m.FixedMonths))
	pdf.Ln(6)
	pdf.Cell(60, 8, "Monthly payment: "+formatGBP(m.MonthlyPayment))
	pdf.Ln(6)
	pdf.Cell(60, 8, "Total paid: "+formatGBP(totalPaid))
	pdf.Ln(6)
	pdf.Cell(60, 8, "Total interest: "+formatGBP(totalInterest))
	pdf.Ln(6)
	pdf.Cell(60, 8, "Total principal: "+formatGBP(totalPrincipal))
	pdf.Ln(6)
	pdf.Cell(60, 8, "Remaining balance: "+formatGBP(remaining))
	pdf.Ln(12)

	pdf.Cell(20, 8, "Month")
	pdf.Cell(30, 8, "Interest")
	pdf.Cell(30, 8, "Principal")
	pdf.Cell(30, 8, "Balance")
	pdf.Ln(8)

	for _, d := range data {
		pdf.Cell(20, 8, strconv.Itoa(d.Month))
		pdf.Cell(30, 8, formatGBP(d.InterestPayment))
		pdf.Cell(30, 8, formatGBP(d.PrincipalPayment))
		pdf.Cell(30, 8, formatGBP(d.Balance))
		pdf.Ln(8)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
