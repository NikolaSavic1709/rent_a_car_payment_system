import VehicleTile from "@/components/VehicleTile"
import Head from "next/head"
import { useEffect, useState } from "react"
import style from '../../styles/Vehicles.module.css'
import formStyle from '../../styles/FormComponents.module.css'
import { useRouter } from "next/router"
import { VEHICLE_TYPE, USER_TYPE } from "@/helpers/enums"
import { Dialog } from "@/components/widgets/Dialog"
import VehicleCreation from "@/components/VehicleCreation"
import axios from 'axios'
import { BACK_BASE_URL } from "@/values/Enviroment"
import { ERROR } from "@/values/Errors"
import { getUserRole } from "@/helpers/auth"
import { Popup } from "@/components/widgets/Popup"

export default function Vehicles() {
    const router = useRouter()
    const [vehicles, setVehicles] = useState([])
    const [selectedVehicles, setSelectedVehicles] = useState([])
    const [mode, setMode] = useState(VEHICLE_TYPE.ALL)
    const [userType, setUserType] = useState(USER_TYPE.ADMIN)
    const [isDialogOpen, setIsDialogOpen] = useState(false)
    const [haveError, setHaveError] = useState(false)
    const [errorMessage, setErrorMesagge] = useState('')

    useEffect(() => filterVehicles(vehicles), [mode])
    useEffect(() => {
        setUserType(getUserRole())
        getVehicles()
    }, [])
    useEffect(() => {
        if (!isDialogOpen) getVehicles()
    }, [isDialogOpen])

    function getVehicles() {
        axios.get(`${BACK_BASE_URL}/vehicles`)
            .then(response => { 
                setVehicles(response.data)
                filterVehicles(response.data)
            })
            .catch(_error => {})
    }

    function filterVehicles(data) {
        if (mode == VEHICLE_TYPE.ALL) setSelectedVehicles(data)
        else if (mode == VEHICLE_TYPE.TRUCKS) setSelectedVehicles(data.filter(vehicle => vehicle.category == VEHICLE_TYPE.TRUCKS))
        else setSelectedVehicles(data.filter(vehicle => vehicle.category != VEHICLE_TYPE.TRUCKS))
    }

    async function create(data) {
        try {
            const response = await axios.post(`${BACK_BASE_URL}/vehicles`, data)
            if (response.status === 201) setIsDialogOpen(false)
            else setErrorMesagge(ERROR.VEHICLE_CREATION)
        } catch (error) { setErrorMesagge(ERROR.VEHICLE_CREATION) }
    }

    return (
        <>
            <Head>
                <title>Rent a car - Vehicles</title>
                <meta name="viewport" content="width=device-width, initial-scale=1" />
            </Head>
            <div className="header">
                <div className="headerTitle">Vehicles</div>
                <div className={formStyle.chipGroup}>
                    <div
                        className={`${mode === VEHICLE_TYPE.ALL ? formStyle.selectedChip : formStyle.chip}`}
                        onClick={() => setMode(VEHICLE_TYPE.ALL)}
                    >
                        All
                    </div>
                    <div
                        className={`${mode === VEHICLE_TYPE.CARS ? formStyle.selectedChip : formStyle.chip}`}
                        onClick={() => setMode(VEHICLE_TYPE.CARS)}
                    >
                        Cars
                    </div>
                    <div 
                        className={`${mode === VEHICLE_TYPE.TRUCKS ? formStyle.selectedChip : formStyle.chip}`}
                        onClick={() => setMode(VEHICLE_TYPE.TRUCKS)}
                    >
                        Trucks
                    </div>
                    {userType == USER_TYPE.ADMIN && (
                        <div 
                            className={`${formStyle.chip}`}
                            onClick={() => {setIsDialogOpen(true)}}
                        >
                        <span className={`material-icons-outlined`}>add</span>
                        Add new
                    </div>
                    )}
                </div>
            </div>
            <div className={style.wrapper}>
                {selectedVehicles && selectedVehicles.map((vehicle, index) => (
                        <VehicleTile 
                            key={index} 
                            vehicle={vehicle} 
                            onClick={() => router.push(`/vehicles/${vehicle.id}`)}
                        />
                ))}
            </div>
            <Dialog
                width={700}
                isOpen={isDialogOpen}
            > 
                <VehicleCreation 
                    onCreateClick={(data) => create(data)}
                    onQuitClick={() => setIsDialogOpen(false)}
                />
            </Dialog>
            <Popup
                open={haveError}
                onClose={() => setHaveError(false)}
                message={errorMessage}
                severity="error"
            />
        </>
    );
    
}