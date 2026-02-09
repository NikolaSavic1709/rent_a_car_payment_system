# PayPal Payment Microservice

A Go-based microservice for handling PayPal payment processing in the Rent-a-Car Payment System.

## Overview

This microservice integrates with PayPal's REST API to enable PayPal Checkout payments. It handles the complete payment flow including order creation, payment approval, and status callbacks to the Payment Service Provider (PSP).

## Features

- **PayPal REST API Integration**: Uses PayPal SDK for sandbox/production environments
- **Order Management**: Create and manage PayPal orders
- **Payment Callbacks**: Success and cancel callback handling
- **Status Tracking**: Retrieve payment status and transaction details
- **Database Persistence**: PostgreSQL for storing payment transactions
- **Normalized Status**: Maps PayPal statuses to PSP-compatible statuses

## Architecture

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│     PSP     │─────▶│   PayPal    │─────▶│   PayPal    │
│  (Caller)   │      │Microservice │      │  REST API   │
└─────────────┘      └─────────────┘      └─────────────┘
                            │
                            ▼
                     ┌─────────────┐
                     │ PostgreSQL  │
                     │  Database   │
                     └─────────────┘
```

## API Endpoints

### Payment Management

- `POST /payment` - Create a new PayPal order
- `GET /payment-status/:paymentId` - Get payment status
- `GET /payment-success` - PayPal success callback handler
- `GET /payment-cancel` - PayPal cancel/failure callback handler

### Health & Diagnostics

- `GET /health` - Health check endpoint
- `GET /` - Service information

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `PORT` | Service port | `8080` |
| `DB_HOST` | PostgreSQL host | `psql_paypal` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `paypal_user` |
| `DB_PASSWORD` | Database password | `paypal_password` |
| `DB_NAME` | Database name | `paypal_db` |
| `PAYPAL_CLIENT_ID` | PayPal Client ID | `your-client-id` |
| `PAYPAL_SECRET` | PayPal Secret | `your-secret` |
| `PAYPAL_MODE` | PayPal mode (sandbox/live) | `sandbox` |
| `PSP_CALLBACK_URL` | PSP callback endpoint | `http://psp_service:8080/payment-callback` |
| `SUCCESS_URL` | Success redirect URL | `http://localhost:3001/paypal?status=success` |
| `CANCEL_URL` | Cancel redirect URL | `http://localhost:3001/paypal?status=cancel` |

## Payment Flow

1. **Initiate Payment**: PSP calls `POST /payment` with amount, currency, transaction details
2. **Create Order**: Service creates PayPal order and returns approval URL
3. **User Approval**: User is redirected to PayPal, logs in, and approves payment
4. **Callback**: PayPal redirects to success/cancel URL with order details
5. **Capture Payment**: Service captures the payment and updates database
6. **PSP Notification**: Service sends callback to PSP with payment status
7. **Final Redirect**: User is redirected back to merchant frontend

## Database Schema

### paypal_payments

Stores PayPal payment transactions.

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL | Primary key |
| `payment_id` | UUID | Unique payment identifier |
| `transaction_id` | UUID | PSP transaction ID |
| `merchant_order_id` | UUID | Merchant order ID |
| `merchant_id` | INTEGER | Merchant identifier |
| `paypal_order_id` | VARCHAR | PayPal order ID |
| `amount` | DECIMAL | Payment amount |
| `currency` | VARCHAR | Currency code (USD, EUR, etc.) |
| `status` | INTEGER | Payment status (0-4) |
| `payer_email` | VARCHAR | PayPal payer email |
| `payer_id` | VARCHAR | PayPal payer ID |
| `created_at` | TIMESTAMP | Creation timestamp |
| `approved_at` | TIMESTAMP | Approval timestamp |
| `completed_at` | TIMESTAMP | Completion timestamp |

### Payment Statuses

- `0` - Pending (order created, awaiting approval)
- `1` - Approved (user approved, awaiting capture)
- `2` - Completed (payment captured successfully)
- `3` - Cancelled (user cancelled payment)
- `4` - Failed (payment processing failed)

## Development

### Prerequisites

- Go 1.24+
- PostgreSQL 13+
- PayPal Sandbox Account

### Getting Started

1. Clone the repository
2. Set up environment variables
3. Run database migrations:
   ```bash
   psql -h localhost -U paypal_user -d paypal_db -f schema.sql
   ```
4. Run the service:
   ```bash
   make run
   # or with air for hot reload
   air
   ```

### Testing

```bash
# Run tests
make test

# Run with coverage
go test -cover ./...
```

### Building

```bash
# Build binary
make build

# Build Docker image
docker build -t paypal-microservice .
```

## Docker Deployment

```bash
# Run with docker-compose
docker-compose up -d

# View logs
docker-compose logs -f paypal_service
```

## Integration with PSP

The PayPal microservice expects the following request format from PSP:

```json
{
  "transactionId": "uuid",
  "merchantOrderId": "uuid",
  "merchantId": 12345,
  "amount": 100.00,
  "currency": "USD"
}
```

And returns:

```json
{
  "paymentId": "uuid",
  "paypalOrderId": "PAYPAL-ORDER-ID",
  "approvalUrl": "https://www.sandbox.paypal.com/checkoutnow?token=...",
  "status": "pending"
}
```

## Callback to PSP

When payment is completed/cancelled, the service sends a callback to PSP:

```json
{
  "transactionId": "uuid",
  "merchantOrderId": "uuid",
  "status": 0,
  "paypalOrderId": "PAYPAL-ORDER-ID",
  "amount": 100.00,
  "currency": "USD",
  "paypalTimestamp": "2026-02-09T12:00:00Z"
}
```

## Security Considerations

- All PayPal API calls use OAuth 2.0 authentication
- Secrets are stored in environment variables
- Database credentials are encrypted
- CORS is configured for specific origins only
- Sandbox mode for development/testing

## License

MIT

## Support

For issues and questions, please open an issue in the repository.
