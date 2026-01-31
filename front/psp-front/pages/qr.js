import { useState, useEffect } from 'react';
import axios from 'axios';

export default function QRCodePage() {
    const [qrCode, setQrCode] = useState(null);
    const [error, setError] = useState(null);

    useEffect(() => {
        console.log(localStorage.getItem('qrRef'));
        const qrRef = localStorage.getItem('qrRef');
        if (qrRef) {
            const fetchQRCode = async () => {
                try {
                    const response = await axios.post('https://nbs.rs/QRcode/api/qr/v1/gen', {
                        "K": "PR",
                        "V": "01",
                        "C": "1",
                        "R": "845000000040484987",
                        "N": "JP EPS BEOGRAD\r\nBALKANSKA 13",
                        "I": "RSD3596,13",
                        "P": "MRĐO MAČKATOVIĆ\r\nŽUPSKA 13\r\nBEOGRAD 6",
                        "SF": "189",
                        "S": "UPLATA PO RAČUNU ZA EL. ENERGIJU",
                        "RO": qrRef
                    }, {
                        responseType: 'arraybuffer'
                    });

                    const base64Image = btoa(
                        new Uint8Array(response.data).reduce((data, byte) => data + String.fromCharCode(byte), '')
                    );

                    setQrCode(`data:image/png;base64,${base64Image}`);
                } catch (error) {
                    console.error('Error fetching QR code:', error);
                    setError('Failed to generate QR code. Please try again later.');
                }
            };

            fetchQRCode();
        } else {
            setError('QR reference is missing. Unable to generate QR code.');
        }
    }, []);

    return (
        <div className="page" style={{ textAlign: 'center', padding: '20px' }}>
            <h1>QR Code Page</h1>
            {error ? (
                <p style={{ color: 'red' }}>{error}</p>
            ) : qrCode ? (
                <>
                    <img src={qrCode} alt="QR Code" style={{ margin: '20px auto', display: 'block' }} />
                    <p>Scan me</p>
                </>
            ) : (
                <p>Loading QR Code...</p>
            )}
        </div>
    );
}