const { test, expect } = require('@playwright/test');

const BASE = 'http://localhost:3114';

test.describe('Multiplayer Restart — New Game button restart in same lobby', () => {

  test('1. Host restart flow — game restart puts room back in lobby with same players', async ({ request }) => {
    const createResp = await request.post(`${BASE}/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 1, nickname: 'Alice' }
    });
    expect(createResp.ok()).toBe(true);
    const createData = await createResp.json();
    const roomCode = createData.roomCode;
    const aliceID = createData.playerID;
    const aliceToken = createData.token;

    const joinResp = await request.post(`${BASE}/api/multiplayer/join`, {
      data: { roomCode, nickname: 'Bob' }
    });
    expect(joinResp.ok()).toBe(true);
    const joinData = await joinResp.json();
    const bobID = joinData.playerID;
    expect(joinData.players.length).toBe(2);

    const startResp = await request.post(`${BASE}/api/multiplayer/start`, {
      data: { roomCode, playerID: aliceID }
    });
    expect(startResp.ok()).toBe(true);

    const aliceStateResp = await request.post(`${BASE}/api/multiplayer/join`, {
      data: { roomCode, playerID: aliceID }
    });
    const aliceState = await aliceStateResp.json();
    expect(aliceState.state).toBe('playing');
    expect(aliceState.wordGames).toBeDefined();
    expect(aliceState.wordGames.length).toBe(1);
    const targetWord = aliceState.wordGames[0].target;

    const aliceGuessResp = await request.post(`${BASE}/api/multiplayer/guess`, {
      data: { roomCode, playerID: aliceID, word: targetWord }
    });
    expect(aliceGuessResp.ok()).toBe(true);

    const bobStateResp = await request.post(`${BASE}/api/multiplayer/join`, {
      data: { roomCode, playerID: bobID }
    });
    const bobState = await bobStateResp.json();
    expect(bobState.state).toBe('playing');
    const bobTarget = bobState.wordGames[0].target;

    const bobGuessResp = await request.post(`${BASE}/api/multiplayer/guess`, {
      data: { roomCode, playerID: bobID, word: bobTarget }
    });
    expect(bobGuessResp.ok()).toBe(true);

    const restartResp = await request.post(`${BASE}/api/multiplayer/restart`, {
      data: { roomCode, playerID: aliceID, token: aliceToken }
    });
    expect(restartResp.ok()).toBe(true);

    const afterRestartResp = await request.post(`${BASE}/api/multiplayer/join`, {
      data: { roomCode, playerID: aliceID }
    });
    const afterRestart = await afterRestartResp.json();
    expect(afterRestart.state).toBe('lobby');
    expect(afterRestart.players.length).toBe(2);
    expect(afterRestart.mode).toBe('progressif');
    expect(afterRestart.wordCount).toBe(1);
  });

  test('2. Non-host cannot restart — Bob gets 403', async ({ request }) => {
    const createResp = await request.post(`${BASE}/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 1, nickname: 'Alice' }
    });
    const createData = await createResp.json();
    const roomCode = createData.roomCode;
    const aliceID = createData.playerID;

    const joinResp = await request.post(`${BASE}/api/multiplayer/join`, {
      data: { roomCode, nickname: 'Bob' }
    });
    const bobData = await joinResp.json();
    const bobID = bobData.playerID;

    await request.post(`${BASE}/api/multiplayer/start`, {
      data: { roomCode, playerID: aliceID }
    });

    const restartResp = await request.post(`${BASE}/api/multiplayer/restart`, {
      data: { roomCode, playerID: bobID }
    });
    expect(restartResp.status()).toBe(403);
  });

  test('3. Player can leave after restart — Bob leaves, Alice remains', async ({ request }) => {
    const createResp = await request.post(`${BASE}/api/multiplayer/create`, {
      data: { mode: 'progressif', wordCount: 1, nickname: 'Alice' }
    });
    const createData = await createResp.json();
    const roomCode = createData.roomCode;
    const aliceID = createData.playerID;
    const aliceToken = createData.token;

    const joinResp = await request.post(`${BASE}/api/multiplayer/join`, {
      data: { roomCode, nickname: 'Bob' }
    });
    const bobData = await joinResp.json();
    const bobID = bobData.playerID;

    await request.post(`${BASE}/api/multiplayer/start`, {
      data: { roomCode, playerID: aliceID }
    });

    const aliceStateResp = await request.post(`${BASE}/api/multiplayer/join`, {
      data: { roomCode, playerID: aliceID }
    });
    const aliceState = await aliceStateResp.json();
    const targetWord = aliceState.wordGames[0].target;

    await request.post(`${BASE}/api/multiplayer/guess`, {
      data: { roomCode, playerID: aliceID, word: targetWord }
    });

    const bobStateResp = await request.post(`${BASE}/api/multiplayer/join`, {
      data: { roomCode, playerID: bobID }
    });
    const bobState = await bobStateResp.json();
    const bobTarget = bobState.wordGames[0].target;

    await request.post(`${BASE}/api/multiplayer/guess`, {
      data: { roomCode, playerID: bobID, word: bobTarget }
    });

    await request.post(`${BASE}/api/multiplayer/restart`, {
      data: { roomCode, playerID: aliceID, token: aliceToken }
    });

    const leaveResp = await request.post(`${BASE}/api/multiplayer/leave`, {
      data: { roomCode, playerID: bobID }
    });
    expect(leaveResp.ok()).toBe(true);

    const aliceAfterResp = await request.post(`${BASE}/api/multiplayer/join`, {
      data: { roomCode, playerID: aliceID }
    });
    expect(aliceAfterResp.ok()).toBe(true);
    const aliceAfter = await aliceAfterResp.json();
    expect(aliceAfter.state).toBe('lobby');
    expect(aliceAfter.players.length).toBe(1);
  });

});
