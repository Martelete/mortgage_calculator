package main

import (
	"bytes"
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
	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func mortgageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		principal, _ := strconv.ParseFloat(r.FormValue("principal"), 64)
		ratePercent, _ := strconv.ParseFloat(r.FormValue("rate"), 64)
		fixedMonths, _ := strconv.Atoi(r.FormValue("months"))
		monthlyPayment, _ := strconv.ParseFloat(r.FormValue("monthly"), 64)

		m := Mortgage{
			Principal:      principal,
			AnnualRate:     ratePercent,
			FixedMonths:    fixedMonths,
			MonthlyPayment: monthlyPayment,
		}

		breakdown := GenerateMonthlyBreakdown(m)

		var totalInterest, totalPrincipal float64
		for _, d := range breakdown {
			totalInterest += d.InterestPayment
			totalPrincipal += d.PrincipalPayment
		}
		totalPaid := totalInterest + totalPrincipal
		remaining := breakdown[len(breakdown)-1].Balance

		data := PageData{
			Mortgage:       m,
			Breakdown:      breakdown,
			TotalPaid:      totalPaid,
			TotalInterest:  totalInterest,
			TotalPrincipal: totalPrincipal,
			Remaining:      remaining,
		}

		tmpl := template.Must(
			template.New("index.html").
				Funcs(template.FuncMap{
					"gbp": formatGBP,
				}).
				ParseFiles("index.html"),
		)

		tmpl.Execute(w, data)
		return
	}

	// Show empty form
	tmpl := template.Must(
		template.New("index.html").
			Funcs(template.FuncMap{
				"gbp": formatGBP,
			}).
			ParseFiles("index.html"),
	)

	tmpl.Execute(w, nil)
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

func downloadPDFHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	principal, _ := strconv.ParseFloat(r.FormValue("principal"), 64)
	ratePercent, _ := strconv.ParseFloat(r.FormValue("rate"), 64)
	fixedMonths, _ := strconv.Atoi(r.FormValue("months"))
	monthlyPayment, _ := strconv.ParseFloat(r.FormValue("monthly"), 64)

	m := Mortgage{
		Principal:      principal,
		AnnualRate:     ratePercent,
		FixedMonths:    fixedMonths,
		MonthlyPayment: monthlyPayment,
	}

	breakdown := GenerateMonthlyBreakdown(m)

	pdfBytes, err := GeneratePDFBytes(m, breakdown)
	if err != nil {
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}

	// Send PDF to browser
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=mortgage_breakdown.pdf")
	w.Write(pdfBytes)
}

func formatGBP(v float64) string {
	s := fmt.Sprintf("%.2f", v)

	parts := strings.Split(s, ".")
	intPart := parts[0]
	decPart := parts[1]

	n := len(intPart)
	for i := n - 3; i > 0; i -= 3 {
		intPart = intPart[:i] + "," + intPart[i:]
	}

	return "£" + intPart + "." + decPart
}

// Generate PDF in memory
func GeneratePDFBytes(m Mortgage, data []MonthlyData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddUTF8Font("DejaVu", "B", "fonts/DejaVuSans-Bold.ttf")
	pdf.SetFont("DejaVu", "B", 14)
	pdf.Cell(40, 10, "Mortgage Breakdown")
	pdf.Ln(12)

	// Mortgage Summary
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

	// Fixed-period Breakdown Table
	pdf.SetFont("DejaVu", "B", 11)
	pdf.Cell(20, 8, "Month")
	pdf.Cell(30, 8, "Interest")
	pdf.Cell(30, 8, "Principal")
	pdf.Cell(30, 8, "Balance")
	pdf.Ln(8)

	pdf.SetFont("DejaVu", "B", 11)
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
