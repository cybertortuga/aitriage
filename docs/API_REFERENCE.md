# AITriage Enterprise API Reference

The AITriage Enterprise API provides programmatic access to the platform's security findings, engagements, and products.

## Base URL
All API requests must be made to `/api/`

## Authentication
AITriage uses `HttpOnly` cookies for authentication to protect against XSS. 

1. Call `POST /api/login` with your credentials.
2. The server responds with a `Set-Cookie: token=...; HttpOnly; Path=/;` header.
3. Your HTTP client (or browser) will automatically include this cookie in subsequent requests.

### POST /api/login
Authenticates a user and issues a JWT session.

**Rate Limit**: 5 failed attempts per 10 minutes per IP.

**Request:**
```json
{
  "username": "admin",
  "password": "yourpassword"
}
```

**Response (200 OK):**
```json
{
  "ok": true,
  "user_id": 1,
  "username": "admin",
  "role": "superadmin",
  "is_admin": true,
  "token": "eyJhb..."
}
```

---

## Products

### GET /api/products
Lists all configured products.

**Response (200 OK):**
```json
{
  "products": [
    {
      "id": 1,
      "name": "Backend Services",
      "description": "Core microservices",
      "type": "api",
      "is_active": true,
      "tags": ["go", "grpc"]
    }
  ]
}
```

### POST /api/products
Creates a new product. (Requires `manager` or higher).

**Request:**
```json
{
  "name": "Frontend Application",
  "description": "React Dashboard",
  "type": "web",
  "tags": "react,typescript"
}
```

---

## Engagements

### GET /api/engagements
Lists all engagements, optionally filtered by `product_id` and `status`.

**Query Parameters:**
- `product_id` (optional): Filter by product.
- `status` (optional): Filter by status (`planned`, `in_progress`, `completed`).

### POST /api/engagements
Creates a new engagement. (Requires `manager` or higher).

**Request:**
```json
{
  "product_id": 1,
  "name": "Q3 Security Audit",
  "engagement_type": "sast_scan",
  "status": "planned",
  "start_date": "2026-07-01T00:00:00Z"
}
```

---

## Findings (Triage & Kanban)

### GET /api/findings
Lists findings across all engagements. Supports filtering by status for Kanban boards.

**Query Parameters:**
- `engagement_id` (optional)
- `status` (optional): `open`, `triage`, `resolved`, `false_positive`
- `severity` (optional): `CRITICAL`, `HIGH`, `MEDIUM`, `LOW`

### PUT /api/findings/{id}
Updates the status of a specific finding. Used heavily by the Kanban board to drag-and-drop issues.

**Request:**
```json
{
  "status": "resolved"
}
```

---

## Scans & Analysis

### POST /api/scan
Triggers a synchronized security scan against a specified target directory.

**Request:**
```json
{
  "path": "/project",
  "engagement_id": 1
}
```

**Response:**
Returns the complete scan report and automatically ingests the findings into the specified database engagement.

### GET /api/health
Health check endpoint. Evaluates both the web service and the backend database connection.

**Response (200 OK):**
```json
{
  "ok": true,
  "tools": {
    "semgrep": true,
    "bandit": true,
    "gitleaks": true,
    "trivy": true
  }
}
```
