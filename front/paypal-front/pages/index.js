import { useEffect, useState } from 'react';
import { useRouter } from 'next/router';
import Head from 'next/head';
import styles from '@/styles/PaymentPage.module.css';
import { PAYPAL_SERVICE_URL } from '@/values/Environment';

export default function PaymentPage() {
  const router = useRouter();
  const { merchantOrderId, amount, currency } = router.query;
  
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [paymentData, setPaymentData] = useState(null);

  useEffect(() => {
    if (!merchantOrderId) return;

    // Fetch payment details from PayPal service
    const fetchPaymentDetails = async () => {
      try {
        setLoading(true);
        
        // In a real implementation, you would fetch from backend
        // For now, we'll construct the payment initiation
        const response = await fetch(`${PAYPAL_SERVICE_URL}/payment`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            merchantOrderId,
            amount: parseFloat(amount) || 100.00,
            currency: currency || 'USD',
            description: `Order ${merchantOrderId}`
          })
        });

        if (!response.ok) {
          throw new Error('Failed to create PayPal payment');
        }

        const data = await response.json();
        setPaymentData(data);
        
        // Auto-redirect to PayPal after 2 seconds
        setTimeout(() => {
          if (data.approvalUrl) {
            window.location.href = data.approvalUrl;
          }
        }, 2000);

      } catch (err) {
        console.error('Payment error:', err);
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchPaymentDetails();
  }, [merchantOrderId, amount, currency]);

  const handleManualRedirect = () => {
    if (paymentData?.approvalUrl) {
      window.location.href = paymentData.approvalUrl;
    }
  };

  return (
    <>
      <Head>
        <title>PayPal Payment - Rent a Car</title>
        <meta name="description" content="PayPal payment processing" />
      </Head>

      <div className={styles.container}>
        <div className={styles.card}>
          <div className={styles.header}>
            <div className={styles.logo}>
              <img src="/paypal-logo.png" alt="PayPal" className={styles.logoImage} />
              <h1>PayPal Payment</h1>
            </div>
          </div>

          <div className={styles.content}>
            {loading && (
              <div className={styles.loadingContainer}>
                <div className={styles.spinner}></div>
                <p className={styles.loadingText}>Preparing your payment...</p>
                <p className={styles.subText}>You will be redirected to PayPal shortly</p>
              </div>
            )}

            {error && (
              <div className={styles.errorContainer}>
                <div className={styles.errorIcon}>✗</div>
                <h2>Payment Error</h2>
                <p className={styles.errorMessage}>{error}</p>
                <button 
                  onClick={() => router.back()} 
                  className={styles.backButton}
                >
                  Go Back
                </button>
              </div>
            )}

            {!loading && !error && paymentData && (
              <div className={styles.successContainer}>
                <div className={styles.checkIcon}>✓</div>
                <h2>Payment Ready</h2>
                
                <div className={styles.paymentDetails}>
                  <div className={styles.detailRow}>
                    <span className={styles.label}>Amount:</span>
                    <span className={styles.value}>
                      {currency} {amount}
                    </span>
                  </div>
                  <div className={styles.detailRow}>
                    <span className={styles.label}>Order ID:</span>
                    <span className={styles.value}>{merchantOrderId}</span>
                  </div>
                </div>

                <p className={styles.redirectText}>
                  Redirecting to PayPal in a moment...
                </p>

                <button 
                  onClick={handleManualRedirect}
                  className={styles.continueButton}
                >
                  Continue to PayPal Now
                </button>

                <p className={styles.secureText}>
                  🔒 Secure payment powered by PayPal
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </>
  );
}
