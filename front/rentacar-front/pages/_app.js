import "@/styles/globals.css"
import Layout from "@/components/widgets/Layout"
import axios from "axios"
import { BACK_BASE_URL } from "@/values/Enviroment"
import { getUserAccessToken } from "@/helpers/auth"

axios.interceptors.request.use(
  (config) => {
    if (config.url.includes('/login') || config.url.includes('/register')) {
      console.log('login or register')
      return config
    }
    const token = getUserAccessToken()
    if (token && !config.headers['skip']) {
      config.headers['Authorization'] = `Bearer ${token.replace(/"/g, '')}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error);
  }
)


// axios.interceptors.response.use(
//   response => { return response },
//   async function (error) {
//       const originalRequest = error.config;
//       if (getUserAccessToken() == null) return
//       if (error.response && error.response.status === 401 && !originalRequest._retry) {
//           originalRequest._retry = true
//           const tokenEndpoint = `${BACK_BASE_URL}/api/user/refreshToken`
//           const accessToken = getUserAccessToken()
//           const refreshToken = getUserRefreshToken()

//       return axios.post(tokenEndpoint, {
//           'accessToken': accessToken,
//           'refreshToken': refreshToken,
//       }, {
//           headers: {
//               'Content-Type': 'application/json',
//               'skip': true
//           },
//       }).then(response => {
//           if (response.status === 200) {
//               localStorage.setItem('accessToken', response.data['accessToken'])
//               localStorage.setItem('refreshToken', response.data['refreshToken'])
//               return axios(originalRequest)
//           } else {
//               logOut();
//           }
//       }).catch(_err => logOut())
//     }
//     return Promise.reject(error)
//   }
// )

export default function App({ Component, pageProps }) {
  return (
      <Layout>
          <Component {...pageProps} />
      </Layout>
  )
}
