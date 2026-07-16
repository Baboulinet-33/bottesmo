# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: bottesmo.spec.js >> Multiplayer API >> 9. Results screen persists after player-finished SSE broadcast
- Location: bottesmo.spec.js:441:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: true
Received: false
```

# Test source

```ts
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
  429 |     expect(aliceRanking.wordResults.length).toBe(2);
  430 | 
  431 |     for (const wr of aliceRanking.wordResults) {
  432 |       expect(wr).toHaveProperty('target');
  433 |       expect(wr).toHaveProperty('attempts');
  434 |       expect(wr).toHaveProperty('results');
  435 |       expect(typeof wr.target).toBe('string');
  436 |       expect(Array.isArray(wr.attempts)).toBe(true);
  437 |       expect(Array.isArray(wr.results)).toBe(true);
  438 |     }
  439 |   });
  440 | 
  441 |   test('9. Results screen persists after player-finished SSE broadcast', async ({ page, request }) => {
  442 |     // Create room with Alice (solo — only one player so the game can start)
  443 |     const createResp = await request.post(`/api/multiplayer/create`, {
  444 |       data: { mode: 'progressif', wordCount: 2, nickname: 'Alice' }
  445 |     });
  446 |     const createData = await createResp.json();
  447 |     const code = createData.roomCode;
  448 |     const aliceID = createData.playerID;
  449 | 
  450 |     // Start and join the game so we can retrieve word targets for the API guesses
  451 |     const startResp = await request.post(`/api/multiplayer/start`, {
  452 |       data: { roomCode: code, playerID: aliceID }
  453 |     });
> 454 |     expect(startResp.ok()).toBe(true);
      |                            ^ Error: expect(received).toBe(expected) // Object.is equality
  455 | 
  456 |     const joinResp = await request.post(`/api/multiplayer/join`, {
  457 |       data: { roomCode: code, playerID: aliceID }
  458 |     });
  459 |     const joinData = await joinResp.json();
  460 |     expect(joinData.wordGames.length).toBe(2);
  461 | 
  462 |     // Submit all winning guesses via API (simulates Alice finishing)
  463 |     let lastGuessData = null;
  464 |     for (let i = 0; i < 2; i++) {
  465 |       const target = joinData.wordGames[i].target;
  466 |       const guessResp = await request.post(`/api/multiplayer/guess`, {
  467 |         data: { roomCode: code, playerID: aliceID, word: target }
  468 |       });
  469 |       expect(guessResp.ok()).toBe(true);
  470 |       lastGuessData = await guessResp.json();
  471 |     }
  472 | 
  473 |     // Navigate to /multiplayer and inject state via evaluate to verify DOM behaviour
  474 |     await page.goto(`/multiplayer`);
  475 | 
  476 |     // Inject the rankings from the HTTP response directly into renderRankings (simulating what
  477 |     // the real client does after POST /api/multiplayer/guess returns playerFinished=true)
  478 |     const rankings = lastGuessData.rankings;
  479 |     await page.evaluate((rankings) => {
  480 |       // Show the results screen
  481 |       document.querySelectorAll('.screen').forEach(s => s.style.display = 'none');
  482 |       const resultsScreen = document.getElementById('screen-results');
  483 |       if (resultsScreen) resultsScreen.style.display = '';
  484 | 
  485 |       // Call renderRankings with the full payload (includes wordResults)
  486 |       if (typeof renderRankings === 'function') {
  487 |         renderRankings(rankings);
  488 |       }
  489 |     }, rankings);
  490 | 
  491 |     // Assert results screen is visible
  492 |     const resultsScreen = page.locator('#screen-results');
  493 |     await expect(resultsScreen).toBeVisible();
  494 | 
  495 |     // Assert player tabs and word results are populated
  496 |     await expect(page.locator('#player-tabs')).not.toBeEmpty();
  497 |     await expect(page.locator('#player-word-results')).not.toBeEmpty();
  498 |     const initialTabCount = await page.locator('#player-tabs .player-tab').count();
  499 |     expect(initialTabCount).toBeGreaterThanOrEqual(1);
  500 |     const initialCardCount = await page.locator('#player-word-results .word-result').count();
  501 |     expect(initialCardCount).toBeGreaterThanOrEqual(1);
  502 | 
  503 |     // Now simulate the `player-finished` SSE broadcast arriving with stripped rankings
  504 |     // (no wordResults) — this is what used to wipe the detail view
  505 |     const strippedRankings = rankings.map(r => ({ ...r, wordResults: undefined }));
  506 |     await page.evaluate((strippedRankings) => {
  507 |       if (typeof renderRankings === 'function') {
  508 |         renderRankings(strippedRankings);
  509 |       }
  510 |     }, strippedRankings);
  511 | 
  512 |     // Re-assert — under the old code tabs/word-results would be empty; with the fix they persist
  513 |     await expect(resultsScreen).toBeVisible();
  514 |     const tabCountAfter = await page.locator('#player-tabs .player-tab').count();
  515 |     expect(tabCountAfter).toBeGreaterThanOrEqual(1);
  516 |     const cardCountAfter = await page.locator('#player-word-results .word-result').count();
  517 |     expect(cardCountAfter).toBeGreaterThanOrEqual(1);
  518 |   });
  519 | });
  520 | 
  521 | test.describe('GET /api/status', () => {
  522 |   test('1. Returns 200 with status, version, and dictionaries', async ({ request }) => {
  523 |     const resp = await request.get(`/api/status`);
  524 |     expect(resp.ok()).toBe(true);
  525 |     const data = await resp.json();
  526 | 
  527 |     expect(data.status).toBe('ok');
  528 |     expect(typeof data.version).toBe('string');
  529 |     expect(data.version.length).toBeGreaterThan(0);
  530 |     expect(Array.isArray(data.dictionaries)).toBe(true);
  531 |     expect(data.dictionaries.length).toBe(2);
  532 |   });
  533 | 
  534 |   test('2. Dictionaries have name, word_count, sha256', async ({ request }) => {
  535 |     const resp = await request.get(`/api/status`);
  536 |     const data = await resp.json();
  537 | 
  538 |     for (const dict of data.dictionaries) {
  539 |       expect(dict).toHaveProperty('name');
  540 |       expect(dict).toHaveProperty('word_count');
  541 |       expect(dict).toHaveProperty('sha256');
  542 |       expect(typeof dict.name).toBe('string');
  543 |       expect(dict.word_count).toBeGreaterThan(0);
  544 |       expect(dict.sha256.length).toBeGreaterThan(0);
  545 |     }
  546 |   });
  547 | 
  548 |   test('3. Dictionaries include words and words_full', async ({ request }) => {
  549 |     const resp = await request.get(`/api/status`);
  550 |     const data = await resp.json();
  551 | 
  552 |     const names = data.dictionaries.map(d => d.name);
  553 |     expect(names).toContain('words');
  554 |     expect(names).toContain('words_full');
```