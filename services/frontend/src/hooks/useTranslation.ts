import locales from '../locales.json';

type Translations = typeof locales['en'];

const getT = (): Translations => {
  const browserTag = typeof window !== 'undefined' ? navigator.language : 'en';
  const primaryLang = new Intl.Locale(browserTag).language;
  return (locales as Record<string, Translations>)[primaryLang] ?? locales['en'];
};

export const useTranslation = () => {
  const browserTag = typeof window !== 'undefined' ? navigator.language : 'en';
  const langCode = new Intl.Locale(browserTag).language;
  return { t: getT(), langCode };
};

export { getT };
