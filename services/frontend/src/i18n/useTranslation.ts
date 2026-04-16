import enTranslations from './en.json';

type Translations = typeof enTranslations;

const SUPPORTED_LANGS = ['en', 'de', 'pl', 'ru', 'es', 'it', 'hu'] as const;

let translations: Translations = enTranslations;
let langCode = 'en';

export const loadTranslations = async (): Promise<void> => {
  const browserTag = typeof window !== 'undefined' ? navigator.language : 'en';
  const primary = new Intl.Locale(browserTag).language;
  langCode = SUPPORTED_LANGS.includes(primary as (typeof SUPPORTED_LANGS)[number]) ? primary : 'en';

  if (langCode === 'en') return;

  const module = await import(`./${langCode}.json`);
  translations = module.default;
};

export const useTranslation = () => ({ t: translations, langCode });

export const getT = (): Translations => translations;
