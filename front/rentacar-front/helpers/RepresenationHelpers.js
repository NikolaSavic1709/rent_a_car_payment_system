import { VEHICLE_CATEGORY } from "./enums"

export function getCategoryIconName(category) {
    if (category == VEHICLE_CATEGORY.SUV) return 'directions_car'
    else if (category == VEHICLE_CATEGORY.Sedan) return 'directions_car'
    else if (category == VEHICLE_CATEGORY.Hatchback) return 'directions_car'
    else if (category == VEHICLE_CATEGORY.Coupe) return 'directions_car'
    else if (category == VEHICLE_CATEGORY.Pickup) return 'fire_truck'
    else return 'directions_car'
}

export function getCategoryLabel(category) {
    if (category == VEHICLE_CATEGORY.SUV) return 'SUV'
    else if (category == VEHICLE_CATEGORY.Sedan) return 'Sedan'
    else if (category == VEHICLE_CATEGORY.Hatchback) return 'Hatchback'
    else if (category == VEHICLE_CATEGORY.Coupe) return 'Coupe'
    else if (category == VEHICLE_CATEGORY.Pickup) return 'Pickup'
    else return 'Car'
}

export function isValidEmail(email) {
    const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailPattern.test(email);
}

export function formatDate(dateString) {
    const date = new Date(dateString)
    const day = String(date.getDate()).padStart(2, '0') 
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const year = date.getFullYear()
    return `${day}.${month}.${year}`
}

export function isDateInPast(dateString) {
    const date = new Date(dateString)
    const now = new Date()
    return date < now
}