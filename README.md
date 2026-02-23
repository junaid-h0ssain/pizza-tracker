# Pizza Tracker

A real-time pizza order tracking web application built with Go. Customers submit orders and follow their progress through a live status tracker. Admins manage all orders from a protected dashboard, and status updates are pushed to the customer's page instantly without a manual refresh.

---

## Features

**Customer**
- Place an order with name, phone, address, and one or more pizzas (type, size, special instructions)
- Get redirected to a personal tracking page after ordering
- See a live progress bar: Order Placed, Preparing, Baking, Quality Check, Ready
- Page updates automatically when the admin changes the order status

**Admin**
- Log in securely at `/login`
- View all orders sorted by newest first
- Update order status via a dropdown — the customer's page updates instantly via SSE
- Delete orders with a confirmation prompt
- See a real-time badge when new orders arrive

---

## Tech Stack

| Technology | Purpose |
|---|---|
| Go | Primary language |
| Gin | HTTP web framework — routing, middleware, HTML rendering |
| GORM | ORM for working with SQLite using Go structs |
| SQLite | Embedded database, no separate server required |
| gin-contrib/sessions | Session management backed by the SQLite database |
| golang.org/x/crypto | bcrypt password hashing for admin accounts |
| go-playground/validator | Struct-tag-based form validation with custom rules |
| teris-io/shortid | Short unique ID generation for orders |
| Server-Sent Events (SSE) | Real-time push notifications from server to browser |
| Tailwind CSS (CDN) | Utility-first CSS styling via CDN, no build step required |
| Go html/template | Server-side HTML templating |



---

## Prerequisites

- Go 1.25 or later
- GCC or a C compiler (required by the SQLite driver, which uses CGO)

On Debian/Ubuntu:
```bash
sudo apt install gcc
```

On macOS (with Xcode Command Line Tools already installed, no extra step needed).

---

## Running Locally

1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/pizza-tracker.git
   cd pizza-tracker
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run the application:
   ```bash
   go run ./cmd
   ```

   The server starts at `http://localhost:8080` by default.

4. Create the first admin account (see below), then visit `http://localhost:8080`.

---

## Environment Variables

All variables are optional. The application falls back to the defaults shown below.

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Port the server listens on |
| `DATABASE_URL` | `./data/orders.db` | Path to the SQLite database file |
| `SESSION_SECRET_KEY` | `pizza-order-secret-key` | Secret key used to sign session cookies |

Example:
```bash
PORT=9090 SESSION_SECRET_KEY=my-secret go run ./cmd
```

---

## Creating the First Admin User

The application does not ship with any admin accounts. You need to create one directly in the database after the first run (which creates the schema automatically).

Using the SQLite CLI:

```bash
# Install sqlite3 if needed: sudo apt install sqlite3
sqlite3 data/orders.db
```

Inside the SQLite shell, insert a bcrypt-hashed password. You can generate one with a small Go snippet:

```go
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte("your-password"), bcrypt.DefaultCost)
    fmt.Println(string(hash))
}
```

Then insert the user:
```sql
INSERT INTO users (username, password) VALUES ('admin', '<bcrypt-hash>');
.quit
```

After that, log in at `http://localhost:8080/login`.

---

