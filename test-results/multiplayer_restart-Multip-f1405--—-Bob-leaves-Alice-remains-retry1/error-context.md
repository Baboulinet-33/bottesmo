# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: multiplayer_restart.spec.js >> Multiplayer Restart — New Game button restart in same lobby >> 3. Player can leave after restart — Bob leaves, Alice remains
- Location: multiplayer_restart.spec.js:93:3

# Error details

```
TypeError: Cannot read properties of undefined (reading '0')
```

# Test source

```ts
  16  |       data: { roomCode, nickname: 'Bob' }
  17  |     });
  18  |     expect(joinResp.ok()).toBe(true);
  19  |     const joinData = await joinResp.json();
  20  |     const bobID = joinData.playerID;
  21  |     expect(joinData.players.length).toBe(2);
  22  | 
  23  |     const startResp = await request.post(`/api/multiplayer/start`, {
  24  |       data: { roomCode, playerID: aliceID }
  25  |     });
  26  |     expect(startResp.ok()).toBe(true);
  27  | 
  28  |     const aliceStateResp = await request.post(`/api/multiplayer/join`, {
  29  |       data: { roomCode, playerID: aliceID }
  30  |     });
  31  |     const aliceState = await aliceStateResp.json();
  32  |     expect(aliceState.state).toBe('playing');
  33  |     expect(aliceState.wordGames).toBeDefined();
  34  |     expect(aliceState.wordGames.length).toBe(1);
  35  |     const targetWord = aliceState.wordGames[0].target;
  36  | 
  37  |     const aliceGuessResp = await request.post(`/api/multiplayer/guess`, {
  38  |       data: { roomCode, playerID: aliceID, word: targetWord }
  39  |     });
  40  |     expect(aliceGuessResp.ok()).toBe(true);
  41  | 
  42  |     const bobStateResp = await request.post(`/api/multiplayer/join`, {
  43  |       data: { roomCode, playerID: bobID }
  44  |     });
  45  |     const bobState = await bobStateResp.json();
  46  |     expect(bobState.state).toBe('playing');
  47  |     const bobTarget = bobState.wordGames[0].target;
  48  | 
  49  |     const bobGuessResp = await request.post(`/api/multiplayer/guess`, {
  50  |       data: { roomCode, playerID: bobID, word: bobTarget }
  51  |     });
  52  |     expect(bobGuessResp.ok()).toBe(true);
  53  | 
  54  |     const restartResp = await request.post(`/api/multiplayer/restart`, {
  55  |       data: { roomCode, playerID: aliceID, token: aliceToken }
  56  |     });
  57  |     expect(restartResp.ok()).toBe(true);
  58  | 
  59  |     const afterRestartResp = await request.post(`/api/multiplayer/join`, {
  60  |       data: { roomCode, playerID: aliceID }
  61  |     });
  62  |     const afterRestart = await afterRestartResp.json();
  63  |     expect(afterRestart.state).toBe('lobby');
  64  |     expect(afterRestart.players.length).toBe(2);
  65  |     expect(afterRestart.mode).toBe('progressif');
  66  |     expect(afterRestart.wordCount).toBe(1);
  67  |   });
  68  | 
  69  |   test('2. Non-host cannot restart — Bob gets 403', async ({ request }) => {
  70  |     const createResp = await request.post(`/api/multiplayer/create`, {
  71  |       data: { mode: 'progressif', wordCount: 1, nickname: 'Alice' }
  72  |     });
  73  |     const createData = await createResp.json();
  74  |     const roomCode = createData.roomCode;
  75  |     const aliceID = createData.playerID;
  76  | 
  77  |     const joinResp = await request.post(`/api/multiplayer/join`, {
  78  |       data: { roomCode, nickname: 'Bob' }
  79  |     });
  80  |     const bobData = await joinResp.json();
  81  |     const bobID = bobData.playerID;
  82  | 
  83  |     await request.post(`/api/multiplayer/start`, {
  84  |       data: { roomCode, playerID: aliceID }
  85  |     });
  86  | 
  87  |     const restartResp = await request.post(`/api/multiplayer/restart`, {
  88  |       data: { roomCode, playerID: bobID }
  89  |     });
  90  |     expect(restartResp.status()).toBe(403);
  91  |   });
  92  | 
  93  |   test('3. Player can leave after restart — Bob leaves, Alice remains', async ({ request }) => {
  94  |     const createResp = await request.post(`/api/multiplayer/create`, {
  95  |       data: { mode: 'progressif', wordCount: 1, nickname: 'Alice' }
  96  |     });
  97  |     const createData = await createResp.json();
  98  |     const roomCode = createData.roomCode;
  99  |     const aliceID = createData.playerID;
  100 |     const aliceToken = createData.token;
  101 | 
  102 |     const joinResp = await request.post(`/api/multiplayer/join`, {
  103 |       data: { roomCode, nickname: 'Bob' }
  104 |     });
  105 |     const bobData = await joinResp.json();
  106 |     const bobID = bobData.playerID;
  107 | 
  108 |     await request.post(`/api/multiplayer/start`, {
  109 |       data: { roomCode, playerID: aliceID }
  110 |     });
  111 | 
  112 |     const aliceStateResp = await request.post(`/api/multiplayer/join`, {
  113 |       data: { roomCode, playerID: aliceID }
  114 |     });
  115 |     const aliceState = await aliceStateResp.json();
> 116 |     const targetWord = aliceState.wordGames[0].target;
      |                                            ^ TypeError: Cannot read properties of undefined (reading '0')
  117 | 
  118 |     await request.post(`/api/multiplayer/guess`, {
  119 |       data: { roomCode, playerID: aliceID, word: targetWord }
  120 |     });
  121 | 
  122 |     const bobStateResp = await request.post(`/api/multiplayer/join`, {
  123 |       data: { roomCode, playerID: bobID }
  124 |     });
  125 |     const bobState = await bobStateResp.json();
  126 |     const bobTarget = bobState.wordGames[0].target;
  127 | 
  128 |     await request.post(`/api/multiplayer/guess`, {
  129 |       data: { roomCode, playerID: bobID, word: bobTarget }
  130 |     });
  131 | 
  132 |     await request.post(`/api/multiplayer/restart`, {
  133 |       data: { roomCode, playerID: aliceID, token: aliceToken }
  134 |     });
  135 | 
  136 |     const leaveResp = await request.post(`/api/multiplayer/leave`, {
  137 |       data: { roomCode, playerID: bobID }
  138 |     });
  139 |     expect(leaveResp.ok()).toBe(true);
  140 | 
  141 |     const aliceAfterResp = await request.post(`/api/multiplayer/join`, {
  142 |       data: { roomCode, playerID: aliceID }
  143 |     });
  144 |     expect(aliceAfterResp.ok()).toBe(true);
  145 |     const aliceAfter = await aliceAfterResp.json();
  146 |     expect(aliceAfter.state).toBe('lobby');
  147 |     expect(aliceAfter.players.length).toBe(1);
  148 |   });
  149 | 
  150 | });
  151 | 
```