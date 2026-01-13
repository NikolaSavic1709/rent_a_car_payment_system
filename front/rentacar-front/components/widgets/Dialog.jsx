import { useEffect, useRef, useState } from "react"
import style from '../../styles/Dialog.module.css'

export function Dialog({ 
    isOpen=false, 
    width=300,
    onCloseModal, 
    children 
}) {
    const ref = useRef()
    useEffect(() => {
        if (isOpen) ref.current?.showModal()
        else {
            setTimeout(() => {
                ref.current?.close();
            }, 300);
        }
    }, [isOpen])

    return(
        <dialog
            className={`${style.dialog} ${isOpen ? style.show : style.hide}`}
            style={{ width: `${width}px` }}
            ref={ref}
            onCancel={onCloseModal}
        >
            {children}
        </dialog>
    )
}