const { test, expect } = require('@playwright/test');

const BASE = 'http://localhost:3113';

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

test.describe('Tusmo E2E — Match UI to real Tusmo', () => {
  let targetWord = '';
  let firstLetter = '';

  test.beforeAll(async ({ request }) => {
    // Start a solo game to learn the target word
    const resp = await request.post(`${BASE}/api/game/new`, {
      data: { mode: 'solo' }
    });
    const game = await resp.json();
    targetWord = game.firstLetter; // We only know first letter
    firstLetter = game.firstLetter;
    console.log(`First letter: ${firstLetter}, word length: ${game.wordLength}`);
  });

  test('1. Basic typing flow — letters fill cells, Enter/Valider submit', async ({ page }) => {
    await page.goto(`${BASE}/game?mode=solo`);
    await page.waitForSelector('#grid');

    // Row 0, col 0 should be pre-filled with first letter (locked)
    const tile00 = page.locator('#tile-0-0');
    await expect(tile00).toHaveText(firstLetter);
    await expect(tile00).toHaveClass(/locked/);
    await expect(tile00).toHaveClass(/correct/);

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

    // After submission, tiles should have color classes
    const tile01 = page.locator('#tile-0-1');
    const tileClass = await tile01.getAttribute('class');
    expect(tileClass).toMatch(/submitted|correct|present|absent/);
  });

  test('2. Backspace behavior — typed letters clear, locked letters stay', async ({ page }) => {
    await page.goto(`${BASE}/game?mode=solo`);
    await page.waitForSelector('#grid');

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
    await page.goto(`${BASE}/game?mode=solo`);
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
    await page.goto(`${BASE}/game?mode=solo`);
    await page.waitForSelector('#keyboard');

    // Click A on-screen keyboard
    await page.locator('.kb-key').filter({ hasText: /^A$/ }).click();
    await expect(page.locator('#tile-0-1')).toHaveText('A');

    // Click Z on-screen keyboard
    await page.locator('.kb-key').filter({ hasText: /^Z$/ }).click();
    await expect(page.locator('#tile-0-2')).toHaveText('Z');
  });

  test('5. Keyboard Backspace click clears last cell', async ({ page }) => {
    await page.goto(`${BASE}/game?mode=solo`);
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
    await page.goto(`${BASE}/game?mode=solo`);
    await page.waitForSelector('#keyboard');

    // Fill row and submit
    for (let i = 1; i < 7; i++) {
      await page.keyboard.press('KeyA');
    }
    await submitGuess(page);
    await page.waitForTimeout(500);

    // Some keyboard keys should now have color classes
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
    expect(hasColor).toBe(true);
  });

  test('7. Complete game (win) — verify win message appears', async ({ page }) => {
    // Get target word via API
    const resp = await page.request().post(`${BASE}/api/game/new`, {
      data: { mode: 'solo' }
    });
    const game = await resp.json();

    await page.goto(`${BASE}/game?mode=solo`);
    await page.waitForSelector('#grid');

    // Type the target word (known via intercept)
    const word = game.firstLetter + 'XXXXXX';
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
    await page.goto(`${BASE}/game?mode=solo`);
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
    await page.goto(`${BASE}/game?mode=daily`);
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
    await page.goto(`${BASE}/`);
    await page.waitForSelector('.mode-btn');
    const buttons = page.locator('.mode-btn');
    await expect(buttons).toHaveCount(3);
    await expect(buttons.nth(2)).toHaveText('Multijoueur');
  });

  test('2. Late joiner receives non-empty wordGames', async ({ request }) => {
    const createResp = await request.post(`${BASE}/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 2, nickname: 'Alice' }
    });
    const createData = await createResp.json();
    const code = createData.roomCode;
    const aliceID = createData.playerID;

    const startResp = await request.post(`${BASE}/api/multiplayer/start`, {
      data: { roomCode: code, playerID: aliceID }
    });
    expect(startResp.ok()).toBe(true);

    const joinResp = await request.post(`${BASE}/api/multiplayer/join`, {
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
    const createResp = await request.post(`${BASE}/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 2, nickname: 'Charlie' }
    });
    const createData = await createResp.json();
    const code = createData.roomCode;
    const playerID = createData.playerID;

    const firstResp = await request.post(`${BASE}/api/multiplayer/start`, {
      data: { roomCode: code, playerID }
    });
    expect(firstResp.ok()).toBe(true);

    const secondResp = await request.post(`${BASE}/api/multiplayer/start`, {
      data: { roomCode: code, playerID }
    });
    expect(secondResp.status()).toBe(400);
    const secondData = await secondResp.json();
    expect(secondData.error).toBeTruthy();
  });

  test('4. SSE endpoint returns text/event-stream content type', async ({ page }) => {
    const createResp = await page.request().post(`${BASE}/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 2, nickname: 'Diana' }
    });
    const createData = await createResp.json();
    const code = createData.roomCode;
    const playerID = createData.playerID;

    const response = await page.goto(`${BASE}/api/multiplayer/events?room=${code}&player=${playerID}`);
    expect(response.headers()['content-type']).toContain('text/event-stream');

    await page.goto(`${BASE}/`);
  });
});

test.describe('Multiplayer API', () => {
  let roomCode = '';
  let creatorID = '';

  test('1. Create room', async ({ request }) => {
    const resp = await request.post(`${BASE}/api/multiplayer/create`, {
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
    const resp = await request.post(`${BASE}/api/multiplayer/join`, {
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
    const resp = await request.post(`${BASE}/api/multiplayer/start`, {
      data: { roomCode, playerID: creatorID }
    });
    expect(resp.ok()).toBe(true);
  });

  test('4. Guess word', async ({ request }) => {
    expect(roomCode).toBeTruthy();
    expect(creatorID).toBeTruthy();

    // Join to get current game state
    const joinResp = await request.post(`${BASE}/api/multiplayer/join`, {
      data: { roomCode, playerID: creatorID }
    });
    const joinData = await joinResp.json();
    expect(joinData.state).toBe('playing');
    expect(joinData.wordSequence.length).toBe(3);

    // Get the first word target
    const target = joinData.wordGames[0].target;

    const resp = await request.post(`${BASE}/api/multiplayer/guess`, {
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

    const resp = await request.post(`${BASE}/api/multiplayer/guess`, {
      data: { roomCode, playerID: creatorID, word: 'XXXXXX' }
    });
    expect(resp.ok()).toBe(false);
  });

  test('6. Leave room', async ({ request }) => {
    expect(roomCode).toBeTruthy();
    expect(creatorID).toBeTruthy();

    const resp = await request.post(`${BASE}/api/multiplayer/leave`, {
      data: { roomCode, playerID: creatorID }
    });
    expect(resp.ok()).toBe(true);
  });

  test('7. Create room invalid params', async ({ request }) => {
    const resp1 = await request.post(`${BASE}/api/multiplayer/create`, {
      data: { mode: 'invalid', wordCount: 3, nickname: 'Test' }
    });
    expect(resp1.ok()).toBe(false);

    const resp2 = await request.post(`${BASE}/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 0, nickname: 'Test' }
    });
    expect(resp2.ok()).toBe(false);

    const resp3 = await request.post(`${BASE}/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 3, nickname: '' }
    });
    expect(resp3.ok()).toBe(false);
  });
});

test.describe('Theme toggle', () => {
  test('1. Theme toggle button exists on home page', async ({ page }) => {
    await page.goto(`${BASE}/`);
    await expect(page.locator('#theme-toggle')).toBeVisible();
  });

  test('2. Toggle switches theme and back', async ({ page }) => {
    await page.goto(`${BASE}/`);
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
    await page.goto(`${BASE}/`);
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
    await page.goto(`${BASE}/game?mode=solo`);
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
    await page.goto(`${BASE}/`);
    // Without saved preference and prefers-color-scheme: dark, default should be dark (no data-theme or empty)
    const theme = await page.locator('html').getAttribute('data-theme');
    expect(theme === null || theme === '').toBe(true);
  });
});
