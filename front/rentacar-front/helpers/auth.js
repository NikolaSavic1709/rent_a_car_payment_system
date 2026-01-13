import Router from "next/router"

export function putUserAccessToken(token) {
    window.localStorage.setItem('accessToken', JSON.stringify(token))
}

export function putUserRefreshToken(token) {
    window.localStorage.setItem('refreshToken', JSON.stringify(token))
}

export function getUserAccessToken() {
    if (typeof window !== 'undefined') {
        return window.localStorage.getItem('accessToken')
    }
    return null;
}

export function getUserRefreshToken() {
    return window.localStorage.getItem('refreshToken')
}

export function getUserRole() {
    const roleString = parseJwt(getUserAccessToken())['role']
    try {
        if (roleString != undefined) return roleString
    } catch { return null }
    return null
}

function parseJwt(token) {
    if (token == null) return {}
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(window.atob(base64).split('').map(function (c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
    }).join(''))
    return JSON.parse(jsonPayload)
}

export function logOut() {
    window.localStorage.clear()
}