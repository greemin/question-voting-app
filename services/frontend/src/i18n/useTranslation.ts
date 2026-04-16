import enTranslations from './en.json';

type Translations = typeof enTranslations;

const loaders: Record<string, () => Promise<{ default: Translations }>> = {
  de: () => import('./de.json'),
  pl: () => import('./pl.json'),
  ru: () => import('./ru.json'),
  es: () => import('./es.json'),
  it: () => import('./it.json'),
  hu: () => import('./hu.json'),
};

let translations: Translations = enTranslations;
let langCode = 'en';

export const loadTranslations = async (): Promise<void> => {
  const browserTag = typeof window !== 'undefined' ? navigator.language : 'en';
  const primary = new Intl.Locale(browserTag).language;
  langCode = primary in loaders ? primary : 'en';

  if (langCode === 'en') return;

  const module = await loaders[langCode]();
  translations = module.default;
};

export const useTranslation = () => ({ t: translations, langCode });

export const getT = (): Translations => translations;
