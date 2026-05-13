package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image/png"
	"net/http"
	"strings"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/minio/minio-go/v7"

	"github.com/thxnxt11/payment_test/models"
)

type ServiceError struct {
	Status  int
	Message string
}

func (err *ServiceError) Error() string {
	return err.Message
}

type QRService struct {
	minioClient *minio.Client
	bucket      string
	urlExpiry   time.Duration
}

func NewQRService(client *minio.Client, bucket string, urlExpiry time.Duration) *QRService {
	return &QRService{
		minioClient: client,
		bucket:      bucket,
		urlExpiry:   urlExpiry,
	}
}

func (service *QRService) Generate(ctx context.Context, req models.QRRequest) (models.QRResponse, error) {
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "promptpay"
	}

	amount, err := normalizeAmount(req.Amount)
	if err != nil {
		return models.QRResponse{}, &ServiceError{Status: http.StatusBadRequest, Message: err.Error()}
	}

	var payload string
	switch mode {
	case "promptpay":
		promptPayID, promptPayType, err := normalizePromptPayID(req.PromptPayID)
		if err != nil {
			return models.QRResponse{}, &ServiceError{Status: http.StatusBadRequest, Message: err.Error()}
		}
		payload = buildPromptPayPayload(promptPayID, promptPayType, amount)
	case "biller":
		billerID, err := normalizeBillerID(req.BillerID)
		if err != nil {
			return models.QRResponse{}, &ServiceError{Status: http.StatusBadRequest, Message: err.Error()}
		}
		reference1, err := normalizeReference(req.Reference1, "reference1", true)
		if err != nil {
			return models.QRResponse{}, &ServiceError{Status: http.StatusBadRequest, Message: err.Error()}
		}
		reference2, err := normalizeReference(req.Reference2, "reference2", false)
		if err != nil {
			return models.QRResponse{}, &ServiceError{Status: http.StatusBadRequest, Message: err.Error()}
		}
		payload = buildBillerPayload(billerID, reference1, reference2, amount)
	default:
		return models.QRResponse{}, &ServiceError{Status: http.StatusBadRequest, Message: "mode must be promptpay or biller"}
	}

	pngBytes, err := renderQRCode(payload)
	if err != nil {
		return models.QRResponse{}, &ServiceError{Status: http.StatusInternalServerError, Message: "failed to render QR"}
	}

	objectName := fmt.Sprintf("qr/%s-%s.png", time.Now().UTC().Format("20060102-150405"), randomSuffix())
	_, err = service.minioClient.PutObject(ctx, service.bucket, objectName, bytes.NewReader(pngBytes), int64(len(pngBytes)), minio.PutObjectOptions{
		ContentType: "image/png",
	})
	if err != nil {
		return models.QRResponse{}, &ServiceError{Status: http.StatusInternalServerError, Message: "failed to upload QR"}
	}

	presignedURL, err := service.minioClient.PresignedGetObject(ctx, service.bucket, objectName, service.urlExpiry, nil)
	if err != nil {
		return models.QRResponse{}, &ServiceError{Status: http.StatusInternalServerError, Message: "failed to sign QR URL"}
	}

	return models.QRResponse{
		Payload:    payload,
		Object:     objectName,
		QRCodeURL:  presignedURL.String(),
		ExpiresIn:  int(service.urlExpiry.Seconds()),
		CreatedUTC: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func normalizePromptPayID(raw string) (string, string, error) {
	clean := strings.TrimSpace(raw)
	clean = strings.ReplaceAll(clean, "-", "")
	clean = strings.ReplaceAll(clean, " ", "")
	if clean == "" {
		return "", "", errors.New("promptpay_id is required")
	}
	if !isDigits(clean) {
		return "", "", errors.New("promptpay_id must be digits only")
	}

	if len(clean) == 10 {
		if !strings.HasPrefix(clean, "0") {
			return "", "", errors.New("phone promptpay_id must start with 0")
		}
		return "0066" + clean[1:], "01", nil
	}
	if len(clean) == 13 {
		return clean, "02", nil
	}
	return "", "", errors.New("promptpay_id must be 10 or 13 digits")
}

func normalizeBillerID(raw string) (string, error) {
	clean := strings.TrimSpace(raw)
	clean = strings.ReplaceAll(clean, "-", "")
	clean = strings.ReplaceAll(clean, " ", "")
	if clean == "" {
		return "", errors.New("biller_id is required")
	}
	if !isDigits(clean) {
		return "", errors.New("biller_id must be digits only")
	}
	if len(clean) != 15 {
		return "", errors.New("biller_id must be 15 digits")
	}
	return clean, nil
}

func normalizeReference(raw string, field string, required bool) (string, error) {
	clean := strings.ToUpper(strings.TrimSpace(raw))
	if clean == "" {
		if required {
			return "", fmt.Errorf("%s is required", field)
		}
		return "", nil
	}
	if len(clean) > 20 {
		return "", fmt.Errorf("%s must be 20 characters or less", field)
	}
	if !isAlphaNum(clean) {
		return "", fmt.Errorf("%s must be A-Z or 0-9 only", field)
	}
	return clean, nil
}

func normalizeAmount(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	if strings.HasPrefix(value, ".") {
		value = "0" + value
	}
	if strings.Count(value, ".") > 1 {
		return "", errors.New("amount format is invalid")
	}
	parts := strings.Split(value, ".")
	if !isDigits(parts[0]) {
		return "", errors.New("amount must be numeric")
	}
	if len(parts) == 2 {
		if len(parts[1]) > 2 {
			return "", errors.New("amount supports up to 2 decimals")
		}
		if parts[1] == "" {
			parts[1] = "00"
		}
		if !isDigits(parts[1]) {
			return "", errors.New("amount must be numeric")
		}
		integer := strings.TrimLeft(parts[0], "0")
		if integer == "" {
			integer = "0"
		}
		return fmt.Sprintf("%s.%s", integer, padRight(parts[1], 2)), nil
	}
	integer := strings.TrimLeft(parts[0], "0")
	if integer == "" {
		integer = "0"
	}
	return fmt.Sprintf("%s.00", integer), nil
}

func padRight(value string, size int) string {
	if len(value) >= size {
		return value
	}
	return value + strings.Repeat("0", size-len(value))
}

func isDigits(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isAlphaNum(value string) bool {
	for _, r := range value {
		if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') {
			continue
		}
		return false
	}
	return true
}

func buildPromptPayPayload(promptPayID string, promptPayType string, amount string) string {
	const (
		payloadFormat = "01"
		merchantCat  = "0000"
		currencyCode = "764"
		countryCode  = "TH"
		merchantName = "PromptPay"
		merchantCity = "Bangkok"
		promptPayAID = "A000000677010111"
	)

	pointOfInit := "11"
	if amount != "" {
		pointOfInit = "12"
	}

	merchantAccount := tlv("00", promptPayAID) + tlv(promptPayType, promptPayID)
	fields := []string{
		tlv("00", payloadFormat),
		tlv("01", pointOfInit),
		tlv("29", merchantAccount),
		tlv("52", merchantCat),
		tlv("53", currencyCode),
	}
	if amount != "" {
		fields = append(fields, tlv("54", amount))
	}
	fields = append(fields,
		tlv("58", countryCode),
		tlv("59", merchantName),
		tlv("60", merchantCity),
		"6304",
	)

	payload := strings.Join(fields, "")
	crc := crc16(payload)
	return payload + fmt.Sprintf("%04X", crc)
}

func buildBillerPayload(billerID string, reference1 string, reference2 string, amount string) string {
	const (
		payloadFormat = "01"
		merchantCat  = "0000"
		currencyCode = "764"
		countryCode  = "TH"
		merchantName = "PromptPay"
		merchantCity = "Bangkok"
		billerAID    = "A000000677010112"
	)

	pointOfInit := "11"
	if amount != "" {
		pointOfInit = "12"
	}

	merchantAccount := tlv("00", billerAID) + tlv("01", billerID) + tlv("02", reference1)
	if reference2 != "" {
		merchantAccount += tlv("03", reference2)
	}

	fields := []string{
		tlv("00", payloadFormat),
		tlv("01", pointOfInit),
		tlv("30", merchantAccount),
		tlv("52", merchantCat),
		tlv("53", currencyCode),
	}
	if amount != "" {
		fields = append(fields, tlv("54", amount))
	}
	fields = append(fields,
		tlv("58", countryCode),
		tlv("59", merchantName),
		tlv("60", merchantCity),
		"6304",
	)

	payload := strings.Join(fields, "")
	crc := crc16(payload)
	return payload + fmt.Sprintf("%04X", crc)
}

func tlv(id, value string) string {
	return fmt.Sprintf("%s%02d%s", id, len(value), value)
}

func crc16(payload string) uint16 {
	crc := uint16(0xFFFF)
	for i := 0; i < len(payload); i++ {
		crc ^= uint16(payload[i]) << 8
		for j := 0; j < 8; j++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

func renderQRCode(payload string) ([]byte, error) {
	qrCode, err := qr.Encode(payload, qr.H, qr.Auto)
	if err != nil {
		return nil, err
	}
	scaled, err := barcode.Scale(qrCode, 512, 512)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	if err := png.Encode(&buffer, scaled); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func randomSuffix() string {
	bytesBuf := make([]byte, 4)
	if _, err := rand.Read(bytesBuf); err != nil {
		return "00000000"
	}
	return hex.EncodeToString(bytesBuf)
}
