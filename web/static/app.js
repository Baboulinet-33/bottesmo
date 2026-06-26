let gameState = null;

const DEFAULT_STATS = '{"played":0,"won":0,"streak":0,"maxStreak":0,"lastResult":""}';

document.addEventListener('DOMContentLoaded', () => {
    const gameDiv = document.getElementById('game');
    if (gameDiv) {
        const mode = gameDiv.dataset.mode;
        startGame(mode);
        return;
    }

    const homeDiv = document.getElementById('home');
    if (homeDiv) {
        document.querySelectorAll('.mode-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const mode = btn.dataset.mode;
                window.location.href = '/game?mode=' + mode;
            });
        });
    }
});

function startGame(mode) {
    fetch('/api/game/new', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ mode: mode })
    })
    .then(res => res.json())
    .then(data => {
        gameState = data;
        gameState.attempts = [];
        gameState.won = false;
        gameState.gameOver = false;
        gameState.currentRow = 0;
        gameState.foundLetters = [{position: 0, letter: data.firstLetter}];
        gameState.lockedPositions = new Set([0]);
        gameState.letterStatuses = {};
        initGrid(data.wordLength, data.firstLetter);
        setupCellInput(data.wordLength);
        renderKeyboard();
    })
    .catch(err => {
        showMessage('Erreur de connexion au serveur', 'error');
    });
}

function initGrid(wordLength, firstLetter) {
    const grid = document.getElementById('grid');
    grid.innerHTML = '';

    for (let row = 0; row < 6; row++) {
        const rowDiv = document.createElement('div');
        rowDiv.className = 'row';
        rowDiv.id = 'row-' + row;

        for (let col = 0; col < wordLength; col++) {
            const tile = document.createElement('div');
            tile.className = 'tile';
            tile.id = 'tile-' + row + '-' + col;
            rowDiv.appendChild(tile);
        }
        grid.appendChild(rowDiv);
    }

    for (const fp of gameState.foundLetters) {
        const tile = document.getElementById('tile-0-' + fp.position);
        tile.textContent = fp.letter;
        tile.classList.add('locked', 'correct');
    }

    gameState.currentCol = firstUnlockedPosition(0);
    addCursor();
}

function firstUnlockedPosition(startFrom) {
    const wordLength = gameState.wordLength;
    let col = startFrom;
    while (col < wordLength && gameState.lockedPositions.has(col)) {
        col++;
    }
    return col;
}

function addCursor() {
    removeCursor();
    const row = gameState.currentRow;
    const col = gameState.currentCol;
    const tile = document.getElementById('tile-' + row + '-' + col);
    if (tile) tile.classList.add('cursor');
}

function removeCursor() {
    document.querySelectorAll('.tile.cursor').forEach(t => t.classList.remove('cursor'));
}

function setupCellInput(wordLength) {
    const btn = document.getElementById('submit-btn');
    btn.disabled = false;

    document.addEventListener('keydown', (e) => {
        if (gameState.gameOver) return;
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

    btn.onclick = () => {
        if (!gameState.gameOver) submitCurrentWord();
    };
}

function handleLetter(letter) {
    const row = gameState.currentRow;
    const col = gameState.currentCol;
    const wordLength = gameState.wordLength;

    if (col >= wordLength) return;

    const tile = document.getElementById('tile-' + row + '-' + col);
    tile.textContent = letter;
    tile.classList.remove('locked', 'correct', 'present', 'absent');

    gameState.currentCol = col + 1;
    addCursor();
}

function handleBackspace() {
    const row = gameState.currentRow;
    const col = gameState.currentCol;

    if (col <= 0) return;

    const newCol = col - 1;
    const tile = document.getElementById('tile-' + row + '-' + newCol);

    const foundLetter = gameState.foundLetters.find(fp => fp.position === newCol);
    if (foundLetter) {
        tile.textContent = foundLetter.letter;
        tile.classList.add('locked', 'correct');
        tile.classList.remove('present', 'absent');
    } else {
        tile.textContent = '';
        tile.classList.remove('locked', 'correct', 'submitted', 'present', 'absent');
    }

    gameState.currentCol = newCol;
    addCursor();
}

function submitCurrentWord() {
    const row = gameState.currentRow;
    const wordLength = gameState.wordLength;
    let word = '';

    for (let col = 0; col < wordLength; col++) {
        const tile = document.getElementById('tile-' + row + '-' + col);
        const letter = tile.textContent.trim();
        if (!letter) {
            showMessage('Le mot doit faire ' + wordLength + ' lettres', 'error');
            resetCurrentRow();
            return;
        }
        word += letter;
    }

    submitGuess(word);
}

function enableInput() {
    const btn = document.getElementById('submit-btn');
    btn.disabled = false;
    addCursor();
}

function resetCurrentRow() {
    const row = gameState.currentRow;
    const wordLength = gameState.wordLength;
    for (let col = 0; col < wordLength; col++) {
        const tile = document.getElementById('tile-' + row + '-' + col);
        const foundLetter = gameState.foundLetters.find(fp => fp.position === col);
        if (foundLetter) {
            tile.textContent = foundLetter.letter;
            tile.classList.add('locked', 'correct');
            tile.classList.remove('present', 'absent');
        } else {
            tile.textContent = '';
            tile.classList.remove('locked', 'correct', 'submitted', 'present', 'absent');
        }
    }
    gameState.currentCol = firstUnlockedPosition(0);
    addCursor();
}

function submitGuess(word) {
    const firstLetter = gameState.firstLetter;
    if (word[0] !== firstLetter) {
        showMessage('Le mot doit commencer par ' + firstLetter, 'error');
        resetCurrentRow();
        enableInput();
        return;
    }

    const btn = document.getElementById('submit-btn');
    btn.disabled = true;

    fetch('/api/game/guess', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ gameId: gameState.id, word: word })
    })
    .then(res => res.json())
    .then(data => {
        if (data.error) {
            showMessage(data.error, 'error');
            resetCurrentRow();
            enableInput();
            return;
        }

        const row = gameState.attempts.length;
        updateGrid(row, data.results);
        updateKeyboard(data.results);
        gameState.attempts.push(word);

        if (data.gameOver) {
            if (data.won) {
                showMessage('Bravo ! Vous avez trouvé le mot !', 'win');
            } else {
                showMessage('Perdu ! Le mot était : ' + data.word, 'lose');
            }
            btn.disabled = true;
            updateStats(data.won);
            addReplayButton();
        } else {
            prepareNextRow(data.results);
            enableInput();
        }
    })
    .catch(err => {
        showMessage('Erreur de connexion', 'error');
        btn.disabled = false;
        addCursor();
    });
}

function prepareNextRow(results) {
    const wordLength = gameState.wordLength;

    for (let col = 0; col < results.length; col++) {
        if (results[col].Status === 0) {
            const existing = gameState.foundLetters.find(f => f.position === col);
            if (!existing) {
                const letter = String.fromCharCode(results[col].Letter);
                gameState.foundLetters.push({position: col, letter: letter});
                gameState.lockedPositions.add(col);
            }
        }
    }

    gameState.currentRow++;

    for (const fp of gameState.foundLetters) {
        const tile = document.getElementById('tile-' + gameState.currentRow + '-' + fp.position);
        tile.textContent = fp.letter;
        tile.classList.add('locked', 'correct');
    }

    gameState.currentCol = firstUnlockedPosition(0);
}

function updateGrid(row, results) {
    for (let col = 0; col < results.length; col++) {
        const tile = document.getElementById('tile-' + row + '-' + col);
        const r = results[col];
        tile.textContent = String.fromCharCode(r.Letter);

        const statusClasses = ['correct', 'present', 'absent'];
        const statusClass = statusClasses[r.Status] || 'absent';

        setTimeout(() => {
            tile.classList.add('submitted', statusClass);
        }, col * 100);
    }
}

function updateKeyboard(results) {
    for (const r of results) {
        const letter = String.fromCharCode(r.Letter);
        const status = r.Status;
        if (gameState.letterStatuses[letter] === undefined || status < gameState.letterStatuses[letter]) {
            gameState.letterStatuses[letter] = status;
        }
    }

    const statusClasses = ['correct', 'present', 'absent'];
    document.querySelectorAll('.kb-key').forEach(key => {
        const letter = key.dataset.key;
        const status = gameState.letterStatuses[letter];
        key.classList.remove('correct', 'present', 'absent');
        if (status !== undefined) {
            key.classList.add(statusClasses[status]);
        }
    });
}

function renderKeyboard() {
    const container = document.getElementById('keyboard');
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
                btn.dataset.key = key;
            } else {
                btn.textContent = key;
                btn.dataset.key = key;
            }
            btn.addEventListener('click', () => handleKeyClick(key));
            rowDiv.appendChild(btn);
        }
        container.appendChild(rowDiv);
    }
}

function handleKeyClick(key) {
    if (gameState.gameOver) return;
    if (key === 'Enter') {
        submitCurrentWord();
    } else if (key === 'Backspace') {
        handleBackspace();
    } else {
        handleLetter(key);
    }
}

function showMessage(msg, type) {
    const el = document.getElementById('message');
    el.textContent = msg;
    el.className = type;
}

function updateStats(won) {
    const stats = JSON.parse(localStorage.getItem('tusmo-stats') || DEFAULT_STATS);
    stats.played++;
    if (won) {
        stats.won++;
        stats.streak++;
        if (stats.streak > stats.maxStreak) stats.maxStreak = stats.streak;
        stats.lastResult = 'won';
    } else {
        stats.streak = 0;
        stats.lastResult = 'lost';
    }
    localStorage.setItem('tusmo-stats', JSON.stringify(stats));
    displayStats(stats);
}

function displayStats(stats) {
    if (!stats) {
        stats = JSON.parse(localStorage.getItem('tusmo-stats') || DEFAULT_STATS);
    }
    const el = document.getElementById('stats');
    el.innerHTML = 'Parties: ' + stats.played + ' | Victoires: ' + stats.won + ' | Séries: ' + stats.streak;
}

function addReplayButton() {
    const msg = document.getElementById('message');
    const btn = document.createElement('button');
    btn.textContent = 'Rejouer';
    btn.className = 'mode-btn';
    btn.style.marginTop = '1rem';
    btn.addEventListener('click', () => {
        if (gameState.mode === 'daily') {
            window.location.href = '/game?mode=daily';
        } else {
            window.location.reload();
        }
    });
    msg.appendChild(document.createElement('br'));
    msg.appendChild(btn);
}
