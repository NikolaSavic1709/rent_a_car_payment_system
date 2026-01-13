import VehicleDetails from "@/components/VehicleDetails"
import Head from "next/head"
import { useRouter } from "next/router"
import { useEffect, useState } from "react"
import style from "../../styles/Vehicles.module.css"
import VehiclePriceTile from "@/components/VehiclePriceTile"
import PaymentTile from "@/components/PaymentTile"
import axios from 'axios'
import { BACK_BASE_URL } from "@/values/Enviroment"
import { ERROR } from "@/values/Errors"
import { Popup } from "@/components/widgets/Popup"
import { getUserRole } from "@/helpers/auth"
import { PAYMANT_TYPE, USER_TYPE } from "@/helpers/enums"

export default function Vehicle() {
    const router = useRouter()
    const [role, setRole] = useState(null)
    const { id } = router.query
    const [vehicle, setVehicle] = useState(null)
    const [days, setDays] = useState(1)
    const [haveError, setHaveError] = useState(false)
    const [errorMessage, setErrorMesagge] = useState('')

    useEffect(() => {setRole(getUserRole())}, [router])
    useEffect(() => {if (id) getVehicle()}, [id])

    function getVehicle() {
        axios.get(`${BACK_BASE_URL}/vehicles/${id}`)
            .then(response => setVehicle(response.data))
            .catch(_error => {})
    }

    function onBuyClick(paymentMethod) {
        var payment = {
            vehicleId: Number(id),
            method: paymentMethod,
            days: Number(days)
        };
        if (paymentMethod===PAYMANT_TYPE.QR){
            window.location.href = "http://localhost:3001/qr";
            return
        }
        if (paymentMethod===PAYMANT_TYPE.CRYPTO){
            window.location.href = "http://localhost:3001/crypto";
            return
        }
        if (paymentMethod===PAYMANT_TYPE.PAYPAL){
            window.location.href = "http://localhost:3001/paypal";
            return
        }
        console.log(payment);
        axios.post(`${BACK_BASE_URL}/vehicles/purchase`, payment)
            .then(response => {
                console.log(response.data);
                // Redirect to the URL from the response
                if (response.data && response.data.redirectUrl) {
                    window.location.href = response.data.redirectUrl;
                } else {
                    console.error('responseUrl is missing in the response');
                }
            })
            .catch(_error => {
                setHaveError(true);
                setErrorMesagge(ERROR.PAYMENT_INITIALISATION);
            });
    }

    return (
        <>
            <Head>
                <title>Rent a car - Vehicle details</title>
                <meta name="viewport" content="width=device-width, initial-scale=1" />
            </Head>
                {vehicle &&
                    <div className={style.vehicleDetailsPageTemplate}>
                        <VehicleDetails vehicle={vehicle} />
                        {role != USER_TYPE.ADMIN && (
                            <div>
                                <VehiclePriceTile vehicle={vehicle} onDaysChange={setDays} />
                                <div className="spacer-h-m"/>
                                <PaymentTile onBuyClick={onBuyClick} />
                            </div>
                        )}
                    </div>
                }
                <Popup
                    open={haveError}
                    onClose={() => setHaveError(false)}
                    message={errorMessage}
                    severity="error"
                />
        </>
    )
}