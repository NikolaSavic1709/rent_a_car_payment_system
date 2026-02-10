import { useEffect, useState } from 'react'
import { useRouter } from 'next/router'
import { CircularProgress } from '@mui/material'
import Head from 'next/head'
import style from '../styles/CardPage.module.css'

export default function PayPalPage() {
    const router = useRouter()
    const [error, setError] = useState(null)
    const [merchantOrderId, setMerchantOrderId] = useState(null)

    useEffect(() => {
        // Extract merchantOrderId from URL query parameters
        const url = window.location.href
        const params = new URLSearchParams(new URL(url).search)
        const mId = params.get('merchantOrderId')
        
        if (!mId) {
            setError('Missing payment information. Please try again.')
            return
        }

        setMerchantOrderId(mId)

        // Redirect to PayPal frontend with merchantOrderId
        // PayPal frontend will handle payment creation and redirect to PayPal
        const paypalFrontURL = `http://localhost:3003?merchantOrderId=${mId}`
        
        // Small delay to show loading state
        setTimeout(() => {
            window.location.href = paypalFrontURL
        }, 500)
        
    }, [])

    return (
        <>
            <Head>
                <title>PayLink - Redirecting to PayPal</title>
                <meta name="description" content="Redirecting to PayPal payment" />
                <meta name="viewport" content="width=device-width, initial-scale=1" />
            </Head>
            <div className="page">
                <div className={style.container}>
                    <div className={style.card}>
                        {error ? (
                            <div>
                                <h2 style={{ color: '#d32f2f', marginBottom: '1rem' }}>Error</h2>
                                <p>{error}</p>
                                <button 
                                    onClick={() => router.back()}
                                    style={{
                                        marginTop: '1rem',
                                        padding: '0.5rem 1rem',
                                        backgroundColor: '#0070ba',
                                        color: 'white',
                                        border: 'none',
                                        borderRadius: '4px',
                                        cursor: 'pointer'
                                    }}
                                >
                                    Go Back
                                </button>
                            </div>
                        ) : (
                            <div style={{ textAlign: 'center' }}>
                                <CircularProgress 
                                    size={60}
                                    sx={{ color: '#0070ba', marginBottom: '1.5rem' }}
                                />
                                <h2 style={{ marginBottom: '0.5rem' }}>Redirecting to PayPal</h2>
                                <p style={{ color: '#666' }}>Please wait while we prepare your payment...</p>
                                {merchantOrderId && (
                                    <p style={{ fontSize: '0.85rem', color: '#999', marginTop: '1rem' }}>
                                        Order: {merchantOrderId}
                                    </p>
                                )}
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </>
    )
}