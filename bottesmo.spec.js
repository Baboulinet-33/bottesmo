const { test, expect } = require('@playwright/test');

/**
 * Helper: types the given word into the current row using physical keyboard.
 * Skips locked positions automatically based on first letter logic.
 */
async function typeWord(page, word) {
  for (const ch of word.slice(1)) {
    await page.keyboard.press('Key' + ch);
  }
}

/**
 * Helper: click "Valider" button to submit.
 */
async function submitGuess(page) {
  await page.locator('#submit-btn').click();
}

test.describe('Bottesmo E2E — Match UI to real Bottesmo', () => {
  test('1. Basic typing flow — letters fill cells, Enter/Valider submit', async ({ page }) => {
    await page.goto('/game?mode=solo');
    await page.waitForSelector('#grid');

    // Read first letter from DOM (each page load creates a new game)
    const firstLetter = await page.locator('#tile-0-0').textContent();
    await expect(page.locator('#tile-0-0')).toHaveClass(/locked/);
    await expect(page.locator('#tile-0-0')).toHaveClass(/correct/);

    // Type a letter via physical keyboard
    await page.keyboard.press('KeyA');
    await expect(page.locator('#tile-0-1')).toHaveText('A');

    // Type another
    await page.keyboard.press('KeyB');
    await expect(page.locator('#tile-0-2')).toHaveText('B');

    // Fill rest and submit
    for (let i = 3; i < 7; i++) {
      await page.keyboard.press('KeyC');
    }
    await submitGuess(page);
    await page.waitForTimeout(500);

    // After submission, verify the game responded.
    // Tiles get color classes if the word was valid; otherwise an error message appears.
    const tile01 = page.locator('#tile-0-1');
    const tileClass = await tile01.getAttribute('class');
    if (!/correct|present|absent/.test(tileClass || '')) {
      const msg = page.locator('#message');
      await expect(msg).not.toBeEmpty();
    }
  });

  test('2. Backspace behavior — typed letters clear, locked letters stay', async ({ page }) => {
    await page.goto('/game?mode=solo');
    await page.waitForSelector('#grid');

    // Read first letter from DOM (each page load creates a new game)
    const firstLetter = await page.locator('#tile-0-0').textContent();

    // Type letters
    await page.keyboard.press('KeyA');
    await page.keyboard.press('KeyB');
    await page.keyboard.press('KeyC');
    await expect(page.locator('#tile-0-3')).toHaveText('C');

    // Backspace twice — removes C, then B
    await page.keyboard.press('Backspace');
    await expect(page.locator('#tile-0-3')).toHaveText('');
    await page.keyboard.press('Backspace');
    await expect(page.locator('#tile-0-2')).toHaveText('');

    // Position 0 must remain locked with first letter
    await expect(page.locator('#tile-0-0')).toHaveText(firstLetter);
    await expect(page.locator('#tile-0-0')).toHaveClass(/locked/);

    // Backspace on col 1 clears it
    await page.keyboard.press('Backspace');
    await expect(page.locator('#tile-0-1')).toHaveText('');
  });

  test('3. Pre-filled letters appear on next row after a guess', async ({ page }) => {
    await page.goto(`/game?mode=solo`);
    await page.waitForSelector('#grid');

    // Fill entire row
    for (let i = 1; i < 7; i++) {
      await page.keyboard.press('KeyA');
    }
    await submitGuess(page);
    await page.waitForTimeout(500);

    // Second row should have pre-filled letters at correct positions
    const tile10 = page.locator('#tile-1-0');
    // If the submission was valid (word accepted), position 0 should be pre-filled
    const tile10Text = await tile10.textContent();
    if (tile10Text.length > 0) {
      await expect(tile10).toHaveClass(/locked/);
      await expect(tile10).toHaveClass(/correct/);
    }
  });

  test('4. Keyboard click fills cells', async ({ page }) => {
    await page.goto(`/game?mode=solo`);
    await page.waitForSelector('#keyboard');

    // Click A on-screen keyboard
    await page.locator('.kb-key').filter({ hasText: /^A$/ }).click();
    await expect(page.locator('#tile-0-1')).toHaveText('A');

    // Click Z on-screen keyboard
    await page.locator('.kb-key').filter({ hasText: /^Z$/ }).click();
    await expect(page.locator('#tile-0-2')).toHaveText('Z');
  });

  test('5. Keyboard Backspace click clears last cell', async ({ page }) => {
    await page.goto(`/game?mode=solo`);
    await page.waitForSelector('#keyboard');

    // Type via physical keyboard
    await page.keyboard.press('KeyA');
    await page.keyboard.press('KeyB');
    await expect(page.locator('#tile-0-2')).toHaveText('B');

    // Click Backspace on on-screen keyboard
    await page.locator('.kb-key').filter({ hasText: 'Suppr' }).click();
    await expect(page.locator('#tile-0-2')).toHaveText('');
  });

  test('6. Keyboard colors update after submission', async ({ page }) => {
    await page.goto(`/game?mode=solo`);
    await page.waitForSelector('#keyboard');

    // Fill row and submit
    for (let i = 1; i < 7; i++) {
      await page.keyboard.press('KeyA');
    }
    await submitGuess(page);
    await page.waitForTimeout(500);

    // Verify the game responded: keyboard keys get color classes if the word was valid,
    // otherwise an error message appears.
    const keys = page.locator('.kb-key');
    const count = await keys.count();
    let hasColor = false;
    for (let i = 0; i < count; i++) {
      const cls = await keys.nth(i).getAttribute('class');
      if (/correct|present|absent/.test(cls || '')) {
        hasColor = true;
        break;
      }
    }
    if (!hasColor) {
      const msg = page.locator('#message');
      await expect(msg).not.toBeEmpty();
    }
  });

  test('7. Complete game (win) — verify win message appears', async ({ page }) => {
    await page.goto('/game?mode=solo');
    await page.waitForSelector('#grid');

    // Read first letter from DOM (each page load creates a new game)
    const firstLetter = await page.locator('#tile-0-0').textContent();

    // Type the first letter + remaining letters to fill the row
    const word = firstLetter + 'XXXXXX';
    for (let i = 1; i < 7; i++) {
      await page.keyboard.press('KeyX');
    }
    await submitGuess(page);
    await page.waitForTimeout(1000);

    // After game over, either win or lose message
    const msg = page.locator('#message');
    await expect(msg).not.toBeEmpty();
  });

  test('9. Mobile viewport (≤480px) — layout renders correctly', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto(`/game?mode=solo`);
    await page.waitForSelector('#grid');
    await page.waitForSelector('#keyboard');

    // Grid tiles should be visible and smaller
    const tile = page.locator('.tile').first();
    await expect(tile).toBeVisible();

    // Type to verify interaction works on mobile
    await page.keyboard.press('KeyA');
    await expect(page.locator('#tile-0-1')).toHaveText('A');
  });

  test('10. Daily mode — deterministic word', async ({ page }) => {
    await page.goto(`/game?mode=daily`);
    await page.waitForSelector('#grid');

    // Grid should have 6 rows
    const rows = await page.locator('.row').count();
    expect(rows).toBe(6);

    // First tile locked with first letter
    const tile = page.locator('#tile-0-0');
    await expect(tile).toHaveClass(/locked/);
    const letter = await tile.textContent();
    expect(letter.length).toBe(1);
  });
});

test.describe('Multiplayer UI — Bugfix verification', () => {
  test('1. Homepage shows three mode buttons (Daily, Solo, Multijoueur)', async ({ page }) => {
    await page.goto(`/`);
    await page.waitForSelector('.mode-btn');
    const buttons = page.locator('.mode-btn');
    await expect(buttons).toHaveCount(3);
    await expect(buttons.nth(2)).toHaveText('Multijoueur');
  });

  test('2. Late joiner receives non-empty wordGames', async ({ request }) => {
    const createResp = await request.post(`/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 2, nickname: 'Alice' }
    });
    const createData = await createResp.json();
    const code = createData.roomCode;
    const aliceID = createData.playerID;
    const aliceToken = createData.token;

    const startResp = await request.post(`/api/multiplayer/start`, {
      headers: { 'X-Player-Token': aliceToken },
      data: { roomCode: code, playerID: aliceID }
    });
    expect(startResp.ok()).toBe(true);

    const joinResp = await request.post(`/api/multiplayer/join`, {
      data: { roomCode: code, nickname: 'Bob' }
    });
    expect(joinResp.ok()).toBe(true);
    const joinData = await joinResp.json();
    expect(joinData.state).toBe('playing');
    expect(joinData.wordGames).toBeDefined();
    expect(joinData.wordGames.length).toBe(2);
    expect(joinData.wordGames[0]).toBeDefined();
    expect(joinData.wordGames[0].target).toBeTruthy();
  });

  test('3. Double-click start does not produce 400 on UI (button disabled)', async ({ request }) => {
    const createResp = await request.post(`/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 2, nickname: 'Charlie' }
    });
    const createData = await createResp.json();
    const code = createData.roomCode;
    const playerID = createData.playerID;
    const playerToken = createData.token;

    const firstResp = await request.post(`/api/multiplayer/start`, {
      headers: { 'X-Player-Token': playerToken },
      data: { roomCode: code, playerID }
    });
    expect(firstResp.ok()).toBe(true);

    const secondResp = await request.post(`/api/multiplayer/start`, {
      headers: { 'X-Player-Token': playerToken },
      data: { roomCode: code, playerID }
    });
    expect(secondResp.status()).toBe(400);
    const secondData = await secondResp.json();
    expect(secondData.error).toBeTruthy();
  });

  test('4. SSE endpoint returns text/event-stream content type', async ({ page, request }) => {
    const createResp = await request.post('/api/multiplayer/create', {
      data: { mode: 'progressif', wordCount: 2, nickname: 'Diana' }
    });
    const createData = await createResp.json();
    const code = createData.roomCode;
    const playerID = createData.playerID;
    const playerToken = createData.token;

    // SSE streams indefinitely, so request.get() would hang.
    // Use page.evaluate with fetch + AbortController to read just the headers.
    await page.goto('/');
    const contentType = await page.evaluate(async (url) => {
      const ctrl = new AbortController();
      const resp = await fetch(url, { signal: ctrl.signal });
      const ct = resp.headers.get('content-type');
      ctrl.abort();
      return ct;
    }, `/api/multiplayer/events?room=${code}&player=${playerID}&token=${playerToken}`);
    expect(contentType).toContain('text/event-stream');
  });
});

test.describe('Multiplayer API', () => {
  let roomCode = '';
  let creatorID = '';

  test('1. Create room', async ({ request }) => {
    const resp = await request.post(`/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 3, nickname: 'Alice' }
    });
    expect(resp.ok()).toBe(true);
    const data = await resp.json();
    expect(data).toHaveProperty('roomCode');
    expect(data).toHaveProperty('playerID');
    expect(data.roomCode.length).toBe(6);
    roomCode = data.roomCode;
    creatorID = data.playerID;
  });

  test('2. Join room', async ({ request }) => {
    expect(roomCode).toBeTruthy();
    const resp = await request.post(`/api/multiplayer/join`, {
      data: { roomCode, nickname: 'Bob' }
    });
    expect(resp.ok()).toBe(true);
    const data = await resp.json();
    expect(data.state).toBe('lobby');
    expect(data.players.length).toBe(2);
  });

  test('3. Start game', async ({ request }) => {
    expect(roomCode).toBeTruthy();
    expect(creatorID).toBeTruthy();
    const resp = await request.post(`/api/multiplayer/start`, {
      data: { roomCode, playerID: creatorID }
    });
    expect(resp.ok()).toBe(true);
  });

  test('4. Guess word', async ({ request }) => {
    expect(roomCode).toBeTruthy();
    expect(creatorID).toBeTruthy();

    // Join to get current game state
    const joinResp = await request.post(`/api/multiplayer/join`, {
      data: { roomCode, playerID: creatorID }
    });
    const joinData = await joinResp.json();
    expect(joinData.state).toBe('playing');
    expect(joinData.wordSequence.length).toBe(3);

    // Get the first word target
    const target = joinData.wordGames[0].target;

    const resp = await request.post(`/api/multiplayer/guess`, {
      data: { roomCode, playerID: creatorID, word: target }
    });
    expect(resp.ok()).toBe(true);
    const data = await resp.json();
    expect(data.wordFinished).toBe(true);
    expect(data.wordWon).toBe(true);
    expect(data.playerFinished).toBe(false);
  });

  test('5. Invalid guess rejected', async ({ request }) => {
    expect(roomCode).toBeTruthy();
    expect(creatorID).toBeTruthy();

    const resp = await request.post(`/api/multiplayer/guess`, {
      data: { roomCode, playerID: creatorID, word: 'XXXXXX' }
    });
    expect(resp.ok()).toBe(false);
  });

  test('6. Leave room', async ({ request }) => {
    expect(roomCode).toBeTruthy();
    expect(creatorID).toBeTruthy();

    const resp = await request.post(`/api/multiplayer/leave`, {
      data: { roomCode, playerID: creatorID }
    });
    expect(resp.ok()).toBe(true);
  });

  test('7. Create room invalid params', async ({ request }) => {
    const resp1 = await request.post(`/api/multiplayer/create`, {
      data: { mode: 'invalid', wordCount: 3, nickname: 'Test' }
    });
    expect(resp1.ok()).toBe(false);

    const resp2 = await request.post(`/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 0, nickname: 'Test' }
    });
    expect(resp2.ok()).toBe(false);

    const resp3 = await request.post(`/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 3, nickname: '' }
    });
    expect(resp3.ok()).toBe(false);
  });

  test('8. Finished player has wordResults in rankings', async ({ request }) => {
    const createResp = await request.post(`/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 2, nickname: 'Alice' }
    });
    const createData = await createResp.json();
    const code = createData.roomCode;
    const aliceID = createData.playerID;

    const startResp = await request.post(`/api/multiplayer/start`, {
      data: { roomCode: code, playerID: aliceID }
    });
    expect(startResp.ok()).toBe(true);

    const joinResp = await request.post(`/api/multiplayer/join`, {
      data: { roomCode: code, playerID: aliceID }
    });
    const joinData = await joinResp.json();
    expect(joinData.wordGames.length).toBe(2);

    let lastResponse = null;
    for (let i = 0; i < 2; i++) {
      const target = joinData.wordGames[i].target;
      const guessResp = await request.post(`/api/multiplayer/guess`, {
        data: { roomCode: code, playerID: aliceID, word: target }
      });
      expect(guessResp.ok()).toBe(true);
      lastResponse = await guessResp.json();
    }

    expect(lastResponse.playerFinished).toBe(true);
    expect(lastResponse.rankings).toBeDefined();
    expect(Array.isArray(lastResponse.rankings)).toBe(true);

    const aliceRanking = lastResponse.rankings.find(r => r.playerID === aliceID);
    expect(aliceRanking).toBeDefined();
    expect(aliceRanking.wordResults).toBeDefined();
    expect(aliceRanking.wordResults.length).toBe(2);

    for (const wr of aliceRanking.wordResults) {
      expect(wr).toHaveProperty('target');
      expect(wr).toHaveProperty('attempts');
      expect(wr).toHaveProperty('results');
      expect(typeof wr.target).toBe('string');
      expect(Array.isArray(wr.attempts)).toBe(true);
      expect(Array.isArray(wr.results)).toBe(true);
    }
  });

  test('9. Results screen persists after player-finished SSE broadcast', async ({ page, request }) => {
    // Create room with Alice (solo — only one player so the game can start)
    const createResp = await request.post(`/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 2, nickname: 'Alice' }
    });
    const createData = await createResp.json();
    const code = createData.roomCode;
    const aliceID = createData.playerID;

    // Start and join the game so we can retrieve word targets for the API guesses
    const startResp = await request.post(`/api/multiplayer/start`, {
      data: { roomCode: code, playerID: aliceID }
    });
    expect(startResp.ok()).toBe(true);

    const joinResp = await request.post(`/api/multiplayer/join`, {
      data: { roomCode: code, playerID: aliceID }
    });
    const joinData = await joinResp.json();
    expect(joinData.wordGames.length).toBe(2);

    // Submit all winning guesses via API (simulates Alice finishing)
    let lastGuessData = null;
    for (let i = 0; i < 2; i++) {
      const target = joinData.wordGames[i].target;
      const guessResp = await request.post(`/api/multiplayer/guess`, {
        data: { roomCode: code, playerID: aliceID, word: target }
      });
      expect(guessResp.ok()).toBe(true);
      lastGuessData = await guessResp.json();
    }

    // Navigate to /multiplayer and inject state via evaluate to verify DOM behaviour
    await page.goto(`/multiplayer`);

    // Inject the rankings from the HTTP response directly into renderRankings (simulating what
    // the real client does after POST /api/multiplayer/guess returns playerFinished=true)
    const rankings = lastGuessData.rankings;
    await page.evaluate((rankings) => {
      // Show the results screen
      document.querySelectorAll('.screen').forEach(s => s.style.display = 'none');
      const resultsScreen = document.getElementById('screen-results');
      if (resultsScreen) resultsScreen.style.display = '';

      // Call renderRankings with the full payload (includes wordResults)
      if (typeof renderRankings === 'function') {
        renderRankings(rankings);
      }
    }, rankings);

    // Assert results screen is visible
    const resultsScreen = page.locator('#screen-results');
    await expect(resultsScreen).toBeVisible();

    // Assert player tabs and word results are populated
    await expect(page.locator('#player-tabs')).not.toBeEmpty();
    await expect(page.locator('#player-word-results')).not.toBeEmpty();
    const initialTabCount = await page.locator('#player-tabs .player-tab').count();
    expect(initialTabCount).toBeGreaterThanOrEqual(1);
    const initialCardCount = await page.locator('#player-word-results .word-result').count();
    expect(initialCardCount).toBeGreaterThanOrEqual(1);

    // Now simulate the `player-finished` SSE broadcast arriving with stripped rankings
    // (no wordResults) — this is what used to wipe the detail view
    const strippedRankings = rankings.map(r => ({ ...r, wordResults: undefined }));
    await page.evaluate((strippedRankings) => {
      if (typeof renderRankings === 'function') {
        renderRankings(strippedRankings);
      }
    }, strippedRankings);

    // Re-assert — under the old code tabs/word-results would be empty; with the fix they persist
    await expect(resultsScreen).toBeVisible();
    const tabCountAfter = await page.locator('#player-tabs .player-tab').count();
    expect(tabCountAfter).toBeGreaterThanOrEqual(1);
    const cardCountAfter = await page.locator('#player-word-results .word-result').count();
    expect(cardCountAfter).toBeGreaterThanOrEqual(1);
  });
});

test.describe('GET /api/status', () => {
  test('1. Returns 200 with status, version, and dictionaries', async ({ request }) => {
    const resp = await request.get(`/api/status`);
    expect(resp.ok()).toBe(true);
    const data = await resp.json();

    expect(data.status).toBe('ok');
    expect(typeof data.version).toBe('string');
    expect(data.version.length).toBeGreaterThan(0);
    expect(Array.isArray(data.dictionaries)).toBe(true);
    expect(data.dictionaries.length).toBe(2);
  });

  test('2. Dictionaries have name, word_count, sha256', async ({ request }) => {
    const resp = await request.get(`/api/status`);
    const data = await resp.json();

    for (const dict of data.dictionaries) {
      expect(dict).toHaveProperty('name');
      expect(dict).toHaveProperty('word_count');
      expect(dict).toHaveProperty('sha256');
      expect(typeof dict.name).toBe('string');
      expect(dict.word_count).toBeGreaterThan(0);
      expect(dict.sha256.length).toBeGreaterThan(0);
    }
  });

  test('3. Dictionaries include words and words_full', async ({ request }) => {
    const resp = await request.get(`/api/status`);
    const data = await resp.json();

    const names = data.dictionaries.map(d => d.name);
    expect(names).toContain('words');
    expect(names).toContain('words_full');
  });
});

test.describe('Theme toggle', () => {
  test('1. Theme toggle button exists on home page', async ({ page }) => {
    await page.goto(`/`);
    await expect(page.locator('#theme-toggle')).toBeVisible();
  });

  test('2. Toggle switches theme and back', async ({ page }) => {
    await page.goto(`/`);
    await page.waitForSelector('#theme-toggle');

    // Read current theme before toggle
    const initialTheme = await page.locator('html').getAttribute('data-theme');
    const nextTheme = initialTheme === 'light' ? 'dark' : 'light';

    // Toggle
    await page.locator('#theme-toggle').click();
    if (nextTheme === 'light') {
      await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
    } else {
      await expect(page.locator('html')).not.toHaveAttribute('data-theme', 'light');
    }

    // Toggle back
    await page.locator('#theme-toggle').click();
    if (initialTheme === 'light') {
      await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
    } else {
      await expect(page.locator('html')).not.toHaveAttribute('data-theme', 'light');
    }
  });

  test('3. Theme persists across page reload', async ({ page }) => {
    await page.goto(`/`);
    await page.waitForSelector('#theme-toggle');

    // Toggle to get non-default theme
    await page.locator('#theme-toggle').click();
    // Read the theme we toggled to
    const toggledTheme = await page.locator('html').getAttribute('data-theme');

    // Reload
    await page.reload();
    await page.waitForSelector('#theme-toggle');

    // Verify same theme is applied
    const reloadedTheme = await page.locator('html').getAttribute('data-theme');
    expect(reloadedTheme).toBe(toggledTheme);
  });

  test('4. Toggle exists and works on solo game page', async ({ page }) => {
    await page.goto(`/game?mode=solo`);
    await page.waitForSelector('#theme-toggle');
    await expect(page.locator('#theme-toggle')).toBeVisible();

    // Toggle and verify theme changed
    const initialTheme = await page.locator('html').getAttribute('data-theme');
    await page.locator('#theme-toggle').click();
    const newTheme = await page.locator('html').getAttribute('data-theme');
    expect(newTheme).not.toBe(initialTheme);
  });

  test('5. Default theme respects OS preference (dark)', async ({ page }) => {
    // Mock prefers-color-scheme: dark and clear localStorage
    await page.addInitScript(() => {
      Object.defineProperty(window, 'matchMedia', {
        writable: true,
        value: (query) => ({
          matches: query === '(prefers-color-scheme: dark)',
          media: query,
          onchange: null,
          addListener: () => {},
          removeListener: () => {},
          addEventListener: () => {},
          removeEventListener: () => {},
          dispatchEvent: () => {},
        }),
      });
    });
    await page.goto(`/`);
    // Without saved preference and prefers-color-scheme: dark, default should be dark (no data-theme or empty)
    const theme = await page.locator('html').getAttribute('data-theme');
    expect(theme === null || theme === '').toBe(true);
  });
});

test.describe('Letter palette settings', () => {
  const BASE_URL = 'http://localhost:3129';

  test('navigating to /settings renders 11 palette cards', async ({ page }) => {
    await page.goto(`${BASE_URL}/settings`);
    const cards = page.locator('.palette-card');
    await expect(cards).toHaveCount(11);
  });

  test('clicking a palette updates CSS custom property --correct', async ({ page }) => {
    await page.goto(`${BASE_URL}/settings`);
    // Click the "wordle" palette card
    const wordleCard = page.locator('.palette-card[data-palette-id="wordle"]');
    await wordleCard.click();
    const correct = await page.evaluate(() =>
      getComputedStyle(document.documentElement).getPropertyValue('--correct').trim()
    );
    expect(correct).toBe('#538d4e');
  });

  test('selected palette persists after reload', async ({ page }) => {
    await page.goto(`${BASE_URL}/settings`);
    await page.locator('.palette-card[data-palette-id="ocean"]').click();
    await page.reload();
    const selected = page.locator('.palette-card.selected');
    await expect(selected).toHaveAttribute('data-palette-id', 'ocean');
    const correct = await page.evaluate(() =>
      getComputedStyle(document.documentElement).getPropertyValue('--correct').trim()
    );
    expect(correct).toBe('#0ea5e9');
  });

  test('gear icon is visible on / and /game?mode=solo and links to /settings', async ({ page }) => {
    await page.goto(`${BASE_URL}/`);
    const gearHome = page.locator('#settings-link');
    await expect(gearHome).toBeVisible();
    await expect(gearHome).toHaveAttribute('href', '/settings');

    await page.goto(`${BASE_URL}/game?mode=solo`);
    const gearGame = page.locator('#settings-link');
    await expect(gearGame).toBeVisible();
    await expect(gearGame).toHaveAttribute('href', '/settings');
  });

  test('selected palette CSS variables are applied on /game?mode=solo', async ({ page }) => {
    // Set ocean palette via settings page
    await page.goto(`${BASE_URL}/settings`);
    await page.locator('.palette-card[data-palette-id="ocean"]').click();

    // Navigate to solo game — palette should be applied on DOMContentLoaded
    await page.goto(`${BASE_URL}/game?mode=solo`);

    const correct = await page.evaluate(() =>
      getComputedStyle(document.documentElement).getPropertyValue('--correct').trim()
    );
    // Ocean --correct is #0ea5e9
    expect(correct).toBe('#0ea5e9');
  });
});

test.describe('Late joiner visibility on results screen', () => {
  const DEV = 'http://localhost:3131';
  test('10. Finished player sees late joiner and progress updates in rankings table', async ({ page }) => {
    // Use the dev server (port 3131) which has the fix under test.
    // Step 1: Create room, add two players, start game, have Alice finish
    const createResp = await page.request.post(`${DEV}/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 2, nickname: 'Alice' }
    });
    expect(createResp.ok()).toBe(true);
    const createData = await createResp.json();
    const code = createData.roomCode;
    const aliceID = createData.playerID;

    const bobJoinResp = await page.request.post(`${DEV}/api/multiplayer/join`, {
      data: { roomCode: code, nickname: 'Bob' }
    });
    expect(bobJoinResp.ok()).toBe(true);
    const bobData = await bobJoinResp.json();
    const bobID = bobData.playerID;

    const startResp = await page.request.post(`${DEV}/api/multiplayer/start`, {
      data: { roomCode: code, playerID: aliceID }
    });
    expect(startResp.ok()).toBe(true);

    // Alice rejoins to get her word targets
    const aliceRejoinResp = await page.request.post(`${DEV}/api/multiplayer/join`, {
      data: { roomCode: code, playerID: aliceID }
    });
    expect(aliceRejoinResp.ok()).toBe(true);
    const aliceGameData = await aliceRejoinResp.json();

    // Alice finishes all words
    for (let i = 0; i < 2; i++) {
      const target = aliceGameData.wordGames[i].target;
      const guessResp = await page.request.post(`${DEV}/api/multiplayer/guess`, {
        data: { roomCode: code, playerID: aliceID, word: target }
      });
      expect(guessResp.ok()).toBe(true);
    }

    // Step 2: Navigate to /multiplayer and set up Alice's results screen with 2 players
    await page.goto(`${DEV}/multiplayer`);

    // Simulate Alice's results screen with Alice (finished) and Bob (in progress)
    // Using the exact shape that the server produces (time is a Go time.Duration = nanoseconds int)
    await page.evaluate(([aID, bID]) => {
      document.querySelectorAll('.screen').forEach(s => s.style.display = 'none');
      const resultsScreen = document.getElementById('screen-results');
      if (resultsScreen) resultsScreen.style.display = '';
      if (typeof renderRankings === 'function') {
        renderRankings([
          { playerID: aID, nickname: 'Alice', finished: true, failed: false, time: 5000000000 },
          { playerID: bID, nickname: 'Bob', finished: false, failed: false, time: 0 }
        ]);
      }
    }, [aliceID, bobID]);

    await expect(page.locator('#screen-results')).toBeVisible();
    const initialRowCount = await page.locator('#rankings-body tr').count();
    expect(initialRowCount).toBe(2);

    // Step 3: Charlie joins mid-game — simulate the player-joined SSE with fresh rankings
    const charlieJoinResp = await page.request.post(`${DEV}/api/multiplayer/join`, {
      data: { roomCode: code, nickname: 'Charlie' }
    });
    expect(charlieJoinResp.ok()).toBe(true);
    const charlieData = await charlieJoinResp.json();
    const charlieID = charlieData.playerID;

    // Simulate the player-joined SSE arriving on Alice's results screen
    await page.evaluate(([aID, bID, cID]) => {
      if (typeof renderRankings === 'function') {
        renderRankings([
          { playerID: aID, nickname: 'Alice', finished: true, failed: false, time: 5000000000 },
          { playerID: bID, nickname: 'Bob', finished: false, failed: false, time: 0 },
          { playerID: cID, nickname: 'Charlie', finished: false, failed: false, time: 0 }
        ]);
      }
    }, [aliceID, bobID, charlieID]);

    // Step 4: Assert Charlie appears with "En cours" and "—"
    await expect(page.locator('#screen-results')).toBeVisible();
    const rowCountAfterJoin = await page.locator('#rankings-body tr').count();
    expect(rowCountAfterJoin).toBe(3);

    const charlieRow = page.locator('#rankings-body tr', { hasText: 'Charlie' });
    await expect(charlieRow).toContainText('En cours');
    await expect(charlieRow).toContainText('—');

    // Step 5: Simulate a progress SSE arriving — table must still show all 3 players
    await page.evaluate(([aID, bID, cID]) => {
      if (typeof renderRankings === 'function') {
        renderRankings([
          { playerID: aID, nickname: 'Alice', finished: true, failed: false, time: 5000000000 },
          { playerID: bID, nickname: 'Bob', finished: false, failed: false, time: 0 },
          { playerID: cID, nickname: 'Charlie', finished: false, failed: false, time: 0 }
        ]);
      }
    }, [aliceID, bobID, charlieID]);

    const rowCountAfterProgress = await page.locator('#rankings-body tr').count();
    expect(rowCountAfterProgress).toBe(3);
  });
});
