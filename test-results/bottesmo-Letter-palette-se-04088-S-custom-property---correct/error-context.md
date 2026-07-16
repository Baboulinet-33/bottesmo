# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: bottesmo.spec.js >> Letter palette settings >> clicking a palette updates CSS custom property --correct
- Location: bottesmo.spec.js:652:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:3129/settings
Call log:
  - navigating to "http://localhost:3129/settings", waiting until "load"

```

# Test source

```ts
  553 |     expect(names).toContain('words');
  554 |     expect(names).toContain('words_full');
  555 |   });
  556 | });
  557 | 
  558 | test.describe('Theme toggle', () => {
  559 |   test('1. Theme toggle button exists on home page', async ({ page }) => {
  560 |     await page.goto(`/`);
  561 |     await expect(page.locator('#theme-toggle')).toBeVisible();
  562 |   });
  563 | 
  564 |   test('2. Toggle switches theme and back', async ({ page }) => {
  565 |     await page.goto(`/`);
  566 |     await page.waitForSelector('#theme-toggle');
  567 | 
  568 |     // Read current theme before toggle
  569 |     const initialTheme = await page.locator('html').getAttribute('data-theme');
  570 |     const nextTheme = initialTheme === 'light' ? 'dark' : 'light';
  571 | 
  572 |     // Toggle
  573 |     await page.locator('#theme-toggle').click();
  574 |     if (nextTheme === 'light') {
  575 |       await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
  576 |     } else {
  577 |       await expect(page.locator('html')).not.toHaveAttribute('data-theme', 'light');
  578 |     }
  579 | 
  580 |     // Toggle back
  581 |     await page.locator('#theme-toggle').click();
  582 |     if (initialTheme === 'light') {
  583 |       await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
  584 |     } else {
  585 |       await expect(page.locator('html')).not.toHaveAttribute('data-theme', 'light');
  586 |     }
  587 |   });
  588 | 
  589 |   test('3. Theme persists across page reload', async ({ page }) => {
  590 |     await page.goto(`/`);
  591 |     await page.waitForSelector('#theme-toggle');
  592 | 
  593 |     // Toggle to get non-default theme
  594 |     await page.locator('#theme-toggle').click();
  595 |     // Read the theme we toggled to
  596 |     const toggledTheme = await page.locator('html').getAttribute('data-theme');
  597 | 
  598 |     // Reload
  599 |     await page.reload();
  600 |     await page.waitForSelector('#theme-toggle');
  601 | 
  602 |     // Verify same theme is applied
  603 |     const reloadedTheme = await page.locator('html').getAttribute('data-theme');
  604 |     expect(reloadedTheme).toBe(toggledTheme);
  605 |   });
  606 | 
  607 |   test('4. Toggle exists and works on solo game page', async ({ page }) => {
  608 |     await page.goto(`/game?mode=solo`);
  609 |     await page.waitForSelector('#theme-toggle');
  610 |     await expect(page.locator('#theme-toggle')).toBeVisible();
  611 | 
  612 |     // Toggle and verify theme changed
  613 |     const initialTheme = await page.locator('html').getAttribute('data-theme');
  614 |     await page.locator('#theme-toggle').click();
  615 |     const newTheme = await page.locator('html').getAttribute('data-theme');
  616 |     expect(newTheme).not.toBe(initialTheme);
  617 |   });
  618 | 
  619 |   test('5. Default theme respects OS preference (dark)', async ({ page }) => {
  620 |     // Mock prefers-color-scheme: dark and clear localStorage
  621 |     await page.addInitScript(() => {
  622 |       Object.defineProperty(window, 'matchMedia', {
  623 |         writable: true,
  624 |         value: (query) => ({
  625 |           matches: query === '(prefers-color-scheme: dark)',
  626 |           media: query,
  627 |           onchange: null,
  628 |           addListener: () => {},
  629 |           removeListener: () => {},
  630 |           addEventListener: () => {},
  631 |           removeEventListener: () => {},
  632 |           dispatchEvent: () => {},
  633 |         }),
  634 |       });
  635 |     });
  636 |     await page.goto(`/`);
  637 |     // Without saved preference and prefers-color-scheme: dark, default should be dark (no data-theme or empty)
  638 |     const theme = await page.locator('html').getAttribute('data-theme');
  639 |     expect(theme === null || theme === '').toBe(true);
  640 |   });
  641 | });
  642 | 
  643 | test.describe('Letter palette settings', () => {
  644 |   const BASE_URL = 'http://localhost:3129';
  645 | 
  646 |   test('navigating to /settings renders 11 palette cards', async ({ page }) => {
  647 |     await page.goto(`${BASE_URL}/settings`);
  648 |     const cards = page.locator('.palette-card');
  649 |     await expect(cards).toHaveCount(11);
  650 |   });
  651 | 
  652 |   test('clicking a palette updates CSS custom property --correct', async ({ page }) => {
> 653 |     await page.goto(`${BASE_URL}/settings`);
      |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:3129/settings
  654 |     // Click the "wordle" palette card
  655 |     const wordleCard = page.locator('.palette-card[data-palette-id="wordle"]');
  656 |     await wordleCard.click();
  657 |     const correct = await page.evaluate(() =>
  658 |       getComputedStyle(document.documentElement).getPropertyValue('--correct').trim()
  659 |     );
  660 |     expect(correct).toBe('#538d4e');
  661 |   });
  662 | 
  663 |   test('selected palette persists after reload', async ({ page }) => {
  664 |     await page.goto(`${BASE_URL}/settings`);
  665 |     await page.locator('.palette-card[data-palette-id="ocean"]').click();
  666 |     await page.reload();
  667 |     const selected = page.locator('.palette-card.selected');
  668 |     await expect(selected).toHaveAttribute('data-palette-id', 'ocean');
  669 |     const correct = await page.evaluate(() =>
  670 |       getComputedStyle(document.documentElement).getPropertyValue('--correct').trim()
  671 |     );
  672 |     expect(correct).toBe('#0ea5e9');
  673 |   });
  674 | 
  675 |   test('gear icon is visible on / and /game?mode=solo and links to /settings', async ({ page }) => {
  676 |     await page.goto(`${BASE_URL}/`);
  677 |     const gearHome = page.locator('#settings-link');
  678 |     await expect(gearHome).toBeVisible();
  679 |     await expect(gearHome).toHaveAttribute('href', '/settings');
  680 | 
  681 |     await page.goto(`${BASE_URL}/game?mode=solo`);
  682 |     const gearGame = page.locator('#settings-link');
  683 |     await expect(gearGame).toBeVisible();
  684 |     await expect(gearGame).toHaveAttribute('href', '/settings');
  685 |   });
  686 | 
  687 |   test('selected palette CSS variables are applied on /game?mode=solo', async ({ page }) => {
  688 |     // Set ocean palette via settings page
  689 |     await page.goto(`${BASE_URL}/settings`);
  690 |     await page.locator('.palette-card[data-palette-id="ocean"]').click();
  691 | 
  692 |     // Navigate to solo game — palette should be applied on DOMContentLoaded
  693 |     await page.goto(`${BASE_URL}/game?mode=solo`);
  694 | 
  695 |     const correct = await page.evaluate(() =>
  696 |       getComputedStyle(document.documentElement).getPropertyValue('--correct').trim()
  697 |     );
  698 |     // Ocean --correct is #0ea5e9
  699 |     expect(correct).toBe('#0ea5e9');
  700 |   });
  701 | });
  702 | 
  703 | test.describe('Late joiner visibility on results screen', () => {
  704 |   const DEV = 'http://localhost:3131';
  705 |   test('10. Finished player sees late joiner and progress updates in rankings table', async ({ page }) => {
  706 |     // Use the dev server (port 3131) which has the fix under test.
  707 |     // Step 1: Create room, add two players, start game, have Alice finish
  708 |     const createResp = await page.request.post(`${DEV}/api/multiplayer/create`, {
  709 |       data: { mode: 'progressif', wordCount: 2, nickname: 'Alice' }
  710 |     });
  711 |     expect(createResp.ok()).toBe(true);
  712 |     const createData = await createResp.json();
  713 |     const code = createData.roomCode;
  714 |     const aliceID = createData.playerID;
  715 | 
  716 |     const bobJoinResp = await page.request.post(`${DEV}/api/multiplayer/join`, {
  717 |       data: { roomCode: code, nickname: 'Bob' }
  718 |     });
  719 |     expect(bobJoinResp.ok()).toBe(true);
  720 |     const bobData = await bobJoinResp.json();
  721 |     const bobID = bobData.playerID;
  722 | 
  723 |     const startResp = await page.request.post(`${DEV}/api/multiplayer/start`, {
  724 |       data: { roomCode: code, playerID: aliceID }
  725 |     });
  726 |     expect(startResp.ok()).toBe(true);
  727 | 
  728 |     // Alice rejoins to get her word targets
  729 |     const aliceRejoinResp = await page.request.post(`${DEV}/api/multiplayer/join`, {
  730 |       data: { roomCode: code, playerID: aliceID }
  731 |     });
  732 |     expect(aliceRejoinResp.ok()).toBe(true);
  733 |     const aliceGameData = await aliceRejoinResp.json();
  734 | 
  735 |     // Alice finishes all words
  736 |     for (let i = 0; i < 2; i++) {
  737 |       const target = aliceGameData.wordGames[i].target;
  738 |       const guessResp = await page.request.post(`${DEV}/api/multiplayer/guess`, {
  739 |         data: { roomCode: code, playerID: aliceID, word: target }
  740 |       });
  741 |       expect(guessResp.ok()).toBe(true);
  742 |     }
  743 | 
  744 |     // Step 2: Navigate to /multiplayer and set up Alice's results screen with 2 players
  745 |     await page.goto(`${DEV}/multiplayer`);
  746 | 
  747 |     // Simulate Alice's results screen with Alice (finished) and Bob (in progress)
  748 |     // Using the exact shape that the server produces (time is a Go time.Duration = nanoseconds int)
  749 |     await page.evaluate(([aID, bID]) => {
  750 |       document.querySelectorAll('.screen').forEach(s => s.style.display = 'none');
  751 |       const resultsScreen = document.getElementById('screen-results');
  752 |       if (resultsScreen) resultsScreen.style.display = '';
  753 |       if (typeof renderRankings === 'function') {
```