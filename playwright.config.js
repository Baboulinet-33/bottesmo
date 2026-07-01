const { defineConfig } = require('@playwright/test');

module.exports = defineConfig({
  testMatch: ['tusmo.spec.js', 'multiplayer_restart.spec.js'],
  use: {
    baseURL: 'http://localhost:3106',
    headless: true,
  },
  timeout: 30000,
  retries: 1,
});
