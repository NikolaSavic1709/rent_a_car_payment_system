import Head from "next/head"
import formStyle from '../styles/Form.module.css'
import { useEffect, useState } from "react"
import style from '../styles/SubscriptionsPage.module.css'
import { Button } from "@mui/material"
import axios from "axios"
import { BACK_BASE_URL } from "@/values/Enviroment"
import { Popup } from "@/components/widgets/Popup"

export default function Subscription() {
    const [showPassword, setShowPassword] = useState(false)
    const [openPopup, setOpenPopup] = useState(false)
    const [popupMessage, setPopupMessage] = useState('')
    const [popupSeverity, setPopupSeverity] = useState('error')
    const [form, setForm] = useState({
        merchantId: null,
        merchantPassword: '',
        methods: []
    })

    useEffect(() => {
        if (form.merchantId) return
        const url = window.location.href
        const params = new URLSearchParams(new URL(url).search)
        const id = Number(params.get('merchantId'))
        axios.get(`${BACK_BASE_URL}/subscription/${id}`, form)
            .then(response => {
                setForm({ ...form, merchantId: id, methods: response.data ? response.data : []})
            })
            .catch(_error => {console.log(_error)})
    }, [])

    function toggleInList(str) {
        var list = form.methods
        const index = list.indexOf(str)
        if (index === -1) list.push(str)
        else list.splice(index, 1)
        setForm({ ...form, methods: list})
    }

    function isButtonDisabled() {
        return form.merchantPassword == ''
    }

    function subscribe() {
        axios.post(`${BACK_BASE_URL}/subscription`, form)
            .then(response => {
                setPopupMessage('You have successfully subscribed to the selected payment methods.')
                setPopupSeverity('success')
                setOpenPopup(true)
            })
            .catch(_error => {
                setPopupMessage('You have not subscribed to the selected payment methods. Please try again.')
                setPopupSeverity('error')
                setOpenPopup(true)
            })
    }

    return(
        <>
            <Head>
                <title>PayLink - Payment Subscriptons</title>
                <meta name="viewport" content="width=device-width, initial-scale=1" />
                <link rel="icon" href="/favicon.ico" />
            </Head>
            <div>
                <div className="paymentTitle w-full">Payment</div>
                <div className="paymentTitle w-full">Subscriptions</div>
                <div className={style.description}>Choose the payment methods you want to subscribe to.</div>
                <div className='spacer-h-s' />
                <div className={style.wrapper}>
                    <div className={style.rowWrapper}>
                        <div 
                            className={`${formStyle.bigChip} ${form.methods.includes(PAYMENT_METHOD.CARD) ? formStyle.bigChipSelected : ''}`}
                            onClick={() => toggleInList(PAYMENT_METHOD.CARD)}
                        >
                            Card
                        </div>
                        <div 
                            className={`${formStyle.bigChip} ${form.methods.includes(PAYMENT_METHOD.QR_CODE) ? formStyle.bigChipSelected : ''}`}
                            onClick={() => toggleInList(PAYMENT_METHOD.QR_CODE)}
                        >
                            QR Code
                        </div>
                    </div>
                    <div className={style.rowWrapper}>
                        <div 
                            className={`${formStyle.bigChip} ${form.methods.includes(PAYMENT_METHOD.PAYPAL) ? formStyle.bigChipSelected : ''}`}
                            onClick={() => toggleInList(PAYMENT_METHOD.PAYPAL)}
                        >
                            PayPal
                        </div>
                        <div 
                            className={`${formStyle.bigChip} ${form.methods.includes(PAYMENT_METHOD.CRYPTO) ? formStyle.bigChipSelected : ''}`}
                            onClick={() => toggleInList(PAYMENT_METHOD.CRYPTO)}
                        >
                            Crypto
                        </div>
                    </div>
                    <div>
                        <div className={`${formStyle.inputWrapper} w-full`}>
                            <input
                                className={formStyle.input}
                                value={form.password}
                                placeholder='Merchant Password'
                                type={showPassword ? "text" : "password"}
                                onChange={(e) => setForm({ ...form, merchantPassword: e.target.value})} 
                            />
                            <span onClick={() => setShowPassword(prevState => !prevState)} className={`material-icons-outlined  ${formStyle.inputIcon}`}>{showPassword ? 'visibility_off' : 'visibility'}</span> 
                        </div>
                    </div>
                    <Button
                        disableRipple
                        className={`${formStyle.button} ${formStyle.raisedButton} w-full`}
                        onClick={subscribe}
                        disabled={isButtonDisabled()}
                    >
                        SUBSCRIBE
                    </Button>
                </div>
                <Popup
                    open={openPopup}
                    onClose={() => setOpenPopup(false)}
                    message={popupMessage}
                    severity={popupSeverity}
                />
            </div>
        </>
    )
}

const PAYMENT_METHOD = {
    CARD: 0,
    PAYPAL: 1,
    CRYPTO: 2,
    QR_CODE: 4,
}