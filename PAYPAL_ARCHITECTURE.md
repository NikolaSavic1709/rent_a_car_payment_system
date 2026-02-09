# PayPal Payment System Architecture

## System Overview

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         RENT-A-CAR PAYMENT SYSTEM                        │
└──────────────────────────────────────────────────────────────────────────┘

┌─────────────┐
│   Webshop   │ (Port 8080)
│  (Frontend) │
└──────┬──────┘
       │ 1. POST /payment
       │    paymentMethod: "PAYPAL"
       ▼
┌─────────────┐
│    Nginx    │ (Port 80/443)
│   Reverse   │
│    Proxy    │
└──────┬──────┘
       │ 2. Route to PSP
       ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         PSP SERVICE (Port 8080)                     │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  PaymentHandler                                               │  │
│  │  - Validates merchant                                         │  │
│  │  - Creates transaction (status: InProgress)                   │  │
│  │  - Routes by paymentMethod                                    │  │
│  └────────────────────────┬─────────────────────────────────────┘  │
│                           │                                         │
│    ┌──────────────────────┼──────────────────────────────┐         │
│    │                      │                              │         │
│    ▼                      ▼                              ▼         │
│  ┌─────────┐      ┌──────────────┐              ┌──────────┐      │
│  │  Card   │      │    PayPal    │              │  Crypto  │      │
│  │ Handler │      │   Handler    │ ◄── Focus    │ Handler  │      │
│  └─────────┘      └──────┬───────┘              └──────────┘      │
│                           │ 3. Forward request                     │
│                           │    to PayPal Service                   │
└───────────────────────────┼────────────────────────────────────────┘
                            │
                            ▼
    ┌────────────────────────────────────────────────────────────┐
    │         PAYPAL MICROSERVICE (Port 8088)                    │
    │  ┌──────────────────────────────────────────────────────┐  │
    │  │  CreatePaymentHandler                                 │  │
    │  │  1. Receives request from PSP                         │  │
    │  │  2. Creates PayPal Order via REST API                 │  │
    │  │  3. Stores in PostgreSQL (status: pending)            │  │
    │  │  4. Returns approval URL to PSP                       │  │
    │  └──────────────────────────────────────────────────────┘  │
    │           │                                   ▲             │
    │           │ 4. Return                         │ 6. Callback│
    │           │    approvalURL                    │    from     │
    │           ▼                                   │    PayPal   │
    │  ┌──────────────────────────────────────────────────────┐  │
    │  │  PaymentSuccessHandler / PaymentCancelHandler        │  │
    │  │  1. Receives callback from PayPal                    │  │
    │  │  2. Captures payment (if success)                    │  │
    │  │  3. Updates database                                 │  │
    │  │  4. Sends callback to PSP                            │  │
    │  │  5. Redirects user to frontend                       │  │
    │  └──────────────────────────────────────────────────────┘  │
    │           │                                                 │
    │           │ Database                                        │
    │           ▼                                                 │
    │  ┌──────────────────────┐                                  │
    │  │   PostgreSQL         │                                  │
    │  │   paypal_payments    │                                  │
    │  │   (Port 5439)        │                                  │
    │  └──────────────────────┘                                  │
    └────────────┬──────────────────────────────────────────────┘
                 │
                 │ 5. HTTPS API Call
                 ▼
    ┌────────────────────────────────────────────────────────┐
    │           PAYPAL SANDBOX API                           │
    │  https://api-m.sandbox.paypal.com/v2/checkout/orders   │
    │                                                         │
    │  Operations:                                            │
    │  - Create Order                                         │
    │  - Capture Payment                                      │
    │  - Get Order Details                                    │
    └────────────┬───────────────────────────────────────────┘
                 │
                 │ 6. User Redirect
                 ▼
    ┌────────────────────────────────────────────────────────┐
    │         USER BROWSER                                   │
    │  https://www.sandbox.paypal.com/checkoutnow?token=...  │
    │                                                         │
    │  User Actions:                                          │
    │  1. Login to PayPal                                     │
    │  2. Review payment details                              │
    │  3. Click "Pay Now" OR "Cancel"                         │
    └────────────────────────────────────────────────────────┘
                 │
                 │ 7. Redirect back
                 └─────────┐
                           │
                           ▼
              Success: /payment-success?token=XXX
              Cancel:  /payment-cancel?token=XXX

```

## Sequence Diagram

```
User         Webshop      PSP        PayPal        PayPal        PayPal
                        Service    Service       Sandbox        DB
 │              │          │           │             │            │
 │ 1. Book Car  │          │           │             │            │
 │─────────────>│          │           │             │            │
 │              │          │           │             │            │
 │              │ 2. POST  │           │             │            │
 │              │ /payment │           │             │            │
 │              │ PAYPAL   │           │             │            │
 │              │─────────>│           │             │            │
 │              │          │           │             │            │
 │              │          │ 3. Create │             │            │
 │              │          │ Transaction            │            │
 │              │          │ (InProgress)           │            │
 │              │          │           │             │            │
 │              │          │ 4. POST   │             │            │
 │              │          │ /payment  │             │            │
 │              │          │──────────>│             │            │
 │              │          │           │             │            │
 │              │          │           │ 5. POST     │            │
 │              │          │           │ /v2/checkout│            │
 │              │          │           │ /orders     │            │
 │              │          │           │────────────>│            │
 │              │          │           │             │            │
 │              │          │           │<────────────│            │
 │              │          │           │ Order +     │            │
 │              │          │           │ approvalURL │            │
 │              │          │           │             │            │
 │              │          │           │ 6. INSERT   │            │
 │              │          │           │ payment     │            │
 │              │          │           │────────────────────────> │
 │              │          │           │             │            │
 │              │          │<──────────│             │            │
 │              │          │ approvalURL            │            │
 │              │          │           │             │            │
 │              │<─────────│           │             │            │
 │              │ paymentURL           │             │            │
 │              │           │           │             │            │
 │<─────────────│           │           │             │            │
 │ Redirect to PayPal      │           │             │            │
 │─────────────────────────────────────────────────>│            │
 │              │           │           │             │            │
 │              │ 7. Login & Approve    │             │            │
 │<─────────────────────────────────────────────────>│            │
 │              │           │           │             │            │
 │              │           │           │<────────────│            │
 │              │           │           │ Redirect:   │            │
 │              │           │           │ /payment-   │            │
 │──────────────────────────────────────>success     │            │
 │              │           │           │             │            │
 │              │           │           │ 8. GET      │            │
 │              │           │           │ order       │            │
 │              │           │           │────────────>│            │
 │              │           │           │<────────────│            │
 │              │           │           │             │            │
 │              │           │           │ 9. POST     │            │
 │              │           │           │ capture     │            │
 │              │           │           │────────────>│            │
 │              │           │           │<────────────│            │
 │              │           │           │             │            │
 │              │           │           │ 10. UPDATE  │            │
 │              │           │           │ status:     │            │
 │              │           │           │ completed   │            │
 │              │           │           │────────────────────────> │
 │              │           │           │             │            │
 │              │           │<──────────│             │            │
 │              │           │ 11. PUT   │             │            │
 │              │           │ /callback │             │            │
 │              │           │ Successful│             │            │
 │              │<──────────│           │             │            │
 │              │ 12. UPDATE│           │             │            │
 │              │ transaction           │             │            │
 │              │ (Successful)          │             │            │
 │              │           │           │             │            │
 │<─────────────│           │           │             │            │
 │ 13. Redirect │           │           │             │            │
 │ Success Page │           │           │             │            │
 │              │           │           │             │            │
```

## Payment Status Flow

```
┌─────────────┐
│  INITIATED  │ ← Webshop creates payment request
└──────┬──────┘
       │
       │ PSP creates transaction
       ▼
┌─────────────┐
│ REDIRECTED  │ ← User sent to PayPal approval URL
│ (InProgress)│
└──────┬──────┘
       │
       │ User approves on PayPal
       ▼
┌─────────────┐
│  APPROVED   │ ← PayPal redirects to success callback
│ (InProgress)│
└──────┬──────┘
       │
       │ PayPal Service captures payment
       ▼
┌─────────────┐
│   SUCCESS   │ ← Callback sent to PSP
│ (Successful)│   PSP notifies webshop
└─────────────┘


Cancellation Path:
┌─────────────┐
│ REDIRECTED  │
└──────┬──────┘
       │
       │ User clicks Cancel on PayPal
       ▼
┌─────────────┐
│ CANCELLED   │ ← Callback sent with Failed status
│  (Failed)   │
└─────────────┘
```

## Data Flow

### 1. Payment Initiation
```
Webshop → PSP
{
  merchantId: 1,
  merchantOrderId: "uuid",
  amount: 100.00,
  currency: "USD",
  paymentMethod: "PAYPAL"
}

PSP → PayPal Service
{
  transactionId: "uuid",
  merchantOrderId: "uuid",
  merchantId: 1,
  amount: 100.00,
  currency: "USD"
}

PayPal Service → PayPal API
{
  intent: "CAPTURE",
  purchase_units: [{
    amount: {
      currency_code: "USD",
      value: "100.00"
    }
  }]
}
```

### 2. Payment Approval
```
PayPal API → PayPal Service
{
  id: "PAYPAL-ORDER-ID",
  status: "CREATED",
  links: [{
    rel: "approve",
    href: "https://sandbox.paypal.com/checkoutnow?token=XXX"
  }]
}

PayPal Service → PSP
{
  paymentId: "uuid",
  paypalOrderId: "PAYPAL-ORDER-ID",
  approvalUrl: "https://...",
  status: "pending"
}
```

### 3. Payment Completion
```
PayPal Service → PayPal API (Capture)
POST /v2/checkout/orders/{order-id}/capture

PayPal API → PayPal Service
{
  id: "PAYPAL-ORDER-ID",
  status: "COMPLETED",
  purchase_units: [{
    payments: {
      captures: [{
        id: "CAPTURE-ID",
        status: "COMPLETED"
      }]
    }
  }],
  payer: {
    email_address: "buyer@example.com",
    payer_id: "PAYER-ID"
  }
}

PayPal Service → PSP (Callback)
{
  acquirerOrderId: "payment-uuid",
  acquirerTimestamp: "2026-02-09T20:35:00Z",
  merchantOrderId: "uuid",
  transactionId: "uuid",
  status: 0  // Successful
}
```

## Network Architecture

```
Docker Networks:
┌─────────────────────────────────────────────────────────┐
│                    shared_network                       │
│  ┌────────┐  ┌──────┐  ┌────────┐  ┌────────┐          │
│  │ Nginx  │  │ PSP  │  │ PayPal │  │ Crypto │          │
│  │        │  │      │  │        │  │        │          │
│  └────────┘  └──────┘  └────────┘  └────────┘          │
└─────────────────────────────────────────────────────────┘

┌─────────────────────┐
│  paypal_network     │
│  ┌────────┐         │
│  │ PayPal │         │
│  │ Service│         │
│  └────┬───┘         │
│       │             │
│  ┌────▼──────────┐  │
│  │ PostgreSQL    │  │
│  │ (psql_paypal) │  │
│  └───────────────┘  │
└─────────────────────┘
```

## Port Mapping

| Service | Internal Port | External Port |
|---------|---------------|---------------|
| Nginx | 80 | 80 |
| Webshop | 8080 | 8081 |
| PSP | 8080 | 8082 |
| PayPal Service | 8080 | **8088** |
| Crypto Service | 8080 | 8087 |
| PostgreSQL (PayPal) | 5432 | **5439** |

## Security Layers

```
┌─────────────────────────────────────────────────────────┐
│                    Internet/User                        │
└───────────────────────┬─────────────────────────────────┘
                        │
                        │ HTTPS (Production)
                        ▼
                  ┌──────────┐
                  │  Nginx   │ ← SSL Termination
                  │  Proxy   │   Rate Limiting
                  └────┬─────┘   IP Filtering
                       │
                       │ Internal HTTP
                       ▼
              ┌─────────────────┐
              │   PSP Service   │ ← Merchant Authentication
              │                 │   Request Validation
              └────┬────────┬───┘   Transaction Tracking
                   │        │
        ┌──────────┘        └──────────┐
        │                               │
        ▼                               ▼
┌───────────────┐            ┌──────────────────┐
│ PayPal Service│            │ Other Services   │
│               │            │ (Card, Crypto)   │
└───────┬───────┘            └──────────────────┘
        │
        │ OAuth 2.0
        ▼
┌──────────────┐
│ PayPal API   │ ← PayPal's Security
│   Sandbox    │   (Client ID + Secret)
└──────────────┘
```

## Deployment Architecture

```
Production Environment:

┌─────────────────────────────────────────────────────────┐
│                    Load Balancer                        │
└───────────────────────┬─────────────────────────────────┘
                        │
         ┌──────────────┼──────────────┐
         │              │              │
         ▼              ▼              ▼
    ┌────────┐    ┌────────┐    ┌────────┐
    │Instance│    │Instance│    │Instance│
    │   1    │    │   2    │    │   3    │
    └───┬────┘    └───┬────┘    └───┬────┘
        │             │              │
        └──────┬──────┴──────┬───────┘
               │             │
        ┌──────▼──────┐ ┌───▼────────┐
        │  PostgreSQL │ │   Redis    │
        │  (Primary)  │ │   Cache    │
        └──────┬──────┘ └────────────┘
               │
        ┌──────▼──────┐
        │ PostgreSQL  │
        │  (Replica)  │
        └─────────────┘
```

This architecture ensures:
- High availability
- Horizontal scalability
- Data redundancy
- Fast response times
- Secure communication
