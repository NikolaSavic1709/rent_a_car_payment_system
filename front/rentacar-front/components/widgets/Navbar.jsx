import Link from "next/link"
import style from "../../styles/Navbar.module.css"
import { useEffect, useState } from "react"
import { useRouter } from "next/router"
import { getUserRole, logOut } from "@/helpers/auth"
import { USER_TYPE } from "@/helpers/enums"
import axios from "axios"
import { BACK_BASE_URL } from "@/values/Enviroment"

export default function Navbar() {
    const router = useRouter()
    const [role, setRole] = useState(null)
    const [selectedOption, setSelectedOption] = useState(PAGE.VEHICLES)

    useEffect(() => {setRole(getUserRole())}, [router])

    return(
        <div className={style.wrapper}>
            <div className={style.topWrapper}>
                {/* <img 
                    className={style.logo}
                    src="/rentacar_small.svg" 
                    alt="Rentacar Logo" 
                    width="35" 
                    height="35" 
                /> */}
                {(role == USER_TYPE.ADMIN || USER_TYPE.CUSTOMER) && role != null && (
                    <Link 
                        className={`${style.optionWrapper}`}
                        href={`/${PAGE.VEHICLES}`}
                        onClick={() => setSelectedOption(PAGE.VEHICLES)}
                    >
                        <span className={`material-icons-outlined ${style.option} ${selectedOption === PAGE.VEHICLES ? style.selectedOption : ''}`}>category</span>
                        <div className={style.hoverText}>Vehicles</div>
                    </Link>
                )}
                {role == USER_TYPE.CUSTOMER && (
                    <Link 
                        className={`${style.optionWrapper}`}
                        href={`/${PAGE.ACCOUNT}`}
                        onClick={() => setSelectedOption(PAGE.ACCOUNT)}
                    >
                        <span className={`material-icons-outlined ${style.option} ${selectedOption === PAGE.ACCOUNT ? style.selectedOption : ''}`}>person</span>
                        <div className={style.hoverText}>Account</div>
                    </Link>
                )}
                {role == USER_TYPE.ADMIN && (
                    <div 
                        className={`${style.optionWrapper}`}
                        onClick={() => {
                            axios.get(`${BACK_BASE_URL}/subscription`)
                                .then(response => window.open(response.data.url, '_blank'))
                                .catch(_error => {})
                        }}
                    >
                        <span className={`material-icons-outlined ${style.option}`}>settings</span>
                        <div className={style.hoverText}>Payment Subscriptions</div>
                    </div>
                )}
            </div>
            {role != null && (
                <Link 
                    className={style.optionWrapper}
                    href={`/`}
                    onClick={() => {
                        setSelectedOption(PAGE.VEHICLES) //next time when log in it vehicles opt will be selected
                        setRole(null)
                        logOut()
                    }}
                >
                    <span className={`material-icons-outlined ${style.option}`}>exit_to_app</span>
                    <div className={style.hoverText}>Logout</div>
                </Link>
            )}
        </div>
    )
}

const PAGE = {
    VEHICLES: 'vehicles',
    ACCOUNT: 'account',
    LOGOUT: 'logout'
}