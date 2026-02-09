import { useState, useEffect, useCallback } from 'react'
import Head from 'next/head'
import axios from 'axios'
import { QRCodeSVG } from 'qrcode.react'
import { CircularProgress, Button } from '@mui/material'
import style from '../styles/CryptoPage.module.css'
import { PSP_BASE_URL } from '@/values/Environment'
import { PAYMENT_STATUS } from '@/helpers/enums'

export default function CryptoPaymentPage() {
    const [merchantOrderId, setMerchantOrderId] = useState(null)
    const [paymentData, setPaymentData] = useState(null)
    const [paymentStatus, setPaymentStatus] = useState(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)
    const [timeRemaining, setTimeRemaining] = useState(null)
    const [copied, setCopied] = useState(false)

    // Extract merchantOrderId from URL
    useEffect(() => {
        const params = new URLSearchParams(window.location.search)
        const id = params.get('merchantOrderId')
        if (id) {
            setMerchantOrderId(id)
        } else {
            setError('Invalid payment link')
            setLoading(false)
        }
    }, [])

    // Fetch payment details from PSP
    const fetchPaymentDetails = useCallback(async () => {
        if (!merchantOrderId) return

        try {
            const response = await axios.get(
                `${PSP_BASE_URL}/crypto-payment-details?merchantOrderId=${merchantOrderId}`
            )
            setPaymentData(response.data)
            setLoading(false)
        } catch (err) {
            console.error('Error fetching payment details:', err)
            setError(err.response?.data?.error || 'Failed to load payment details')
            setLoading(false)
        }
    }, [merchantOrderId])

    // Poll payment status
    const pollPaymentStatus = useCallback(async () => {
        if (!paymentData?.paymentId) return

        try {
            const response = await axios.get(
                `${PSP_BASE_URL}/crypto-status?paymentId=${paymentData.paymentId}`
            )
            setPaymentStatus(response.data)

            // Redirect on confirmed/expired/failed
            if (response.data.status === PAYMENT_STATUS.CONFIRMED) {
                setTimeout(() => {
                    window.location.href = `http://localhost:3000/payment?merchantOrderId=${merchantOrderId}`
                }, 2000)
            } else if (
                response.data.status === PAYMENT_STATUS.EXPIRED ||
                response.data.status === PAYMENT_STATUS.FAILED
            ) {
                setTimeout(() => {
                    window.location.href = `http://localhost:3000/payment?merchantOrderId=${merchantOrderId}`
                }, 3000)
            }
        } catch (err) {
            console.error('Error polling status:', err)
        }
    }, [paymentData, merchantOrderId])

    // Initialize: fetch payment details
    useEffect(() => {
        if (merchantOrderId) {
            fetchPaymentDetails()
        }
    }, [merchantOrderId, fetchPaymentDetails])

    // Start polling when we have payment data
    useEffect(() => {
        if (!paymentData) return

        pollPaymentStatus()
        const interval = setInterval(pollPaymentStatus, 10000) // Poll every 10 seconds

        return () => clearInterval(interval)
    }, [paymentData, pollPaymentStatus])

    // Countdown timer
    useEffect(() => {
        if (!paymentData?.expiryTime) return

        const updateTimer = () => {
            const now = new Date()
            const expiry = new Date(paymentData.expiryTime)
            const diff = expiry - now

            if (diff <= 0) {
                setTimeRemaining(0)
                return
            }

            const minutes = Math.floor(diff / 60000)
            const seconds = Math.floor((diff % 60000) / 1000)
            setTimeRemaining(`${minutes}:${seconds.toString().padStart(2, '0')}`)
        }

        updateTimer()
        const interval = setInterval(updateTimer, 1000)

        return () => clearInterval(interval)
    }, [paymentData])

    const copyAddress = () => {
        if (paymentData?.destinationAddress) {
            navigator.clipboard.writeText(paymentData.destinationAddress)
            setCopied(true)
            setTimeout(() => setCopied(false), 2000)
        }
    }

    const getStatusDisplay = () => {
        if (!paymentStatus) return { text: 'Waiting for payment', className: style.statusPending }

        switch (paymentStatus.status) {
            case PAYMENT_STATUS.PENDING:
                return { text: 'Waiting for payment', className: style.statusPending }
            case PAYMENT_STATUS.CONFIRMING:
                return {
                    text: `Confirming (${paymentStatus.confirmations}/${paymentData?.requiredConfirmations || 0})`,
                    className: style.statusConfirming
                }
            case PAYMENT_STATUS.CONFIRMED:
                return { text: 'Payment confirmed', className: style.statusConfirmed }
            case PAYMENT_STATUS.EXPIRED:
                return { text: 'Payment expired', className: style.statusExpired }
            case PAYMENT_STATUS.FAILED:
                return { text: 'Payment failed', className: style.statusFailed }
            default:
                return { text: paymentStatus.status, className: style.statusPending }
        }
    }

    const getConfirmationProgress = () => {
        if (!paymentStatus || !paymentData) return 0
        return (paymentStatus.confirmations / paymentData.requiredConfirmations) * 100
    }

    if (loading) {
        return (
            <>
                <Head>
                    <title>Loading Payment...</title>
                </Head>
                <div className="page">
                    <div className={style.loadingContainer}>
                        <CircularProgress size={60} sx={{ color: '#228aba' }} />
                        <div className={style.loadingText}>Loading payment details...</div>
                    </div>
                </div>
            </>
        )
    }

    if (error) {
        return (
            <>
                <Head>
                    <title>Payment Error</title>
                </Head>
                <div className="page">
                    <div className={style.errorContainer}>
                        <div className={style.errorTitle}>Payment Error</div>
                        <div className={style.errorMessage}>{error}</div>
                        <Button
                            className={style.actionButton}
                            onClick={() => window.location.href = 'http://localhost:3000'}
                        >
                            Return to Shop
                        </Button>
                    </div>
                </div>
            </>
        )
    }

    if (paymentStatus?.status === PAYMENT_STATUS.CONFIRMED) {
        return (
            <>
                <Head>
                    <title>Payment Successful</title>
                </Head>
                <div className="page">
                    <div className={style.successContainer}>
                        <span className={`material-icons-outlined ${style.successIcon}`}>
                            check_circle
                        </span>
                        <div className={style.successTitle}>Payment Confirmed!</div>
                        <div className={style.successMessage}>
                            Your payment has been confirmed. Redirecting...
                        </div>
                        <CircularProgress size={40} sx={{ color: '#228aba' }} />
                    </div>
                </div>
            </>
        )
    }

    const statusDisplay = getStatusDisplay()
    const paymentURI = `${paymentData?.currency?.toLowerCase()}:${paymentData?.destinationAddress}?amount=${paymentData?.amount}`

    return (
        <>
            <Head>
                <title>Crypto Payment - {paymentData?.currency}</title>
                <meta name="viewport" content="width=device-width, initial-scale=1" />
            </Head>
            <div className="page">
                <div className={style.container}>
                    <div className={style.header}>
                        <div className={style.title}>Cryptocurrency Payment</div>
                        <div className={style.subtitle}>
                            Complete your payment using {paymentData?.currency}
                        </div>
                    </div>

                    <div className={style.paymentSection}>
                        <div className={style.sectionTitle}>
                            <span className="material-icons-outlined">payment</span>
                            Payment Details
                        </div>
                        <div className={style.paymentDetails}>
                            <div className={style.detailRow}>
                                <span className={style.detailLabel}>Amount</span>
                                <span className={style.detailValue}>
                                    {paymentData?.amount} {paymentData?.currency}
                                </span>
                            </div>
                            <div className={style.detailRow}>
                                <span className={style.detailLabel}>Currency</span>
                                <span className={style.detailValue}>{paymentData?.currency}</span>
                            </div>
                            <div className={style.detailRow}>
                                <span className={style.detailLabel}>Confirmations Required</span>
                                <span className={style.detailValue}>
                                    {paymentData?.requiredConfirmations}
                                </span>
                            </div>
                        </div>

                        {timeRemaining !== null && timeRemaining !== 0 && (
                            <div className={style.timerSection}>
                                <span className="material-icons-outlined" style={{ color: '#856404' }}>
                                    schedule
                                </span>
                                <span className={style.timerText}>Time remaining:</span>
                                <span className={style.timerValue}>{timeRemaining}</span>
                            </div>
                        )}

                        <div className={style.addressSection}>
                            <div className={style.addressLabel}>Send payment to:</div>
                            <div className={style.addressValue}>
                                <span>{paymentData?.destinationAddress}</span>
                                <Button
                                    className={style.copyButton}
                                    onClick={copyAddress}
                                    size="small"
                                >
                                    {copied ? 'Copied!' : 'Copy'}
                                </Button>
                            </div>
                        </div>
                    </div>

                    <div className={style.paymentSection}>
                        <div className={style.sectionTitle}>
                            <span className="material-icons-outlined">qr_code</span>
                            Scan QR Code
                        </div>
                        <div className={style.qrSection}>
                            <div className={style.qrCode}>
                                <QRCodeSVG value={paymentURI} size={200} />
                            </div>
                            <div className={style.qrInstruction}>
                                Scan with your crypto wallet app
                            </div>
                        </div>
                    </div>

                    <div className={style.statusSection}>
                        <div className={style.sectionTitle}>
                            <span className="material-icons-outlined">info</span>
                            Payment Status
                        </div>
                        <div className={`${style.statusBadge} ${statusDisplay.className}`}>
                            {statusDisplay.text}
                        </div>
                        {paymentStatus?.status === PAYMENT_STATUS.CONFIRMING && (
                            <>
                                <div className={style.progressBar}>
                                    <div
                                        className={style.progressFill}
                                        style={{ width: `${getConfirmationProgress()}%` }}
                                    />
                                </div>
                                <div className={style.confirmationsInfo}>
                                    {paymentStatus.confirmations} of {paymentData?.requiredConfirmations} confirmations
                                </div>
                            </>
                        )}
                        <div className={style.helpText}>
                            Status updates automatically every 10 seconds
                        </div>
                    </div>

                    <div className={style.warningBox}>
                        <span className={`material-icons-outlined ${style.warningIcon}`}>
                            warning
                        </span>
                        <div className={style.warningText}>
                            Only send {paymentData?.currency} to this address. Sending any other
                            cryptocurrency will result in permanent loss of funds.
                        </div>
                    </div>
                </div>
            </div>
        </>
    )
}
