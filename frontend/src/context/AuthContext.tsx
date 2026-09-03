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

interface AuthContextType {
  user: UserProfile
  language: Language
  setLanguage: (lang: Language) => void
  t: TranslationDict
  switchAccountType: (type: 'b2b' | 'b2c') => void
  login: (profile: UserProfile) => void
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
    setIsAuthModalOpen(false)
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
