import { VEHICLE_CATEGORY, VEHICLE_TYPE } from "@/helpers/enums"
import { useState } from "react"
import style from '../styles/Vehicles.module.css'
import formStyle from '../styles/FormComponents.module.css'
import { Button } from "@mui/material"
import VehicleCategorySelector from "./VehicleCategorySelector"

export default function VehicleCreation({
    onCreateClick,
    onQuitClick
}) {
    const [mode, setMode] = useState(VEHICLE_TYPE.CARS)
    const [selectedCategories, setSelectedCategories] = useState()
    const [form, setForm] = useState({
        name: '',
        description: '',
        price: ''
    })

    function onQuit() {
        if (onQuitClick) onQuitClick()
        clearForm()
    }

    function onCreate() {
        var category = selectedCategories[0].category 
        const data = {
            category: category,
            name: form.name,
            description: form.description,
            price: Number(form.price)
        }
        if (onCreateClick) onCreateClick(data)
        clearForm()
    }

    function clearForm() {
        setMode(VEHICLE_TYPE.CARS)
        setForm({
            name: '',
            description: '',
            price: ''
        })
    }

    function isButtonDisabled() {
        return form.name == '' || form.description == '' || form.price == ''
    }

    return(
        <div>
            <div className="header">
                <div className="headerTitle">Add vehicle</div>
                <div className={formStyle.chipGroup}>
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
                </div>
            </div>
            
            <VehicleCategorySelector mode={mode} onChange={(categories) => setSelectedCategories(categories)} />
            <div className="spacer-h-s"/>
            <div className={style.doubleInputWrapper}>
                <div className={`${formStyle.inputWrapper} w-full`}>
                    <input 
                        className={formStyle.input}
                        value={form.name}
                        placeholder='Name'
                        onChange={(e) => {setForm({...form, name: e.target.value})}}  
                    />
                </div>
                <div className={`${formStyle.inputWrapper} w-full`}>
                    <input 
                        className={formStyle.input}
                        value={form.price}
                        type="number"
                        placeholder='Price'
                        onChange={(e) => {setForm({...form, price: e.target.value})}}  
                    />
                </div>
            </div>
            <div className="spacer-h-s"/>
            <textarea
                className={formStyle.textArea}
                value={form.description}
                onChange={(e) => {setForm({...form, description: e.target.value})}}  
                placeholder="Description"
            />
            <div className="spacer-h-xs"/>
            <div className={formStyle.rightBuuttonsWrapper}>
                <Button 
                        disableRipple 
                        className={`${formStyle.button} ${formStyle.outlinedButton}`}
                        onClick={onQuit}
                    >
                        QUIT
                </Button>
                <Button 
                    disableRipple 
                    className={`${formStyle.button} ${formStyle.raisedButton}`}
                    onClick={onCreate}
                    disabled={isButtonDisabled()}
                >
                    Create
                </Button>
            </div>
        </div>
    )
}