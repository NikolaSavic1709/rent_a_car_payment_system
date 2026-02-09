# PayPal Microservice Architecture

## Overview

The PayPal microservice is a standalone payment processing service that integrates with PayPal's REST API to handle PayPal Checkout payments. It follows the same architectural patterns as the existing crypto and bank microservices in the rent-a-car payment system.

## Architecture Diagram

```
┌──────────────┐
│   Webshop    │
│  (Customer)  │
└──────┬───────┘
       │ 1. Purchase Request
       ▼
┌──────────────┐
│     PSP      │
│  (Router)    │
└──────┬───────┘
       │ 2. Forward PayPal Payment
       ▼
┌──────────────┐      ┌──────────────┐
│   PayPal     │──3──▶│   PayPal     │
│Microservice  │      │   REST API   │
└──────┬───────┘      └──────────────┘
       │                      │
       │ 4. Approval URL      │ 5. User Approval
       ▼                      ▼
┌──────────────┐      ┌──────────────┐
│  PSP Frontend│──────│   Customer   │
│  (3001)      │      │   Browser    │
└──────────────┘      └──────────────┘
       │                      │
       │◀─────6. Redirect────┘
       │
       ▼
┌──────────────┐
│   PayPal     │
│Microservice  │──7. Callback──▶ PSP ──8. Notify──▶ Webshop
│ (Success)    │
└──────┬───────┘
       │
       │ 9. User Redirect
       ▼
┌──────────────┐
│  PSP Frontend│
│(Success Page)│
└──────────────┘
```

## Payment Flow

### 1. Payment Initiation
- PSP receives payment request from webshop
- PSP routes to PayPal microservice based on `paymentMethod: "PAYPAL"`
- Request includes: `transactionId`, `merchantOrderId`, `amount`, `currency`

### 2. Order Creation
- PayPal microservice calls PayPal REST API to create order
- PayPal returns `order_id` and `approval_url`
- Service stores payment in PostgreSQL database with status `Pending`

### 3. User Approval
- Service returns approval URL to PSP
- PSP redirects user to PayPal login/approval page
- User logs into PayPal and approves payment

### 4. Callback Handling
- PayPal redirects user back to service with order details
- Service captures the payment via PayPal API
- Database updated with status `Completed` and capture ID

### 5. PSP Notification
- Service sends callback to PSP with payment status
- PSP updates its transaction database
- PSP notifies webshop of successful payment

### 6. User Redirection
- Service redirects user to PSP frontend success page
- User sees payment confirmation

## API Endpoints

### Payment Management

#### POST /payment
Creates a new PayPal order.

**Request:**
```json
{
  "transactionId": "uuid",
  "merchantOrderId": "uuid",
  "merchantId": 12345,
  "amount": 100.00,
  "currency": "USD",
  "description": "Rent-a-Car Vehicle Purchase"
}
```

**Response:**
```json
{
  "paymentId": "uuid",
  "paypalOrderId": "PAYPAL-ORDER-ID",
  "approvalUrl": "https://www.sandbox.paypal.com/checkoutnow?token=...",
  "status": "pending"
}
```

#### GET /payment-status/:paymentId
Retrieves payment status.

**Response:**
```json
{
  "paymentId": "uuid",
  "transactionId": "uuid",
  "paypalOrderId": "PAYPAL-ORDER-ID",
  "status": "completed",
  "amount": 100.00,
  "currency": "USD",
  "payerEmail": "buyer@example.com",
  "createdAt": "2026-02-09T12:00:00Z",
  "completedAt": "2026-02-09T12:05:00Z"
}
```

### Callback Endpoints

#### GET /payment-success
Handles successful PayPal payment callback.

**Query Parameters:**
- `token`: PayPal order ID
- `PayerID`: PayPal payer ID

**Process:**
1. Retrieve payment from database
2. Capture payment via PayPal API
3. Update database status to `Completed`
4. Send callback to PSP
5. Redirect user to success URL

#### GET /payment-cancel
Handles cancelled PayPal payment.

**Query Parameters:**
- `token`: PayPal order ID

**Process:**
1. Retrieve payment from database
2. Update database status to `Cancelled`
3. Send callback to PSP
4. Redirect user to cancel URL

## Database Schema

### paypal_payments Table

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL | Primary key |
| `payment_id` | UUID | Unique payment identifier |
| `transaction_id` | UUID | PSP transaction ID |
| `merchant_order_id` | UUID | Merchant order ID |
| `merchant_id` | INTEGER | Merchant identifier |
| `paypal_order_id` | VARCHAR(255) | PayPal order ID |
| `paypal_capture_id` | VARCHAR(255) | PayPal capture ID |
| `amount` | DECIMAL(10,2) | Payment amount |
| `currency` | VARCHAR(3) | Currency code |
| `status` | INTEGER | Payment status (0-4) |
| `payer_email` | VARCHAR(255) | Payer email |
| `payer_id` | VARCHAR(255) | PayPal payer ID |
| `payer_name` | VARCHAR(255) | Payer full name |
| `created_at` | TIMESTAMP | Creation time |
| `approved_at` | TIMESTAMP | Approval time |
| `completed_at` | TIMESTAMP | Completion time |
| `cancelled_at` | TIMESTAMP | Cancellation time |
| `approval_url` | TEXT | PayPal approval URL |
| `description` | TEXT | Payment description |
| `failure_reason` | TEXT | Failure reason if failed |

### Payment Status Enum

```go
const (
    Pending   = 0  // Order created, awaiting approval
    Approved  = 1  // User approved, awaiting capture
    Completed = 2  // Payment captured successfully
    Cancelled = 3  // User cancelled payment
    Failed    = 4  // Payment processing failed
)
```

## Components

### 1. Database Layer (`internal/database/`)

**database.go**
- PostgreSQL connection management
- CRUD operations for payments
- Health check functionality

**models.go**
- Data models: `PayPalPayment`, `PaymentRequest`, `PaymentResponse`
- Status enums and constants
- Callback structures

### 2. Server Layer (`internal/server/`)

**server.go**
- HTTP server initialization
- PayPal client integration
- Dependency injection

**routes.go**
- Route registration
- CORS configuration
- Middleware setup

**api.go**
- Payment creation handler
- Payment status handler
- Success/cancel callback handlers

**paypal_client.go**
- PayPal SDK wrapper
- Order creation
- Payment capture
- Order retrieval

**callback.go**
- PSP callback sender
- Status mapping
- URL builders

### 3. Main Application (`cmd/api/`)

**main.go**
- Application entry point
- Database initialization
- Server startup
- Graceful shutdown

## Integration with PSP

### Request from PSP to PayPal Service

```http
POST http://paypal_service:8080/payment
Content-Type: application/json

{
  "transactionId": "uuid-from-psp",
  "merchantOrderId": "uuid-from-webshop",
  "merchantId": 12345,
  "amount": 150.00,
  "currency": "USD",
  "description": "Vehicle rental payment"
}
```

### Callback from PayPal Service to PSP

```http
PUT http://psp_service:8080/payment-callback
Content-Type: application/json

{
  "transactionId": "uuid-from-psp",
  "merchantOrderId": "uuid-from-webshop",
  "status": 0,
  "paypalOrderId": "PAYPAL-ORDER-ID",
  "amount": 150.00,
  "currency": "USD",
  "paypalTimestamp": "2026-02-09T12:05:00Z"
}
```

### Status Mapping

PayPal Status → PSP Status:
- `Completed` → `Successful (0)`
- `Approved` → `InProgress (1)`
- `Cancelled` → `Failed (2)`
- `Failed` → `Failed (2)`
- `Pending` → `InProgress (1)`

## Configuration

### Environment Variables

See [.env.example](./.env.example) for all configuration options.

**Critical Variables:**
- `PAYPAL_CLIENT_ID`: From PayPal Developer Dashboard
- `PAYPAL_SECRET`: From PayPal Developer Dashboard
- `PAYPAL_MODE`: `sandbox` or `live`
- `PSP_CALLBACK_URL`: PSP endpoint for callbacks

### PayPal Sandbox Setup

1. Create account at https://developer.paypal.com
2. Go to Dashboard → Apps & Credentials
3. Create new Sandbox app
4. Copy Client ID and Secret
5. Set in `.env` file

### Docker Deployment

**Port Allocation:**
- External: `8087` (host machine)
- Internal: `8080` (container)
- Database: `5439:5432` (host:container)

**Network:**
- Shares network with other services for inter-service communication
- PostgreSQL accessible at `psql_paypal:5432` from service

## Security Considerations

### Authentication
- OAuth 2.0 with PayPal API
- Access tokens refreshed automatically by SDK

### Data Protection
- Database credentials in environment variables
- PayPal secrets never logged or exposed
- HTTPS for all PayPal API calls

### CORS
- Restricted to specific frontend origins
- Credentials allowed for authenticated requests

### Sandbox Mode
- All testing uses PayPal Sandbox
- No real money transactions
- Test credentials from PayPal Developer Dashboard

## Testing

### Unit Tests
```bash
make test
```

### Integration Tests
Uses testcontainers for PostgreSQL:
- Database operations
- Payment creation
- Status updates

### Manual Testing
1. Start service: `docker-compose up -d`
2. Create payment via API
3. Visit approval URL in browser
4. Log in with PayPal Sandbox account
5. Approve payment
6. Verify callback to PSP
7. Check database status

### Test Accounts
Create sandbox accounts at:
https://developer.paypal.com/dashboard/accounts

## Error Handling

### Payment Creation Errors
- Invalid amount/currency → 400 Bad Request
- PayPal API failure → 500 Internal Server Error
- Database error → 500 Internal Server Error

### Callback Errors
- Missing order ID → Redirect to cancel URL
- Payment not found → Redirect to cancel URL
- Capture failure → Update status to Failed, notify PSP

### PSP Callback Errors
- Non-blocking (async goroutine)
- Logged but doesn't affect user flow
- Retries not implemented (future enhancement)

## Monitoring & Logging

### Log Levels
- Info: Payment creation, status updates, callbacks
- Error: PayPal API failures, database errors
- Debug: Request/response details (development only)

### Health Check
```http
GET /health
```

Returns database connection status and statistics.

## Future Enhancements

1. **Webhook Support**: PayPal IPN for real-time notifications
2. **Refunds**: Support for payment refunds
3. **Subscriptions**: Recurring payment support
4. **Multi-Currency**: Dynamic currency conversion
5. **Retry Logic**: Automatic PSP callback retries
6. **Metrics**: Prometheus metrics for monitoring
7. **Rate Limiting**: Protect against abuse

## Comparison with Other Services

| Feature | Bank Service | Crypto Service | PayPal Service |
|---------|--------------|----------------|----------------|
| External API | Bank Gateway | Blockchain | PayPal REST API |
| Approval Flow | Card Form | QR/Wallet | Redirect |
| Confirmation | Instant | Multi-step | Instant (after approval) |
| User Action | Enter card | Send crypto | Login & approve |
| Callback | Yes | Yes | Yes |
| Database | PostgreSQL | PostgreSQL | PostgreSQL |
| Port | 8081-8083 | 8086 | 8087 |

## Development Workflow

### Local Development
```bash
# Install dependencies
make deps

# Run with hot reload
make dev

# Format code
make fmt

# Run tests
make test
```

### Docker Development
```bash
# Build and run
docker-compose up --build

# View logs
docker-compose logs -f paypal_service

# Stop
docker-compose down
```

## Troubleshooting

### PayPal API Authentication Failed
- Verify `PAYPAL_CLIENT_ID` and `PAYPAL_SECRET`
- Check `PAYPAL_MODE` is set correctly
- Ensure credentials are from same environment (sandbox/live)

### Callback Not Received by PSP
- Check `PSP_CALLBACK_URL` configuration
- Verify PSP service is running
- Check network connectivity between services
- Review PSP logs for callback reception

### Database Connection Failed
- Verify PostgreSQL container is running
- Check database credentials in `.env`
- Ensure database initialized with `schema.sql`
- Review database health check logs

### Payment Not Captured
- Check PayPal Sandbox account has funds
- Verify order was approved by user
- Review PayPal API response in logs
- Check order status in PayPal Dashboard

## License

MIT License - See project root for details.
