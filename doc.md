# 🍕 Pizza Tracker — Project Documentation

A full-stack, real-time pizza order tracking web application built with Go. Customers place orders and watch their status update live; admins manage all orders from a protected dashboard.

---

## Table of Contents

1. [What It Does](#what-it-does)
2. [Tech Stack](#tech-stack)
3. [Project Structure](#project-structure)
4. [How It Works — Big Picture](#how-it-works--big-picture)
5. [Database Models](#database-models)
6. [Application Layers](#application-layers)
   - [Configuration (`utils.go`)](#configuration-utilsgo)
   - [Validators (`validators.go`)](#validators-validatorsgo)
   - [Handlers (`handlers.go`)](#handlers-handlersgo)
   - [Customer Flow (`customer.go`)](#customer-flow-customergo)
   - [Admin Flow (`admin.go`)](#admin-flow-admingo)
   - [Middleware (`middleware.go`)](#middleware-middlewarego)
   - [Real-Time Notifications (`notifications.go`, `events.go`)](#real-time-notifications-notificationsgo-eventsgo)
   - [Routes (`routes.go`)](#routes-routesgo)
   - [Entry Point (`main.go`)](#entry-point-maingo)
7. [Templates](#templates)
8. [Real-Time Updates (SSE) Explained](#real-time-updates-sse-explained)
9. [Authentication & Sessions Explained](#authentication--sessions-explained)
10. [API / Route Reference](#api--route-reference)
11. [Getting Started](#getting-started)
12. [Environment Variables](#environment-variables)
13. [Creating the First Admin User](#creating-the-first-admin-user)
14. [Data Flow Diagrams](#data-flow-diagrams)

---

## What It Does

### Customer Features
- Fill out an order form with their name, phone, address, and one or more pizzas (type, size, special instructions).
- Get redirected to a personal tracking page after placing the order.
- See a visual progress bar (Order Placed → Preparing → Baking → Quality Check → Ready).
- **Receive automatic live updates** — the page refreshes itself the moment an admin changes the order status.

### Admin Features
- Log in securely at `/login` (bcrypt-hashed passwords, database-backed sessions).
- View all orders in a table sorted by newest first.
- Change any order's status via a dropdown — the customer's page updates instantly.
- Delete orders (with a confirmation dialog).
- See a **real-time badge** when new orders arrive, without needing to refresh.

---

## Tech Stack

| Package | Purpose |
|---|---|
| [`github.com/gin-gonic/gin`](https://github.com/gin-gonic/gin) | HTTP web framework — routing, middleware, template rendering |
| [`gorm.io/gorm`](https://gorm.io) | ORM — work with SQLite using Go structs instead of raw SQL |
| [`gorm.io/driver/sqlite`](https://gorm.io/docs/connecting_to_the_database.html) | SQLite driver for GORM |
| [`github.com/gin-contrib/sessions`](https://github.com/gin-contrib/sessions) | Session middleware for Gin |
| [`github.com/gin-contrib/sessions/gorm`](https://github.com/gin-contrib/sessions) | GORM-backed session store (persists sessions to SQLite) |
| [`golang.org/x/crypto/bcrypt`](https://pkg.go.dev/golang.org/x/crypto/bcrypt) | Secure password hashing |
| [`github.com/go-playground/validator/v10`](https://github.com/go-playground/validator) | Struct-tag-based form validation with custom rules |
| [`github.com/teris-io/shortid`](https://github.com/teris-io/shortid) | Generates short unique IDs for orders (e.g. `aBc123XyZ`) |
| Tailwind CSS (CDN) | Utility-first CSS framework for styling |

---

## Project Structure

```
pizza-tracker/
├── cmd/                    # Application code (main package)
│   ├── main.go             # Entry point — wires everything together
│   ├── handlers.go         # Handler struct — bundles all dependencies
│   ├── routes.go           # URL routing
│   ├── customer.go         # Customer order & tracking handlers
│   ├── admin.go            # Admin login, dashboard, order management handlers
│   ├── middleware.go       # Auth middleware (protects admin routes)
│   ├── notifications.go    # NotificationManager — tracks SSE clients
│   ├── events.go           # SSE HTTP handlers (stream events to browsers)
│   ├── validators.go       # Custom form validators (pizza type/size)
│   └── utils.go            # Config, template loading, session helpers
├── models/                 # Database models
│   ├── models.go           # DB initialization, AutoMigrate
│   ├── orders.go           # Order & OrderItem structs + DB methods
│   └── user.go             # User struct + authentication methods
├── tmp/                    # HTML templates
│   ├── base.tmpl           # Shared HTML head/foot ({{define "top"}} / {{define "bottom"}})
│   ├── orders.tmpl         # Customer order form
│   ├── customer.tmpl       # Customer order tracking page
│   ├── admin.tmpl          # Admin dashboard
│   ├── login.tmpl          # Admin login page
│   └── static/
│       └── pizza.svg       # Spinning pizza SVG
├── data/
│   └── orders.db           # SQLite database file (auto-created)
├── go.mod
├── go.sum
└── doc.md                  # This file
```

---

## How It Works — Big Picture

```
Browser (Customer)          Go Server                      SQLite DB
      |                         |                               |
      |  GET /                  |                               |
      |------------------------>|                               |
      |  HTML order form        |                               |
      |<------------------------|                               |
      |                         |                               |
      |  POST /new-order        |                               |
      |------------------------>|  INSERT order + items         |
      |                         |------------------------------>|
      |                         |  Notify "admin:new_orders"    |
      |  Redirect /customer/:id |                               |
      |<------------------------|                               |
      |                         |                               |
      |  GET /customer/:id      |                               |
      |------------------------>|  SELECT order WHERE id=?      |
      |                         |------------------------------>|
      |  Tracking page + SSE    |                               |
      |<------------------------|                               |
      |  (SSE connection open)  |                               |

Browser (Admin)             Go Server
      |  POST /admin/order/:id/update                           |
      |------------------------>|  UPDATE orders SET status=?   |
      |                         |------------------------------>|
      |                         |  Notify "order:<id>"          |
      |                         |       |                        |
      |                         |  SSE message sent to customer |
      |  Customer page reloads <--------------------------------|
```

---

## Database Models

### `Order` (`models/orders.go`)

```
orders table
┌──────────────┬──────────────────────────────────────────┐
│ id           │ Short unique string (e.g. "aBc123XyZ")   │
│ status       │ One of the OrderStatuses values           │
│ customer_name│ Customer's full name                      │
│ phone        │ Customer's phone number                   │
│ address      │ Delivery address                          │
│ created_at   │ Timestamp (auto-managed by GORM)          │
└──────────────┴──────────────────────────────────────────┘
```

An `Order` has many `OrderItem`s (one per pizza).

### `OrderItem` (`models/orders.go`)

```
order_items table
┌──────────────┬──────────────────────────────────┐
│ id           │ Short unique string               │
│ order_id     │ Foreign key → orders.id           │
│ size         │ "Small" / "Medium" / "Large" / …  │
│ item_name    │ Pizza type (e.g. "Pepperoni")      │
│ instructions │ Optional special instructions      │
└──────────────┴──────────────────────────────────┘
```

### `User` (`models/user.go`)

```
users table
┌──────────┬──────────────────────────┐
│ id       │ String primary key        │
│ username │ Unique index              │
│ password │ bcrypt hash (never plain) │
└──────────┴──────────────────────────┘
```

### Predefined constants in `models/orders.go`

```go
OrderStatuses = []string{"Order Placed", "Preparing", "Baking", "Quality Check", "Ready"}
PizzaTypes    = []string{"Margherita", "Pepperoni", "Vegetarian", "Hawaiian", ...}
PizzaSizes    = []string{"Small", "Medium", "Large", "X-Large"}
```

These are the only accepted values. Custom validators enforce them at the HTTP layer.

---

## Application Layers

### Configuration (`utils.go`)

`loadConfig()` reads environment variables with sensible defaults:

```go
Config{
    Port:             "8080",
    DBPath:           "./data/orders.db",
    SessionSecretKey: "pizza-order-secret-key",
}
```

`loadTemplates()` parses all `*.tmpl` files from the `tmp/` directory and registers two template helper functions:

| Helper | Usage | Example |
|---|---|---|
| `add` | Integer addition | `{{add $index 1}}` → `1` when index is `0` |
| `json` | Serialize Go value to safe JS | `{{json .Statuses}}` → `["Order Placed","Preparing",...]` |

`setupSessionStore()` creates a GORM-backed cookie session store with security settings:
- **MaxAge**: 24 hours
- **HttpOnly**: Prevents JavaScript cookie access (XSS protection)
- **SameSite Strict**: Prevents CSRF attacks
- **Secure**: HTTPS-only cookies (set to `false` for local HTTP development if needed)

---

### Validators (`validators.go`)

Two custom validators are registered with Gin's validator engine at startup:

| Tag | Validates |
|---|---|
| `valid_pizza_type` | Value exists in `models.PizzaTypes` |
| `valid_pizza_size` | Value exists in `models.PizzaSizes` |

Used in `OrderRequest`:
```go
Sizes      []string `form:"size"  binding:"required,min=1,dive,valid_pizza_size"`
PizzaTypes []string `form:"pizza" binding:"required,min=1,dive,valid_pizza_type"`
```

`dive` tells the validator to validate each element of the slice individually.

---

### Handlers (`handlers.go`)

The `Handler` struct is the central dependency container:

```go
type Handler struct {
    orders              *models.OrderModel       // DB access for orders
    users               *models.UserModel        // DB access for users
    notificationManager *NotificationManager     // SSE client registry
}
```

`NewHandler(dbModel)` wires everything together once during startup. All handler methods are defined on `*Handler`, giving every handler access to all dependencies without passing them as function parameters.

---

### Customer Flow (`customer.go`)

| Function | What it does |
|---|---|
| `ServeNewOrderForm` | Renders the order form (`orders.tmpl`) with pizza type/size lists |
| `HandleNewOrderPost` | Validates form, creates `Order` + `[]OrderItem` in DB, notifies admins via SSE, redirects to tracking page |
| `serveCustomer` | Loads the order from DB, renders `customer.tmpl` with order data and all statuses |

**Key data structures:**

```go
// Passed to orders.tmpl
type OrderFormData struct {
    PizzaTypes []string
    PizzaSizes []string
}

// Passed to customer.tmpl
type CustomerData struct {
    Title    string
    Order    models.Order
    Statuses []string   // Used to build the progress bar
}

// Parsed from the POST form
type OrderRequest struct {
    Name         string
    Phone        string
    Address      string
    Sizes        []string   // One entry per pizza
    PizzaTypes   []string   // Parallel to Sizes
    Instructions []string   // Parallel to Sizes
}
```

When the form submits, the browser sends parallel arrays:
`size=Small&size=Large&pizza=Pepperoni&pizza=Hawaiian&instructions=Extra+cheese&instructions=`

The handler zips these arrays into `[]OrderItem` before saving.

---

### Admin Flow (`admin.go`)

| Function | What it does |
|---|---|
| `HandleLoginGet` | Renders `login.tmpl` |
| `HandleLoginPost` | Validates credentials via `AuthenticateUser` (bcrypt), stores `userID` and `username` in session, redirects to `/admin` |
| `HandleLogout` | Clears the session, redirects to `/login` |
| `ServeAdminDashboard` | Fetches all orders (newest first), renders `admin.tmpl` |
| `HandleOrderPut` | Updates order status in DB, sends SSE notification to the customer's tracking page, redirects back to dashboard |
| `HandleOrderDelete` | Deletes order and all its items from DB, redirects back to dashboard |

**Password security:** `AuthenticateUser` in `models/user.go` always returns a generic `"invalid credentials"` error whether the username doesn't exist or the password is wrong. This prevents attackers from discovering valid usernames.

---

### Middleware (`middleware.go`)

`AuthMiddleware()` is applied to the entire `/admin` route group:

1. Reads `userID` from the session.
2. If missing → redirect to `/login` and abort.
3. Looks up the user in the database.
4. If not found (user was deleted) → clear session, redirect to `/login`, abort.
5. Stores the `user` object in `c.Set("user", user)` for downstream handlers.
6. Calls `c.Next()` to proceed.

---

### Real-Time Notifications (`notifications.go`, `events.go`)

#### `NotificationManager` (`notifications.go`)

Uses a **publish-subscribe** pattern to track active SSE connections:

```
clients map:
  "order:aBc123XyZ"  →  { chan1: true, chan2: true }
  "admin:new_orders" →  { chan3: true }
```

Each browser connection gets its own `chan string` (buffered with capacity 10).

| Method | Thread-safety | What it does |
|---|---|---|
| `AddClient(key, ch)` | Write lock | Registers a new channel under a key |
| `RemoveClient(key, ch)` | Write lock | Removes channel, closes it to stop the SSE stream |
| `Notify(key, msg)` | Read lock | Sends message to all channels under a key (non-blocking: skips full channels) |

`sync.RWMutex` is used because: multiple goroutines read the map concurrently when sending notifications, but writes (add/remove) need exclusive access.

#### SSE Handlers (`events.go`)

**`notificationHandler`** — for customers at `/notifications?orderId=xxx`
1. Validates the `orderId` query parameter and confirms the order exists.
2. Creates a channel and registers it as `"order:{orderId}"`.
3. Defers cleanup (unregisters channel on disconnect).
4. Calls `streamSSE` to start streaming.

**`adminNotificationHandler`** — for admins at `/admin/notifications`
1. Creates a channel and registers it as `"admin:new_orders"`.
2. Defers cleanup.
3. Calls `streamSSE` to start streaming.

**`streamSSE`** — shared helper
Sets the required HTTP headers (`Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`) then uses Gin's `c.Stream()` to loop until the channel closes:
```go
c.Stream(func(w io.Writer) bool {
    if msg, ok := <-client; ok {
        c.SSEvent("message", msg)
        return true   // keep streaming
    }
    return false      // channel closed → stop streaming
})
```

---

### Routes (`routes.go`)

```
GET  /                         → ServeNewOrderForm
POST /new-order                → HandleNewOrderPost
GET  /customer/:id             → serveCustomer
GET  /notifications            → notificationHandler (SSE)

GET  /login                    → HandleLoginGet
POST /login                    → HandleLoginPost
POST /logout                   → HandleLogout

GET  /admin                    → ServeAdminDashboard       ⟵ Auth required
POST /admin/order/:id/update   → HandleOrderPut            ⟵ Auth required
POST /admin/order/:id/delete   → HandleOrderDelete         ⟵ Auth required
GET  /admin/notifications      → adminNotificationHandler  ⟵ Auth required

GET  /static/*filepath         → Serves tmp/static/
```

Session middleware (`sessions.Sessions`) is applied globally so every request has access to session data.

---

### Entry Point (`main.go`)

Startup sequence (order matters):

```
1. loadConfig()               → Read env vars
2. slog setup                 → Structured logger
3. models.InitDB()            → Open SQLite, AutoMigrate tables
4. setupSessionStore()        → GORM-backed cookie session store
5. NewHandler()               → Bundle DB models + NotificationManager
6. gin.Default()              → Create router (Logger + Recovery middleware)
7. loadTemplates()            → Parse tmp/*.tmpl, register helpers
8. setupRoutes()              → Register all URL handlers
9. RegisterValidators()       → Register custom pizza validators
10. router.Run()              → Start HTTP server
```

---

## Templates

All templates share structure via `base.tmpl`:

```
base.tmpl   defines: {{define "top"}}  ...  {{end}}
                      {{define "bottom"}} ... {{end}}

other templates use:
    {{template "top" .}}
    ... page content ...
    {{template "bottom" .}}
```

| Template | Data passed | Purpose |
|---|---|---|
| `orders.tmpl` | `OrderFormData` (PizzaTypes, PizzaSizes) | Customer order form with dynamic pizza fields |
| `customer.tmpl` | `CustomerData` (Title, Order, Statuses) | Order tracking page with progress bar and SSE |
| `admin.tmpl` | `AdminDashboardData` (Orders, Statuses, Username) | Admin order table with status dropdowns and SSE badge |
| `login.tmpl` | `LoginData` (Error) | Admin login form |

### Template helper usage examples

```html
<!-- Add 1 to a zero-based index -->
Pizza #{{add $index 1}}

<!-- Render Go slice as JavaScript array -->
const statuses = {{json .Statuses}};

<!-- Conditional rendering -->
{{if $pizza.Instructions}}{{$pizza.Instructions}}{{else}}None{{end}}

<!-- Access root data inside a range -->
{{range $.Statuses}}...{{end}}
```

---

## Real-Time Updates (SSE) Explained

### How the customer tracking page updates automatically

1. The `customer.tmpl` opens an SSE connection on page load:
   ```js
   const eventSrc = new EventSource(`/notifications?orderId=${orderId}`);
   eventSrc.onmessage = () => location.reload();
   ```
2. When an admin changes the order status, `HandleOrderPut` calls:
   ```go
   h.notificationManager.Notify("order:"+orderID, "order_updated")
   ```
3. The `NotificationManager` finds all channels registered under `"order:{id}"` and sends the message.
4. The `streamSSE` goroutine reads from the channel and writes an SSE event to the HTTP response.
5. The browser's `EventSource` fires `onmessage`, triggering `location.reload()`.
6. The page reloads with the new status from the database.

### How the admin badge works

1. `admin.tmpl` opens an SSE connection on page load:
   ```js
   const eventSrc = new EventSource("/admin/notifications");
   eventSrc.onmessage = e => {
       newOrdersCount++;
       newOrdersBadge.textContent = `${newOrdersCount} new order(s)`;
       newOrdersBadge.classList.remove('hidden');
   };
   ```
2. When a customer places a new order, `HandleNewOrderPost` calls:
   ```go
   h.notificationManager.Notify("admin:new_orders", "new_order")
   ```
3. The badge appears/increments without a page reload.

---

## Authentication & Sessions Explained

### Login flow

```
POST /login
  → ShouldBind validates form
  → users.AuthenticateUser(username, password)
      → DB lookup by username
      → bcrypt.CompareHashAndPassword(storedHash, providedPassword)
  → SetSessionValue(c, "userID", user.ID)   // saves to DB via gormstore
  → SetSessionValue(c, "username", user.Username)
  → Redirect to /admin
```

### Session storage

Sessions are stored in the SQLite database (the same file as application data) via `gin-contrib/sessions/gorm`. The session cookie only contains an encrypted session ID — no user data is sent to the browser.

### Protecting admin routes

```go
admin := router.Group("/admin")
admin.Use(h.AuthMiddleware())
```

`AuthMiddleware` runs before every admin handler. It reads `userID` from the session and aborts with a redirect to `/login` if the user is not authenticated.

---

## API / Route Reference

### Public

| Method | Path | Description |
|---|---|---|
| `GET` | `/` | Order form |
| `POST` | `/new-order` | Submit new order |
| `GET` | `/customer/:id` | Order tracking page |
| `GET` | `/notifications?orderId=xxx` | SSE stream for customer tracking |
| `GET` | `/login` | Admin login form |
| `POST` | `/login` | Admin login submit |
| `POST` | `/logout` | Admin logout |

### Admin (authentication required)

| Method | Path | Description |
|---|---|---|
| `GET` | `/admin` | Admin dashboard |
| `POST` | `/admin/order/:id/update` | Update order status (`status` form field) |
| `POST` | `/admin/order/:id/delete` | Delete order and its items |
| `GET` | `/admin/notifications` | SSE stream for new order notifications |

---

## Getting Started

### Prerequisites

- Go 1.21+
- `gcc` (required by the SQLite driver for CGO)

### Run the app

```bash
# Clone and enter the project
cd pizza-tracker

# Download dependencies
go mod tidy

# Run the server
go run cmd/*.go
```

Server starts at **http://localhost:8080**

### Run in production mode

```bash
GIN_MODE=release go run cmd/*.go
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `DATABASE_URL` | `./data/orders.db` | Path to SQLite database file |
| `SESSION_SECRET_KEY` | `pizza-order-secret-key` | Secret key for encrypting session cookies |

> **Important:** Change `SESSION_SECRET_KEY` to a long random string in production. Changing it invalidates all existing sessions.

> **Note:** The `Secure: true` session option requires HTTPS. For local HTTP development, this may need to be temporarily set to `false` in `utils.go`.

---

## Creating the First Admin User

The application does not have a self-registration flow for admins. Create the first user manually with a bcrypt hash:

**Step 1:** Generate a bcrypt hash for your password at https://bcrypt-generator.com (use cost 12).

**Step 2:** Insert the user into the database:

```bash
sqlite3 data/orders.db 'INSERT INTO users (id, username, password) VALUES ("admin-user", "admin", "$2a$12$YOUR_HASH_HERE");'
```

**Step 3:** Verify:

```bash
sqlite3 -header -column data/orders.db "SELECT id, username FROM users;"
```

**Step 4:** Log in at http://localhost:8080/login

---

## Data Flow Diagrams

### New Order Submission

```
Customer fills form
        │
        ▼
POST /new-order
        │
        ▼
ShouldBind + custom validators
  (valid_pizza_type, valid_pizza_size)
        │
    pass?
   /     \
  No      Yes
  │         │
  ▼         ▼
400 JSON   Build []OrderItem
           from parallel form arrays
                │
                ▼
           DB: INSERT order + items
           (BeforeCreate hook generates shortid)
                │
                ▼
           Notify "admin:new_orders"
           (SSE badge on admin dashboard)
                │
                ▼
           Redirect → /customer/:id
```

### Admin Status Update

```
Admin selects status from dropdown
(onchange="this.form.submit()")
        │
        ▼
POST /admin/order/:id/update
        │
        ▼
AuthMiddleware checks session
        │
        ▼
DB: UPDATE orders SET status = ?
        │
        ▼
Notify "order:{id}"
        │        \
        │         SSE message sent to
        │         customer's browser
        │              │
        ▼              ▼
Redirect /admin    location.reload()
                   on customer page
```

### SSE Connection Lifecycle

```
Browser opens EventSource("/notifications?orderId=xxx")
        │
        ▼
Server creates channel ch (buffered, cap 10)
Registers ch under key "order:xxx"
Sets SSE headers, enters streaming loop
        │
        │  ←── waits on channel
        │
   [Admin updates status]
        │
        ▼
Notify("order:xxx", "order_updated")
  sends "order_updated" into ch
        │
        ▼
streamSSE reads from ch
c.SSEvent("message", "order_updated") written to response
        │
        ▼
Browser receives event → location.reload()
        │
   [Browser closes tab / navigates away]
        │
        ▼
HTTP connection closes
Gin's c.Stream() returns
defer runs: RemoveClient("order:xxx", ch)
  → channel closed, goroutine exits
```

