import style from '../../styles/Payment.module.css'

export default function PaymentFail(){
    return (
        <div className={style.page}>
            <span className={`material-icons-outlined ${style.icon}`}>close</span>
            <div>
                <div className={style.title}>Payment Failed</div>
                <div className={style.description}>Your payment was not successfully processed. Please try again.</div>
            </div>
        </div>
    )
}