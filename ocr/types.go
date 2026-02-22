package ocr

type Item struct {
	Description string `json:"description"`
	Qty         int    `json:"qty"`
	UnitPrice   string `json:"unit_price"`
	GSTRate     string `json:"gst_rate"`
	Subtotal    string `json:"subtotal"`
}

type Invoice struct {
	Seller          string `json:"seller"`
	ABN             string `json:"abn"`
	DocumentDate    string `json:"document_date"`
	OrderNo         string `json:"order_no"`
	OrderDate       string `json:"order_date"`
	BillingAddress  string `json:"billing_address"`
	DeliveryAddress string `json:"delivery_address"`
	Items           []Item `json:"items"`
	ShippingCharges string `json:"shipping_charges"`
	Total           string `json:"total"`
}

type Result struct {
	Invoice      Invoice `json:"invoice"`
	Model        string  `json:"model"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	RawResponse  string  `json:"raw_response"`
}
