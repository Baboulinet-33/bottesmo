let vsState = {
    roomId: null,
    playerId: null,
    isCreator: false,
    players: [],
    words: [],
    wordCount: 0,
    lengthMode: 'progressive',
    currentWord: 0,
    attempts: [],
    won: false,
    failed: false,
    eventSource: null,
};

document.addEventListener('DOMContentLoaded', () => {
    const vsApp = document.getElementById('vs-app');
    if (!vsApp) return;

    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');

    if (code) {
        showVSSection('lobby');
        document.getElementById('vs-share-link').value = window.location.href;
        autoJoin(code);
    } else {
        showVSSection('create');
        setupCreateForm();
    }

    setupNicknameEdit();
});

function showVSSection(section) {
    ['vs-create', 'vs-lobby', 'vs-game', 'vs-results'].forEach(id => {
        document.getElementById(id).style.display = id === 'vs-' + section ? 'block' : 'none';
    });
}

function getNickname() {
    let nick = localStorage.getItem('tusmo-pseudo');
    if (!nick) {
        nick = prompt('Choisissez votre pseudo :', '');
        if (!nick) nick = 'Joueur';
        localStorage.setItem('tusmo-pseudo', nick);
    }
    return nick;
}

function setupNicknameEdit() {
    const input = document.getElementById('vs-nickname-input');
    const saveBtn = document.getElementById('vs-nickname-save');
    input.value = getNickname();
    saveBtn.addEventListener('click', () => {
        const val = input.value.trim();
        if (val) {
            localStorage.setItem('tusmo-pseudo', val);
        }
    });
}

function setupCreateForm() {
    const form = document.getElementById('vs-create-form');
    form.addEventListener('submit', (e) => {
        e.preventDefault();
        const mode = form.querySelector('input[name="lengthMode"]:checked').value;
        const count = parseInt(document.getElementById('wordCount').value, 10);
        createRoom(mode, count);
    });
}

function createRoom(mode, count) {
    fetch('/api/vs/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ lengthMode: mode, wordCount: count })
    })
    .then(res => res.json())
    .then(data => {
        vsState.roomId = data.roomId;
        vsState.wordCount = count;
        vsState.lengthMode = mode;
        vsState.isCreator = true;
        vsState.playerId = 'pending';
        vsState.players = [];

        joinRoom(data.code, getNickname());
    })
    .catch(() => showVSMessage('Erreur de connexion', 'error'));
}

function joinRoom(code, nickname) {
    fetch('/api/vs/join', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code: code, nickname: nickname })
    })
    .then(res => res.json())
    .then(data => {
        if (data.error) {
            showVSMessage(data.error, 'error');
            return;
        }
        vsState.roomId = data.roomId;
        vsState.playerId = data.playerId;
        vsState.isCreator = data.isCreator;
        vsState.players = data.players;

        document.getElementById('vs-share-link').value = window.location.origin + '/vs/join?code=' + code;

        showVSSection('lobby');
        renderLobby();
        connectSSE(data.roomId, data.playerId);
    })
    .catch(() => showVSMessage('Erreur de connexion', 'error'));
}

function autoJoin(code) {
    const nickname = getNickname();
    joinRoom(code, nickname);
}

function connectSSE(roomId, playerId) {
    if (vsState.eventSource) {
        vsState.eventSource.close();
    }

    const es = new EventSource('/api/vs/events?roomId=' + roomId + '&playerId=' + playerId);
    vsState.eventSource = es;

    es.addEventListener('player-joined', (e) => {
        const data = JSON.parse(e.data);
        vsState.players.push({ id: data.playerId, nickname: data.nickname });
        renderLobby();
    });

    es.addEventListener('game-started', (e) => {
        const data = JSON.parse(e.data);
        vsState.words = data.words;
        vsState.players = data.players;
        vsState.currentWord = 0;
        vsState.attempts = [];
        vsState.won = false;
        vsState.failed = false;
        startVSGame();
    });

    es.addEventListener('progress', (e) => {
        const data = JSON.parse(e.data);
        updateProgress(data);
    });

    es.addEventListener('game-over', (e) => {
        const data = JSON.parse(e.data);
        showResults(data.rankings);
    });

    es.addEventListener('rematch', (e) => {
        const data = JSON.parse(e.data);
        handleRematch(data);
    });

    es.onerror = () => {
        setTimeout(() => {
            if (vsState.roomId && vsState.playerId) {
                connectSSE(vsState.roomId, vsState.playerId);
            }
        }, 3000);
    };
}

function renderLobby() {
    const list = document.getElementById('vs-player-list');
    list.innerHTML = '';
    vsState.players.forEach(p => {
        const li = document.createElement('li');
        li.textContent = p.nickname;
        if (p.id === vsState.playerId) {
            li.textContent += ' (vous)';
        }
        list.appendChild(li);
    });

    document.getElementById('vs-player-count').textContent = vsState.players.length;

    const startBtn = document.getElementById('vs-start-btn');
    if (vsState.isCreator) {
        startBtn.style.display = 'inline-block';
        startBtn.disabled = vsState.players.length < 2;
        startBtn.onclick = startGame;
    } else {
        startBtn.style.display = 'none';
    }
}

function startGame() {
    fetch('/api/vs/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ roomId: vsState.roomId, playerId: vsState.playerId })
    })
    .then(res => res.json())
    .then(data => {
        if (data.error) {
            showVSMessage(data.error, 'error');
            return;
        }
    })
    .catch(() => showVSMessage('Erreur de connexion', 'error'));
}

function startVSGame() {
    showVSSection('game');
    setupVSGameUI();
    initVSGrid();
    enableVSInput();
    renderProgress();
}

function setupVSGameUI() {
    const total = vsState.words.length;
    document.getElementById('vs-word-label').textContent = 'Mot 1/' + total;
    const input = document.getElementById('vs-guess-input');
    const btn = document.getElementById('vs-submit-btn');
    input.disabled = false;
    input.value = '';
    input.focus();
    btn.disabled = false;
    btn.onclick = submitVSGuess;
    input.onkeydown = (e) => {
        if (e.key === 'Enter') submitVSGuess();
    };
    document.getElementById('vs-message').textContent = '';
    document.getElementById('vs-message').className = '';
}

function initVSGrid() {
    const grid = document.getElementById('vs-grid');
    grid.innerHTML = '';

    for (let row = 0; row < 6; row++) {
        const rowDiv = document.createElement('div');
        rowDiv.className = 'row';
        rowDiv.id = 'vs-row-' + row;

        const wordLen = vsState.words[vsState.currentWord].length;
        for (let col = 0; col < wordLen; col++) {
            const tile = document.createElement('div');
            tile.className = 'tile';
            tile.id = 'vs-tile-' + row + '-' + col;
            if (row === 0 && col === 0) {
                tile.textContent = vsState.words[vsState.currentWord][0];
                tile.classList.add('first');
            }
            rowDiv.appendChild(tile);
        }
        grid.appendChild(rowDiv);
    }

    const input = document.getElementById('vs-guess-input');
    input.maxLength = wordLen;
    input.placeholder = wordLen + ' lettres';
}

function enableVSInput() {
    const input = document.getElementById('vs-guess-input');
    const btn = document.getElementById('vs-submit-btn');
    input.disabled = false;
    btn.disabled = false;
    input.value = '';
    input.focus();

    input.oninput = () => {
        input.value = input.value.replace(/[^a-zA-Z]/g, '').toUpperCase();
    };
}

function submitVSGuess() {
    const input = document.getElementById('vs-guess-input');
    const btn = document.getElementById('vs-submit-btn');
    const word = input.value.trim().toUpperCase();
    const wordLen = vsState.words[vsState.currentWord].length;

    if (word.length !== wordLen) {
        showVSMessage('Le mot doit faire ' + wordLen + ' lettres', 'error');
        return;
    }

    const firstLetter = vsState.words[vsState.currentWord][0];
    if (word[0] !== firstLetter) {
        showVSMessage('Le mot doit commencer par ' + firstLetter, 'error');
        return;
    }

    input.disabled = true;
    btn.disabled = true;

    fetch('/api/vs/guess', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ roomId: vsState.roomId, playerId: vsState.playerId, word: word })
    })
    .then(res => res.json())
    .then(data => {
        if (data.error) {
            showVSMessage(data.error, 'error');
            enableVSInput();
            return;
        }

        const row = vsState.attempts.length;
        updateVSGrid(row, data.results);
        vsState.attempts.push(word);

        if (data.allComplete) {
            if (data.won) {
                showVSMessage('Bravo ! Vous avez trouvé le mot !', 'win');
            } else if (data.failed) {
                showVSMessage('Perdu ! Le mot était : ' + vsState.words[vsState.currentWord], 'lose');
            }
            input.disabled = true;
            btn.disabled = true;
        } else if (data.wordComplete) {
            vsState.currentWord = data.currentWord;
            vsState.attempts = [];
            const total = vsState.words.length;
            document.getElementById('vs-word-label').textContent = 'Mot ' + (data.currentWord + 1) + '/' + total;
            initVSGrid();
            enableVSInput();
            updateVSWordProgress(data.currentWord);
        } else {
            enableVSInput();
        }
    })
    .catch(() => {
        showVSMessage('Erreur de connexion', 'error');
        enableVSInput();
    });
}

function updateVSGrid(row, results) {
    for (let col = 0; col < results.length; col++) {
        const tile = document.getElementById('vs-tile-' + row + '-' + col);
        const r = results[col];
        tile.textContent = String.fromCharCode(r.Letter);

        const statusClasses = ['correct', 'present', 'absent'];
        const statusClass = statusClasses[r.Status] || 'absent';

        setTimeout(() => {
            tile.classList.add('submitted', statusClass);
        }, col * 100);
    }
}

function updateVSWordProgress(wordIdx) {
}

function renderProgress() {
    const list = document.getElementById('vs-progress-list');
    list.innerHTML = '';

    vsState.players.forEach(p => {
        const div = document.createElement('div');
        div.className = 'vs-player-progress';
        div.id = 'vs-progress-' + p.id;

        const nameSpan = document.createElement('span');
        nameSpan.className = 'vs-player-name';
        nameSpan.textContent = p.nickname;
        if (p.id === vsState.playerId) {
            nameSpan.textContent += ' (vous)';
        }
        div.appendChild(nameSpan);

        const dotsDiv = document.createElement('div');
        dotsDiv.className = 'vs-word-dots';
        for (let i = 0; i < vsState.words.length; i++) {
            const dot = document.createElement('span');
            dot.className = 'vs-dot';
            dot.id = 'vs-dot-' + p.id + '-' + i;
            dotsDiv.appendChild(dot);
        }
        div.appendChild(dotsDiv);

        list.appendChild(div);
    });
}

function updateProgress(data) {
    const playerId = data.playerId;
    const currentWord = data.currentWord || 0;

    for (let i = 0; i < vsState.words.length; i++) {
        const dot = document.getElementById('vs-dot-' + playerId + '-' + i);
        if (!dot) continue;

        if (i < currentWord) {
            dot.className = 'vs-dot completed';
        } else if (i === currentWord) {
            dot.className = 'vs-dot active';
        } else {
            dot.className = 'vs-dot';
        }
    }

    if (data.allComplete) {
        const dot = document.getElementById('vs-dot-' + playerId + '-' + (vsState.words.length - 1));
        if (dot) {
            dot.className = data.won ? 'vs-dot completed' : 'vs-dot failed';
        }
    }
}

function showResults(rankings) {
    showVSSection('results');
    const tbody = document.getElementById('vs-ranking-body');
    tbody.innerHTML = '';

    rankings.forEach(r => {
        const tr = document.createElement('tr');
        const rankTd = document.createElement('td');
        rankTd.textContent = '#' + r.Rank;
        tr.appendChild(rankTd);

        const nameTd = document.createElement('td');
        nameTd.textContent = r.Nickname;
        tr.appendChild(nameTd);

        const timeTd = document.createElement('td');
        if (r.Failed) {
            timeTd.textContent = 'Échec';
        } else {
            const ms = r.CompletedTime / 1000000;
            const seconds = Math.floor(ms / 1000);
            const millis = Math.floor(ms % 1000);
            timeTd.textContent = seconds + 's ' + millis + 'ms';
        }
        tr.appendChild(timeTd);
        tr.className = r.PlayerID === vsState.playerId ? 'vs-current-player' : '';
        tbody.appendChild(tr);
    });

    document.getElementById('vs-rematch-btn').onclick = rematch;
}

function rematch() {
    fetch('/api/vs/rematch', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ roomId: vsState.roomId, playerId: vsState.playerId })
    })
    .then(res => res.json())
    .then(data => {
        if (data.error) {
            showVSMessage(data.error, 'error');
            return;
        }
        window.location.href = '/vs/join?code=' + data.code;
    })
    .catch(() => showVSMessage('Erreur de connexion', 'error'));
}

function handleRematch(data) {
    if (!data.playerIds || !data.playerIds[vsState.playerId]) {
        window.location.href = data.shareUrl;
        return;
    }
}

function showVSMessage(msg, type) {
    const el = document.getElementById('vs-message');
    el.textContent = msg;
    el.className = type;
}
