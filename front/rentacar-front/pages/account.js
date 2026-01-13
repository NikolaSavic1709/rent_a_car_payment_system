import { useEffect, useState } from "react"
import style from '../styles/Vehicles.module.css'
import VehicleTile from "@/components/VehicleTile"
import { useRouter } from "next/router"
import { isDateInPast } from "@/helpers/RepresenationHelpers"
import Head from "next/head"
import { BACK_BASE_URL } from "@/values/Enviroment"
import axios from 'axios'

export default function Account() {
    const router = useRouter()
    const [data, setData] = useState(null)

    useEffect(() => getUserInfo(), [])

    function getUserInfo() {
        axios.get(`${BACK_BASE_URL}/user/info`)
            .then(response => setData(response.data))
            .catch(_error => {})
    }

    return(
        <>
            <Head>
                <title>Rent a car - Account</title>
                <meta name="viewport" content="width=device-width, initial-scale=1" />
            </Head>
            {data && (
                <div>
                    <div className="headerTitle">{data.fullname}</div>
                    <div className="spacer-h-m"/>
                    <div className="headerSubtitle">My vehicles</div>
                    <div className="spacer-h-s"/>
                    <div className={style.wrapper}>
                        {data.payments && data.payments.map((vehicle, index) => (
                                <VehicleTile 
                                    key={index} 
                                    isViewOnly={!isDateInPast(vehicle.expirationDate)}
                                    vehicle={vehicle} 
                                    onClick={() =>  {
                                        if (isDateInPast(vehicle.expirationDate)) router.push(`/vehicles/${vehicle.id}`)
                                    }}
                                />
                        ))}
                    </div>
                </div>
            )}
        </>
    )
}