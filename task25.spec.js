const { test, expect, chromium } = require('@playwright/test');

const BASE = 'http://localhost:3125';

test('no #game-timer in DOM on multiplayer page', async ({ page }) => {
    await page.goto(`${BASE}/multiplayer`);
    const timerCount = await page.locator('#game-timer').count();
    expect(timerCount).toBe(0);
});

test('renderRankings uses M:SS.mmm format', async ({ request }) => {
    // Verify by checking the JS source on the server
    const resp = await request.get(`${BASE}/static/multiplayer.js`);
    const src = await resp.text();

    // Should NOT have old format
    expect(src).not.toContain('startTimer');
    expect(src).not.toContain('game-timer');
    expect(src).not.toContain('r.time / 1e9');

    // Should have new format
    expect(src).toContain('r.time / 1e6');
    expect(src).toContain('totalMs % 1000');
    expect(src).toMatch(/\$\{m\}:\$\{s\}\.\$\{ms\}/);
});
