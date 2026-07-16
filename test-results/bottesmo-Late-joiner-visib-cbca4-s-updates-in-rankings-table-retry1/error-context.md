# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: bottesmo.spec.js >> Late joiner visibility on results screen >> 10. Finished player sees late joiner and progress updates in rankings table
- Location: bottesmo.spec.js:705:3

# Error details

```
Error: apiRequestContext.post: connect ECONNREFUSED ::1:3131
Call log:
  - → POST http://localhost:3131/api/multiplayer/create
    - user-agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/149.0.7827.55 Safari/537.36
    - accept: */*
    - accept-encoding: gzip,deflate,br
    - content-type: application/json
    - content-length: 54

```