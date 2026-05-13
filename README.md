# PromptPay QR Generator

โปรเจกต์ตัวอย่างสำหรับสร้าง PromptPay QR (แบบปกติและ Bill Payment) ด้วย Go backend และ Next.js frontend พร้อมเก็บไฟล์ QR ลง MinIO

## โครงสร้าง

- backend: Go API (CRC16 + TLV) + MinIO
- frontend/my-app: Next.js UI

## ความต้องการระบบ

- Go 1.24+
- Node.js 20+
- MinIO (Docker หรือ binary)

## ตั้งค่า Backend

คัดลอก .env ตัวอย่างและปรับค่า

```
cd backend
copy .env.example .env
```

รัน MinIO (Docker ตัวอย่าง)

```
docker run -p 9000:9000 -p 9001:9001 --name minio -e MINIO_ROOT_USER=minioadmin -e MINIO_ROOT_PASSWORD=minioadmin -v C:\minio-data:/data quay.io/minio/minio server /data --console-address ":9001"
```

รัน backend

```
cd backend
go mod tidy
go run .
```

## ตั้งค่า Frontend

```
cd frontend\my-app
copy .env.local.example .env.local
npm install
npm run dev
```

เปิด UI: http://localhost:3000

## API

- POST /api/qr

### PromptPay (phone / national ID)

```json
{
  "mode": "promptpay",
  "promptpay_id": "0812345678",
  "amount": "100.00"
}
```

### Bill Payment (biller_id + reference)

```json
{
  "mode": "biller",
  "biller_id": "123456789012345",
  "reference1": "INV20260001",
  "reference2": "A1",
  "amount": "250.00"
}
```

## หมายเหตุ

- หาก backend รันพอร์ตอื่น ให้ปรับ `NEXT_PUBLIC_API_BASE_URL` ใน frontend
