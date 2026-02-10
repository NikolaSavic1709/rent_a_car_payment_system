import Head from "next/head"
import { useEffect, useState } from "react"
import axios from 'axios'
import { useRouter } from 'next/router'
import { BACK_BASE_URL } from '@/values/Enviroment'

export default function Home() {
  const router = useRouter()
  const [merchants, setMerchants] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null
    if (!token) {
      router.push('/login')
      return
    }

    async function fetchMerchants() {
      try {
        const resp = await axios.get(`${BACK_BASE_URL}/merchants`, {
          headers: { Authorization: `Bearer ${token}` }
        })
        if (resp.status === 200 && resp.data?.merchants) {
          // resp.data.merchants is expected to be an array of objects {username, merchantId}
          setMerchants(resp.data.merchants)
        } else {
          setError('Failed to load merchants')
        }
      } catch (err) {
        if (err?.response?.status === 401) {
          try { localStorage.removeItem('token') } catch (e) {}
          router.push('/login')
        } else {
          setError('Server error')
        }
      } finally {
        setLoading(false)
      }
    }

    fetchMerchants()
  }, [])

  function signOut() {
    try { localStorage.removeItem('token') } catch (e) {}
    router.push('/login')
  }

  return (
    <>
      <Head>
        <title>PayLink</title>
      </Head>
      <main style={{padding: 24}}>
        <div style={{display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16}}>
          <h1>Merchants</h1>
          <button onClick={signOut}>Sign out</button>
        </div>

        {loading && <div>Loading...</div>}
        {error && <div style={{color: 'red'}}>{error}</div>}

        <div style={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: 12}}>
          {merchants.map((m, i) => (
            <div
              key={m.merchantId ?? i}
              style={{border: '1px solid #ddd', padding: 12, borderRadius: 6, cursor: 'pointer'}}
              onClick={() => router.push(`/subscription?merchantId=${m.merchantId}`)}
            >
              <div style={{fontWeight: 600}}>Username</div>
              <div>{m.username}</div>
            </div>
          ))}
        </div>
      </main>
    </>
  )
}
