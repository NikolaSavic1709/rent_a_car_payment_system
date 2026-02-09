# PayPal Integration Documentation

## Overview

This document describes the complete PayPal payment integration for the rent-a-car payment system. The integration follows a redirect-based checkout flow using PayPal's REST API in sandbox mode.

## Architecture

### Components

1. **Webshop** - Initiates payment requests
2. **PSP Service** - Payment Service Provider (orchestrator)
3. **PayPal Microservice** - Handles PayPal-specific operations
4. **PayPal Frontend** - User interface for payment processing
5. **PayPal Sandbox** - External PayPal test environment

### Payment Flow

```
┌─────────┐      ┌─────────┐      ┌──────────────┐      ┌──────────────┐
│ Webshop │─────>│   PSP   │─────>│    PayPal    │─────>│    PayPal    │
│         │      │ Service │      │ Microservice │      │   Frontend   │
└─────────┘      └─────────┘      └──────────────┘      └──────┬───────┘
     ▲                ▲                    │                     │
     │                │                    │                     │
     │                └────────────────────┘                     │
     │                     (callback)                            │
     │                                                            │
     │                                                            ▼
     │                                                   ┌──────────────┐
     │                                                   │    PayPal    │
     │                                                   │   Sandbox    │
     │                                                   └──────┬───────┘
     │                                                          │
     └──────────────────────────────────────────────────────────┘
                          (user redirect after approval)
```

### Detailed Flow

#### 1. INITIATED → REDIRECTED

1. **Webshop** sends payment request to **PSP**:
   ```json
   POST /payment
   {
     "merchantId": 1,
     "merchantPassword": "hashed_password",
     "merchantOrderId": "uuid",
     "amount": 100.00,
     "currency": "USD",
     "paymentMethod": "PAYPAL"
   }
   ```

2. **PSP** creates transaction (status: `InProgress`)

3. **PSP** forwards to **PayPal Microservice**:
   ```json
   POST http://paypal_service:8080/payment
   {
     "transactionId": "uuid",
     "merchantOrderId": "uuid",
     "merchantId": 1,
     "amount": 100.00,
     "currency": "USD",
     "description": "Order uuid"
   }
   ```

4. **PayPal Microservice** creates PayPal order via REST API

5. **PayPal Microservice** returns approval URL to **PSP**:
   ```json
   {
     "paymentId": "uuid",
     "paypalOrderId": "PAYPAL-ORDER-ID",
     "approvalUrl": "https://www.sandbox.paypal.com/checkoutnow?token=...",
     "status": "pending"
   }
   ```

6. **PSP** returns payment URL to **Webshop**:
   ```json
   {
     "paymentURL": "https://www.sandbox.paypal.com/checkoutnow?token=...",
     "tokenId": "uuid",
     "token": "payment-id",
     "tokenExp": "2026-02-09T20:30:00Z"
   }
   ```

7. **Webshop** redirects user to PayPal

#### 2. REDIRECTED → APPROVED

8. User logs into PayPal sandbox account

9. User approves payment

10. PayPal redirects to: `http://localhost:8088/payment-success?token=ORDER-ID&PayerID=PAYER-ID`

#### 3. APPROVED → SUCCESS/FAILED

11. **PayPal Microservice** receives success callback

12. **PayPal Microservice** captures payment via PayPal API

13. **PayPal Microservice** updates local database (status: `Completed`)

14. **PayPal Microservice** sends callback to **PSP**:
    ```json
    PUT http://psp_service:8080/payment-callback
    {
      "acquirerOrderId": "payment-uuid",
      "acquirerTimestamp": "2026-02-09T20:35:00Z",
      "merchantOrderId": "uuid",
      "transactionId": "uuid",
      "status": 0  // Successful
    }
    ```

15. **PSP** updates transaction status (`Successful`)

16. **PSP** forwards to **Webshop** redirect URL

17. **Webshop** displays success page to user

### Cancellation Flow

If user cancels:
- PayPal redirects to: `http://localhost:8088/payment-cancel?token=ORDER-ID`
- PayPal Microservice updates status to `Cancelled`
- Sends callback to PSP with status: `Failed`
- User redirected to failure page

## Configuration

### Environment Variables

#### Core `.env`
```env
# PayPal Service
PAYPAL_PORT=8080
PAYPAL_DB_HOST=psql_paypal
PAYPAL_DB_PORT=5439
PAYPAL_DB_DATABASE=paypal_db
PAYPAL_DB_USERNAME=paypal_user
PAYPAL_DB_PASSWORD=paypal_password
PAYPAL_DB_SCHEMA=public
PAYPAL_APP_ENV=local

# PayPal API Credentials (Get from https://developer.paypal.com/dashboard/)
PAYPAL_CLIENT_ID=your-sandbox-client-id
PAYPAL_SECRET=your-sandbox-secret
PAYPAL_MODE=sandbox

# Service URLs
PAYPAL_SERVICE_URL=http://paypal_service:8080
PSP_CALLBACK_URL=http://psp_service:8080/payment-callback
```

### PayPal Sandbox Setup

1. Go to https://developer.paypal.com/dashboard/
2. Create a sandbox account
3. Create REST API app
4. Copy Client ID and Secret
5. Update `.env` with credentials

### Test Accounts

PayPal provides test accounts for sandbox:
- **Buyer Account**: Use PayPal-generated test account
- **Seller Account**: Your sandbox business account

Test card numbers: https://developer.paypal.com/tools/sandbox/card-testing/

## Database Schema

### PayPal Microservice

```sql
CREATE TABLE paypal_payments (
    id SERIAL PRIMARY KEY,
    payment_id UUID UNIQUE NOT NULL,
    transaction_id UUID NOT NULL,
    merchant_order_id UUID NOT NULL,
    merchant_id INTEGER NOT NULL,
    
    paypal_order_id VARCHAR(255) UNIQUE NOT NULL,
    paypal_capture_id VARCHAR(255),
    
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status INTEGER NOT NULL DEFAULT 0,
    
    payer_email VARCHAR(255),
    payer_id VARCHAR(255),
    payer_name VARCHAR(255),
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMP,
    completed_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    
    approval_url TEXT,
    description TEXT,
    failure_reason TEXT
);
```

### PSP Service

Transaction table already supports PayPal:
- `payment_method` field: Set to `"PAYPAL"`
- Existing status tracking: `InProgress` → `Successful` / `Failed`

## API Endpoints

### PSP Service

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/payment` | Initiate payment (all methods) |
| PUT | `/payment-callback` | Receive payment status callbacks |

### PayPal Microservice

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/payment` | Create PayPal order |
| GET | `/payment-status/:paymentId` | Get payment status |
| GET | `/payment-success` | PayPal success callback |
| GET | `/payment-cancel` | PayPal cancel callback |
| GET | `/health` | Health check |

## Status Mapping

### PayPal → PSP Status

| PayPal Status | PSP Status | Description |
|---------------|------------|-------------|
| `pending` | `InProgress` | Order created, awaiting approval |
| `approved` | `InProgress` | User approved, awaiting capture |
| `completed` | `Successful` | Payment captured successfully |
| `cancelled` | `Failed` | User cancelled payment |
| `failed` | `Failed` | Payment processing failed |

## Error Handling

### Common Errors

1. **Invalid Credentials**
   - Error: PayPal API authentication fails
   - Solution: Check `PAYPAL_CLIENT_ID` and `PAYPAL_SECRET`

2. **Capture Failed**
   - Error: Order approved but capture fails
   - Solution: Check order is in approved state, not already captured

3. **Callback Timeout**
   - Error: PSP doesn't receive callback
   - Solution: Check `PSP_CALLBACK_URL` is accessible from PayPal service

4. **Database Connection**
   - Error: Cannot store payment
   - Solution: Verify PostgreSQL container is running

## Testing

### Manual Testing

1. **Start Services**:
   ```bash
   cd core
   docker-compose up paypal_service psql_paypal psp_service
   ```

2. **Create Payment Request**:
   ```bash
   curl -X POST http://localhost/payment \
     -H "Content-Type: application/json" \
     -d '{
       "merchantId": 1,
       "merchantPassword": "password",
       "merchantOrderId": "550e8400-e29b-41d4-a716-446655440000",
       "amount": 50.00,
       "currency": "USD",
       "paymentMethod": "PAYPAL",
       "merchantTimestamp": "2026-02-09T20:00:00Z",
       "paymentDeadline": "2026-02-09T21:00:00Z"
     }'
   ```

3. **Response**:
   ```json
   {
     "paymentURL": "https://www.sandbox.paypal.com/checkoutnow?token=ABC123",
     "tokenId": "uuid",
     "token": "payment-uuid",
     "tokenExp": "2026-02-09T20:30:00Z"
   }
   ```

4. **Open Payment URL** in browser

5. **Login with sandbox buyer account**

6. **Approve payment**

7. **Verify Success** - Check webshop redirect

### Check Payment Status

```bash
curl http://localhost:8088/payment-status/{payment-id}
```

## Constraints & Business Rules

1. **Sandbox Only**: No real money transactions
2. **Currency Support**: USD, EUR, GBP (check PayPal docs for full list)
3. **Amount Limits**: Sandbox has no limits, production has per-transaction limits
4. **Session Timeout**: 30 minutes for payment approval
5. **Idempotency**: Duplicate order creation prevented by PayPal
6. **Callback Reliability**: Asynchronous, may retry on failure

## Integration with Existing Flows

### Card Payments (Unchanged)
- `paymentMethod: "CREDIT_CARD"` → Bank Gateway flow
- No impact on existing card processing

### Crypto Payments (Unchanged)
- `paymentMethod: "CRYPTO"` → Crypto microservice flow
- No impact on existing crypto processing

### QR Payments (Unchanged)
- `paymentMethod: "QR"` → QR code flow
- No impact on existing QR processing

## Monitoring & Logging

### PayPal Microservice Logs

```bash
docker logs paypal_service -f
```

Key log entries:
- `Creating PayPal payment` - Order creation
- `PayPal payment created` - Order created successfully
- `PayPal success callback` - User approved
- `PayPal payment captured` - Payment completed
- `PSP callback successful` - Notification sent

### PSP Service Logs

```bash
docker logs psp_service -f
```

Key log entries:
- Payment method routing
- Callback receipt
- Transaction status updates

## Security Considerations

1. **Credentials**: Store PayPal credentials securely, never commit to git
2. **HTTPS**: Production must use HTTPS for all callbacks
3. **Validation**: Verify callback signatures (implement in production)
4. **Rate Limiting**: Implement request limits on PSP endpoints
5. **Database**: Use prepared statements (already implemented with GORM)

## Production Deployment

### Checklist

- [ ] Replace sandbox credentials with production credentials
- [ ] Set `PAYPAL_MODE=live`
- [ ] Update callback URLs to production domains (HTTPS)
- [ ] Configure proper error handling and retry logic
- [ ] Set up monitoring and alerting
- [ ] Test with small amounts first
- [ ] Verify webhook signatures
- [ ] Implement proper logging and audit trails

## Troubleshooting

### Payment Not Creating

1. Check PayPal service logs for API errors
2. Verify credentials are correct
3. Test PayPal API directly: https://developer.paypal.com/api/rest/

### Callback Not Received

1. Verify PSP service is accessible from PayPal service
2. Check network configuration (docker networks)
3. Verify `PSP_CALLBACK_URL` is correct
4. Check PSP logs for incoming requests

### Database Errors

1. Check PostgreSQL container is running: `docker ps | grep psql_paypal`
2. Verify schema is loaded: `docker exec psql_paypal psql -U paypal_user -d paypal_db -c "\dt"`
3. Check connection parameters match `.env`

## Support & Resources

- **PayPal Developer Docs**: https://developer.paypal.com/docs/api/overview/
- **Sandbox Testing**: https://developer.paypal.com/tools/sandbox/
- **API Reference**: https://developer.paypal.com/api/rest/
- **Status Codes**: https://developer.paypal.com/api/rest/reference/orders/v2/errors/

## Changelog

- **2026-02-09**: Initial PayPal integration implementation
  - PayPal microservice created
  - PSP extended to support PayPal routing
  - Redirect-based checkout flow implemented
  - Callback mechanism established
  - Docker configuration added
