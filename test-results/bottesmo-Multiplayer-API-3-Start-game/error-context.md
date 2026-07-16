# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: bottesmo.spec.js >> Multiplayer API >> 3. Start game
- Location: bottesmo.spec.js:322:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: true
Received: false
```

# Test source

```ts
  228 | 
  229 |     const startResp = await request.post(`/api/multiplayer/start`, {
  230 |       headers: { 'X-Player-Token': aliceToken },
  231 |       data: { roomCode: code, playerID: aliceID }
  232 |     });
  233 |     expect(startResp.ok()).toBe(true);
  234 | 
  235 |     const joinResp = await request.post(`/api/multiplayer/join`, {
  236 |       data: { roomCode: code, nickname: 'Bob' }
  237 |     });
  238 |     expect(joinResp.ok()).toBe(true);
  239 |     const joinData = await joinResp.json();
  240 |     expect(joinData.state).toBe('playing');
  241 |     expect(joinData.wordGames).toBeDefined();
  242 |     expect(joinData.wordGames.length).toBe(2);
  243 |     expect(joinData.wordGames[0]).toBeDefined();
  244 |     expect(joinData.wordGames[0].target).toBeTruthy();
  245 |   });
  246 | 
  247 |   test('3. Double-click start does not produce 400 on UI (button disabled)', async ({ request }) => {
  248 |     const createResp = await request.post(`/api/multiplayer/create`, {
  249 |       data: { mode: 'progressif', wordCount: 2, nickname: 'Charlie' }
  250 |     });
  251 |     const createData = await createResp.json();
  252 |     const code = createData.roomCode;
  253 |     const playerID = createData.playerID;
  254 |     const playerToken = createData.token;
  255 | 
  256 |     const firstResp = await request.post(`/api/multiplayer/start`, {
  257 |       headers: { 'X-Player-Token': playerToken },
  258 |       data: { roomCode: code, playerID }
  259 |     });
  260 |     expect(firstResp.ok()).toBe(true);
  261 | 
  262 |     const secondResp = await request.post(`/api/multiplayer/start`, {
  263 |       headers: { 'X-Player-Token': playerToken },
  264 |       data: { roomCode: code, playerID }
  265 |     });
  266 |     expect(secondResp.status()).toBe(400);
  267 |     const secondData = await secondResp.json();
  268 |     expect(secondData.error).toBeTruthy();
  269 |   });
  270 | 
  271 |   test('4. SSE endpoint returns text/event-stream content type', async ({ page, request }) => {
  272 |     const createResp = await request.post('/api/multiplayer/create', {
  273 |       data: { mode: 'progressif', wordCount: 2, nickname: 'Diana' }
  274 |     });
  275 |     const createData = await createResp.json();
  276 |     const code = createData.roomCode;
  277 |     const playerID = createData.playerID;
  278 |     const playerToken = createData.token;
  279 | 
  280 |     // SSE streams indefinitely, so request.get() would hang.
  281 |     // Use page.evaluate with fetch + AbortController to read just the headers.
  282 |     await page.goto('/');
  283 |     const contentType = await page.evaluate(async (url) => {
  284 |       const ctrl = new AbortController();
  285 |       const resp = await fetch(url, { signal: ctrl.signal });
  286 |       const ct = resp.headers.get('content-type');
  287 |       ctrl.abort();
  288 |       return ct;
  289 |     }, `/api/multiplayer/events?room=${code}&player=${playerID}&token=${playerToken}`);
  290 |     expect(contentType).toContain('text/event-stream');
  291 |   });
  292 | });
  293 | 
  294 | test.describe('Multiplayer API', () => {
  295 |   let roomCode = '';
  296 |   let creatorID = '';
  297 | 
  298 |   test('1. Create room', async ({ request }) => {
  299 |     const resp = await request.post(`/api/multiplayer/create`, {
  300 |       data: { mode: 'progressif', wordCount: 3, nickname: 'Alice' }
  301 |     });
  302 |     expect(resp.ok()).toBe(true);
  303 |     const data = await resp.json();
  304 |     expect(data).toHaveProperty('roomCode');
  305 |     expect(data).toHaveProperty('playerID');
  306 |     expect(data.roomCode.length).toBe(6);
  307 |     roomCode = data.roomCode;
  308 |     creatorID = data.playerID;
  309 |   });
  310 | 
  311 |   test('2. Join room', async ({ request }) => {
  312 |     expect(roomCode).toBeTruthy();
  313 |     const resp = await request.post(`/api/multiplayer/join`, {
  314 |       data: { roomCode, nickname: 'Bob' }
  315 |     });
  316 |     expect(resp.ok()).toBe(true);
  317 |     const data = await resp.json();
  318 |     expect(data.state).toBe('lobby');
  319 |     expect(data.players.length).toBe(2);
  320 |   });
  321 | 
  322 |   test('3. Start game', async ({ request }) => {
  323 |     expect(roomCode).toBeTruthy();
  324 |     expect(creatorID).toBeTruthy();
  325 |     const resp = await request.post(`/api/multiplayer/start`, {
  326 |       data: { roomCode, playerID: creatorID }
  327 |     });
> 328 |     expect(resp.ok()).toBe(true);
      |                       ^ Error: expect(received).toBe(expected) // Object.is equality
  329 |   });
  330 | 
  331 |   test('4. Guess word', async ({ request }) => {
  332 |     expect(roomCode).toBeTruthy();
  333 |     expect(creatorID).toBeTruthy();
  334 | 
  335 |     // Join to get current game state
  336 |     const joinResp = await request.post(`/api/multiplayer/join`, {
  337 |       data: { roomCode, playerID: creatorID }
  338 |     });
  339 |     const joinData = await joinResp.json();
  340 |     expect(joinData.state).toBe('playing');
  341 |     expect(joinData.wordSequence.length).toBe(3);
  342 | 
  343 |     // Get the first word target
  344 |     const target = joinData.wordGames[0].target;
  345 | 
  346 |     const resp = await request.post(`/api/multiplayer/guess`, {
  347 |       data: { roomCode, playerID: creatorID, word: target }
  348 |     });
  349 |     expect(resp.ok()).toBe(true);
  350 |     const data = await resp.json();
  351 |     expect(data.wordFinished).toBe(true);
  352 |     expect(data.wordWon).toBe(true);
  353 |     expect(data.playerFinished).toBe(false);
  354 |   });
  355 | 
  356 |   test('5. Invalid guess rejected', async ({ request }) => {
  357 |     expect(roomCode).toBeTruthy();
  358 |     expect(creatorID).toBeTruthy();
  359 | 
  360 |     const resp = await request.post(`/api/multiplayer/guess`, {
  361 |       data: { roomCode, playerID: creatorID, word: 'XXXXXX' }
  362 |     });
  363 |     expect(resp.ok()).toBe(false);
  364 |   });
  365 | 
  366 |   test('6. Leave room', async ({ request }) => {
  367 |     expect(roomCode).toBeTruthy();
  368 |     expect(creatorID).toBeTruthy();
  369 | 
  370 |     const resp = await request.post(`/api/multiplayer/leave`, {
  371 |       data: { roomCode, playerID: creatorID }
  372 |     });
  373 |     expect(resp.ok()).toBe(true);
  374 |   });
  375 | 
  376 |   test('7. Create room invalid params', async ({ request }) => {
  377 |     const resp1 = await request.post(`/api/multiplayer/create`, {
  378 |       data: { mode: 'invalid', wordCount: 3, nickname: 'Test' }
  379 |     });
  380 |     expect(resp1.ok()).toBe(false);
  381 | 
  382 |     const resp2 = await request.post(`/api/multiplayer/create`, {
  383 |       data: { mode: 'progressif', wordCount: 0, nickname: 'Test' }
  384 |     });
  385 |     expect(resp2.ok()).toBe(false);
  386 | 
  387 |     const resp3 = await request.post(`/api/multiplayer/create`, {
  388 |       data: { mode: 'progressif', wordCount: 3, nickname: '' }
  389 |     });
  390 |     expect(resp3.ok()).toBe(false);
  391 |   });
  392 | 
  393 |   test('8. Finished player has wordResults in rankings', async ({ request }) => {
  394 |     const createResp = await request.post(`/api/multiplayer/create`, {
  395 |       data: { mode: 'progressif', wordCount: 2, nickname: 'Alice' }
  396 |     });
  397 |     const createData = await createResp.json();
  398 |     const code = createData.roomCode;
  399 |     const aliceID = createData.playerID;
  400 | 
  401 |     const startResp = await request.post(`/api/multiplayer/start`, {
  402 |       data: { roomCode: code, playerID: aliceID }
  403 |     });
  404 |     expect(startResp.ok()).toBe(true);
  405 | 
  406 |     const joinResp = await request.post(`/api/multiplayer/join`, {
  407 |       data: { roomCode: code, playerID: aliceID }
  408 |     });
  409 |     const joinData = await joinResp.json();
  410 |     expect(joinData.wordGames.length).toBe(2);
  411 | 
  412 |     let lastResponse = null;
  413 |     for (let i = 0; i < 2; i++) {
  414 |       const target = joinData.wordGames[i].target;
  415 |       const guessResp = await request.post(`/api/multiplayer/guess`, {
  416 |         data: { roomCode: code, playerID: aliceID, word: target }
  417 |       });
  418 |       expect(guessResp.ok()).toBe(true);
  419 |       lastResponse = await guessResp.json();
  420 |     }
  421 | 
  422 |     expect(lastResponse.playerFinished).toBe(true);
  423 |     expect(lastResponse.rankings).toBeDefined();
  424 |     expect(Array.isArray(lastResponse.rankings)).toBe(true);
  425 | 
  426 |     const aliceRanking = lastResponse.rankings.find(r => r.playerID === aliceID);
  427 |     expect(aliceRanking).toBeDefined();
  428 |     expect(aliceRanking.wordResults).toBeDefined();
```