# ใช้ Go 1.24 alpine ตาม go.mod ของฟลุ๊ค
FROM golang:1.24-alpine

# ติดตั้งเครื่องมือที่จำเป็น
RUN apk add --no-cache git build-base

WORKDIR /app

# 1. จัดการ dependencies
COPY go.mod go.sum ./
RUN go mod download

# 2. ก๊อปปี้โค้ดทั้งหมด
COPY . .

# 3. Build โปรเจกต์ (ถ้า main.go อยู่ที่ root ใช้ ./main.go)
RUN go build -o main ./main.go

# 4. เปิดพอร์ต 8080
EXPOSE 8080

# 5. สั่งรัน
CMD ["./main"]