const BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost';
const PAYPAL_SERVICE_URL = process.env.NEXT_PUBLIC_PAYPAL_SERVICE_URL || 'http://localhost:8088';
const PSP_SERVICE_URL = process.env.NEXT_PUBLIC_PSP_SERVICE_URL || 'http://localhost';

export { BASE_URL, PAYPAL_SERVICE_URL, PSP_SERVICE_URL };
