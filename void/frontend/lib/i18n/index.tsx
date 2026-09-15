'use client';

import React, { createContext, useContext, useEffect, useMemo, useState } from 'react';
import en from './en.json';
import fa from './fa.json';
import zh from './zh.json';

export type Locale = 'en' | 'fa' | 'zh';

const dictionaries: Record<Locale, any> = { en, fa, zh };

export const localeDirection: Record<Locale, 'ltr' | 'rtl'> = {
  en: 'ltr',
  fa: 'rtl',
  zh: 'ltr',
};

export const localeFontClass: Record<Locale, string> = {
  en: 'font-sans',
  fa: 'font-fa',
  zh: 'font-zh',
};

interface I18nContextValue {
  locale: Locale;
  dir: 'ltr' | 'rtl';
  setLocale: (l: Locale) => void;
  t: (path: string) => any;
}

const I18nContext = createContext<I18nContextValue | null>(null);

function resolve(dict: any, path: string): any {
  return path.split('.').reduce((acc, key) => (acc && acc[key] !== undefined ? acc[key] : undefined), dict);
}

const STORAGE_KEY = 'void.locale';

export function I18nProvider({ children }: { children: React.ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>('en');

  useEffect(() => {
    const saved = typeof window !== 'undefined' ? (window.localStorage.getItem(STORAGE_KEY) as Locale | null) : null;
    if (saved && dictionaries[saved]) {
      setLocaleState(saved);
    }
  }, []);

  useEffect(() => {
    const dir = localeDirection[locale];
    document.documentElement.setAttribute('lang', locale);
    document.documentElement.setAttribute('dir', dir);
    document.body.classList.remove('font-sans', 'font-fa', 'font-zh');
    document.body.classList.add(localeFontClass[locale]);
    window.localStorage.setItem(STORAGE_KEY, locale);
  }, [locale]);

  const setLocale = (l: Locale) => setLocaleState(l);

  const t = useMemo(() => {
    const dict = dictionaries[locale];
    return (path: string) => {
      const value = resolve(dict, path);
      return value === undefined ? path : value;
    };
  }, [locale]);

  const value: I18nContextValue = { locale, dir: localeDirection[locale], setLocale, t };

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n() {
  const ctx = useContext(I18nContext);
  if (!ctx) throw new Error('useI18n must be used within I18nProvider');
  return ctx;
}
