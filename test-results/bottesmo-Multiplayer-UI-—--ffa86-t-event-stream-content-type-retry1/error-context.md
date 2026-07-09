# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: bottesmo.spec.js >> Multiplayer UI — Bugfix verification >> 4. SSE endpoint returns text/event-stream content type
- Location: bottesmo.spec.js:275:3

# Error details

```
TypeError: page.request is not a function
```

# Test source

```ts
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
  220 | test.describe('Multiplayer UI — Bugfix verification', () => {
  221 |   test('1. Homepage shows three mode buttons (Daily, Solo, Multijoueur)', async ({ page }) => {
  222 |     await page.goto(`${BASE}/`);
  223 |     await page.waitForSelector('.mode-btn');
  224 |     const buttons = page.locator('.mode-btn');
  225 |     await expect(buttons).toHaveCount(3);
  226 |     await expect(buttons.nth(2)).toHaveText('Multijoueur');
  227 |   });
  228 | 
  229 |   test('2. Late joiner receives non-empty wordGames', async ({ request }) => {
  230 |     const createResp = await request.post(`${BASE}/api/multiplayer/create`, {
  231 |       data: { mode: 'progressif', wordCount: 2, nickname: 'Alice' }
  232 |     });
  233 |     const createData = await createResp.json();
  234 |     const code = createData.roomCode;
  235 |     const aliceID = createData.playerID;
  236 | 
  237 |     const startResp = await request.post(`${BASE}/api/multiplayer/start`, {
  238 |       data: { roomCode: code, playerID: aliceID }
  239 |     });
  240 |     expect(startResp.ok()).toBe(true);
  241 | 
  242 |     const joinResp = await request.post(`${BASE}/api/multiplayer/join`, {
  243 |       data: { roomCode: code, nickname: 'Bob' }
  244 |     });
  245 |     expect(joinResp.ok()).toBe(true);
  246 |     const joinData = await joinResp.json();
  247 |     expect(joinData.state).toBe('playing');
  248 |     expect(joinData.wordGames).toBeDefined();
  249 |     expect(joinData.wordGames.length).toBe(2);
  250 |     expect(joinData.wordGames[0]).toBeDefined();
  251 |     expect(joinData.wordGames[0].target).toBeTruthy();
  252 |   });
  253 | 
  254 |   test('3. Double-click start does not produce 400 on UI (button disabled)', async ({ request }) => {
  255 |     const createResp = await request.post(`${BASE}/api/multiplayer/create`, {
  256 |       data: { mode: 'progressif', wordCount: 2, nickname: 'Charlie' }
  257 |     });
  258 |     const createData = await createResp.json();
  259 |     const code = createData.roomCode;
  260 |     const playerID = createData.playerID;
  261 | 
  262 |     const firstResp = await request.post(`${BASE}/api/multiplayer/start`, {
  263 |       data: { roomCode: code, playerID }
  264 |     });
  265 |     expect(firstResp.ok()).toBe(true);
  266 | 
  267 |     const secondResp = await request.post(`${BASE}/api/multiplayer/start`, {
  268 |       data: { roomCode: code, playerID }
  269 |     });
  270 |     expect(secondResp.status()).toBe(400);
  271 |     const secondData = await secondResp.json();
  272 |     expect(secondData.error).toBeTruthy();
  273 |   });
  274 | 
  275 |   test('4. SSE endpoint returns text/event-stream content type', async ({ page }) => {
> 276 |     const createResp = await page.request().post(`${BASE}/api/multiplayer/create`, {
      |                                   ^ TypeError: page.request is not a function
  277 |       data: { mode: 'progressif', wordCount: 2, nickname: 'Diana' }
  278 |     });
  279 |     const createData = await createResp.json();
  280 |     const code = createData.roomCode;
  281 |     const playerID = createData.playerID;
  282 | 
  283 |     const response = await page.goto(`${BASE}/api/multiplayer/events?room=${code}&player=${playerID}`);
  284 |     expect(response.headers()['content-type']).toContain('text/event-stream');
  285 | 
  286 |     await page.goto(`${BASE}/`);
  287 |   });
  288 | });
  289 | 
  290 | test.describe('Multiplayer API', () => {
  291 |   let roomCode = '';
  292 |   let creatorID = '';
  293 | 
  294 |   test('1. Create room', async ({ request }) => {
  295 |     const resp = await request.post(`${BASE}/api/multiplayer/create`, {
  296 |       data: { mode: 'progressif', wordCount: 3, nickname: 'Alice' }
  297 |     });
  298 |     expect(resp.ok()).toBe(true);
  299 |     const data = await resp.json();
  300 |     expect(data).toHaveProperty('roomCode');
  301 |     expect(data).toHaveProperty('playerID');
  302 |     expect(data.roomCode.length).toBe(6);
  303 |     roomCode = data.roomCode;
  304 |     creatorID = data.playerID;
  305 |   });
  306 | 
  307 |   test('2. Join room', async ({ request }) => {
  308 |     expect(roomCode).toBeTruthy();
  309 |     const resp = await request.post(`${BASE}/api/multiplayer/join`, {
  310 |       data: { roomCode, nickname: 'Bob' }
  311 |     });
  312 |     expect(resp.ok()).toBe(true);
  313 |     const data = await resp.json();
  314 |     expect(data.state).toBe('lobby');
  315 |     expect(data.players.length).toBe(2);
  316 |   });
  317 | 
  318 |   test('3. Start game', async ({ request }) => {
  319 |     expect(roomCode).toBeTruthy();
  320 |     expect(creatorID).toBeTruthy();
  321 |     const resp = await request.post(`${BASE}/api/multiplayer/start`, {
  322 |       data: { roomCode, playerID: creatorID }
  323 |     });
  324 |     expect(resp.ok()).toBe(true);
  325 |   });
  326 | 
  327 |   test('4. Guess word', async ({ request }) => {
  328 |     expect(roomCode).toBeTruthy();
  329 |     expect(creatorID).toBeTruthy();
  330 | 
  331 |     // Join to get current game state
  332 |     const joinResp = await request.post(`${BASE}/api/multiplayer/join`, {
  333 |       data: { roomCode, playerID: creatorID }
  334 |     });
  335 |     const joinData = await joinResp.json();
  336 |     expect(joinData.state).toBe('playing');
  337 |     expect(joinData.wordSequence.length).toBe(3);
  338 | 
  339 |     // Get the first word target
  340 |     const target = joinData.wordGames[0].target;
  341 | 
  342 |     const resp = await request.post(`${BASE}/api/multiplayer/guess`, {
  343 |       data: { roomCode, playerID: creatorID, word: target }
  344 |     });
  345 |     expect(resp.ok()).toBe(true);
  346 |     const data = await resp.json();
  347 |     expect(data.wordFinished).toBe(true);
  348 |     expect(data.wordWon).toBe(true);
  349 |     expect(data.playerFinished).toBe(false);
  350 |   });
  351 | 
  352 |   test('5. Invalid guess rejected', async ({ request }) => {
  353 |     expect(roomCode).toBeTruthy();
  354 |     expect(creatorID).toBeTruthy();
  355 | 
  356 |     const resp = await request.post(`${BASE}/api/multiplayer/guess`, {
  357 |       data: { roomCode, playerID: creatorID, word: 'XXXXXX' }
  358 |     });
  359 |     expect(resp.ok()).toBe(false);
  360 |   });
  361 | 
  362 |   test('6. Leave room', async ({ request }) => {
  363 |     expect(roomCode).toBeTruthy();
  364 |     expect(creatorID).toBeTruthy();
  365 | 
  366 |     const resp = await request.post(`${BASE}/api/multiplayer/leave`, {
  367 |       data: { roomCode, playerID: creatorID }
  368 |     });
  369 |     expect(resp.ok()).toBe(true);
  370 |   });
  371 | 
  372 |   test('7. Create room invalid params', async ({ request }) => {
  373 |     const resp1 = await request.post(`${BASE}/api/multiplayer/create`, {
  374 |       data: { mode: 'invalid', wordCount: 3, nickname: 'Test' }
  375 |     });
  376 |     expect(resp1.ok()).toBe(false);
```