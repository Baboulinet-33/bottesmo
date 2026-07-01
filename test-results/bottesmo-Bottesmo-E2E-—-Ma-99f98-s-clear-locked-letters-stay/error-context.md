# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: bottesmo.spec.js >> Bottesmo E2E — Match UI to real Bottesmo >> 2. Backspace behavior — typed letters clear, locked letters stay
- Location: bottesmo.spec.js:68:3

# Error details

```
Error: expect(locator).toHaveText(expected) failed

Locator:  locator('#tile-0-0')
Expected: "E"
Received: "C"
Timeout:  5000ms

Call log:
  - Expect "toHaveText" with timeout 5000ms
  - waiting for locator('#tile-0-0')
    14 × locator resolved to <div id="tile-0-0" class="tile locked correct">C</div>
       - unexpected value "C"

```

```yaml
- text: C
```

# Test source

```ts
  1   | const { test, expect } = require('@playwright/test');
  2   | 
  3   | const BASE = 'http://localhost:3118';
  4   | 
  5   | /**
  6   |  * Helper: types the given word into the current row using physical keyboard.
  7   |  * Skips locked positions automatically based on first letter logic.
  8   |  */
  9   | async function typeWord(page, word) {
  10  |   for (const ch of word.slice(1)) {
  11  |     await page.keyboard.press('Key' + ch);
  12  |   }
  13  | }
  14  | 
  15  | /**
  16  |  * Helper: click "Valider" button to submit.
  17  |  */
  18  | async function submitGuess(page) {
  19  |   await page.locator('#submit-btn').click();
  20  | }
  21  | 
  22  | test.describe('Bottesmo E2E — Match UI to real Bottesmo', () => {
  23  |   let targetWord = '';
  24  |   let firstLetter = '';
  25  | 
  26  |   test.beforeAll(async ({ request }) => {
  27  |     // Start a solo game to learn the target word
  28  |     const resp = await request.post(`${BASE}/api/game/new`, {
  29  |       data: { mode: 'solo' }
  30  |     });
  31  |     const game = await resp.json();
  32  |     targetWord = game.firstLetter; // We only know first letter
  33  |     firstLetter = game.firstLetter;
  34  |     console.log(`First letter: ${firstLetter}, word length: ${game.wordLength}`);
  35  |   });
  36  | 
  37  |   test('1. Basic typing flow — letters fill cells, Enter/Valider submit', async ({ page }) => {
  38  |     await page.goto(`${BASE}/game?mode=solo`);
  39  |     await page.waitForSelector('#grid');
  40  | 
  41  |     // Row 0, col 0 should be pre-filled with first letter (locked)
  42  |     const tile00 = page.locator('#tile-0-0');
  43  |     await expect(tile00).toHaveText(firstLetter);
  44  |     await expect(tile00).toHaveClass(/locked/);
  45  |     await expect(tile00).toHaveClass(/correct/);
  46  | 
  47  |     // Type a letter via physical keyboard
  48  |     await page.keyboard.press('KeyA');
  49  |     await expect(page.locator('#tile-0-1')).toHaveText('A');
  50  | 
  51  |     // Type another
  52  |     await page.keyboard.press('KeyB');
  53  |     await expect(page.locator('#tile-0-2')).toHaveText('B');
  54  | 
  55  |     // Fill rest and submit
  56  |     for (let i = 3; i < 7; i++) {
  57  |       await page.keyboard.press('KeyC');
  58  |     }
  59  |     await submitGuess(page);
  60  |     await page.waitForTimeout(500);
  61  | 
  62  |     // After submission, tiles should have color classes
  63  |     const tile01 = page.locator('#tile-0-1');
  64  |     const tileClass = await tile01.getAttribute('class');
  65  |     expect(tileClass).toMatch(/submitted|correct|present|absent/);
  66  |   });
  67  | 
  68  |   test('2. Backspace behavior — typed letters clear, locked letters stay', async ({ page }) => {
  69  |     await page.goto(`${BASE}/game?mode=solo`);
  70  |     await page.waitForSelector('#grid');
  71  | 
  72  |     // Type letters
  73  |     await page.keyboard.press('KeyA');
  74  |     await page.keyboard.press('KeyB');
  75  |     await page.keyboard.press('KeyC');
  76  |     await expect(page.locator('#tile-0-3')).toHaveText('C');
  77  | 
  78  |     // Backspace twice — removes C, then B
  79  |     await page.keyboard.press('Backspace');
  80  |     await expect(page.locator('#tile-0-3')).toHaveText('');
  81  |     await page.keyboard.press('Backspace');
  82  |     await expect(page.locator('#tile-0-2')).toHaveText('');
  83  | 
  84  |     // Position 0 must remain locked with first letter
> 85  |     await expect(page.locator('#tile-0-0')).toHaveText(firstLetter);
      |                                             ^ Error: expect(locator).toHaveText(expected) failed
  86  |     await expect(page.locator('#tile-0-0')).toHaveClass(/locked/);
  87  | 
  88  |     // Backspace on col 1 clears it
  89  |     await page.keyboard.press('Backspace');
  90  |     await expect(page.locator('#tile-0-1')).toHaveText('');
  91  |   });
  92  | 
  93  |   test('3. Pre-filled letters appear on next row after a guess', async ({ page }) => {
  94  |     await page.goto(`${BASE}/game?mode=solo`);
  95  |     await page.waitForSelector('#grid');
  96  | 
  97  |     // Fill entire row
  98  |     for (let i = 1; i < 7; i++) {
  99  |       await page.keyboard.press('KeyA');
  100 |     }
  101 |     await submitGuess(page);
  102 |     await page.waitForTimeout(500);
  103 | 
  104 |     // Second row should have pre-filled letters at correct positions
  105 |     const tile10 = page.locator('#tile-1-0');
  106 |     // If the submission was valid (word accepted), position 0 should be pre-filled
  107 |     const tile10Text = await tile10.textContent();
  108 |     if (tile10Text.length > 0) {
  109 |       await expect(tile10).toHaveClass(/locked/);
  110 |       await expect(tile10).toHaveClass(/correct/);
  111 |     }
  112 |   });
  113 | 
  114 |   test('4. Keyboard click fills cells', async ({ page }) => {
  115 |     await page.goto(`${BASE}/game?mode=solo`);
  116 |     await page.waitForSelector('#keyboard');
  117 | 
  118 |     // Click A on-screen keyboard
  119 |     await page.locator('.kb-key').filter({ hasText: /^A$/ }).click();
  120 |     await expect(page.locator('#tile-0-1')).toHaveText('A');
  121 | 
  122 |     // Click Z on-screen keyboard
  123 |     await page.locator('.kb-key').filter({ hasText: /^Z$/ }).click();
  124 |     await expect(page.locator('#tile-0-2')).toHaveText('Z');
  125 |   });
  126 | 
  127 |   test('5. Keyboard Backspace click clears last cell', async ({ page }) => {
  128 |     await page.goto(`${BASE}/game?mode=solo`);
  129 |     await page.waitForSelector('#keyboard');
  130 | 
  131 |     // Type via physical keyboard
  132 |     await page.keyboard.press('KeyA');
  133 |     await page.keyboard.press('KeyB');
  134 |     await expect(page.locator('#tile-0-2')).toHaveText('B');
  135 | 
  136 |     // Click Backspace on on-screen keyboard
  137 |     await page.locator('.kb-key').filter({ hasText: 'Suppr' }).click();
  138 |     await expect(page.locator('#tile-0-2')).toHaveText('');
  139 |   });
  140 | 
  141 |   test('6. Keyboard colors update after submission', async ({ page }) => {
  142 |     await page.goto(`${BASE}/game?mode=solo`);
  143 |     await page.waitForSelector('#keyboard');
  144 | 
  145 |     // Fill row and submit
  146 |     for (let i = 1; i < 7; i++) {
  147 |       await page.keyboard.press('KeyA');
  148 |     }
  149 |     await submitGuess(page);
  150 |     await page.waitForTimeout(500);
  151 | 
  152 |     // Some keyboard keys should now have color classes
  153 |     const keys = page.locator('.kb-key');
  154 |     const count = await keys.count();
  155 |     let hasColor = false;
  156 |     for (let i = 0; i < count; i++) {
  157 |       const cls = await keys.nth(i).getAttribute('class');
  158 |       if (/correct|present|absent/.test(cls || '')) {
  159 |         hasColor = true;
  160 |         break;
  161 |       }
  162 |     }
  163 |     expect(hasColor).toBe(true);
  164 |   });
  165 | 
  166 |   test('7. Complete game (win) — verify win message appears', async ({ page }) => {
  167 |     // Get target word via API
  168 |     const resp = await page.request().post(`${BASE}/api/game/new`, {
  169 |       data: { mode: 'solo' }
  170 |     });
  171 |     const game = await resp.json();
  172 | 
  173 |     await page.goto(`${BASE}/game?mode=solo`);
  174 |     await page.waitForSelector('#grid');
  175 | 
  176 |     // Type the target word (known via intercept)
  177 |     const word = game.firstLetter + 'XXXXXX';
  178 |     for (let i = 1; i < 7; i++) {
  179 |       await page.keyboard.press('KeyX');
  180 |     }
  181 |     await submitGuess(page);
  182 |     await page.waitForTimeout(1000);
  183 | 
  184 |     // After game over, either win or lose message
  185 |     const msg = page.locator('#message');
```