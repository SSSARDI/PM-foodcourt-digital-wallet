# PM Foodcourt Digital Wallet – Backend

Backend API สำหรับระบบ Foodcourt Digital Wallet
พัฒนาด้วย **Go (Golang)** และ **MySQL**

---

# 📦 Tech Stack

* Go
* MySQL
* net/http (หรือ framework ที่ใช้)
* UUID
* bcrypt (password hashing)

---

# 🛠 Prerequisites

ต้องติดตั้งก่อน:

* Go (1.21+ แนะนำ)
* MySQL Server
* Git

เช็คเวอร์ชัน:

```bash
go version
mysql --version
```

---

# 🗄 Database Setup

## 1️⃣ สร้าง Database

ไปที่โฟลเดอร์โปรเจกต์:

```bash
cd "/mnt/c/Users/SSSARDI/Desktop/PM-food court/PM-foodcourt-digital-wallet"
```

Import schema:

```bash
mysql -u root -p -h 127.0.0.1 < pkg/database/001_init.sql
```

ตรวจสอบ:

```bash
mysql -u root -p -h 127.0.0.1
```

```sql
SHOW DATABASES;
USE foodcourt;
SHOW TABLES;
```

---

# ⚙️ Environment Configuration

ในไฟล์ database connection (เช่น `internal/database/mysql.go`)

ตั้งค่า DSN:

```go
dsn := "root:YOUR_PASSWORD@tcp(127.0.0.1:3306)/foodcourt?parseTime=true"
```

ถ้าใช้ Docker ต้องเปลี่ยน host ตาม service name

---

# 📥 Install Dependencies

```bash
go mod tidy
```

---

# 🚀 Run Backend Server

```bash
go run .
```

หรือถ้ามี main.go:

```bash
go run main.go
```

Server จะรันที่:

```
http://localhost:8080
```

---

# 🧪 Example API Testing

### Register

```bash
curl -X POST http://localhost:8080/register \
-d "full_name=Test User" \
-d "email=test@mail.com" \
-d "password=1234"
```

### Get Wallet Balance

```bash
curl http://localhost:8080/wallet/balance?user_id=UUID
```

---

# 🔐 Important Notes

* Password ต้อง hash ด้วย bcrypt
* ทุก flow เกี่ยวกับเงินต้องใช้ SQL Transaction
* ห้าม update balance แบบ `balance = amount`
* ใช้ `balance = balance ± amount`

---

# 📂 Project Structure (Example)

```
internal/
  ├── database/
  ├── repository/
  ├── service/
  ├── handler/
pkg/
  └── database/
      └── 001_init.sql
main.go
go.mod
```

---

# 🔥 Production Notes

* เพิ่ม index ที่ wallet_id, user_id
* ใช้ ENV file สำหรับ password
* หลีกเลี่ยง hardcode DB credentials
* ใช้ HTTPS ใน production

---

# 👨‍💻 Developer

Natthanon PUMPUANG
PM Foodcourt Digital Wallet Project

---
