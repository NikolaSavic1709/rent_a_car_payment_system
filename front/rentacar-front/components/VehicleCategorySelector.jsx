import { VEHICLE_CATEGORY, VEHICLE_TYPE } from "@/helpers/enums"
import { useEffect, useState } from "react"
import style from '../styles/Vehicles.module.css'
import formStyle from '../styles/FormComponents.module.css'
import { getCategoryLabel } from "@/helpers/RepresenationHelpers"

export default function VehicleCategorySelector({ mode, onChange }) {
    const [m, setM] = useState()
    const [selectedCategories, setSelectedCategories] = useState([{ category: VEHICLE_CATEGORY.SUV, description: "" }])

    useEffect(() => setM(mode), [mode])
    useEffect(() => {if (onChange) onChange(selectedCategories)}, [selectedCategories])

    useEffect(() => {
        setM(mode);
        if (mode === VEHICLE_TYPE.CARS && selectedCategories.length > 1) {
            setSelectedCategories([selectedCategories[0]])
        }
    }, [mode]);

    const handleCategoryClick = (category) => {
        if (m === VEHICLE_TYPE.CARS) {
            setSelectedCategories([{ category, description: "" }])
        }
    }

    const handleDescriptionChange = (category, value) => {
        setSelectedCategories((prev) =>
            prev.map((item) =>
                item.category === category ? { ...item, description: value } : item
            )
        )
    }

    return (
        <div>
            <div className={style.headerWrapper}>
                {/* Chips */}
                <div className={formStyle.chipGroup}>
                    {Object.values(VEHICLE_CATEGORY)
                        .filter((category) => category !== VEHICLE_CATEGORY.TRUCKS)
                        .map((category) => (
                            <div
                                key={category}
                                className={selectedCategories.some((item) => item.category === category)? formStyle.selectedChip: formStyle.chip}
                                onClick={() => handleCategoryClick(category)}
                            >
                                {getCategoryLabel(category)}
                            </div>
                        ))}
                </div>
            </div>
            {/* Descriptions */}
        </div>
    )
}