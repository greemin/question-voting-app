import { test, expect, Page } from '@playwright/test';

// Injects a script before page load that tracks all WebSocket instances on window._wsInstances.
async function trackWebSockets(page: Page) {
  await page.addInitScript(() => {
    (window as any)._wsInstances = [];
    const OrigWS = window.WebSocket;
    function TrackedWS(this: WebSocket, url: string, protocols?: string | string[]) {
      const ws = new OrigWS(url, protocols);
      (window as any)._wsInstances.push(ws);
      return ws;
    }
    TrackedWS.prototype = OrigWS.prototype;
    TrackedWS.CONNECTING = OrigWS.CONNECTING;
    TrackedWS.OPEN = OrigWS.OPEN;
    TrackedWS.CLOSING = OrigWS.CLOSING;
    TrackedWS.CLOSED = OrigWS.CLOSED;
    (window as any).WebSocket = TrackedWS;
  });
}

async function closeAllWebSockets(page: Page) {
  await page.evaluate(() => {
    ((window as any)._wsInstances as WebSocket[]).forEach(ws => ws.close());
  });
}

test('document title reflects APP_NAME on home page and session page', async ({ page }) => {
  const appName = process.env.APP_NAME || 'Question Voting App';
  const sessionSlug = `e2e-title-${Date.now()}`;

  await page.goto('/');
  expect(await page.title()).toBe(appName);

  await page.getByRole('textbox').fill(sessionSlug);
  await page.getByRole('button', { name: /new voting session/i }).click();
  await page.waitForURL(new RegExp(`.*${sessionSlug}.*`));

  await expect(page).toHaveTitle(`${appName} - ${sessionSlug}`);

  // Cleanup
  page.once('dialog', dialog => dialog.accept());
  await page.getByRole('button', { name: /end session/i }).click();
});

test('Host and User interact in real-time', async ({ browser }) => {
  // Create two isolated browser contexts (simulating two different users)
  const hostContext = await browser.newContext();
  const userContext = await browser.newContext();

  const hostPage = await hostContext.newPage();
  const userPage = await userContext.newPage();

  const sessionSlug = `e2e-test-${Date.now()}`;

  try {
    // 1. Host goes to the homepage and creates a session
    await hostPage.goto('/');
    
    // 💡 NOTE: Adjust these locators if your actual UI text/placeholders differ
    await hostPage.getByRole('textbox').fill(sessionSlug);
    await hostPage.getByRole('button', { name: /new voting session/i }).click();

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
  } finally {
    // 6. Teardown: Clean up the session so we don't litter the database.
    try {
      // Accept the window.confirm dialog when it appears
      hostPage.once('dialog', dialog => dialog.accept());
      await hostPage.getByRole('button', { name: /end session/i }).click();
    } catch (e) {
      console.log('Cleanup: Could not click End Session (test may have failed early).');
    }
  }
});

test('live updates resume after WebSocket reconnection', async ({ browser }) => {
  const hostContext = await browser.newContext();
  const userContext = await browser.newContext();

  const hostPage = await hostContext.newPage();
  const userPage = await userContext.newPage();

  // Track WS instances so we can close them to simulate an nginx timeout.
  await trackWebSockets(hostPage);

  const sessionSlug = `e2e-reconnect-${Date.now()}`;

  try {
    // 1. Host creates session
    await hostPage.goto('/');
    await hostPage.getByRole('textbox').fill(sessionSlug);
    await hostPage.getByRole('button', { name: /new voting session/i }).click();
    await hostPage.waitForURL(new RegExp(`.*${sessionSlug}.*`));
    const sessionUrl = hostPage.url();

    // 2. User joins
    await userPage.goto(sessionUrl);

    // 3. Baseline: first question appears on host via WS
    const firstQuestion = 'Question before reconnect';
    await userPage.getByRole('textbox').fill(firstQuestion);
    await userPage.getByRole('button', { name: /ask|submit/i }).click();
    await expect(hostPage.getByText(firstQuestion)).toBeVisible();

    // 4. Simulate nginx idle timeout — close all WS connections on the host page
    await closeAllWebSockets(hostPage);

    // 5. Wait for reconnect (3s timer + buffer)
    await hostPage.waitForTimeout(4000);

    // 6. Second question must appear on host in real-time via the reconnected WS
    const secondQuestion = 'Question after reconnect';
    await userPage.getByRole('textbox').fill(secondQuestion);
    await userPage.getByRole('button', { name: /ask|submit/i }).click();
    await expect(hostPage.getByText(secondQuestion)).toBeVisible();
  } finally {
    try {
      hostPage.once('dialog', dialog => dialog.accept());
      await hostPage.getByRole('button', { name: /end session/i }).click();
    } catch (e) {
      console.log('Cleanup: Could not click End Session.');
    }
  }
});