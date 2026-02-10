import { useEffect } from 'react';
import { useRouter } from 'next/router';
import Head from 'next/head';
import styles from '@/styles/ResultPage.module.css';

export default function CancelPage() {
  const router = useRouter();
  const { token, reason } = router.query;

  const handleBackToShop = () => {
    window.location.href = 'http://localhost:3001';
  };

  const handleRetryPayment = () => {
    router.push('/');
  };

  return (
    <>
      <Head>
        <title>Payment Cancelled - PayPal</title>
        <meta name="description" content="PayPal payment cancelled" />
      </Head>

      <div className={styles.container}>
        <div className={styles.card}>
          <div className={styles.resultContainer}>
            <div className={styles.iconContainer} style={{ backgroundColor: '#f44336' }}>
              <span className={styles.icon}>✗</span>
            </div>

            <h1 className={styles.title}>Payment Cancelled</h1>
            
            <p className={styles.message}>
              {reason === 'missing_order_id' 
                ? 'Invalid payment session. Please try again.'
                : reason === 'payment_not_found'
                ? 'Payment not found. Please contact support.'
                : reason === 'capture_failed'
                ? 'Payment processing failed. Please try again.'
                : 'You have cancelled the PayPal payment.'}
            </p>

            {token && (
              <div className={styles.detailsCard}>
                <h3>Payment Information</h3>
                <div className={styles.details}>
                  <div className={styles.detailRow}>
                    <span className={styles.label}>Order Token:</span>
                    <span className={styles.value}>{token}</span>
                  </div>
                  {reason && (
                    <div className={styles.detailRow}>
                      <span className={styles.label}>Reason:</span>
                      <span className={styles.value}>{reason}</span>
                    </div>
                  )}
                </div>
              </div>
            )}

            <div className={styles.actions}>
              <button 
                onClick={handleRetryPayment}
                className={styles.secondaryButton}
              >
                Try Again
              </button>
              <button 
                onClick={handleBackToShop}
                className={styles.primaryButton}
              >
                Return to Shop
              </button>
            </div>

            <p className={styles.note}>
              No charges were made to your account.
            </p>
          </div>
        </div>
      </div>
    </>
  );
}
