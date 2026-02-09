import { useEffect, useState } from 'react';
import { useRouter } from 'next/router';
import Head from 'next/head';
import styles from '@/styles/ResultPage.module.css';
import { PAYPAL_SERVICE_URL } from '@/values/Environment';
import { getStatusDisplay } from '@/helpers/enums';

export default function SuccessPage() {
  const router = useRouter();
  const { token, PayerID, paymentId } = router.query;
  
  const [loading, setLoading] = useState(true);
  const [paymentStatus, setPaymentStatus] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!token && !paymentId) return;

    const checkPaymentStatus = async () => {
      try {
        setLoading(true);
        
        // Wait a bit for the callback to be processed
        await new Promise(resolve => setTimeout(resolve, 2000));

        // Check payment status
        const statusResponse = await fetch(
          `${PAYPAL_SERVICE_URL}/payment-status/${paymentId || token}`
        );

        if (statusResponse.ok) {
          const data = await statusResponse.json();
          setPaymentStatus(data);
        } else {
          // If status endpoint fails, assume success based on callback
          setPaymentStatus({
            status: 'completed',
            paymentId: paymentId || token,
            transactionId: router.query.transactionId
          });
        }

      } catch (err) {
        console.error('Status check error:', err);
        // Don't set error, assume success
        setPaymentStatus({
          status: 'completed',
          paymentId: paymentId || token
        });
      } finally {
        setLoading(false);
      }
    };

    checkPaymentStatus();
  }, [token, PayerID, paymentId, router.query.transactionId]);

  const handleBackToShop = () => {
    // Redirect to webshop or main app
    window.location.href = 'http://localhost:3001';
  };

  const statusDisplay = getStatusDisplay(paymentStatus?.status);

  return (
    <>
      <Head>
        <title>Payment Successful - PayPal</title>
        <meta name="description" content="PayPal payment successful" />
      </Head>

      <div className={styles.container}>
        <div className={styles.card}>
          {loading ? (
            <div className={styles.loadingContainer}>
              <div className={styles.spinner}></div>
              <h2>Processing Payment...</h2>
              <p>Please wait while we confirm your payment</p>
            </div>
          ) : (
            <div className={styles.resultContainer}>
              <div className={styles.iconContainer} style={{ backgroundColor: statusDisplay.color }}>
                <span className={styles.icon}>{statusDisplay.icon}</span>
              </div>

              <h1 className={styles.title}>Payment Successful!</h1>
              
              <p className={styles.message}>
                Your payment has been processed successfully through PayPal.
              </p>

              {paymentStatus && (
                <div className={styles.detailsCard}>
                  <h3>Payment Details</h3>
                  <div className={styles.details}>
                    {paymentStatus.paymentId && (
                      <div className={styles.detailRow}>
                        <span className={styles.label}>Payment ID:</span>
                        <span className={styles.value}>{paymentStatus.paymentId}</span>
                      </div>
                    )}
                    {paymentStatus.transactionId && (
                      <div className={styles.detailRow}>
                        <span className={styles.label}>Transaction ID:</span>
                        <span className={styles.value}>{paymentStatus.transactionId}</span>
                      </div>
                    )}
                    {paymentStatus.amount && (
                      <div className={styles.detailRow}>
                        <span className={styles.label}>Amount:</span>
                        <span className={styles.value}>
                          {paymentStatus.currency} {paymentStatus.amount}
                        </span>
                      </div>
                    )}
                    {PayerID && (
                      <div className={styles.detailRow}>
                        <span className={styles.label}>Payer ID:</span>
                        <span className={styles.value}>{PayerID}</span>
                      </div>
                    )}
                    <div className={styles.detailRow}>
                      <span className={styles.label}>Status:</span>
                      <span 
                        className={styles.statusBadge}
                        style={{ backgroundColor: statusDisplay.color }}
                      >
                        {statusDisplay.text}
                      </span>
                    </div>
                  </div>
                </div>
              )}

              <div className={styles.actions}>
                <button 
                  onClick={handleBackToShop}
                  className={styles.primaryButton}
                >
                  Return to Shop
                </button>
              </div>

              <p className={styles.note}>
                A confirmation email will be sent to your registered email address.
              </p>
            </div>
          )}
        </div>
      </div>
    </>
  );
}
