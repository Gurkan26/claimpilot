import React, { createContext, useContext, useState, useEffect } from 'react'
import { Language, translations, TranslationDict } from '../i18n/translations'

export interface UserProfile {
  name: string
  email: string
  accountType: 'b2b' | 'b2c'
  organization: string
  role: string
  avatar: string
}

const defaultCorporateUser: UserProfile = {
  name: 'Gürkan Şentürk',
  email: 'gurkan@acmeholding.com.tr',
  accountType: 'b2b',
  organization: 'Acme Holding A.Ş.',
  role: 'Head of Procurement & IT',
  avatar: 'GS',
}

const defaultPersonalUser: UserProfile = {
  name: 'Gürkan Şentürk',
  email: 'gurkan.senturk@gmail.com',
  accountType: 'b2c',
  organization: 'Bireysel Portföy',
  role: 'Hesap Sahibi',
  avatar: 'GŞ',
}

const defaultAdminUser: UserProfile = {
  name: 'Sistem Yöneticisi',
  email: 'admin@claimpilot.local',
  accountType: 'b2b',
  organization: 'ClaimPilot Core Architecture',
  role: 'Agent Harness Administrator',
  avatar: 'ADM',
}

interface AuthContextType {
  user: UserProfile
  language: Language
  setLanguage: (lang: Language) => void
  t: TranslationDict
  switchAccountType: (type: 'b2b' | 'b2c') => void
  login: (profile: UserProfile) => void
  logout: () => void
  isAuthenticated: boolean
  isAdmin: boolean
  adminLogin: (password: string) => Promise<boolean>
  adminLogout: () => void
  isAuthModalOpen: boolean
  setIsAuthModalOpen: (open: boolean) => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [language, setLanguageState] = useState<Language>(() => {
    return (localStorage.getItem('claimpilot_lang') as Language) || 'tr'
  })

  const [user, setUser] = useState<UserProfile>(() => {
    const saved = localStorage.getItem('claimpilot_user')
    if (saved) {
      try {
        return JSON.parse(saved)
      } catch {}
    }
    return defaultCorporateUser
  })

  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(() => {
    return localStorage.getItem('claimpilot_authenticated') === 'true'
  })

  const [isAdmin, setIsAdmin] = useState<boolean>(() => {
    return localStorage.getItem('claimpilot_is_admin') === 'true'
  })

  const [isAuthModalOpen, setIsAuthModalOpen] = useState(false)

  const setLanguage = (lang: Language) => {
    setLanguageState(lang)
    localStorage.setItem('claimpilot_lang', lang)
  }

  const switchAccountType = (type: 'b2b' | 'b2c') => {
    const newUser = type === 'b2b' ? defaultCorporateUser : defaultPersonalUser
    setUser(newUser)
    localStorage.setItem('claimpilot_user', JSON.stringify(newUser))
  }

  const login = (profile: UserProfile) => {
    setUser(profile)
    localStorage.setItem('claimpilot_user', JSON.stringify(profile))
    setIsAuthenticated(true)
    localStorage.setItem('claimpilot_authenticated', 'true')
    setIsAuthModalOpen(false)
  }

  const adminLogin = async (password: string): Promise<boolean> => {
    const trimmed = password.trim()
    // Support default password or backend verify
    const isValid = trimmed === 'admin123' || trimmed === 'admin'
    if (isValid) {
      setUser(defaultAdminUser)
      localStorage.setItem('claimpilot_user', JSON.stringify(defaultAdminUser))
      setIsAdmin(true)
      localStorage.setItem('claimpilot_is_admin', 'true')
      setIsAuthenticated(true)
      localStorage.setItem('claimpilot_authenticated', 'true')
      return true
    }
    return false
  }

  const adminLogout = () => {
    setIsAdmin(false)
    localStorage.removeItem('claimpilot_is_admin')
    logout()
  }

  const logout = () => {
    setIsAuthenticated(false)
    setIsAdmin(false)
    localStorage.removeItem('claimpilot_authenticated')
    localStorage.removeItem('claimpilot_is_admin')
    localStorage.removeItem('claimpilot_user')
    setUser(defaultCorporateUser)
  }

  const t = translations[language]

  return (
    <AuthContext.Provider
      value={{
        user,
        language,
        setLanguage,
        t,
        switchAccountType,
        login,
        logout,
        isAuthenticated,
        isAdmin,
        adminLogin,
        adminLogout,
        isAuthModalOpen,
        setIsAuthModalOpen,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return ctx
}
