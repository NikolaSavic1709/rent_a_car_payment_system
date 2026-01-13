import axios from 'axios'
import { BACK_BASE_URL } from '@/values/Enviroment'
import React, { useEffect, useState } from 'react'
import { CircularProgress } from '@mui/material'
import style from '../../styles/Payment.module.css'

export default function PaymentProcessing() {

    const [response, setResponse] = useState(null)
    const [timeElapsed, setTimeElapsed] = useState(0)
    const [merchantId, setMerchantId] = useState(null)

    useEffect(() => {
        if (merchantId) return
        const url = window.location.href
        const params = new URLSearchParams(new URL(url).search)
        setMerchantId(params.get('merchantOrderId'))
    }, [])



    useEffect(() => {
        const fetchData = () => {
            var data = {merchantOrderId: merchantId}
            axios.post(`${BACK_BASE_URL}/purchase-status/check`, data)
                .then(response => window.location.href = response.data.url)
                .catch(_error => { })
        }

        const interval = setInterval(() => {
            fetchData()
            setTimeElapsed(prev => prev + 5)
        }, 5000)

        if (response || timeElapsed >= 30) {
            clearInterval(interval);
            window.location.href = `http://localhost:3000/payment/fail`
        }

        return () => clearInterval(interval);
    }, [response, timeElapsed]);


    return (
        <div className={style.page}>
            <CircularProgress 
                size={100}
                sx={{color: '#228aba'}}
            />
            <div>
                <div className={style.title}>Payment processing</div>
                <div className={style.description}>Please wait while we process your transaction.</div>
            </div>
        </div>
    )
}