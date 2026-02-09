# PayPal Payment Integration - Quick Start Guide

## Setup (5 minutes)

### 1. Get PayPal Sandbox Credentials

1. Visit https://developer.paypal.com/dashboard/
2. Log in or create developer account
3. Go to **Apps & Credentials** → **Sandbox**
4. Click **Create App**
5. Copy **Client ID** and **Secret**

### 2. Configure Environment

Edit `core/.env`:

```env
# Replace with your credentials
PAYPAL_CLIENT_ID=YOUR_CLIENT_ID_HERE
PAYPAL_SECRET=YOUR_SECRET_HERE
PAYPAL_MODE=sandbox
```

### 3. Start Services

```bash
cd core
docker-compose up paypal_service psql_paypal psp_service nginx
```

Wait for: `✓ PayPal microservice builds successfully`

## Test Payment (2 minutes)

### 1. Create Payment

```bash
curl -X POST http://localhost/payment \
  -H "Content-Type: application/json" \
  -d '{
    "merchantId": 1,
    "merchantPassword": "test",
    "merchantOrderId": "550e8400-e29b-41d4-a716-446655440000",
    "merchantTimestamp": "2026-02-09T20:00:00Z",
    "paymentDeadline": "2026-02-09T21:00:00Z",
    "amount": 25.00,
    "currency": "USD",
    "paymentMethod": "PAYPAL"
  }'
```

### 2. Response

```json
{
  "paymentURL": "https://www.sandbox.paypal.com/checkoutnow?token=...",
  "tokenId": "uuid",
  "token": "payment-id"
}
```

### 3. Complete Payment

1. Copy `paymentURL` from response
2. Open in browser
3. Log in with PayPal sandbox buyer account
4. Click **Pay Now**
5. You'll be redirected to success page

### 4. Verify

```bash
# Check payment status
curl http://localhost:8088/payment-status/{payment-id}

# Check logs
docker logs paypal_service --tail 50
```

## Payment Methods Comparison

| Method | Endpoint | User Experience |
|--------|----------|----------------|
| **PAYPAL** | PSP → PayPal Service → PayPal | User redirected to PayPal, logs in, approves |
| **CREDIT_CARD** | PSP → Bank Gateway → Bank | User enters card on site, bank processes |
| **CRYPTO** | PSP → Crypto Service → Blockchain | User sends crypto to address |
| **QR** | PSP → Bank (QR) | User scans QR code |

## Supported Currencies

PayPal sandbox supports: USD, EUR, GBP, CAD, AUD, JPY, and more.

Check full list: https://developer.paypal.com/docs/reports/reference/paypal-supported-currencies/

## Payment Lifecycle

```
INITIATED (Webshop creates payment)
    ↓
REDIRECTED (User sent to PayPal)
    ↓
APPROVED (User approves on PayPal)
    ↓
SUCCESS (Payment captured, merchant notified)
```

Or:

```
INITIATED → REDIRECTED → CANCELLED (User clicks cancel)
```

## Common Use Cases

### 1. Rent-a-Car Booking

```javascript
// Frontend (Next.js)
const response = await fetch('/api/payment', {
  method: 'POST',
  body: JSON.stringify({
    merchantId: 1,
    merchantPassword: 'hashed',
    merchantOrderId: bookingId,
    amount: totalPrice,
    currency: 'USD',
    paymentMethod: 'PAYPAL',
    merchantTimestamp: new Date(),
    paymentDeadline: new Date(Date.now() + 30*60*1000)
  })
});

const { paymentURL } = await response.json();
window.location.href = paymentURL; // Redirect to PayPal
```

### 2. Subscription Payment

Same flow, but amount can be recurring subscription fee.

### 3. Refund (Future Enhancement)

PayPal API supports refunds - can be added to PayPal microservice.

## Troubleshooting

### "Invalid credentials"

- Check `PAYPAL_CLIENT_ID` and `PAYPAL_SECRET` in `.env`
- Ensure you're using **Sandbox** credentials, not Live
- Regenerate credentials in PayPal Dashboard if needed

### "Payment not captured"

- Check PayPal service logs: `docker logs paypal_service`
- Verify order is in `approved` state
- Ensure user clicked "Pay Now" on PayPal

### "Callback not received"

- Verify PSP service is running: `docker ps | grep psp`
- Check network connectivity between services
- Review PSP logs: `docker logs psp_service`

### "Database error"

- Check PostgreSQL container: `docker ps | grep psql_paypal`
- Verify schema loaded: `docker exec psql_paypal psql -U paypal_user -d paypal_db -c "\dt"`
- Restart services: `docker-compose restart paypal_service psql_paypal`

## Security Notes

⚠️ **IMPORTANT FOR PRODUCTION:**

1. **Never commit credentials** - Use environment variables or secrets manager
2. **Use HTTPS** - All callback URLs must use HTTPS in production
3. **Validate webhooks** - Implement signature verification
4. **Switch to Live mode** - Set `PAYPAL_MODE=live` with production credentials
5. **Test thoroughly** - Use small amounts first in production

## API Endpoints Reference

### PSP Service (http://localhost or nginx)

- `POST /payment` - Initiate payment (all methods)
- `PUT /payment-callback` - Payment status callback

### PayPal Service (http://localhost:8088)

- `POST /payment` - Create PayPal order
- `GET /payment-status/:id` - Check payment status
- `GET /payment-success` - Success callback (internal)
- `GET /payment-cancel` - Cancel callback (internal)
- `GET /health` - Health check

## Database Queries

### Check Recent Payments

```sql
-- Connect to PayPal database
docker exec -it psql_paypal psql -U paypal_user -d paypal_db

-- View recent payments
SELECT 
  payment_id, 
  paypal_order_id, 
  amount, 
  currency, 
  status,
  created_at 
FROM paypal_payments 
ORDER BY created_at DESC 
LIMIT 10;
```

### Payment Status Values

| Status Code | Status Name | Description |
|-------------|-------------|-------------|
| 0 | `pending` | Order created, awaiting approval |
| 1 | `approved` | User approved, awaiting capture |
| 2 | `completed` | Payment successfully captured |
| 3 | `cancelled` | User cancelled payment |
| 4 | `failed` | Payment processing failed |

## Next Steps

1. **Test different currencies**: Try EUR, GBP instead of USD
2. **Test failure scenarios**: Cancel payment, use expired card
3. **Integrate with frontend**: Build PayPal payment UI
4. **Add refund support**: Extend PayPal service with refund API
5. **Monitor in production**: Set up logging and alerts

## Resources

- **Full Documentation**: See `PAYPAL_INTEGRATION.md`
- **PayPal Sandbox**: https://www.sandbox.paypal.com
- **Developer Dashboard**: https://developer.paypal.com/dashboard/
- **API Docs**: https://developer.paypal.com/docs/api/overview/

## Support

For issues with:
- **PayPal API**: Check PayPal developer docs
- **Integration**: Review logs and this guide
- **Docker**: Verify docker-compose.yml configuration
