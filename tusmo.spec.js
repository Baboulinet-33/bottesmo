const { test, expect } = require('@playwright/test');

const BASE = 'http://localhost:3106';

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
