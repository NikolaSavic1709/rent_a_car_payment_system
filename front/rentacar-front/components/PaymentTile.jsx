import style from '../styles/Vehicles.module.css'
import formStyle from '../styles/FormComponents.module.css'
import { useState } from 'react'
import { PAYMANT_TYPE } from '@/helpers/enums'
import Button from '@mui/material/Button';

export default function PaymentTile({ onBuyClick }) {
    const [paymentType, setPaymentType] = useState(PAYMANT_TYPE.CREDIT_CARD)

    return(
        <div className="tileWrapper">
            <div className={style.priceDesciption}>Payment method:</div>
            <div className='spacer-h-s'/>
            <div className={`${formStyle.chipGroup} ${style.paymentsChipsWrapper}`}>
                <div
                    className={`${paymentType == PAYMANT_TYPE.CREDIT_CARD ? formStyle.selectedChip : formStyle.chip} ${style.paymentChip}`}
                    onClick={() => setPaymentType(PAYMANT_TYPE.CREDIT_CARD)}
                >
                    <span className={`material-icons-outlined`}>credit_card </span>
                    Credit Card
                </div>
                <div
                    className={`${paymentType == PAYMANT_TYPE.QR ? formStyle.selectedChip : formStyle.chip} ${style.paymentChip}`}
                    onClick={() => setPaymentType(PAYMANT_TYPE.QR)}
                >
                    <span className={`material-icons-outlined`}>qr_code </span>
                    QR code
                </div>
                <div
                    className={`${paymentType == PAYMANT_TYPE.PAYPAL ? formStyle.selectedChip : formStyle.chip} ${style.paymentChip}`}
                    onClick={() => setPaymentType(PAYMANT_TYPE.PAYPAL)}
                >
                    <img 
                        className={style.logo}
                        src={paymentType === PAYMANT_TYPE.PAYPAL ? '/paypal_white.svg' : '/paypal.svg'}
                        alt="Rentacar Logo" 
                        width="22" 
                        height="22" 
                    />
                    PayPal
                </div>
                <div
                    className={style.chipInactive}
                    onClick={() => setPaymentType(PAYMANT_TYPE.CRYPTO)}
                >
                    <span className={`material-icons-outlined`}>currency_bitcoin </span>
                    Cryptocurrency
                </div>
            </div>
            <div className='spacer-h-s'/>
            <Button 
                disableRipple 
                className={`${formStyle.button} ${formStyle.raisedButton} w-full`}
                onClick={() => {if(onBuyClick) onBuyClick(paymentType)}}
            >
                CHECKOUT
            </Button>
        </div>
    )
}