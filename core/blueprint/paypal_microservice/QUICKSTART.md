# PayPal Microservice - Quick Start Guide

## Prerequisites

- Go 1.24+
- Docker & Docker Compose
- PayPal Developer Account (for sandbox credentials)

## Step 1: Get PayPal Credentials

1. Go to https://developer.paypal.com/dashboard/
2. Log in or create a developer account
3. Navigate to "Apps & Credentials"
4. Click "Create App" in the Sandbox section
5. Copy the **Client ID** and **Secret**

## Step 2: Configure Environment

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and set your PayPal credentials:
   ```env
   PAYPAL_CLIENT_ID=your-actual-client-id
   PAYPAL_SECRET=your-actual-secret
   PAYPAL_MODE=sandbox
   ```

## Step 3: Run with Docker

```bash
# Build and start services
docker-compose up -d

# View logs
docker-compose logs -f paypal_service

# Check health
curl http://localhost:8087/health
```

## Step 4: Test the Service

### Create a Payment

```bash
curl -X POST http://localhost:8087/payment \
  -H "Content-Type: application/json" \
  -d '{
    "transactionId": "123e4567-e89b-12d3-a456-426614174000",
    "merchantOrderId": "123e4567-e89b-12d3-a456-426614174001",
    "merchantId": 12345,
    "amount": 100.00,
    "currency": "USD",
    "description": "Test Payment"
  }'
```

**Expected Response:**
```json
{
  "paymentId": "uuid",
  "paypalOrderId": "PAYPAL-ORDER-ID",
  "approvalUrl": "https://www.sandbox.paypal.com/checkoutnow?token=...",
  "status": "pending"
}
```

### Complete Payment Flow

1. Open the `approvalUrl` in your browser
2. Log in with a PayPal Sandbox buyer account
3. Approve the payment
4. You'll be redirected back to the success URL
5. Check payment status:

```bash
curl http://localhost:8087/payment-status/{paymentId}
```

## Step 5: Create PayPal Sandbox Test Accounts

1. Go to https://developer.paypal.com/dashboard/accounts
2. Create a **Business Account** (seller)
3. Create a **Personal Account** (buyer)
4. Note the email and password for testing

## Local Development (without Docker)

### Install Dependencies

```bash
make deps
```

### Setup Database

```bash
# Start PostgreSQL locally
docker run -d \
  --name paypal-postgres \
  -e POSTGRES_USER=paypal_user \
  -e POSTGRES_PASSWORD=paypal_password \
  -e POSTGRES_DB=paypal_db \
  -p 5439:5432 \
  postgres:latest

# Run migrations
psql -h localhost -p 5439 -U paypal_user -d paypal_db -f schema.sql
```

### Update .env for Local Development

```env
DB_HOST=localhost
DB_PORT=5439
PORT=8087
```

### Run the Service

```bash
# With hot reload
make dev

# Or build and run
make build
make run
```

## Integration with PSP

### PSP Configuration

In your PSP microservice, add PayPal service configuration:

```env
PAYPAL_SERVICE_URL=http://paypal_service:8080
```

### PSP Payment Handler

The PSP should forward PayPal payments to this service:

```go
if paymentMethod == "PAYPAL" {
    // Forward to PayPal microservice
    resp, err := http.Post(
        "http://paypal_service:8080/payment",
        "application/json",
        requestBody,
    )
}
```

### PSP Callback Endpoint

The PSP must have a callback endpoint to receive payment updates:

```go
router.PUT("/payment-callback", func(c *gin.Context) {
    var callback PayPalCallback
    c.ShouldBindJSON(&callback)
    
    // Update transaction status
    // Notify webshop
})
```

## Troubleshooting

### Service Won't Start

**Check logs:**
```bash
docker-compose logs paypal_service
```

**Common issues:**
- Missing PayPal credentials in `.env`
- Database not ready (wait a few seconds)
- Port 8087 already in use

### PayPal API Errors

**401 Unauthorized:**
- Verify `PAYPAL_CLIENT_ID` and `PAYPAL_SECRET`
- Check credentials are from sandbox (not live)

**422 Unprocessable Entity:**
- Check amount is positive
- Verify currency is supported (USD, EUR, GBP)

### Database Connection Failed

```bash
# Check database is running
docker-compose ps

# Restart database
docker-compose restart psql_paypal

# View database logs
docker-compose logs psql_paypal
```

### Payment Not Completing

1. Check PayPal sandbox account has funds
2. Verify approval URL was opened
3. Check service logs for capture errors
4. Review PayPal Dashboard for order status

## Testing Checklist

- [ ] Service starts successfully
- [ ] Health check returns 200
- [ ] Payment creation returns approval URL
- [ ] User can approve payment on PayPal
- [ ] Payment status updates to "completed"
- [ ] Callback sent to PSP
- [ ] User redirected to success page
- [ ] Database contains payment record

## API Reference

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Service info |
| GET | `/health` | Health check |
| POST | `/payment` | Create payment |
| GET | `/payment-status/:id` | Get payment status |
| GET | `/payment-success` | Success callback |
| GET | `/payment-cancel` | Cancel callback |

### Status Values

- `pending` - Order created, awaiting approval
- `approved` - User approved, awaiting capture
- `completed` - Payment successful
- `cancelled` - User cancelled
- `failed` - Payment failed

## Next Steps

1. **Integrate with PSP**: Update PSP to forward PayPal payments
2. **Update Frontend**: Create PayPal payment UI
3. **Configure Webhooks**: Set up PayPal IPN for real-time updates
4. **Production Setup**: Switch to live credentials when ready

## Support

- **Documentation**: See [README.md](README.md) and [ARCHITECTURE.md](ARCHITECTURE.md)
- **PayPal Docs**: https://developer.paypal.com/docs/
- **Issues**: Create an issue in the repository

## Clean Up

```bash
# Stop services
docker-compose down

# Remove volumes (deletes database data)
docker-compose down -v

# Remove images
docker rmi paypal-microservice:latest
```
