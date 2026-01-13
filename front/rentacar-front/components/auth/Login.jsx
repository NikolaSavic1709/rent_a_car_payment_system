import { useState } from "react"
import style from '../../styles/FormComponents.module.css'
import authStyle from '../../styles/Auth.module.css'
import { Button } from "@mui/material"

export default function Login({ 
    width,
    onSignInClick,
    onSingleSignInClick,
    onSingUpClick 
}) {
    const [showPassword, setShowPassword] = useState(false)
    const [form, setForm] = useState({
        username: '',
        password: ''
    })

    function singUpClick() {
        if (onSingUpClick) onSingUpClick()
        setForm({
            username: '',
            password: ''
        })
    }

    function isButtonDisabled() {
        return form.username == '' || form.password == ''
    }

    return(
        <div style={{ width: width }}>
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
                onClick={() => {if(onSignInClick) onSignInClick(form)}}
                disabled={isButtonDisabled()}
            >
                Sign in
            </Button>
            <div className="spacer-h-s" />
            <Button 
                disableRipple 
                className={`${style.button} ${style.outlinedButton} w-full`}
                onClick={() => {if(onSingleSignInClick) onSingleSignInClick()}}
            >
                <img 
                    className={style.logo}
                    src='/google.svg'
                    alt="Rentacar Logo" 
                    width="16" 
                    height="16" 
                />
                <div className="spacer-h-xs"/>
                Sign in with Google
            </Button>
            <div className="spacer-h-s" />
            <div className={authStyle.doNotHaveWrapper}>
                <div>Do not have account?</div>
                <Button 
                    disableRipple
                    className={`${style.button} ${style.linkButton}`}
                    onClick={singUpClick}
                >
                    Sign up
                </Button>
            </div>
        </div>
    )
}