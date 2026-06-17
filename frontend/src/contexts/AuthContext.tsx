import React, { createContext, useContext, useState, useEffect } from 'react'
import { api, setAccessToken } from '../lib/api'

interface User {
  id: string
  email: string
}

interface AuthContextType {
  user: User | null
  loading: boolean
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string) => Promise<void>
  logout: (global?: boolean) => Promise<void>
  refreshSession: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | null>(null)

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  const fetchCurrentUser = async () => {
    try {
      const res = await api.get<User>('/auth/me')
      setUser(res.data)
    } catch (err) {
      setUser(null)
    }
  }

  const refreshSession = async () => {
    try {
      const res = await api.post<{ access_token: string; expires_in: number }>('/auth/refresh')
      setAccessToken(res.data.access_token)
      await fetchCurrentUser()
    } catch (err) {
      setAccessToken(null)
      setUser(null)
    }
  }

  useEffect(() => {
    const initAuth = async () => {
      await refreshSession()
      setLoading(false)
    }
    initAuth()
  }, [])

  useEffect(() => {
    if (!user) return

    // Refresh every 8 minutes (JWT expires in 10 minutes)
    const interval = setInterval(() => {
      refreshSession()
    }, 8 * 60 * 1000)

    return () => clearInterval(interval)
  }, [user])

  const login = async (email: string, password: string) => {
    const res = await api.post<{ access_token: string; expires_in: number }>('/auth/login', {
      email,
      password,
    })
    setAccessToken(res.data.access_token)
    await fetchCurrentUser()
  }

  const register = async (email: string, password: string) => {
    await api.post('/auth/register', { email, password })
    await login(email, password)
  }

  const logout = async (global = false) => {
    try {
      await api.delete(`/auth/logout${global ? '?global=true' : ''}`)
    } catch (err) {
      console.error('Logout error:', err)
    } finally {
      setAccessToken(null)
      setUser(null)
    }
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        loading,
        isAuthenticated: !!user,
        login,
        register,
        logout,
        refreshSession,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
