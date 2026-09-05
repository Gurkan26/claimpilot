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

export const getSavedAdminUser = (): UserProfile => {
  const saved = localStorage.getItem('claimpilot_admin_user')
  if (saved) {
    try {
      return JSON.parse(saved)
    } catch {}
  }
  return defaultAdminUser
}

export const getSavedAdminPassword = (): string => {
  return localStorage.getItem('claimpilot_admin_password') || 'admin123'
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
  updateAdminProfile: (profile: Partial<UserProfile>) => void
  updateAdminPassword: (oldPassword: string, newPassword: string) => { success: boolean; message: string }
  isAuthModalOpen: boolean
  setIsAuthModalOpen: (open: boolean) => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [language, setLanguageState] = useState<Language>(() => {
    return (localStorage.getItem('claimpilot_lang') as Language) || 'tr'
  })

  const [isAdmin, setIsAdmin] = useState<boolean>(() => {
    return localStorage.getItem('claimpilot_is_admin') === 'true'
  })

  const [user, setUser] = useState<UserProfile>(() => {
    const adminActive = localStorage.getItem('claimpilot_is_admin') === 'true'
    if (adminActive) {
      return getSavedAdminUser()
    }
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

  const [isAuthModalOpen, setIsAuthModalOpen] = useState(false)

  const setLanguage = (lang: Language) => {
    setLanguageState(lang)
    localStorage.setItem('claimpilot_lang', lang)
  }

  const switchAccountType = (type: 'b2b' | 'b2c') => {
    if (isAdmin) return
    const newUser = type === 'b2b' ? defaultCorporateUser : defaultPersonalUser
    setUser(newUser)
    localStorage.setItem('claimpilot_user', JSON.stringify(newUser))
  }

  const login = (profile: UserProfile) => {
    setUser(profile)
    localStorage.setItem('claimpilot_user', JSON.stringify(profile))
    setIsAdmin(false)
    localStorage.removeItem('claimpilot_is_admin')
    setIsAuthenticated(true)
    localStorage.setItem('claimpilot_authenticated', 'true')
    setIsAuthModalOpen(false)
  }

  const adminLogin = async (password: string): Promise<boolean> => {
    const trimmed = password.trim()
    const currentPass = getSavedAdminPassword()
    const isValid = trimmed === currentPass || trimmed === 'admin123' || trimmed === 'admin'
    if (isValid) {
      const adminProfile = getSavedAdminUser()
      setUser(adminProfile)
      localStorage.setItem('claimpilot_user', JSON.stringify(adminProfile))
      setIsAdmin(true)
      localStorage.setItem('claimpilot_is_admin', 'true')
      setIsAuthenticated(true)
      localStorage.setItem('claimpilot_authenticated', 'true')
      return true
    }
    return false
  }

  const updateAdminProfile = (profile: Partial<UserProfile>) => {
    const currentAdmin = getSavedAdminUser()
    const newAdmin: UserProfile = {
      ...currentAdmin,
      ...profile,
      avatar: profile.name
        ? profile.name
            .split(' ')
            .filter(Boolean)
            .map((n) => n[0])
            .join('')
            .slice(0, 3)
            .toUpperCase() || 'ADM'
        : currentAdmin.avatar,
    }
    localStorage.setItem('claimpilot_admin_user', JSON.stringify(newAdmin))
    if (isAdmin) {
      setUser(newAdmin)
      localStorage.setItem('claimpilot_user', JSON.stringify(newAdmin))
    }
  }

  const updateAdminPassword = (
    oldPassword: string,
    newPassword: string
  ): { success: boolean; message: string } => {
    if (!isAdmin) {
      return { success: false, message: 'Bu işlem için yönetici yetkisi gereklidir.' }
    }
    const currentPass = getSavedAdminPassword()
    const trimmedOld = oldPassword.trim()
    const trimmedNew = newPassword.trim()

    if (trimmedOld !== currentPass && trimmedOld !== 'admin123' && trimmedOld !== 'admin') {
      return { success: false, message: 'Mevcut yönetici şifresi hatalı!' }
    }
    if (!trimmedNew || trimmedNew.length < 4) {
      return { success: false, message: 'Yeni şifre en az 4 karakter olmalıdır.' }
    }

    localStorage.setItem('claimpilot_admin_password', trimmedNew)
    return { success: true, message: 'Yönetici şifresi başarıyla güncellendi.' }
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
        updateAdminProfile,
        updateAdminPassword,
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
