import { getCategoryIconName, getCategoryLabel } from "@/helpers/RepresenationHelpers"
import style from "../styles/Vehicles.module.css"
import { VEHICLE_CATEGORY } from "@/helpers/enums"

export default function VehicleDetails({ vehicle }) {

    function getCategoryDescription() {
        return getCategoryLabel(vehicle.category) + ' plan'
    }

    return(
        <div>
            <div className="headerTitle">{vehicle.name}</div>
            <div className="spacer-h-s" />
            {vehicle.category != VEHICLE_CATEGORY.TRUCKS ? (
                <div className={style.categoryDescriptionWrapper}>
                    <span className={`material-icons-outlined ${style.headerItem}`}>
                        {getCategoryIconName(vehicle.category)} 
                    </span>
                    <div className={style.categoryDescriptionext}>{getCategoryDescription()}</div>
                </div>
                ) : (
                    <div>
                        {vehicle.items.map((item, index) => (
                            <div key={index} className={style.categoryDescriptionWrapper}>
                                <span className={`material-icons-outlined ${style.headerItem}`}>
                                    {getCategoryIconName(item.category)} 
                                </span>
                                <div className={style.categoryDescriptionext}>{item.description}</div>
                            </div>
                        ))}
                    </div>
                )}
            <div className="spacer-h-s"/>
            <div className={style.vehicleDesription}>{vehicle.description}</div>
        </div>
    )

}