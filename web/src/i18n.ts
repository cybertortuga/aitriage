import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import enTranslation from './locales/en/translation.json';
import ruTranslation from './locales/ru/translation.json';
import enPages from './locales/en/pages.json';
import ruPages from './locales/ru/pages.json';
import enComponents from './locales/en/components.json';
import ruComponents from './locales/ru/components.json';

const resources = {
  en: {
    translation: { ...enTranslation, components: enComponents },
    pages: enPages,
    components: enComponents,
  },
  ru: {
    translation: { ...ruTranslation, components: ruComponents },
    pages: ruPages,
    components: ruComponents,
  },
};

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false, // React already safes from xss
    },
  });

export default i18n;
