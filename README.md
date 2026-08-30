# 🛒 Bazar List

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Database](https://img.shields.io/badge/Database-MySQL-4479A1?style=flat&logo=mysql&logoColor=white)](https://www.mysql.com/)
[![ORM](https://img.shields.io/badge/ORM-GORM-009688?style=flat)](https://gorm.io/)
[![JWT Auth](https://img.shields.io/badge/Auth-JWT-black?style=flat&logo=jsonwebtokens)](https://jwt.io/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A modern, full-featured web application and RESTful API built in Go (Golang) for managing household shopping and grocery (*bazar*) lists. It features JWT-based user authentication, monthly list organization, budget/expense tracking, item status tracking, and a clean, responsive web interface.

---

## ✨ Features

- **🔐 User Authentication & Authorization**
  - Secure user registration and login with bcrypt password hashing
  - Stateless JWT (JSON Web Token) based session handling
  - Multi-user isolation — each user manages their private lists
- **📅 Monthly List Organization**
  - Create and manage date-tagged shopping lists
  - Automatic grouping and filtering by month (`YYYY-MM`)
  - Built-in pagination and historical month navigation
- **📝 Real-Time Shopping & Item Tracking**
  - Add items with name and price tags
  - Real-time toggle to mark items as purchased/pending
  - Automatic calculation of total cost per list and aggregated monthly expenditure
- **🎨 Modern Responsive Web UI**
  - Fast Single Page Application (SPA) frontend served directly by the backend
  - Clean desktop & mobile-friendly design
  - Intuitive modals for adding/editing lists and items
- **⚡ RESTful API & Modular Architecture**
  - Robust backend built with Gorilla Mux and GORM
  - CORS middleware support
  - Auto-migrated MySQL schema

---

## 🛠️ Tech Stack

- **Backend:** [Go](https://go.dev/) (1.21+)
- **Routing & Middleware:** [Gorilla Mux](https://github.com/gorilla/mux)
- **Database & ORM:** [MySQL](https://www.mysql.com/) with [GORM](https://gorm.io/)
- **Authentication:** [golang-jwt/jwt](https://github.com/golang-jwt/jwt) & [x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- **Frontend:** Vanilla HTML5, CSS3, JavaScript (SPA)

---

## 📁 Project Structure

```
bazar_list/
├── cmd/
│   └── web/                # Web application entry point (main.go)
├── internal/               # Private application code
│   ├── api/                # HTTP handlers, routing & middleware
│   │   ├── auth_handlers.go
│   │   ├── list_handlers.go
│   │   ├── middleware.go
│   │   └── utils.go
│   ├── auth/               # JWT token generation & password hashing
│   │   └── auth.go
│   ├── models/             # Data models (User, ShoppingList, Item)
│   │   └── database.go
│   └── storage/            # Database layer (MySQL & GORM)
│       └── mysql.go
├── web/
│   └── static/             # Frontend assets (index.html, styles, scripts)
├── docs/                   # Detailed documentation and learning guides
├── scripts/                # Setup and deployment helper scripts
├── .env.example            # Example environment configuration
├── Makefile                # Build, run, and development commands
├── go.mod                  # Go module definition
└── README.md
```

---

## 🚀 Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.21 or higher
- [MySQL](https://dev.mysql.com/downloads/) 5.7+ / 8.0+ (or running via Docker)
- [Git](https://git-scm.com/)
- [Make](https://www.gnu.org/software/make/) (optional, for helper commands)

---

### Step 1: Clone the Repository

```bash
git clone https://github.com/sejan/bazarlist.git
cd bazarlist
```

### Step 2: Configure MySQL Database

Create a database for the application:

```sql
-- Connect to MySQL
mysql -u root -p

-- Create database and user
CREATE DATABASE bazarlist CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'bazarlist'@'localhost' IDENTIFIED BY 'bazarlist123';
GRANT ALL PRIVILEGES ON bazarlist.* TO 'bazarlist'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

> **Using Docker instead?**
> ```bash
> docker run --name bazarlist-mysql \
>   -e MYSQL_ROOT_PASSWORD=rootpass \
>   -e MYSQL_DATABASE=bazarlist \
>   -e MYSQL_USER=bazarlist \
>   -e MYSQL_PASSWORD=bazarlist123 \
>   -p 3306:3306 -d mysql:8.0
> ```

### Step 3: Configure Environment Variables

Copy `.env.example` to `.env` and adjust the values if needed:

```bash
cp .env.example .env
```

Example `.env` file:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=bazarlist
DB_PASSWORD=bazarlist123
DB_NAME=bazarlist

# Security & Server
JWT_SECRET=your-super-secret-jwt-key-change-this
PORT=8080
```

### Step 4: Run the Application

```bash
# Download dependencies
go mod download

# Run directly via Make
make run-web

# Or run with Go directly
go run cmd/web/main.go
```

Once started, open your browser and navigate to:
```
http://localhost:8080
```

---

## 📡 REST API Reference

All protected endpoints require an `Authorization` header with the Bearer token:
```
Authorization: Bearer <your_jwt_token>
```

### 🔐 Authentication Endpoints

| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `POST` | `/api/auth/register` | Register a new user | No |
| `POST` | `/api/auth/login` | Log in and obtain JWT | No |

#### Register Request Body
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "secretpassword"
}
```

#### Login Request Body
```json
{
  "email": "john@example.com",
  "password": "secretpassword"
}
```

#### Auth Success Response
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsIn...",
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2026-08-30T12:00:00Z"
  }
}
```

---

### 📋 Shopping Lists Endpoints

| Method | Endpoint | Description | Query Params |
|---|---|---|---|
| `GET` | `/api/lists` | Get paginated lists for a month | `month=YYYY-MM`, `page=1`, `limit=10` |
| `POST` | `/api/lists` | Create a new shopping list | — |
| `GET` | `/api/lists/{id}` | Get list by ID with its items | — |
| `PUT` / `PATCH` | `/api/lists/{id}` | Update shopping list name/date | — |
| `DELETE` | `/api/lists/{id}` | Delete a shopping list | — |

#### Create List Request Body
```json
{
  "name": "Weekly Grocery",
  "date": "2026-08-30"
}
```

---

### 🛒 List Items Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/lists/{id}/items` | Get all items in a list |
| `POST` | `/api/lists/{id}/items` | Add a new item to a list |
| `PUT` / `PATCH` | `/api/lists/{id}/items/{itemId}` | Update item name, price, or purchased status |
| `DELETE` | `/api/lists/{id}/items/{itemId}` | Delete an item from a list |

#### Create Item Request Body
```json
{
  "name": "Milk (2 Liters)",
  "price": 160.00,
  "purchased": false
}
```

---

## ⚙️ Configuration Variables

| Variable | Description | Default |
|---|---|---|
| `DB_HOST` | MySQL database host | `localhost` |
| `DB_PORT` | MySQL database port | `3306` |
| `DB_USER` | MySQL database user | `root` |
| `DB_PASSWORD` | MySQL database password | `""` |
| `DB_NAME` | MySQL database name | `bazarlist` |
| `JWT_SECRET` | Secret key for signing JWT tokens | `bazarlist-secret-key...` |
| `PORT` | HTTP server port | `8080` |

---

## 🛠️ Makefile Commands

| Command | Description |
|---|---|
| `make run-web` | Run the web server directly with environment variables |
| `make build-web` | Compile web binary to `build/bazarlist-web` |
| `make clean` | Remove build artifacts |
| `make fmt` | Format Go source files |
| `make deps` | Download and tidy Go modules |

---

## 📚 Detailed Documentation

For deeper dives into architecture and guides, explore the `docs/` folder:

- **[Architecture Guide](docs/ARCHITECTURE.md)**: System design and architectural patterns.
- **[MySQL & Auth Setup](docs/MYSQL_AUTH_SETUP.md)**: Detailed step-by-step guide for database and user auth setup.
- **[API Reference](docs/API_REFERENCE.md)**: Extended API schema and examples.
- **[Web Application Guide](docs/WEB_APPLICATION.md)**: Overview of UI features and frontend workflow.
- **[Go Learning Roadmap](docs/LEARNING_ROADMAP.md)**: 5-week Go concepts learning path.

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the [issues page](https://github.com/sejan/bazarlist/issues).

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
