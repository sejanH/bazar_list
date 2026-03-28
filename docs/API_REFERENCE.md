# Bazar List API Reference

Complete API documentation for the Bazar List REST API.

## Base URL

```
http://localhost:8080/api
```

## Authentication

Currently, the API does not require authentication. This will be added in a future version.

## Response Format

All API responses are in JSON format.

### Success Response
```json
{
  "data": { ... }
}
```

### Error Response
```json
{
  "error": "Error message here"
}
```

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 404 | Not Found |
| 500 | Internal Server Error |

---

## Endpoints

### Items

#### Get All Items

Get all shopping list items.

```http
GET /api/items
```

**Response:**
```json
[
  {
    "id": 1,
    "name": "Milk",
    "category": "dairy",
    "status": "pending",
    "created_at": "2024-03-28T10:00:00Z",
    "updated_at": "2024-03-28T10:00:00Z"
  }
]
```

---

#### Get Single Item

Get a specific item by ID.

```http
GET /api/items/{id}
```

**Parameters:**
- `id` (integer) - Item ID

**Response:**
```json
{
  "id": 1,
  "name": "Milk",
  "category": "dairy",
  "status": "pending",
  "created_at": "2024-03-28T10:00:00Z",
  "updated_at": "2024-03-28T10:00:00Z"
}
```

---

#### Add Item

Add a new item to the shopping list.

```http
POST /api/items
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "Milk",
  "category": "dairy"
}
```

**Fields:**
- `name` (string, required) - Item name
- `category` (string, optional) - Item category (default: "other")

**Valid Categories:**
- `produce` - Fruits and vegetables
- `dairy` - Milk, cheese, yogurt
- `meat` - Beef, chicken, pork
- `pantry` - Pasta, rice, canned goods
- `frozen` - Frozen foods
- `bakery` - Bread, pastries
- `beverages` - Drinks, juices
- `household` - Cleaning supplies
- `other` - Everything else

**Response (201 Created):**
```json
{
  "id": 1,
  "name": "Milk",
  "category": "dairy",
  "status": "pending",
  "created_at": "2024-03-28T10:00:00Z",
  "updated_at": "2024-03-28T10:00:00Z"
}
```

---

#### Update Item

Update an existing item.

```http
PUT /api/items/{id}
Content-Type: application/json
```

**Parameters:**
- `id` (integer) - Item ID

**Request Body:**
```json
{
  "name": "Organic Milk",
  "category": "dairy"
}
```

**Fields:**
- `name` (string, optional) - New item name
- `category` (string, optional) - New category

**Response (200 OK):**
```json
{
  "id": 1,
  "name": "Organic Milk",
  "category": "dairy",
  "status": "pending",
  "created_at": "2024-03-28T10:00:00Z",
  "updated_at": "2024-03-28T11:00:00Z"
}
```

---

#### Delete Item

Delete an item from the shopping list.

```http
DELETE /api/items/{id}
```

**Parameters:**
- `id` (integer) - Item ID

**Response (200 OK):**
```json
{
  "message": "Item deleted successfully"
}
```

---

#### Complete Item

Mark an item as completed/purchased.

```http
POST /api/items/{id}/complete
```

**Parameters:**
- `id` (integer) - Item ID

**Response (200 OK):**
```json
{
  "id": 1,
  "name": "Milk",
  "category": "dairy",
  "status": "completed",
  "created_at": "2024-03-28T10:00:00Z",
  "updated_at": "2024-03-28T12:00:00Z"
}
```

---

#### Make Item Pending

Mark a completed item as pending.

```http
POST /api/items/{id}/pending
```

**Parameters:**
- `id` (integer) - Item ID

**Response (200 OK):**
```json
{
  "id": 1,
  "name": "Milk",
  "category": "dairy",
  "status": "pending",
  "created_at": "2024-03-28T10:00:00Z",
  "updated_at": "2024-03-28T12:30:00Z"
}
```

---

### Search & Filter

#### Search Items

Search for items by name.

```http
GET /api/search?q=query
```

**Query Parameters:**
- `q` (string, required) - Search query

**Response:**
```json
[
  {
    "id": 1,
    "name": "Milk",
    "category": "dairy",
    "status": "pending",
    "created_at": "2024-03-28T10:00:00Z",
    "updated_at": "2024-03-28T10:00:00Z"
  }
]
```

---

#### Get Items by Category

Get all items in a specific category.

```http
GET /api/items/category/{category}
```

**Parameters:**
- `category` (string) - Category name

**Response:**
```json
[
  {
    "id": 1,
    "name": "Milk",
    "category": "dairy",
    "status": "pending",
    "created_at": "2024-03-28T10:00:00Z",
    "updated_at": "2024-03-28T10:00:00Z"
  },
  {
    "id": 2,
    "name": "Cheese",
    "category": "dairy",
    "status": "pending",
    "created_at": "2024-03-28T10:00:00Z",
    "updated_at": "2024-03-28T10:00:00Z"
  }
]
```

---

### Statistics

#### Get Statistics

Get shopping list statistics.

```http
GET /api/stats
```

**Response:**
```json
{
  "total": 10,
  "pending": 7,
  "completed": 3
}
```

**Fields:**
- `total` (integer) - Total number of items
- `pending` (integer) - Number of pending items
- `completed` (integer) - Number of completed items

---

## Examples

### cURL Examples

#### Add an Item
```bash
curl -X POST http://localhost:8080/api/items \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Apples",
    "category": "produce"
  }'
```

#### Get All Items
```bash
curl http://localhost:8080/api/items
```

#### Search for Items
```bash
curl "http://localhost:8080/api/search?q=milk"
```

#### Complete an Item
```bash
curl -X POST http://localhost:8080/api/items/1/complete
```

#### Delete an Item
```bash
curl -X DELETE http://localhost:8080/api/items/1
```

#### Get Statistics
```bash
curl http://localhost:8080/api/stats
```

### JavaScript/Fetch Examples

#### Add an Item
```javascript
fetch('http://localhost:8080/api/items', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    name: 'Milk',
    category: 'dairy'
  })
})
  .then(response => response.json())
  .then(data => console.log(data));
```

#### Get All Items
```javascript
fetch('http://localhost:8080/api/items')
  .then(response => response.json())
  .then(items => console.log(items));
```

#### Complete an Item
```javascript
fetch('http://localhost:8080/api/items/1/complete', {
  method: 'POST'
})
  .then(response => response.json())
  .then(data => console.log(data));
```

### Python/requests Examples

#### Add an Item
```python
import requests

response = requests.post('http://localhost:8080/api/items', json={
    'name': 'Milk',
    'category': 'dairy'
})
print(response.json())
```

#### Get All Items
```python
import requests

response = requests.get('http://localhost:8080/api/items')
print(response.json())
```

---

## Error Handling

### 400 Bad Request

Invalid request parameters or missing required fields.

```json
{
  "error": "Item name is required"
}
```

### 404 Not Found

Resource not found.

```json
{
  "error": "Item not found"
}
```

### 500 Internal Server Error

Server error occurred.

```json
{
  "error": "Failed to save shopping list"
}
```

---

## Rate Limiting

Currently, there is no rate limiting. This will be added in a future version.

## CORS

The API supports CORS for cross-origin requests. The following headers are set:

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

## Versioning

The API is currently at version 1. Future versions will be prefixed with `/api/v2/`, `/api/v3/`, etc.

## Changelog

### v1.0.0 (2024-03-28)
- Initial API release
- Basic CRUD operations for items
- Search and filter functionality
- Statistics endpoint

---

For more information, see the [Web Application Documentation](WEB_APPLICATION.md).
