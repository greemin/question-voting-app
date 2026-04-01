import { test, expect } from '@playwright/test';

test('Host and User interact in real-time', async ({ browser }) => {
  // Create two isolated browser contexts (simulating two different users)
  const hostContext = await browser.newContext();
  const userContext = await browser.newContext();

  const hostPage = await hostContext.newPage();
  const userPage = await userContext.newPage();

  const sessionSlug = `e2e-test-${Date.now()}`;

  // 1. Host goes to the homepage and creates a session
  await hostPage.goto('/');
  
  // 💡 NOTE: Adjust these locators if your actual UI text/placeholders differ
  await hostPage.getByRole('textbox').fill(sessionSlug);
  await hostPage.getByRole('button', { name: /create/i }).click();

  // Wait for the navigation to the new session page
  await hostPage.waitForURL(new RegExp(`.*${sessionSlug}.*`));
  const sessionUrl = hostPage.url();

  // 2. User joins the same session URL
  await userPage.goto(sessionUrl);
  
  // 3. User submits a question
  const testQuestion = 'How does integration testing work?';
  
  await userPage.getByRole('textbox').fill(testQuestion);
  await userPage.getByRole('button', { name: /ask|submit/i }).click();

  // 4. Assert question appears for the User
  await expect(userPage.getByText(testQuestion)).toBeVisible();

  // 5. Assert question appears for the Host instantly (WebSocket test)
  await expect(hostPage.getByText(testQuestion)).toBeVisible();
});