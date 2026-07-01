const STATUS_CLASSES = ['correct', 'present', 'absent'];
let mp = newMp();

function newMp() {
    return {
        playerID: null, token: null, roomCode: null, mode: null, wordCount: null,
        state: null, creatorID: null, players: [], wordSequence: [],
        wordGames: [], currentWordIdx: 0, currentRow: 0, currentCol: 0,
        foundLetters: [], lockedPositions: new Set(), letterStatuses: {},
        attempts: [], won: false, gameOver: false, wordLength: 0,
        firstLetter: '', maxTries: 6, guessResults: [], eventSource: null,
    };
}

document.addEventListener('DOMContentLoaded', () => {
    const multiDiv = document.getElementById('multiplayer');
    if (!multiDiv) return;

    const params = new URLSearchParams(window.location.search);
    const joinCode = params.get('join');

    const savedNick = localStorage.getItem('tusmo-multi-nickname');
    if (savedNick) {
        document.getElementById('create-nickname').value = savedNick;
        document.getElementById('join-nickname').value = savedNick;
    }

    if (joinCode) {
        document.getElementById('join-code').value = joinCode;
        document.querySelector('#screen-create .multi-form:first-of-type').style.display = 'none';
        document.querySelector('#screen-create .multi-divider').style.display = 'none';
    }

    setupCreateForm();
    setupJoinForm();

    window.addEventListener('beforeunload', () => {
        if (mp.roomCode && mp.playerID) {
            const blob = new Blob([JSON.stringify({
                roomCode: mp.roomCode,
                playerID: mp.playerID
            })], { type: 'application/json' });
            navigator.sendBeacon('/api/multiplayer/leave', blob);
        }
    });
});

function showScreen(screen) {
    document.querySelectorAll('.multi-screen').forEach(s => s.style.display = 'none');
    document.getElementById('screen-' + screen).style.display = 'block';
}

function setupCreateForm() {
    document.getElementById('create-btn').addEventListener('click', createRoom);
}

function setupJoinForm() {
    document.getElementById('join-btn').addEventListener('click', () => {
        const code = document.getElementById('join-code').value.trim();
        const nickname = document.getElementById('join-nickname').value.trim();
        if (!code || !nickname) {
            alert('Veuillez remplir tous les champs');
            return;
        }
        joinRoom(code, nickname);
    });
    document.getElementById('leave-lobby-btn').addEventListener('click', leaveRoom);
    document.getElementById('leave-results-btn').addEventListener('click', leaveRoom);
    document.getElementById('new-game-btn').addEventListener('click', newGame);
    document.getElementById('start-game-btn').addEventListener('click', startGame);
    document.getElementById('copy-btn').addEventListener('click', copyShareLink);
}

function createRoom() {
    const mode = document.getElementById('create-mode').value;
    const wordCount = parseInt(document.getElementById('create-wordcount').value);
    const nickname = document.getElementById('create-nickname').value.trim();

    if (!nickname) {
        alert('Veuillez entrer un pseudo');
        return;
    }

    localStorage.setItem('tusmo-multi-nickname', nickname);

    fetch('/api/multiplayer/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ mode, wordCount, nickname })
    })
    .then(res => res.json())
    .then(data => {
        if (data.error) { alert(data.error); return; }
        mp.playerID = data.playerID;
        mp.token = data.token;
        mp.roomCode = data.roomCode;
        mp.mode = mode;
        mp.wordCount = wordCount;
        mp.creatorID = data.playerID;

        updateLobbyInfo(data.roomCode, mode, wordCount);
        document.getElementById('start-game-btn').style.display = 'block';

        showScreen('lobby');
        setupSSE(data.roomCode, data.playerID);
        updateLobbyPlayers();
    })
    .catch(err => alert('Erreur de connexion'));
}

function joinRoom(code, nickname) {
    localStorage.setItem('tusmo-multi-nickname', nickname);
    const savedID = localStorage.getItem('tusmo-multi-playerid');

    fetch('/api/multiplayer/join', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            roomCode: code,
            nickname: nickname,
            playerID: savedID || ''
        })
    })
    .then(res => res.json())
    .then(data => {
        if (data.error) { alert(data.error); return; }
        mp.playerID = data.playerID;
        mp.token = data.token;
        mp.roomCode = data.roomCode;
        mp.mode = data.mode;
        mp.wordCount = data.wordCount;
        mp.creatorID = data.creatorID;
        mp.players = data.players;

        localStorage.setItem('tusmo-multi-playerid', data.playerID);

        updateLobbyInfo(data.roomCode, data.mode, data.wordCount);

        if (data.creatorID === data.playerID) {
            document.getElementById('start-game-btn').style.display = 'block';
        }

        showScreen('lobby');
        setupSSE(data.roomCode, data.playerID);
        updateLobbyPlayers();

        if (data.state === 'playing') {
            restoreGameState(data);
            showScreen('game');
            startTimer();
        }
    })
    .catch(err => alert('Erreur de connexion'));
}

function restoreGameState(data) {
    mp.wordSequence = data.wordSequence || [];
    mp.currentWordIdx = data.currentWordIdx || 0;
    mp.wordGames = data.wordGames || [];

    loadCurrentWord();
}

function loadCurrentWord() {
    const wg = mp.wordGames[mp.currentWordIdx];
    if (!wg) return;

    mp.wordLength = wg.wordLength;
    mp.firstLetter = wg.target[0];
    mp.maxTries = wg.maxTries;
    mp.attempts = wg.attempts || [];
    mp.won = wg.won;
    mp.gameOver = wg.gameOver;
    mp.guessResults = wg.results || [];
    mp.currentRow = mp.attempts.length;
    mp.foundLetters = [];
    mp.lockedPositions = new Set([0]);
    mp.letterStatuses = {};

    initGrid(mp.wordLength, mp.firstLetter);
    renderKeyboard();

    if (mp.attempts.length > 0) {
        for (let row = 0; row < mp.attempts.length; row++) {
            updateGridRow(row, mp.guessResults[row] || []);
        }
        if (!mp.gameOver) {
            prepareNextRow();
            enableInput();
        }
    }

    document.getElementById('multi-submit-btn').disabled = mp.gameOver;
    updateWordIndicator();

    mp.currentCol = firstUnlockedPosition(0);
    addCursor();
}

function startGame() {
    const btn = document.getElementById('start-game-btn');
    btn.disabled = true;

    fetch('/api/multiplayer/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ roomCode: mp.roomCode, playerID: mp.playerID })
    })
    .then(res => res.json())
    .then(data => {
        if (data.error) {
            alert(data.error);
            btn.disabled = false;
            return;
        }
        if (data.success) {
            loadGame();
        }
    })
    .catch(err => {
        alert('Erreur de connexion');
        btn.disabled = false;
    });
}

function setupSSE(roomCode, playerID) {
    if (mp.eventSource) {
        mp.eventSource.close();
    }

    const url = '/api/multiplayer/events?room=' + encodeURIComponent(roomCode) + '&player=' + encodeURIComponent(playerID);
    mp.eventSource = new EventSource(url);

    mp.eventSource.addEventListener('player-joined', (e) => {
        const data = JSON.parse(e.data);
        mp.players = data.players;
        if (document.getElementById('screen-lobby').style.display !== 'none') {
            updateLobbyPlayers();
        }
        updateProgressPlayers();
    });

    mp.eventSource.addEventListener('player-left', (e) => {
        const data = JSON.parse(e.data);
        mp.players = data.players;
        if (data.newCreatorID) {
            mp.creatorID = data.newCreatorID;
            if (mp.creatorID === mp.playerID) {
                document.getElementById('start-game-btn').style.display = 'block';
            }
        }
        if (document.getElementById('screen-lobby').style.display !== 'none') {
            updateLobbyPlayers();
        }
        updateProgressPlayers();
    });

    mp.eventSource.addEventListener('game-started', (e) => {
        const data = JSON.parse(e.data);
        mp.players = data.players;
        showScreen('game');
        startTimer();
        loadGame();
    });

    mp.eventSource.addEventListener('progress', (e) => {
        const data = JSON.parse(e.data);
        mp.players = data.players;
        updateProgressPlayers();
    });

    mp.eventSource.addEventListener('player-finished', (e) => {
        const data = JSON.parse(e.data);
        renderRankings(data.rankings);
        if (data.playerID === mp.playerID) {
            showScreen('results');
            updateNewGameBtn();
        }
    });

    mp.eventSource.addEventListener('game-over', (e) => {
        const data = JSON.parse(e.data);
        renderRankings(data.rankings);
        showScreen('results');
        updateNewGameBtn();
    });

    mp.eventSource.addEventListener('game-restarted', (e) => {
        const data = JSON.parse(e.data);
        mp.players = data.players;
        updateLobbyPlayers();
        if (mp.creatorID === mp.playerID) {
            document.getElementById('start-game-btn').style.display = 'block';
            document.getElementById('start-game-btn').disabled = false;
        }
        showScreen('lobby');
    });

    mp.eventSource.onerror = () => {
        setTimeout(() => setupSSE(roomCode, playerID), 3000);
    };
}

function fetchGameState() {
    return fetch('/api/multiplayer/join', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            roomCode: mp.roomCode,
            nickname: '',
            playerID: mp.playerID
        })
    }).then(res => res.json());
}

function loadGame() {
    fetchGameState().then(data => {
        if (data.error) { return; }
        mp.wordSequence = data.wordSequence || [];
        mp.currentWordIdx = data.currentWordIdx || 0;
        mp.wordGames = data.wordGames || [];
        mp.players = data.players || [];
        loadCurrentWord();
        updateProgressPlayers();
    });
}

function updateLobbyInfo(roomCode, mode, wordCount) {
    document.getElementById('lobby-code').textContent = roomCode;
    document.getElementById('lobby-share-url').value = window.location.origin + '/multiplayer?join=' + roomCode;
    document.getElementById('lobby-mode').textContent = mode === 'progressif' ? 'Progressif' : 'Aléatoire';
    document.getElementById('lobby-wordcount').textContent = wordCount;
}

function updateLobbyPlayers() {
    const list = document.getElementById('lobby-players');
    list.innerHTML = '';
    mp.players.forEach(p => {
        const li = document.createElement('li');
        li.textContent = (p.id === mp.creatorID ? '👑 ' : '') + p.nickname;
        list.appendChild(li);
    });
    document.getElementById('lobby-playercount').textContent = mp.players.length;
}

function wordStatus(player, wordIdx) {
    if (wordIdx < player.currentWordIdx) return 'won';
    if (wordIdx === player.currentWordIdx) {
        if (player.finished && player.failed) return 'lost';
        return 'current';
    }
    return 'pending';
}

function progressCircles(player) {
    const container = document.createElement('span');
    container.className = 'player-progress-circles';
    for (let i = 0; i < player.totalWords; i++) {
        const dot = document.createElement('span');
        dot.className = 'word-dot ' + wordStatus(player, i);
        container.appendChild(dot);
    }
    return container;
}

function updateProgressPlayers() {
    const list = document.getElementById('progress-players');
    if (!list) return;
    list.innerHTML = '';
    mp.players.forEach(p => {
        const li = document.createElement('li');
        const creatorIcon = document.createTextNode(p.id === mp.creatorID ? '👑 ' : '');
        li.appendChild(creatorIcon);
        li.appendChild(document.createTextNode(p.nickname + ' '));
        li.appendChild(progressCircles(p));
        list.appendChild(li);
    });
}

function updateWordIndicator() {
    const el = document.getElementById('game-word-indicator');
    if (!el) return;
    el.textContent = '';
    const me = mp.players.find(p => p.id === mp.playerID) || { currentWordIdx: mp.currentWordIdx, totalWords: mp.wordCount, finished: false, failed: false };
    for (let i = 0; i < me.totalWords; i++) {
        const dot = document.createElement('span');
        dot.className = 'word-dot ' + wordStatus(me, i);
        el.appendChild(dot);
    }
}

function startTimer() {
    const el = document.getElementById('game-timer');
    if (!el) return;
    const start = Date.now();
    setInterval(() => {
        const elapsed = Math.floor((Date.now() - start) / 1000);
        const m = String(Math.floor(elapsed / 60)).padStart(2, '0');
        const s = String(elapsed % 60).padStart(2, '0');
        el.textContent = m + ':' + s;
    }, 1000);
}

function copyShareLink() {
    const input = document.getElementById('lobby-share-url');
    input.select();
    document.execCommand('copy');
}

function leaveRoom() {
    if (mp.roomCode && mp.playerID) {
        fetch('/api/multiplayer/leave', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ roomCode: mp.roomCode, playerID: mp.playerID })
        }).catch(() => {});
    }
    if (mp.eventSource) {
        mp.eventSource.close();
    }
    mp = newMp();
    document.getElementById('join-code').value = '';
    document.getElementById('join-nickname').value = '';
    showScreen('create');
}

function newGame() {
    fetch('/api/multiplayer/restart', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ roomCode: mp.roomCode, playerID: mp.playerID, token: mp.token })
    })
    .then(res => res.json())
    .then(data => {
        if (data.error) { alert(data.error); return; }
    })
    .catch(err => alert('Erreur de connexion'));
}

function updateNewGameBtn() {
    const newGameBtn = document.getElementById('new-game-btn');
    newGameBtn.style.display = (mp.creatorID === mp.playerID) ? 'block' : 'none';
}

function renderRankings(rankings) {
    const tbody = document.getElementById('rankings-body');
    tbody.innerHTML = '';
    rankings.forEach((r, idx) => {
        const tr = document.createElement('tr');
        const pos = document.createElement('td');
        pos.textContent = idx + 1;
        const name = document.createElement('td');
        name.textContent = r.nickname;
        const time = document.createElement('td');
        if (r.finished) {
            const totalSec = Math.floor(r.time / 1e9);
            const m = String(Math.floor(totalSec / 60)).padStart(2, '0');
            const s = String(totalSec % 60).padStart(2, '0');
            time.textContent = m + ':' + s;
        } else {
            time.textContent = '—';
        }
        const status = document.createElement('td');
        status.textContent = r.failed ? 'Échoué' : (r.finished ? 'Terminé' : 'En cours');
        tr.appendChild(pos);
        tr.appendChild(name);
        tr.appendChild(time);
        tr.appendChild(status);
        tbody.appendChild(tr);
    });
}

// Grid and keyboard (adapted from solo)
function initGrid(wordLength, firstLetter) {
    const grid = document.getElementById('multi-grid');
    grid.innerHTML = '';

    mp.foundLetters = [{ position: 0, letter: firstLetter }];
    mp.lockedPositions = new Set([0]);
    mp.currentRow = 0;

    for (let row = 0; row < 6; row++) {
        const rowDiv = document.createElement('div');
        rowDiv.className = 'row';
        rowDiv.id = 'mrow-' + row;

        for (let col = 0; col < wordLength; col++) {
            const tile = document.createElement('div');
            tile.className = 'tile';
            tile.id = 'mtile-' + row + '-' + col;
            rowDiv.appendChild(tile);
        }
        grid.appendChild(rowDiv);
    }

    const tile00 = document.getElementById('mtile-0-0');
    tile00.textContent = firstLetter;
    tile00.classList.add('locked', 'correct');

    mp.currentCol = firstUnlockedPosition(0);
    addCursor();
}

function firstUnlockedPosition(startFrom) {
    let col = startFrom;
    while (col < mp.wordLength && mp.lockedPositions.has(col)) {
        col++;
    }
    return col;
}

function addCursor() {
    removeCursor();
    const row = mp.currentRow;
    const col = mp.currentCol;
    const tile = document.getElementById('mtile-' + row + '-' + col);
    if (tile) tile.classList.add('cursor');
}

function removeCursor() {
    document.querySelectorAll('.tile.cursor').forEach(t => t.classList.remove('cursor'));
}

function handleKeyClick(key) {
    if (mp.gameOver) return;
    if (key === 'Enter') {
        submitCurrentWord();
    } else if (key === 'Backspace') {
        handleBackspace();
    } else {
        handleLetter(key);
    }
}

document.addEventListener('keydown', (e) => {
    if (!document.getElementById('screen-game') || document.getElementById('screen-game').style.display === 'none') return;
    if (mp.gameOver) return;
    if (e.key === 'Enter') {
        e.preventDefault();
        submitCurrentWord();
    } else if (e.key === 'Backspace') {
        e.preventDefault();
        handleBackspace();
    } else if (e.key.length === 1 && /[a-zA-Z]/.test(e.key)) {
        handleLetter(e.key.toUpperCase());
    }
});

document.getElementById('multi-submit-btn')?.addEventListener('click', () => {
    if (!mp.gameOver) submitCurrentWord();
});

function handleLetter(letter) {
    const row = mp.currentRow;
    const col = mp.currentCol;
    if (col >= mp.wordLength) return;

    const tile = document.getElementById('mtile-' + row + '-' + col);
    tile.textContent = letter;
    tile.classList.remove('locked', 'correct', 'present', 'absent');

    mp.currentCol = col + 1;
    addCursor();
}

function handleBackspace() {
    const row = mp.currentRow;
    const col = mp.currentCol;
    if (col <= 0) return;

    const newCol = col - 1;
    const tile = document.getElementById('mtile-' + row + '-' + newCol);
    resetTile(tile, newCol);

    mp.currentCol = newCol;
    addCursor();
}

function submitCurrentWord() {
    const row = mp.currentRow;
    const wordLength = mp.wordLength;
    let word = '';

    for (let col = 0; col < wordLength; col++) {
        const tile = document.getElementById('mtile-' + row + '-' + col);
        const letter = tile.textContent.trim();
        if (!letter) {
            showMultiMessage('Le mot doit faire ' + wordLength + ' lettres', 'error');
            resetCurrentRow();
            return;
        }
        word += letter;
    }

    if (word[0] !== mp.firstLetter) {
        showMultiMessage('Le mot doit commencer par ' + mp.firstLetter, 'error');
        resetCurrentRow();
        return;
    }

    submitGuess(word);
}

function submitGuess(word) {
    const btn = document.getElementById('multi-submit-btn');
    btn.disabled = true;

    fetch('/api/multiplayer/guess', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            roomCode: mp.roomCode,
            playerID: mp.playerID,
            word: word
        })
    })
    .then(res => res.json())
    .then(data => {
        if (data.error) {
            showMultiMessage(data.error, 'error');
            resetCurrentRow();
            enableInput();
            return;
        }

        const row = mp.currentRow;
        updateGridRow(row, data.results);
        updateKeyboard(data.results);
        mp.attempts.push(word);

        mp.guessResults.push(data.results);

        if (data.wordFinished) {
            if (data.playerFinished) {
                if (data.playerFailed) {
                    showMultiMessage('Perdu ! Vous avez épuisé vos essais.', 'lose');
                } else {
                    showMultiMessage('Bravo ! Vous avez trouvé tous les mots !', 'win');
                }
                btn.disabled = true;
                mp.gameOver = true;
                if (data.rankings) {
                    renderRankings(data.rankings);
                    showScreen('results');
                }
            } else {
                showMultiMessage(data.wordWon ? 'Mot trouvé !' : 'Mot échoué', data.wordWon ? 'win' : 'lose');
                setTimeout(() => {
                    mp.currentWordIdx = data.currentWordIdx;
                    loadNextWord();
                    enableInput();
                }, 1500);
            }
        } else {
            prepareNextRow();
            enableInput();
        }
    })
    .catch(err => {
        showMultiMessage('Erreur de connexion', 'error');
        btn.disabled = false;
        addCursor();
    });
}

function loadNextWord() {
    fetchGameState().then(data => {
        if (data.error) { return; }
        mp.wordGames = data.wordGames || [];
        mp.currentWordIdx = data.currentWordIdx;
        mp.wordSequence = data.wordSequence || [];
        loadCurrentWord();
    });
}

function enableInput() {
    const btn = document.getElementById('multi-submit-btn');
    btn.disabled = false;
    addCursor();
}

function resetTile(tile, col) {
    const foundLetter = mp.foundLetters.find(fp => fp.position === col);
    if (foundLetter) {
        tile.textContent = foundLetter.letter;
        tile.classList.add('locked', 'correct');
        tile.classList.remove('present', 'absent');
    } else {
        tile.textContent = '';
        tile.classList.remove('locked', 'correct', 'submitted', 'present', 'absent');
    }
}

function resetCurrentRow() {
    const row = mp.currentRow;
    const wordLength = mp.wordLength;
    for (let col = 0; col < wordLength; col++) {
        const tile = document.getElementById('mtile-' + row + '-' + col);
        resetTile(tile, col);
    }
    mp.currentCol = firstUnlockedPosition(0);
    addCursor();
}

function prepareNextRow() {
    const wordLength = mp.wordLength;
    const results = mp.guessResults[mp.guessResults.length - 1] || [];

    for (let col = 0; col < results.length; col++) {
        if (results[col].Status === 0) {
            const existing = mp.foundLetters.find(f => f.position === col);
            if (!existing) {
                const letter = String.fromCharCode(results[col].Letter);
                mp.foundLetters.push({ position: col, letter: letter });
                mp.lockedPositions.add(col);
            }
        }
    }

    mp.currentRow++;
    for (const fp of mp.foundLetters) {
        const tile = document.getElementById('mtile-' + mp.currentRow + '-' + fp.position);
        if (tile) {
            tile.textContent = fp.letter;
            tile.classList.add('locked', 'correct');
        }
    }
    mp.currentCol = firstUnlockedPosition(0);
}

function updateGridRow(row, results) {
    for (let col = 0; col < results.length; col++) {
        const tile = document.getElementById('mtile-' + row + '-' + col);
        if (!tile) continue;
        const r = results[col];
        tile.textContent = String.fromCharCode(r.Letter);

        const statusClass = STATUS_CLASSES[r.Status] || 'absent';

        setTimeout(() => {
            tile.classList.add('submitted', statusClass);
        }, col * 100);
    }
}

function updateKeyboard(results) {
    for (const r of results) {
        const letter = String.fromCharCode(r.Letter);
        const status = r.Status;
        if (mp.letterStatuses[letter] === undefined || status < mp.letterStatuses[letter]) {
            mp.letterStatuses[letter] = status;
        }
    }

    document.querySelectorAll('#multi-keyboard .kb-key').forEach(key => {
        const letter = key.dataset.key;
        const status = mp.letterStatuses[letter];
        key.classList.remove('correct', 'present', 'absent');
        if (status !== undefined) {
            key.classList.add(STATUS_CLASSES[status]);
        }
    });
}

function renderKeyboard() {
    const container = document.getElementById('multi-keyboard');
    container.innerHTML = '';

    const rows = [
        ['A', 'Z', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'],
        ['Q', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L', 'M'],
        ['Enter', 'W', 'X', 'C', 'V', 'B', 'N', 'Backspace']
    ];

    for (const row of rows) {
        const rowDiv = document.createElement('div');
        rowDiv.className = 'kb-row';

        for (const key of row) {
            const btn = document.createElement('button');
            btn.className = 'kb-key';
            if (key === 'Enter' || key === 'Backspace') {
                btn.classList.add('special');
                btn.textContent = key === 'Enter' ? 'Entrée' : 'Suppr';
            } else {
                btn.textContent = key;
            }
            btn.dataset.key = key;
            btn.addEventListener('click', () => handleKeyClick(key));
            rowDiv.appendChild(btn);
        }
        container.appendChild(rowDiv);
    }
}

function showMultiMessage(msg, type) {
    const el = document.getElementById('multi-message');
    if (!el) return;
    el.textContent = msg;
    el.className = type;
}
