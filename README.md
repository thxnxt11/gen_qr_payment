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

## มาตรฐาน EMV (TLV + CRC16)

โปรเจกต์นี้สร้าง payload ตามรูปแบบ EMVCo โดยใช้โครงสร้าง TLV (Tag-Length-Value) และปิดท้ายด้วย CRC16

### TLV คืออะไร

- รูปแบบข้อมูล `Tag` + `Length` + `Value`
- `Tag`: ระบุประเภทข้อมูล (เช่น 00 = Payload Format, 54 = Amount)
- `Length`: ความยาวของ `Value` (มักเป็น 2 หลัก)
- `Value`: เนื้อข้อมูลจริง
- ตัวอย่าง `000201` = Tag 00, Length 02, Value 01

### CRC16 คืออะไร

- ค่าตรวจสอบความถูกต้องของข้อมูล (checksum)
- ใช้ CRC-16/CCITT (poly 0x1021, init 0xFFFF)
- คำนวณจาก payload ที่ลงท้ายด้วย `6304`
- ผลลัพธ์เป็นเลขฐาน 16 จำนวน 4 ตัว แล้วต่อท้าย payload

### Tag หลักที่ใช้ในโปรเจกต์นี้

- `00` Payload Format Indicator: ค่าคงที่ `01`
- `01` Point of Initiation Method: `11` = static, `12` = dynamic (มี amount)
- `29` Merchant Account (PromptPay): เก็บ AID + PromptPay ID
- `30` Merchant Account (Bill Payment): เก็บ AID + biller_id + reference
- `52` Merchant Category Code: `0000`
- `53` Transaction Currency: `764` (THB)
- `54` Transaction Amount: ใช้เมื่อมี amount
- `58` Country Code: `TH`
- `59` Merchant Name: `PromptPay`
- `60` Merchant City: `Bangkok`
- `63` CRC: ความยาว 04 (ใส่ค่า CRC16 ต่อท้าย)

### AID ที่ใช้

- PromptPay: `A000000677010111`
- Bill Payment: `A000000677010112`

## หมายเหตุ

- หาก backend รันพอร์ตอื่น ให้ปรับ `NEXT_PUBLIC_API_BASE_URL` ใน frontend
