import { VEHICLE_CATEGORY } from "@/helpers/enums"
import style from "../styles/Vehicles.module.css"
import { formatDate, getCategoryIconName, isDateInPast } from "../helpers/RepresenationHelpers"

export default function VehicleTile({ 
    vehicle,
    isViewOnly = false,
    onClick 
}) {
    return(
        <div 
            className={`${style.vehicleWrapper} ${isViewOnly ? style.isViewOnly : ''}`}
            onClick={onClick ? onClick : undefined}
        >
            <div className={style.headerWrapper}>
                {vehicle?.category == VEHICLE_CATEGORY.TRUCKS ? (
                    <div className={style.headerWrapper}>
                        {vehicle.items.map((item, index) => (
                            <span key={index} className={`material-icons-outlined ${style.headerItem}`}>
                                {getCategoryIconName(item.category)}
                            </span>
                        ))}
                    </div>
                ) : (
                    <span className={`material-icons-outlined ${style.headerItem}`}>
                        {getCategoryIconName(vehicle.category)} 
                    </span>
                )}

                {vehicle.price ? (
                    <div className={`${style.headerItem} ${style.headerItemText}`}>{vehicle.price} €/day</div>
                ) : (
                    <div className={`${style.headerItem} ${style.headerItemText} ${isDateInPast(vehicle.expirationDate) ? style.headerItemInvalidText : '' }`}>
                        {formatDate(vehicle.deadline)}
                    </div>
                )}
            </div>
            <div className={style.vehicleTitle}>{vehicle.name}</div>
            <div className={style.vehicleDescription}>{vehicle.description}</div>
        </div>
    )
}