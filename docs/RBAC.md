# Role-Based Access Control (RBAC)

AITriage Enterprise enforces strict Role-Based Access Control (RBAC) to ensure operational security and segregation of duties.

## Global Roles

The system uses the following hierarchical global roles:

| Role | Access Level | Description |
|---|---|---|
| **superadmin** | **Unrestricted** | Full system access. Cannot be deleted. Can create other users and manage all system configurations. |
| **admin** | **High** | Can manage users (except superadmin), products, engagements, and findings. Cannot access system audit logs or alter superadmin accounts. |
| **manager** | **Medium** | Can create and edit products and engagements, run scans, and triage findings. Cannot manage users. |
| **developer** | **Targeted** | Can only update the status of findings (e.g., mark as fixed) and view assigned products. Cannot create products or engagements. |
| **viewer** | **Read-Only** | Can view dashboards, metrics, read findings, and access executive reports. Cannot execute scans or modify any data. |

## Initial Setup (Superadmin)

Upon the very first launch, AITriage Enterprise automatically initializes the database and creates a default superadmin user.

**Default Superadmin Credentials:**
- **Username:** `admin`
- **Password:** `admin` (Requires immediate change in production)
- **Role:** `superadmin`

> [!CAUTION]
> You must change the default superadmin password immediately upon first logging into the production environment via the Admin Panel -> Users tab.

## Role Enforcement Matrix

Access is enforced globally at the API router level via the `PermissionMiddleware`.

| Endpoint | Method | Allowed Roles |
|---|---|---|
| `/api/login` | POST | All (Unauthenticated) |
| `/api/admin/users` | GET, POST, DELETE | superadmin, admin |
| `/api/scan` | POST | superadmin, admin, manager |
| `/api/triage` | POST | superadmin, admin, manager |
| `/api/products` | GET, POST, PUT, DELETE | superadmin, admin, manager (GET is open to viewer) |
| `/api/findings` | PUT | superadmin, admin, manager, developer |
| `/api/metrics` | GET | superadmin, admin, manager, viewer |
| `/api/audit` | GET | superadmin, admin, manager |

## Token Management

All roles are authenticated via JSON Web Tokens (JWT).
- Tokens are issued upon successful login and stored securely in an `HttpOnly` cookie.
- Tokens have a fixed lifespan of **24 hours**.
- The `JWT_SECRET` must be kept secure. If the secret is compromised or rotated, all existing sessions will immediately invalidate.
