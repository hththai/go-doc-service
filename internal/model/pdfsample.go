package item

import (
	"log"

	"github.com/go-pdf/fpdf"
)

func CreatePDF(storedpath string) {
	pdf := fpdf.New("P", "mm", "A4", "")

	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)

	pdf.Cell(40, 10, "Hello, World")

	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 10, "This is some sample content")

	pdf.Ln(10)
	pdf.MultiCell(0, 5, "This is demonstration", "", "", false)

	err := pdf.OutputFileAndClose(storedpath)

	if err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
	}

	log.Println("PDF 'sample.pdf' generated successfully")
}
