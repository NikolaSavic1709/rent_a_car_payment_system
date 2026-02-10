import Head from 'next/head'
import { useState } from 'react'
import axios from 'axios'
import { useRouter } from 'next/router'
import { BACK_BASE_URL } from '@/values/Enviroment'

export default function LoginPage() {
  const router = useRouter()
  const [form, setForm] = useState({ username: '', password: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function handleSubmit(e) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const resp = await axios.post(`${BACK_BASE_URL}/login`, form)
      if (resp.status === 200) {
        console.log('Login successful')
        console.log('Response data:', resp.data)
        // If backend returns a token, store it and redirect
        // if (resp.data ) {
          try { localStorage.setItem('token', 'token') } catch (e) {}
        // }
        router.push('/')
      } else {
        setError('Login failed')
      }
    } catch (err) {
      setError(err?.response?.data?.message || 'Server error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <Head>
        <title>Login</title>
      </Head>
      <main style={{display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh'}}>
        <form onSubmit={handleSubmit} style={{width: 320}}>
          <h2>Login</h2>
          <div style={{marginBottom: 8}}>
            <input
              placeholder="Username"
              value={form.username}
              onChange={(e) => setForm({...form, username: e.target.value})}
              style={{width: '100%', padding: 8}}
            />
          </div>
          <div style={{marginBottom: 8}}>
            <input
              placeholder="Password"
              type="password"
              value={form.password}
              onChange={(e) => setForm({...form, password: e.target.value})}
              style={{width: '100%', padding: 8}}
            />
          </div>
          {error && <div style={{color: 'red', marginBottom: 8}}>{error}</div>}
          <button type="submit" disabled={loading} style={{width: '100%', padding: 10}}>
            {loading ? 'Signing in...' : 'Sign in'}
          </button>
        </form>
      </main>
    </>
  )
}
