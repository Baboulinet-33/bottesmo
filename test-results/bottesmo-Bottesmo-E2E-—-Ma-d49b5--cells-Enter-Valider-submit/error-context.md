# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: bottesmo.spec.js >> Bottesmo E2E — Match UI to real Bottesmo >> 1. Basic typing flow — letters fill cells, Enter/Valider submit
- Location: bottesmo.spec.js:37:3

# Error details

```
Error: expect(received).toMatch(expected)

Expected pattern: /submitted|correct|present|absent/
Received string:  "tile cursor"
```

# Page snapshot

```yaml
- generic [active] [ref=e1]:
  - banner [ref=e2]:
    - heading "Bottesmo" [level=1] [ref=e3]:
      - link "Bottesmo" [ref=e4] [cursor=pointer]:
        - /url: /
    - button "Changer de thème" [ref=e5] [cursor=pointer]: 🌙
  - main [ref=e6]:
    - generic [ref=e7]:
      - generic [ref=e10]: C
      - button "Valider" [ref=e58] [cursor=pointer]
      - generic [ref=e59]: Mot invalide
      - generic [ref=e60]:
        - generic [ref=e61]:
          - button "A" [ref=e62] [cursor=pointer]
          - button "Z" [ref=e63] [cursor=pointer]
          - button "E" [ref=e64] [cursor=pointer]
          - button "R" [ref=e65] [cursor=pointer]
          - button "T" [ref=e66] [cursor=pointer]
          - button "Y" [ref=e67] [cursor=pointer]
          - button "U" [ref=e68] [cursor=pointer]
          - button "I" [ref=e69] [cursor=pointer]
          - button "O" [ref=e70] [cursor=pointer]
          - button "P" [ref=e71] [cursor=pointer]
        - generic [ref=e72]:
          - button "Q" [ref=e73] [cursor=pointer]
          - button "S" [ref=e74] [cursor=pointer]
          - button "D" [ref=e75] [cursor=pointer]
          - button "F" [ref=e76] [cursor=pointer]
          - button "G" [ref=e77] [cursor=pointer]
          - button "H" [ref=e78] [cursor=pointer]
          - button "J" [ref=e79] [cursor=pointer]
          - button "K" [ref=e80] [cursor=pointer]
          - button "L" [ref=e81] [cursor=pointer]
          - button "M" [ref=e82] [cursor=pointer]
        - generic [ref=e83]:
          - button "Entrée" [ref=e84] [cursor=pointer]
          - button "W" [ref=e85] [cursor=pointer]
          - button "X" [ref=e86] [cursor=pointer]
          - button "C" [ref=e87] [cursor=pointer]
          - button "V" [ref=e88] [cursor=pointer]
          - button "B" [ref=e89] [cursor=pointer]
          - button "N" [ref=e90] [cursor=pointer]
          - button "Suppr" [ref=e91] [cursor=pointer]
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
> 65  |     expect(tileClass).toMatch(/submitted|correct|present|absent/);
      |                       ^ Error: expect(received).toMatch(expected)
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
  85  |     await expect(page.locator('#tile-0-0')).toHaveText(firstLetter);
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
```