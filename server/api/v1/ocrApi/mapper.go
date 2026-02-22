package ocrApi

import (
	"2_Go/internal/document"
	"time"

	"github.com/hththai/ocr"
)

// dateFormats lists the date layouts tried when parsing Invoice.DocumentDate.
// Add more layouts here if new OCR sources use different formats.
var dateFormats = []string{
	"02.01.2006", // DD.MM.YYYY  — e.g. Amazon EU "27.11.2025"
	"02/01/2006", // DD/MM/YYYY
	"01/02/2006", // MM/DD/YYYY  — US format
	"2006-01-02", // ISO 8601
}

// InvoiceToPurchaseInfo maps an OCR-extracted Invoice to a document.PurchaseInfo.
// BuyAt is nil when the date string is empty or cannot be parsed.
func InvoiceToPurchaseInfo(inv ocr.Invoice) document.PurchaseInfo {
	info := document.PurchaseInfo{
		BuyFrom:  inv.Seller,
		BuyPrice: inv.Total,
	}

	for _, layout := range dateFormats {
		if t, err := time.Parse(layout, inv.DocumentDate); err == nil {
			info.BuyAt = &t
			break
		}
	}

	return info
}
