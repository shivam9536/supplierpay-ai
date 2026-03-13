import { createContext, useContext, useState } from 'react'
import { login as loginAPI } from '../services/api'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [token, setToken] = useState(localStorage.getItem('token'))
  const [user, setUser] = useState(
    localStorage.getItem('user_email') || null
  )

  const login = async (email, password) => {
    const res = await loginAPI(email, password)
    const { token: newToken, email: userEmail } = res.data
    localStorage.setItem('token', newToken)
    localStorage.setItem('user_email', userEmail)
    setToken(newToken)
    setUser(userEmail)
  }

  const logout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user_email')
    setToken(null)
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ token, user, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => useContext(AuthContext)
