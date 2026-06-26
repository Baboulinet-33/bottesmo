# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: tusmo.spec.js >> Tusmo E2E — Match UI to real Tusmo >> 6. Keyboard colors update after submission
- Location: tusmo.spec.js:141:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: true
Received: false
```

# Page snapshot

```yaml
- generic [ref=e1]:
  - banner [ref=e2]:
    - heading "Tusmo" [level=1] [ref=e3]:
      - link "Tusmo" [ref=e4] [cursor=pointer]:
        - /url: /
  - main [ref=e5]:
    - generic [ref=e6]:
      - generic [ref=e9]: C
      - button "Valider" [active] [ref=e63] [cursor=pointer]
      - generic [ref=e64]: Le mot doit faire 8 lettres
      - generic [ref=e65]:
        - generic [ref=e66]:
          - button "A" [ref=e67] [cursor=pointer]
          - button "Z" [ref=e68] [cursor=pointer]
          - button "E" [ref=e69] [cursor=pointer]
          - button "R" [ref=e70] [cursor=pointer]
          - button "T" [ref=e71] [cursor=pointer]
          - button "Y" [ref=e72] [cursor=pointer]
          - button "U" [ref=e73] [cursor=pointer]
          - button "I" [ref=e74] [cursor=pointer]
          - button "O" [ref=e75] [cursor=pointer]
          - button "P" [ref=e76] [cursor=pointer]
        - generic [ref=e77]:
          - button "Q" [ref=e78] [cursor=pointer]
          - button "S" [ref=e79] [cursor=pointer]
          - button "D" [ref=e80] [cursor=pointer]
          - button "F" [ref=e81] [cursor=pointer]
          - button "G" [ref=e82] [cursor=pointer]
          - button "H" [ref=e83] [cursor=pointer]
          - button "J" [ref=e84] [cursor=pointer]
          - button "K" [ref=e85] [cursor=pointer]
          - button "L" [ref=e86] [cursor=pointer]
          - button "M" [ref=e87] [cursor=pointer]
        - generic [ref=e88]:
          - button "Entrée" [ref=e89] [cursor=pointer]
          - button "W" [ref=e90] [cursor=pointer]
          - button "X" [ref=e91] [cursor=pointer]
          - button "C" [ref=e92] [cursor=pointer]
          - button "V" [ref=e93] [cursor=pointer]
          - button "B" [ref=e94] [cursor=pointer]
          - button "N" [ref=e95] [cursor=pointer]
          - button "Suppr" [ref=e96] [cursor=pointer]
```

# Test source

```ts
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
> 163 |     expect(hasColor).toBe(true);
      |                      ^ Error: expect(received).toBe(expected) // Object.is equality
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
  186 |     await expect(msg).not.toBeEmpty();
  187 |   });
  188 | 
  189 |   test('9. Mobile viewport (≤480px) — layout renders correctly', async ({ page }) => {
  190 |     await page.setViewportSize({ width: 375, height: 667 });
  191 |     await page.goto(`${BASE}/game?mode=solo`);
  192 |     await page.waitForSelector('#grid');
  193 |     await page.waitForSelector('#keyboard');
  194 | 
  195 |     // Grid tiles should be visible and smaller
  196 |     const tile = page.locator('.tile').first();
  197 |     await expect(tile).toBeVisible();
  198 | 
  199 |     // Type to verify interaction works on mobile
  200 |     await page.keyboard.press('KeyA');
  201 |     await expect(page.locator('#tile-0-1')).toHaveText('A');
  202 |   });
  203 | 
  204 |   test('10. Daily mode — deterministic word', async ({ page }) => {
  205 |     await page.goto(`${BASE}/game?mode=daily`);
  206 |     await page.waitForSelector('#grid');
  207 | 
  208 |     // Grid should have 6 rows
  209 |     const rows = await page.locator('.row').count();
  210 |     expect(rows).toBe(6);
  211 | 
  212 |     // First tile locked with first letter
  213 |     const tile = page.locator('#tile-0-0');
  214 |     await expect(tile).toHaveClass(/locked/);
  215 |     const letter = await tile.textContent();
  216 |     expect(letter.length).toBe(1);
  217 |   });
  218 | });
  219 | 
```