import style from '../../styles/Layout.module.css'
import Navbar from './Navbar'

export default function Layout({ children }) {
    return (
        <div>
            <div className={style.navbar}> <Navbar/> </div>
            <div className={style.content}> {children} </div>
        </div>
    )
}