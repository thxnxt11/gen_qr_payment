package models

type QRRequest struct {
	Mode        string `json:"mode,omitempty"`
	PromptPayID string `json:"promptpay_id,omitempty"`
	Amount      string `json:"amount,omitempty"`
	BillerID    string `json:"biller_id,omitempty"`
	Reference1  string `json:"reference1,omitempty"`
	Reference2  string `json:"reference2,omitempty"`
}

type QRResponse struct {
	Payload    string `json:"payload"`
	Object     string `json:"object"`
	QRCodeURL  string `json:"qr_url"`
	ExpiresIn  int    `json:"expires_in"`
	CreatedUTC string `json:"created_utc"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
