import style from '../styles/Vehicles.module.css'
import formStyle from '../styles/FormComponents.module.css'
import { useEffect, useState } from 'react'

export default function VehiclePriceTile({ vehicle, onDaysChange }) {
    const [days, setDays] = useState(1)
    const [totalPrice, setTotalPrice] = useState(0)
    const [displayedPrice, setDisplayedPrice] = useState(0)

    useEffect(() => {
        if (vehicle) {
            const newPrice = days * vehicle.price
            setTotalPrice(newPrice)
            animatePriceChange(newPrice)
            if (onDaysChange) onDaysChange(days)
        }
    }, [days, vehicle])

    const animatePriceChange = (newPrice) => {
        let startTime = null
        let currentPrice = displayedPrice
        const duration = 500
        const step = (timestamp) => {
            if (!startTime) startTime = timestamp
            const progress = timestamp - startTime
            const priceChange = Math.floor((progress / duration) * (newPrice - currentPrice)) + currentPrice
            setDisplayedPrice(priceChange)

            if (progress < duration) requestAnimationFrame(step) 
            else setDisplayedPrice(newPrice)
        }
        requestAnimationFrame(step)
    }

    function getPriceDescription() {
        if (days == 1) return '€/day'
        else return `€`
    }

    return (
        <div className='tileWrapper'>
            <div className={style.priceDesciption}>Total price:</div>

            <div className={style.priceDaysWrapper}>
                <div className={style.priceText}>
                    {displayedPrice} {getPriceDescription()}
                </div>

                <div className={formStyle.chipGroup}>
                    <div
                        className={`${days === 1 ? formStyle.selectedChip : formStyle.chip}`}
                        onClick={() => setDays(1)}
                    >
                        1
                    </div>
                    <div 
                        className={`${days === 2 ? formStyle.selectedChip : formStyle.chip}`}
                        onClick={() => setDays(2)}
                    >
                        2
                    </div>
                    <div 
                        className={`${days === 3 ? formStyle.selectedChip : formStyle.chip}`}
                        onClick={() => setDays(3)}
                    >
                        3
                    </div>
                    <div 
                        className={`${days === 4 ? formStyle.selectedChip : formStyle.chip}`}
                        onClick={() => setDays(4)}
                    >
                        4
                    </div>
                    <div 
                        className={`${days === 5 ? formStyle.selectedChip : formStyle.chip}`}
                        onClick={() => setDays(5)}
                    >
                        5
                    </div>
                </div>
            </div>
        </div>
    )
}