# MySQL & Authentication Setup Guide

This guide helps you set up MySQL database and authentication for Bazar List.

## 📦 Prerequisites

1. **MySQL Server 5.7+ or 8.0+**
2. **Go 1.21+**
3. **MySQL Client/Connector**

## 🗄️ MySQL Setup

### Option 1: Install MySQL Server

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install mysql-server
sudo mysql_secure_installation
```

**Fedora:**
```bash
sudo dnf install mysql-server
sudo systemctl start mysqld
sudo systemctl enable mysqld
sudo mysql_secure_installation
```

**macOS:**
```bash
brew install mysql
brew services start mysql
```

**Windows:**
Download from: https://dev.mysql.com/downloads/mysql/

### Option 2: Use MySQL Docker Container

```bash
docker run --name bazarlist-mysql \
  -e MYSQL_ROOT_PASSWORD=yourpassword \
  -e MYSQL_DATABASE=bazarlist \
  -p 3306:3306 \
  -d mysql:8.0
```

## 🗃️ Database Setup

### Create Database and User

```sql
-- Connect to MySQL as root
mysql -u root -p

-- Create database
CREATE DATABASE bazarlist CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create user
CREATE USER 'bazarlist'@'localhost' IDENTIFIED BY 'your_password';

-- Grant privileges
GRANT ALL PRIVILEGES ON bazarlist.* TO 'bazarlist'@'localhost';

-- Flush privileges
FLUSH PRIVILEGES;

-- Exit
EXIT;
```

### Verify Setup

```bash
mysql -u bazarlist -p
# Enter password when prompted
# You should see: bazarlist>
```

## 🔧 Environment Configuration

### Option 1: Environment Variables

Create a `.env` file in project root:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=bazarlist
DB_PASSWORD=your_password
DB_NAME=bazarlist

# JWT Secret (change this in production!)
JWT_SECRET=your-super-secret-jwt-key-change-this

# Server Port
PORT=8080
```

### Option 2: Export Environment Variables

```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=bazarlist
export DB_PASSWORD=your_password
export DB_NAME=bazarlist
export JWT_SECRET=your-super-secret-jwt-key-change-this
export PORT=8080
```

### Option 3: Use Defaults

The application uses these defaults if not set:
- `DB_HOST`: `localhost`
- `DB_PORT`: `3306`
- `DB_USER`: `root`
- `DB_PASSWORD`: `""` (empty)
- `DB_NAME`: `bazarlist`
- `JWT_SECRET`: `bazarlist-secret-key-change-in-production`

## 🚀 Running the Application

### 1. Download Dependencies

```bash
cd /home/sejan/Desktop/bazarlist
go mod download
```

### 2. Start MySQL Server

```bash
# Linux
sudo systemctl start mysql

# macOS
brew services start mysql

# Docker
docker start bazarlist-mysql
```

### 3. Run Application

```bash
make run-web
```

You should see:
```
✅ Database connected successfully
🛒 Bazar List Web Server (MySQL + Auth)
📦 Database: MySQL
🌐 Server running at: http://localhost:8080
📄 API: http://localhost:8080/api
```

## 🔐 Authentication Flow

### Registration

1. User clicks "Create Account"
2. Enters username, email, password
3. Frontend sends POST to `/api/auth/register`
4. Backend hashes password with bcrypt
5. Creates user in MySQL
6. Generates JWT token
7. Returns token + user info

### Login

1. User enters username, password
2. Frontend sends POST to `/api/auth/login`
3. Backend finds user by username
4. Verifies password with bcrypt
5. Generates JWT token
6. Returns token + user info

### Authenticated Requests

1. Frontend includes token in `Authorization` header
2. Format: `Bearer <token>`
3. Backend validates token
4. Extracts user ID from token
5. Adds user ID to request context
6. Allows access to protected routes

## 📡 API Endpoints

### Public Endpoints (No Auth Required)

- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user

### Protected Endpoints (Auth Required)

#### Lists
- `GET /api/lists` - Get all user's lists
- `POST /api/lists` - Create new list
- `GET /api/lists/{id}` - Get specific list
- `PUT /api/lists/{id}` - Update list
- `DELETE /api/lists/{id}` - Delete list

#### Items
- `GET /api/lists/{id}/items` - Get items in list
- `POST /api/lists/{id}/items` - Add item to list
- `PUT /api/lists/{id}/items/{itemId}` - Update item
- `DELETE /api/lists/{id}/items/{itemId}` - Delete item

## 📱 Mobile-First Design

The application is optimized for mobile devices:

- **Max Width**: 575px
- **Touch-friendly**: Large buttons and inputs
- **Single Page Application**: Smooth navigation
- **Offline Ready**: LocalStorage for auth tokens
- **Responsive**: Adapts to all screen sizes

### Mobile Features

1. **Login/Register**: Simple forms
2. **List Management**: Tap to open lists
3. **Item Management**: Checkbox for purchased, delete button
4. **Floating Action Button**: Easy access to create lists/items
5. **Smooth Transitions**: No page reloads

## 🗃️ Database Schema

### Users Table
```sql
CREATE TABLE users (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email)
);
```

### Shopping Lists Table
```sql
CREATE TABLE shopping_lists (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_user_id (user_id),
    INDEX idx_date (date)
);
```

### Items Table
```sql
CREATE TABLE items (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    list_id INT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) DEFAULT 0.00,
    purchased BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (list_id) REFERENCES shopping_lists(id),
    INDEX idx_list_id (list_id)
);
```

## 🔒 Security Features

1. **Password Hashing**: bcrypt with cost factor 14
2. **JWT Tokens**: 24-hour expiration
3. **Token Validation**: Middleware checks every request
4. **User Isolation**: Users can only access their own data
5. **CORS**: Configured for cross-origin requests
6. **SQL Injection**: Protected by GORM parameterization

## 🐛 Troubleshooting

### "Access denied for user 'bazarlist'@'localhost'"
```bash
# Grant privileges again
mysql -u root -p -e "GRANT ALL PRIVILEGES ON bazarlist.* TO 'bazarlist'@'localhost';"
mysql -u root -p -e "FLUSH PRIVILEGES;"
```

### "Failed to connect to database"
- Check MySQL is running: `sudo systemctl status mysql`
- Verify credentials in environment variables
- Check MySQL logs: `sudo tail -f /var/log/mysql/error.log`

### "Invalid or expired token"
- Token expires after 24 hours
- User must login again
- Check JWT_SECRET is same on server

### "Failed to migrate database"
- Check database exists: `mysql -u root -p -e "SHOW DATABASES;"`
- Verify user has CREATE TABLE privileges
- Check MySQL version (5.7+ required)

## 📊 Testing the Setup

### Test Database Connection
```bash
mysql -u bazarlist -p bazarlist
# If successful, you'll see: mysql>
```

### Test API Registration
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "testpass123"
  }'
```

### Test API Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }'
```

### Test Protected Endpoint (with token)
```bash
TOKEN="your-jwt-token-here"
curl http://localhost:8080/api/lists \
  -H "Authorization: Bearer $TOKEN"
```

## 🚀 Production Deployment

### Security Checklist

- [ ] Change `JWT_SECRET` environment variable
- [ ] Use strong MySQL passwords
- [ ] Enable SSL/TLS for HTTPS
- [ ] Set up firewall rules
- [ ] Enable MySQL slow query log
- [ ] Set up regular database backups
- [ ] Use environment-specific configs
- [ ] Enable rate limiting
- [ ] Add request logging
- [ ] Set up monitoring

### Environment Variables for Production

```bash
# Production configuration
export DB_HOST=production-db.example.com
export DB_PORT=3306
export DB_USER=bazarlist_prod
export DB_PASSWORD=super-secure-password
export DB_NAME=bazarlist_prod
export JWT_SECRET=use-256-bit-random-string
export PORT=8080
```

## 📚 Additional Resources

- [GORM Documentation](https://gorm.io/docs/)
- [MySQL Documentation](https://dev.mysql.com/doc/)
- [JWT in Go](https://github.com/golang-jwt/jwt)
- [Bcrypt for Go](https://github.com/golang/crypto)

---

**Need Help?**

Check the main [README.md](../README.md) or [WEB_APPLICATION.md](WEB_APPLICATION.md) for more information.
