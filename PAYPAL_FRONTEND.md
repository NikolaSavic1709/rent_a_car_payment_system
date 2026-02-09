# PayPal Frontend Integration Guide

## Quick Start

### 1. Install Dependencies

```bash
cd front/paypal-front
npm install
```

### 2. Start Frontend

```bash
npm run dev
```

Frontend runs on: **http://localhost:3003**

### 3. Test Payment Flow

**Option A: Direct URL Access**
```
http://localhost:3003?merchantOrderId=550e8400-e29b-41d4-a716-446655440000&amount=50&currency=USD
```

**Option B: From Webshop/PSP**

After PSP creates payment, redirect user:
```javascript
const paymentURL = `http://localhost:3003?merchantOrderId=${merchantOrderId}&amount=${amount}&currency=${currency}`;
window.location.href = paymentURL;
```

## Complete Flow Diagram

```
┌────────────────────────────────────────────────────────────────────┐
│                     COMPLETE PAYPAL FLOW                           │
└────────────────────────────────────────────────────────────────────┘

1. USER SELECTS PAYPAL PAYMENT
   ┌──────────────┐
   │   Webshop    │ (localhost:3001)
   │   Frontend   │
   └──────┬───────┘
          │ POST /payment
          │ paymentMethod: "PAYPAL"
          ▼
   ┌──────────────┐
   │     PSP      │ (localhost/nginx)
   │   Backend    │
   └──────┬───────┘
          │ Creates transaction
          │ Forwards to PayPal service
          ▼
   ┌──────────────┐
   │    PayPal    │ (localhost:8088)
   │  Microservice│
   └──────┬───────┘
          │ Creates PayPal order
          │ Returns approval URL
          ▼
   ┌──────────────┐
   │     PSP      │
   │   Returns    │
   │  paymentURL  │
   └──────┬───────┘
          │
          ▼
2. REDIRECT TO PAYPAL FRONTEND
   ┌──────────────┐
   │   Webshop    │
   │   Redirects  │
   └──────┬───────┘
          │ window.location.href = 
          │ "http://localhost:3003?merchantOrderId=..."
          ▼
   ┌──────────────────────────────────┐
   │      PayPal Frontend             │ (localhost:3003)
   │                                  │
   │  ┌────────────────────────────┐ │
   │  │  index.js                   │ │
   │  │  - Receives query params    │ │
   │  │  - Calls PayPal service     │ │
   │  │  - Gets approval URL        │ │
   │  │  - Shows payment details    │ │
   │  └────────────┬───────────────┘ │
   └───────────────┼──────────────────┘
                   │ Auto-redirect after 2s
                   ▼
3. USER APPROVES ON PAYPAL
   ┌──────────────────────────────────┐
   │      PayPal Sandbox              │
   │   (sandbox.paypal.com)           │
   │                                  │
   │  User:                           │
   │  1. Logs in                      │
   │  2. Reviews payment              │
   │  3. Clicks "Pay Now"             │
   │                                  │
   └──────────────┬───────────────────┘
                  │ Redirects to:
                  │ http://localhost:8088/payment-success
                  │ ?token=ORDER-ID&PayerID=PAYER-ID
                  ▼
4. PAYMENT CAPTURED
   ┌──────────────┐
   │    PayPal    │
   │ Microservice │
   │              │
   │ - Captures   │
   │   payment    │
   │ - Updates DB │
   │ - Sends      │
   │   callback   │
   └──────┬───────┘
          │ PUT /payment-callback
          ▼
   ┌──────────────┐
   │     PSP      │
   │  - Updates   │
   │    status    │
   │  - Notifies  │
   │    webshop   │
   └──────┬───────┘
          │ Redirects user to:
          │ http://localhost:3003/success
          │ ?paymentId=...&status=completed
          ▼
5. SUCCESS PAGE DISPLAYED
   ┌──────────────────────────────────┐
   │      PayPal Frontend             │
   │                                  │
   │  ┌────────────────────────────┐ │
   │  │  success.js                 │ │
   │  │  - Shows success message    │ │
   │  │  - Displays payment details │ │
   │  │  - "Return to Shop" button  │ │
   │  └────────────────────────────┘ │
   └──────────────────────────────────┘
```

## Frontend Pages

### 1. Index Page (`/`)

**Purpose**: Payment initiation and PayPal redirect

**URL**: `http://localhost:3003?merchantOrderId=XXX&amount=YYY&currency=ZZZ`

**Flow**:
1. Receives payment parameters from query string
2. Calls PayPal service to create order
3. Displays payment details
4. Auto-redirects to PayPal approval URL after 2 seconds
5. Manual redirect button available

**Query Parameters**:
| Parameter | Required | Example | Description |
|-----------|----------|---------|-------------|
| merchantOrderId | Yes | `550e8400-e29b-41d4-a716-446655440000` | Order ID from webshop |
| amount | No | `50.00` | Payment amount (default: 100.00) |
| currency | No | `USD` | Currency code (default: USD) |

**API Calls**:
```javascript
POST http://localhost:8088/payment
{
  "merchantOrderId": "uuid",
  "amount": 50.00,
  "currency": "USD",
  "description": "Order uuid"
}

Response:
{
  "paymentId": "uuid",
  "paypalOrderId": "PAYPAL-ORDER-ID",
  "approvalUrl": "https://sandbox.paypal.com/checkoutnow?token=...",
  "status": "pending"
}
```

### 2. Success Page (`/success`)

**Purpose**: Display payment success confirmation

**URL**: `http://localhost:3003/success?token=XXX&PayerID=YYY&paymentId=ZZZ`

**Flow**:
1. Receives callback parameters from PayPal
2. Waits 2 seconds for backend processing
3. Fetches payment status from PayPal service
4. Displays success message and details
5. Provides "Return to Shop" button

**Query Parameters**:
| Parameter | Required | Example | Description |
|-----------|----------|---------|-------------|
| token | Yes | `PAYPAL-ORDER-ID` | PayPal order token |
| PayerID | Yes | `PAYER-ID` | PayPal payer identifier |
| paymentId | No | `uuid` | Internal payment ID |

**API Calls**:
```javascript
GET http://localhost:8088/payment-status/{paymentId}

Response:
{
  "paymentId": "uuid",
  "transactionId": "uuid",
  "paypalOrderId": "PAYPAL-ORDER-ID",
  "status": "completed",
  "amount": 50.00,
  "currency": "USD",
  "payerEmail": "buyer@example.com",
  "createdAt": "2026-02-09T20:00:00Z",
  "completedAt": "2026-02-09T20:05:00Z"
}
```

### 3. Cancel Page (`/cancel`)

**Purpose**: Display cancellation message

**URL**: `http://localhost:3003/cancel?token=XXX&reason=YYY`

**Flow**:
1. Receives cancellation parameters
2. Displays cancellation message
3. Provides "Try Again" and "Return to Shop" buttons

**Query Parameters**:
| Parameter | Required | Example | Description |
|-----------|----------|---------|-------------|
| token | Yes | `PAYPAL-ORDER-ID` | PayPal order token |
| reason | No | `capture_failed` | Cancellation reason |

**Possible Reasons**:
- `missing_order_id` - Invalid payment session
- `payment_not_found` - Payment not found in database
- `capture_failed` - Payment capture failed
- (none) - User cancelled payment

## Integration Patterns

### Pattern 1: Direct Integration (Webshop → PayPal Frontend)

**Use Case**: Webshop handles payment entirely

```javascript
// In webshop frontend (React/Next.js)
const handlePayPalPayment = async () => {
  // 1. Create payment in PSP
  const response = await fetch('http://localhost/payment', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      merchantId: 1,
      merchantPassword: 'hashed',
      merchantOrderId: orderId,
      amount: totalAmount,
      currency: 'USD',
      paymentMethod: 'PAYPAL',
      merchantTimestamp: new Date(),
      paymentDeadline: new Date(Date.now() + 30*60*1000)
    })
  });

  const { paymentURL } = await response.json();

  // 2. Redirect to PayPal frontend
  // PSP should return paypal-front URL instead of PayPal approval URL
  window.location.href = paymentURL;
};
```

**PSP Backend Change**:
```javascript
// In PSP paypal.go
response := database.PaymentStartResponse{
  PaymentURL: fmt.Sprintf(
    "http://localhost:3003?merchantOrderId=%s&amount=%.2f&currency=%s",
    merchantOrderId, 
    amount, 
    currency
  ),
  TokenId:    transactionId,
  Token:      paymentResponse.PaymentID.String(),
  TokenExp:   time.Now().Add(30 * time.Minute),
}
```

### Pattern 2: Two-Step Integration (PSP Returns PayPal URL)

**Use Case**: PSP returns PayPal service URL, webshop redirects manually

```javascript
// Webshop gets PayPal service URL from PSP
const { paymentURL } = await createPayment();
// paymentURL = "http://localhost:8088/payment/..."

// Webshop constructs paypal-front URL
const frontendURL = `http://localhost:3003?merchantOrderId=${orderId}&amount=${amount}&currency=${currency}`;
window.location.href = frontendURL;
```

### Pattern 3: Embedded Integration (iFrame - Not Recommended)

Not recommended due to PayPal redirect requirements, but possible:

```html
<!-- In webshop -->
<iframe 
  src="http://localhost:3003?merchantOrderId=..." 
  width="600" 
  height="800"
></iframe>
```

## Environment Configuration

### Development (default)

```javascript
// values/Environment.js
BASE_URL = 'http://localhost'
PAYPAL_SERVICE_URL = 'http://localhost:8088'
PSP_SERVICE_URL = 'http://localhost'
```

### Production

Create `.env.local`:

```env
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
NEXT_PUBLIC_PAYPAL_SERVICE_URL=https://paypal-api.yourdomain.com
NEXT_PUBLIC_PSP_SERVICE_URL=https://psp-api.yourdomain.com
```

## Callback URLs Configuration

### PayPal Microservice Environment

Update `core/.env`:

```env
# Frontend redirect URLs (where PayPal service redirects user after payment)
SUCCESS_URL=http://localhost:3003/success
CANCEL_URL=http://localhost:3003/cancel

# OR in production:
# SUCCESS_URL=https://payment.yourdomain.com/success
# CANCEL_URL=https://payment.yourdomain.com/cancel
```

### PayPal Service Return URLs

Update in `paypal_microservice/internal/server/paypal_client.go`:

```go
ApplicationContext: &paypal.ApplicationContext{
    ReturnURL:  os.Getenv("PAYPAL_RETURN_URL"),  // http://localhost:8088/payment-success
    CancelURL:  os.Getenv("PAYPAL_CANCEL_URL"),  // http://localhost:8088/payment-cancel
    BrandName:  "Rent-a-Car",
    ShippingPreference: "NO_SHIPPING",
}
```

## Testing Checklist

### ✅ Success Flow
- [ ] User lands on paypal-front index page
- [ ] Payment details displayed correctly
- [ ] Auto-redirect to PayPal works
- [ ] Manual "Continue" button works
- [ ] PayPal login page loads
- [ ] Payment approval succeeds
- [ ] Redirect back to success page
- [ ] Payment details shown on success page
- [ ] "Return to Shop" works

### ✅ Cancel Flow
- [ ] User can cancel on PayPal
- [ ] Redirect to cancel page works
- [ ] Cancellation reason displayed
- [ ] "Try Again" redirects to index
- [ ] "Return to Shop" works

### ✅ Error Handling
- [ ] Missing merchantOrderId shows error
- [ ] Invalid amount handled
- [ ] API failure shows error message
- [ ] Network timeout handled
- [ ] PayPal service down shows error

## Common Issues

### 1. CORS Errors

**Problem**: API calls blocked by CORS

**Solution**: Update PayPal service CORS config:

```go
// In paypal_microservice/internal/server/routes.go
r.Use(cors.New(cors.Config{
    AllowOrigins: []string{
        "http://localhost:3000", 
        "http://localhost:3001", 
        "http://localhost:3002",
        "http://localhost:3003",  // Add paypal-front
    },
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
    AllowCredentials: true,
}))
```

### 2. Redirect Loop

**Problem**: Keeps redirecting back to index page

**Solution**: Check PayPal service return URLs match frontend routes

### 3. Payment Status Not Found

**Problem**: Success page can't fetch payment status

**Solution**: 
- Ensure PayPal service is running on port 8088
- Check payment ID is passed correctly
- Verify callback was processed

## Production Deployment

### 1. Build Frontend

```bash
cd front/paypal-front
npm run build
```

### 2. Deploy

**Vercel** (Recommended):
```bash
vercel deploy
```

**Docker**:
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build
EXPOSE 3003
CMD ["npm", "start"]
```

### 3. Update Environment

Set production URLs in deployment platform:
```
NEXT_PUBLIC_PAYPAL_SERVICE_URL=https://paypal-api.yourdomain.com
```

### 4. Configure Callbacks

Update PayPal service `.env`:
```env
SUCCESS_URL=https://payment.yourdomain.com/success
CANCEL_URL=https://payment.yourdomain.com/cancel
```

## Architecture Summary

```
Port Mapping:
┌─────────────────────────────────────────┐
│ 3001 - Webshop Frontend (rent-a-car)   │
│ 3002 - Crypto Payment Frontend         │
│ 3003 - PayPal Payment Frontend ◄─ NEW  │
│ 80   - Nginx (PSP Gateway)              │
│ 8088 - PayPal Microservice              │
│ 5439 - PostgreSQL (PayPal DB)          │
└─────────────────────────────────────────┘

Technology Stack:
┌─────────────────────────────────────────┐
│ Frontend:  Next.js 14 + React 18        │
│ Styling:   CSS Modules                  │
│ State:     React Hooks                  │
│ Routing:   Next.js Router               │
└─────────────────────────────────────────┘
```

## Next Steps

1. **Install dependencies**: `npm install`
2. **Start development**: `npm run dev`
3. **Test locally**: Visit http://localhost:3003
4. **Integrate with webshop**: Update payment redirect
5. **Configure production**: Set environment variables
6. **Deploy**: Use Vercel or Docker

## Resources

- **Frontend Code**: `front/paypal-front/`
- **Backend Integration**: `PAYPAL_INTEGRATION.md`
- **Quick Start**: `PAYPAL_QUICKSTART.md`
- **Architecture**: `PAYPAL_ARCHITECTURE.md`
- **PayPal Docs**: https://developer.paypal.com/docs/
