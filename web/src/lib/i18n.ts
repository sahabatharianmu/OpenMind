import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

import enCommon from '@/locales/en/common.json';
import enLanding from '@/locales/en/landing.json';
import enAuth from '@/locales/en/auth.json';
import enDashboard from '@/locales/en/dashboard.json';
import enAdmin from '@/locales/en/admin.json';

import idCommon from '@/locales/id/common.json';
import idLanding from '@/locales/id/landing.json';
import idAuth from '@/locales/id/auth.json';
import idDashboard from '@/locales/id/dashboard.json';
import idAdmin from '@/locales/id/admin.json';

const savedLang = localStorage.getItem('language') || 'en';

i18n.use(initReactI18next).init({
  resources: {
    en: {
      common: enCommon,
      landing: enLanding,
      auth: enAuth,
      dashboard: enDashboard,
      admin: enAdmin,
    },
    id: {
      common: idCommon,
      landing: idLanding,
      auth: idAuth,
      dashboard: idDashboard,
      admin: idAdmin,
    },
  },
  lng: savedLang,
  fallbackLng: 'en',
  defaultNS: 'common',
  interpolation: {
    escapeValue: false, // React already escapes
  },
});

// Persist language changes
i18n.on('languageChanged', (lng) => {
  localStorage.setItem('language', lng);
});

export default i18n;
