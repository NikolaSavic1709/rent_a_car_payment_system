import { useState } from "react"
import style from '../../styles/FormComponents.module.css'
import authStyle from '../../styles/Auth.module.css'
import { Button } from "@mui/material"

export default function Registration({
    width,
    onSignUpClick,
    onSignInClick
}) {
    const [showPassword, setShowPassword] = useState(false)
    const [form, setForm] = useState({
        name: '',
        email: '',
        username: '',
        password: '',
    })

    function singInClick() {
        if (onSignInClick) onSignInClick()
        setForm({
            name: '',
            email: '',
            username: '',
            password: '',
        })
    }

    function isButtonDisabled() {
        return form.username == '' || form.password == '' || form.name == '' || form.email == ''
    }

    return(
        <div style={{ width: width }}>
             <div className={style.form}>
                <div className={`${style.inputWrapper} w-full`}>
                    <input 
                        className={style.input}
                        value={form.name}
                        placeholder='Full name'
                        onChange={(e) => {setForm({...form, name: e.target.value})}}  
                    />
                </div>
            </div>
            <div className="spacer-h-s" />
            <div className={style.form}>
                <div className={`${style.inputWrapper} w-full`}>
                    <input 
                        className={style.input}
                        value={form.email}
                        placeholder='Email'
                        onChange={(e) => {setForm({...form, email: e.target.value})}}  
                    />
                </div>
            </div>
            <div className="spacer-h-s" />
            <div className={style.form}>
                <div className={`${style.inputWrapper} w-full`}>
                    <input 
                        className={style.input}
                        value={form.username}
                        placeholder='Username'
                        onChange={(e) => {setForm({...form, username: e.target.value})}}  
                    />
                </div>
            </div>
            <div className="spacer-h-s" />
            <div className={style.form}>
                <div className={`${style.inputWrapper} w-full`}>
                    <input 
                        className={style.input}
                        value={form.password}
                        type={showPassword ? "text" : "password"}
                        placeholder='Password'
                        onChange={(e) => { setForm({...form, password: e.target.value}) }}  
                    />
                    <span onClick={() => setShowPassword(prevState => !prevState)} className={`material-icons-outlined  ${style.inputIcon}`}>{showPassword ? 'visibility_off' : 'visibility'}</span> 
                </div>
            </div>
            <div className="spacer-h-s" />
            <Button 
                disableRipple 
                className={`${style.button} ${style.raisedButton} w-full`}
                onClick={() => onSignUpClick(form)}
                disabled={isButtonDisabled()}
            >
                Sign up
            </Button>
            <div className="spacer-h-s" />
            <div className={authStyle.doNotHaveWrapper}>
                <div>Do you have account?</div>
                <Button 
                    disableRipple
                    className={`${style.button} ${style.linkButton}`}
                    onClick={singInClick}
                >
                    Sign in
                </Button>
            </div>
        </div>
    )
}