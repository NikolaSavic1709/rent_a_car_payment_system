import Head from 'next/head'
import { useEffect } from 'react'

export default function Home() {
    useEffect(() => {
        // Redirect to payment page if accessed directly
        const params = new URLSearchParams(window.location.search)
        const merchantOrderId = params.get('merchantOrderId')
        if (merchantOrderId) {
            window.location.href = `/payment?merchantOrderId=${merchantOrderId}`
        }
    }, [])

    return (
        <>
            <Head>
                <title>Crypto Payment Service</title>
                <meta name="description" content="Cryptocurrency payment processing" />
                <meta name="viewport" content="width=device-width, initial-scale=1" />
            </Head>
            <div className="page">
                <h1>Crypto Payment Service</h1>
                <p>Please use a valid payment link to access this service.</p>
            </div>
        </>
    )
}
