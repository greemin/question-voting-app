import { test, expect } from '@playwright/test';

/**
 * Each entry drives one test. The locale is passed directly to a fresh
 * browser context so navigator.language reflects it — the same signal
 * that useTranslation() / getT() reads at runtime.
 *
 * Regional tags (de-AT, es-419) are intentional: they prove that
 * Intl.Locale.language correctly strips the region subtag and falls
 * back to the base dictionary (de, es).
 */
const languages = [
  {
    locale: 'en-US',
    appTitle: 'Question Voting App',
    buttonText: '🚀 Start New Voting Session',
    placeholder: 'Session title (optional)',
    tagline: 'Real-time Q&A for your event',
  },
  {
    locale: 'de-AT', // Austrian German → falls back to 'de' dictionary
    appTitle: 'Frage-Abstimmungs-App',
    buttonText: '🚀 Neue Abstimmungssitzung starten',
    placeholder: 'Sitzungstitel (optional)',
    tagline: 'Echtzeit-Q&A für Ihre Veranstaltung',
  },
  {
    locale: 'pl-PL',
    appTitle: 'Aplikacja do głosowania na pytania',
    buttonText: '🚀 Rozpocznij nową sesję głosowania',
    placeholder: 'Tytuł sesji (opcjonalnie)',
    tagline: 'Q&A w czasie rzeczywistym dla Twojego wydarzenia',
  },
  {
    locale: 'ru-RU',
    appTitle: 'Приложение для голосования за вопросы',
    buttonText: '🚀 Начать новую сессию голосования',
    placeholder: 'Название сессии (необязательно)',
    tagline: 'Q&A в реальном времени для вашего мероприятия',
  },
  {
    locale: 'es-419', // Latin-American Spanish → falls back to 'es' dictionary
    appTitle: 'Aplicación de votación de preguntas',
    buttonText: '🚀 Iniciar nueva sesión de votación',
    placeholder: 'Título de la sesión (opcional)',
    tagline: 'Q&A en tiempo real para tu evento',
  },
  {
    locale: 'it-IT',
    appTitle: 'App di votazione domande',
    buttonText: '🚀 Avvia nuova sessione di votazione',
    placeholder: 'Titolo della sessione (opzionale)',
    tagline: 'Q&A in tempo reale per il tuo evento',
  },
  {
    locale: 'hu-HU',
    appTitle: 'Kérdésszavazó alkalmazás',
    buttonText: '🚀 Új szavazási munkamenet indítása',
    placeholder: 'Munkamenet neve (opcionális)',
    tagline: 'Valós idejű Q&A az eseményéhez',
  },
];

for (const { locale, appTitle, buttonText, placeholder, tagline } of languages) {
  test(`locale ${locale} — homepage renders in correct language`, async ({ browser }) => {
    // Override the browser locale for this context only.
    // navigator.language inside the page will equal the locale string.
    const context = await browser.newContext({ locale });
    const page = await context.newPage();

    await page.goto('/');

    // Heading
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(appTitle);

    // Tagline paragraph
    await expect(page.getByText(tagline, { exact: true })).toBeVisible();

    // Input placeholder (attribute check — not visible text)
    await expect(page.getByRole('textbox')).toHaveAttribute('placeholder', placeholder);

    // CTA button
    await expect(page.getByRole('button', { name: buttonText, exact: true })).toBeVisible();

    await context.close();
  });
}
