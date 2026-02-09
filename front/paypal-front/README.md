# PayPal Payment Frontend

A Next.js frontend application for handling PayPal payments in the Rent-a-Car payment system.

## Overview

This application provides a user interface for:
- Initiating PayPal payments
- Redirecting users to PayPal for approval
- Handling payment success/cancel callbacks
- Displaying payment results

## Features

- ✅ Clean, minimal UI matching existing frontends
- ✅ Automatic redirect to PayPal
- ✅ Success/cancel page handling
- ✅ Real-time payment status display
- ✅ Responsive design
- ✅ Error handling

## Technology Stack

- **Next.js 14** - React framework
- **React 18** - UI library
- **CSS Modules** - Styling

## Project Structure

```
paypal-front/
├── pages/
│   ├── _app.js           # App wrapper
│   ├── _document.js      # HTML document
│   ├── index.js          # Payment initiation page
│   ├── success.js        # Payment success page
│   └── cancel.js         # Payment cancel page
├── styles/
│   ├── globals.css       # Global styles
│   ├── PaymentPage.module.css
│   └── ResultPage.module.css
├── helpers/
│   └── enums.js          # Status enums and helpers
├── values/
│   └── Environment.js    # Environment configuration
└── package.json
```

## Setup

### 1. Install Dependencies

```bash
cd front/paypal-front
npm install
```

### 2. Configure Environment (Optional)

Create `.env.local` for custom URLs:

```env
NEXT_PUBLIC_API_URL=http://localhost
NEXT_PUBLIC_PAYPAL_SERVICE_URL=http://localhost:8088
NEXT_PUBLIC_PSP_SERVICE_URL=http://localhost
```

### 3. Run Development Server

```bash
npm run dev
```

The application will start on http://localhost:3003

## Usage

### Payment Flow

1. **Initiate Payment**
   - Navigate to: `http://localhost:3003?merchantOrderId=ORDER-ID&amount=100&currency=USD`
   - Application creates PayPal payment
   - User automatically redirected to PayPal

2. **PayPal Approval**
   - User logs into PayPal sandbox
   - Reviews payment details
   - Approves or cancels payment

3. **Callback Handling**
   - Success: Redirected to `/success` page
   - Cancel: Redirected to `/cancel` page
   - Payment status displayed

### API Endpoints Used

- `POST /payment` - Create PayPal payment (PayPal Service)
- `GET /payment-status/:id` - Check payment status (PayPal Service)

## Pages

### Index Page (`/`)

Main payment page that:
- Receives payment parameters via query string
- Creates PayPal order
- Displays payment details
- Redirects to PayPal approval URL

**Query Parameters:**
- `merchantOrderId` - Order identifier (required)
- `amount` - Payment amount (default: 100.00)
- `currency` - Currency code (default: USD)

### Success Page (`/success`)

Displays after successful payment:
- Shows payment confirmation
- Displays transaction details
- Provides "Return to Shop" button

**Query Parameters:**
- `token` - PayPal order token
- `PayerID` - PayPal payer identifier
- `paymentId` - Internal payment ID

### Cancel Page (`/cancel`)

Displays when payment is cancelled:
- Shows cancellation message
- Displays reason if available
- Provides retry and return options

**Query Parameters:**
- `token` - PayPal order token
- `reason` - Cancellation reason

## Styling

The application uses CSS Modules for component-scoped styling:

- **globals.css** - Base styles, body background
- **PaymentPage.module.css** - Payment initiation styles
- **ResultPage.module.css** - Success/cancel page styles

### Color Scheme

- Primary: PayPal Blue (`#0070ba`, `#003087`)
- Success: Green (`#4CAF50`)
- Error: Red (`#f44336`)
- Background: Purple Gradient

## Integration

### From Webshop

Redirect user to PayPal frontend:

```javascript
const paymentURL = `http://localhost:3003?merchantOrderId=${orderId}&amount=${amount}&currency=${currency}`;
window.location.href = paymentURL;
```

### From PSP

Return paypal-front URL in payment response:

```javascript
{
  "paymentURL": "http://localhost:3003?merchantOrderId=...",
  "tokenId": "uuid",
  "tokenExp": "2026-02-09T21:00:00Z"
}
```

## Development

### Available Scripts

```bash
npm run dev      # Start development server on port 3003
npm run build    # Build for production
npm run start    # Start production server
npm run lint     # Run ESLint
```

### Port Configuration

Default port: `3003`

To change, update `package.json` scripts:
```json
"dev": "next dev -p 3004"
```

## Production Deployment

### 1. Build Application

```bash
npm run build
```

### 2. Set Environment Variables

```env
NEXT_PUBLIC_API_URL=https://your-domain.com
NEXT_PUBLIC_PAYPAL_SERVICE_URL=https://paypal-api.your-domain.com
```

### 3. Deploy

Deploy the `.next` folder using:
- Vercel (recommended for Next.js)
- Docker
- Traditional hosting

### Docker Deployment (Optional)

Create `Dockerfile`:

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

Build and run:
```bash
docker build -t paypal-front .
docker run -p 3003:3003 paypal-front
```

## Testing

### Manual Testing

1. **Success Flow**
   ```
   http://localhost:3003?merchantOrderId=test-123&amount=50&currency=USD
   → Approve on PayPal
   → See success page
   ```

2. **Cancel Flow**
   ```
   http://localhost:3003?merchantOrderId=test-123&amount=50&currency=USD
   → Cancel on PayPal
   → See cancel page
   ```

### Test PayPal Accounts

Use PayPal sandbox accounts from:
https://developer.paypal.com/dashboard/accounts

## Troubleshooting

### CORS Errors

If API calls fail due to CORS:
- Check PayPal service CORS configuration
- Ensure `http://localhost:3003` is in allowed origins

### Redirect Not Working

If PayPal redirect fails:
- Check `approvalUrl` in API response
- Verify PayPal credentials in backend
- Check browser console for errors

### Payment Status Not Loading

If success page doesn't show status:
- Verify PayPal service is running (port 8088)
- Check network tab for API errors
- Ensure callback was processed

## Architecture

```
User Browser
    │
    ├─> paypal-front (localhost:3003)
    │       │
    │       ├─> Create Payment → PayPal Service (8088)
    │       │                         │
    │       │                         └─> PayPal API
    │       │
    │       └─> Check Status → PayPal Service (8088)
    │
    └─> PayPal Sandbox (approval)
            │
            └─> Callback → PayPal Service
                              │
                              └─> PSP Service
```

## Future Enhancements

- [ ] Add loading skeleton components
- [ ] Implement retry logic for failed requests
- [ ] Add payment history page
- [ ] Support multiple currencies dynamically
- [ ] Add print receipt functionality
- [ ] Implement QR code for payment link
- [ ] Add email notification feature

## Contributing

Follow the existing code style:
- Use functional components
- CSS Modules for styling
- Match naming conventions
- Keep components simple

## License

MIT

## Support

For issues or questions:
- Check documentation: `PAYPAL_INTEGRATION.md`
- Review backend logs: `docker logs paypal_service`
- Verify service status: `docker ps`
