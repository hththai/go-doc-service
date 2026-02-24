package ocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/gen2brain/go-fitz"
)

const invoicePrompt = `Extract the invoice data from this image and return ONLY valid JSON (no markdown, no code fences) with this exact structure:
{
  "seller": "company name",
  "abn": "ABN number",
  "document_date": "date",
  "order_no": "order number",
  "order_date": "date",
  "items": [{"description": "item name", "qty": 1, "unit_price": "price", "gst_rate": "rate", "subtotal": "total"}],
  "total": "numeric amount only, no currency symbol or code (e.g. 456.02 not AUD456.02)"
}`

// Service wraps the OCR logic and holds the Anthropic API key.
type Service struct {
	apiKey string
}

// NewService creates a new OCR Service with the provided Anthropic API key.
func NewService(apiKey string) *Service {
	return &Service{apiKey: apiKey}
}

// ExtractInvoice reads fileBytes (PDF or image), sends them to Claude Haiku,
// and returns the structured Result containing the parsed invoice and token usage.
func (s *Service) ExtractInvoice(ctx context.Context, fileBytes []byte, filename string) (*Result, error) {
	return ExtractInvoice(ctx, s.apiKey, fileBytes, filename)
}

// ExtractInvoice is the package-level function that does the actual extraction.
func ExtractInvoice(ctx context.Context, apiKey string, fileBytes []byte, filename string) (*Result, error) {
	images, err := toBase64Images(fileBytes, filename)
	if err != nil {
		return nil, fmt.Errorf("converting file to images: %w", err)
	}

	var content []anthropic.ContentBlockParamUnion
	for _, img := range images {
		content = append(content, anthropic.ContentBlockParamUnion{
			OfImage: &anthropic.ImageBlockParam{
				Source: anthropic.ImageBlockParamSourceUnion{
					OfBase64: &anthropic.Base64ImageSourceParam{
						Data:      img,
						MediaType: anthropic.Base64ImageSourceMediaTypeImagePNG,
					},
				},
			},
		})
	}
	content = append(content, anthropic.ContentBlockParamUnion{
		OfText: &anthropic.TextBlockParam{
			Text: invoicePrompt,
		},
	})

	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 4096,
		Messages: []anthropic.MessageParam{
			{
				Role:    anthropic.MessageParamRoleUser,
				Content: content,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("calling Claude API: %w", err)
	}

	var responseText string
	for _, block := range message.Content {
		if block.Type == "text" {
			responseText = block.Text
		}
	}

	cleaned := extractJSON(responseText)

	var invoice Invoice
	if err := json.Unmarshal([]byte(cleaned), &invoice); err != nil {
		return &Result{
			Model:        string(message.Model),
			InputTokens:  message.Usage.InputTokens,
			OutputTokens: message.Usage.OutputTokens,
			RawResponse:  responseText,
		}, fmt.Errorf("parsing JSON response: %w", err)
	}

	return &Result{
		Invoice:      invoice,
		Model:        string(message.Model),
		InputTokens:  message.Usage.InputTokens,
		OutputTokens: message.Usage.OutputTokens,
		RawResponse:  responseText,
	}, nil
}

func toBase64Images(data []byte, filename string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".pdf" {
		return pdfToBase64Images(data)
	}
	return []string{base64.StdEncoding.EncodeToString(data)}, nil
}

func pdfToBase64Images(data []byte) ([]string, error) {
	doc, err := fitz.NewFromMemory(data)
	if err != nil {
		return nil, fmt.Errorf("opening PDF: %w", err)
	}
	defer doc.Close()

	var images []string
	for i := 0; i < doc.NumPage(); i++ {
		img, err := doc.Image(i)
		if err != nil {
			return nil, fmt.Errorf("converting page %d: %w", i, err)
		}

		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("encoding page %d to PNG: %w", i, err)
		}

		images = append(images, base64.StdEncoding.EncodeToString(buf.Bytes()))
	}

	return images, nil
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
	}
	return strings.TrimSpace(s)
}
