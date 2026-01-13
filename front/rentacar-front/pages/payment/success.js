import style from '../../styles/Payment.module.css'

export default function PaymentSuccess(){
    return (
        <div className={style.page}>
            <span className={`material-icons-outlined ${style.icon}`}>done</span>
            <div>
                <div className={style.title}>Payment Successful</div>
                <div className={style.description}>Thank you for your payment.</div>
            </div>
        </div>
    )
}