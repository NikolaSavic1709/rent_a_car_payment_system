# Crypto-Front - Cryptocurrency Payment Interface

A Next.js application for processing cryptocurrency payments (BTC, ETH, USDT) in the rent-a-car payment system.

## Features

- 🔐 Secure cryptocurrency payment processing
- 📱 QR code generation for easy mobile wallet scanning
- ⏱️ Real-time payment status updates
- 📊 Confirmation progress tracking
- ⚡ Auto-redirect on payment completion
- 🎨 Responsive design with Material-UI

## Tech Stack

- **Framework:** Next.js 14
- **UI Library:** Material-UI (MUI)
- **HTTP Client:** Axios
- **QR Generation:** qrcode.react
- **Styling:** CSS Modules

## Getting Started

### Prerequisites

- Node.js 18+
- npm or yarn
- Backend services running (PSP and Crypto services)

### Installation

```bash
# Install dependencies
npm install

# Run development server on port 3002
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

The application will be available at [http://localhost:3002](http://localhost:3002)

## Environment Configuration

Configure backend endpoints in `values/Environment.js`:

```javascript
export const PSP_BASE_URL = 'http://localhost:8084'
export const CRYPTO_SERVICE_URL = 'http://localhost:8086'
```

## Usage

### Payment Flow

1. User is redirected from rentacar-front with `merchantOrderId`
2. Application fetches payment details from PSP service
3. Displays:
   - Cryptocurrency wallet address
   - QR code for mobile scanning
   - Payment amount in selected crypto
   - Required confirmations
   - Countdown timer
4. Polls payment status every 10 seconds
5. Updates UI with confirmation progress
6. Auto-redirects on completion

### URL Parameters

- `merchantOrderId` (required): Unique order identifier
- `tokenId` (optional): Payment token

Example:
```
http://localhost:3002/payment?merchantOrderId=123e4567-e89b-12d3-a456-426614174000
```

## Project Structure

```
crypto-front/
├── pages/
│   ├── _app.js           # App wrapper
│   ├── _document.js      # HTML document
│   ├── index.js          # Home/redirect page
│   └── payment.js        # Main payment interface
├── styles/
│   ├── globals.css       # Global styles
│   └── CryptoPage.module.css  # Payment page styles
├── components/           # (Future: reusable components)
├── helpers/
│   └── enums.js         # Payment status enums
├── values/
│   └── Environment.js   # API endpoints
├── public/              # Static assets
├── package.json         # Dependencies (port 3002)
└── next.config.mjs     # Next.js configuration
```

## Payment Statuses

- **Pending:** Waiting for payment
- **Confirming:** Transaction detected, accumulating confirmations
- **Confirmed:** Payment complete
- **Expired:** Payment window expired
- **Failed:** Payment processing failed

## API Integration

### Fetching Payment Details

```javascript
GET /crypto-payment-details?merchantOrderId={uuid}
```

Response:
```json
{
  "paymentId": "uuid",
  "destinationAddress": "tb1q...",
  "amount": 0.0015,
  "currency": "BTC",
  "expiryTime": "2026-02-08T...",
  "requiredConfirmations": 3,
  "status": "pending"
}
```

### Polling Payment Status

```javascript
GET /crypto-status?paymentId={uuid}
```

Response:
```json
{
  "paymentId": "uuid",
  "status": "confirming",
  "confirmations": 1,
  "txHash": "abc123...",
  "blockHeight": 2345678
}
```

## Styling

The application uses CSS Modules for component-scoped styling:

- `globals.css`: Base styles, resets, utilities
- `CryptoPage.module.css`: Payment page specific styles

Key design elements:
- Clean, modern interface
- Status-based color coding
- Responsive layout
- Material Design icons
- Progress indicators

## Development

### Hot Reload

Changes to React components automatically reload in development mode.

### Browser Support

- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)

### Debugging

Use browser DevTools:
- **Console:** Check for errors
- **Network:** Monitor API calls
- **React DevTools:** Inspect component state

## Testing

### Manual Testing

1. Start backend services
2. Start crypto-front: `npm run dev`
3. Navigate to rentacar-front
4. Complete a crypto payment
5. Verify all UI elements display correctly
6. Test status polling
7. Simulate confirmations using backend API
8. Verify auto-redirect

### Test Payment Simulation

Use crypto service test endpoints:

```bash
# Simulate payment detection
curl -X POST http://localhost:8086/simulate-payment \
  -H "Content-Type: application/json" \
  -d '{"paymentId": "your-payment-id"}'

# Add confirmation
curl -X POST http://localhost:8086/simulate-confirmation \
  -H "Content-Type: application/json" \
  -d '{"paymentId": "your-payment-id"}'
```

## Troubleshooting

### Port Already in Use

If port 3002 is occupied:
```bash
# Change port in package.json
"scripts": {
  "dev": "next dev -p 3003"
}
```

### API Connection Issues

1. Verify backend services are running
2. Check CORS configuration in PSP and Crypto services
3. Confirm URLs in `Environment.js` are correct
4. Check browser console for network errors

### QR Code Not Showing

1. Verify `qrcode.react` is installed
2. Check payment details are fetched successfully
3. Inspect console for errors

### Status Not Updating

1. Check polling interval (10 seconds)
2. Verify crypto service is running
3. Check payment ID is correct
4. Monitor Network tab for failed requests

## Security

### Best Practices

- Never display private keys
- Validate all user inputs
- Use HTTPS in production
- Implement rate limiting
- Add CSRF protection
- Sanitize displayed addresses

### Testnet Only

This application is configured for **testnet only**:
- BTC: Bitcoin Testnet
- ETH/USDT: Sepolia Testnet

⚠️ **Do not use for real cryptocurrency transactions**

## Production Deployment

For production:

1. Build optimized bundle:
   ```bash
   npm run build
   ```

2. Update environment variables
3. Configure proper HTTPS
4. Set up monitoring
5. Enable error tracking (Sentry)
6. Add rate limiting
7. Implement caching

## Contributing

When contributing:
1. Follow existing code style
2. Add comments for complex logic
3. Test all payment flows
4. Update documentation
5. Verify CORS and security

## License

Part of the Rent a Car Payment System

## Support

For issues or questions:
- Check backend service logs
- Review ARCHITECTURE.md
- See DEPLOYMENT.md for setup instructions
