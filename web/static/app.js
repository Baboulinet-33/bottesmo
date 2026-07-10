let gameState = null;

const DEFAULT_STATS = '{"played":0,"won":0,"streak":0,"maxStreak":0,"lastResult":""}';
const STATUSES = ['correct', 'present', 'absent'];
const KEYBOARD_ROWS = [
    ['A', 'Z', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'],
    ['Q', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L', 'M'],
    ['Enter', 'W', 'X', 'C', 'V', 'B', 'N', 'Backspace']
];

// ===== Letter Palette Management =====

const LETTER_PALETTES = [
    {
        id: 'original',
        name: 'Bottesmo (défaut)',
        colors: {
            '--correct': '#10b981',
            '--correct-contrast': '#ffffff',
            '--present': '#f59e0b',
            '--present-contrast': '#1e293b',
            '--absent': '#6b7280',
            '--absent-contrast': '#ffffff',
            '--kb-key-absent-bg': '#374151',
            '--kb-key-absent-text': '#9ca3af'
        }
    },
    {
        id: 'wordle',
        name: 'Classique Wordle',
        colors: {
            '--correct': '#538d4e',
            '--correct-contrast': '#ffffff',
            '--present': '#b59f3b',
            '--present-contrast': '#ffffff',
            '--absent': '#3a3a3c',
            '--absent-contrast': '#ffffff',
            '--kb-key-absent-bg': '#3a3a3c',
            '--kb-key-absent-text': '#ffffff'
        }
    },
    {
        id: 'ocean',
        name: 'Océan',
        colors: {
            '--correct': '#0ea5e9',
            '--correct-contrast': '#ffffff',
            '--present': '#06b6d4',
            '--present-contrast': '#ffffff',
            '--absent': '#1e40af',
            '--absent-contrast': '#bfdbfe',
            '--kb-key-absent-bg': '#1e3a5f',
            '--kb-key-absent-text': '#93c5fd'
        }
    },
    {
        id: 'sunset',
        name: 'Coucher de soleil',
        colors: {
            '--correct': '#f97316',
            '--correct-contrast': '#ffffff',
            '--present': '#fbbf24',
            '--present-contrast': '#1c1917',
            '--absent': '#7c3aed',
            '--absent-contrast': '#ede9fe',
            '--kb-key-absent-bg': '#4c1d95',
            '--kb-key-absent-text': '#ddd6fe'
        }
    },
    {
        id: 'foret',
        name: 'Forêt',
        colors: {
            '--correct': '#4d7c0f',
            '--correct-contrast': '#f7fee7',
            '--present': '#ca8a04',
            '--present-contrast': '#fefce8',
            '--absent': '#78350f',
            '--absent-contrast': '#fef3c7',
            '--kb-key-absent-bg': '#422006',
            '--kb-key-absent-text': '#d97706'
        }
    },
    {
        id: 'neon',
        name: 'Néon',
        colors: {
            '--correct': '#39ff14',
            '--correct-contrast': '#000000',
            '--present': '#ff00ff',
            '--present-contrast': '#000000',
            '--absent': '#111111',
            '--absent-contrast': '#39ff14',
            '--kb-key-absent-bg': '#0a0a0a',
            '--kb-key-absent-text': '#39ff14'
        }
    },
    {
        id: 'mono',
        name: 'Monochrome',
        colors: {
            '--correct': '#111111',
            '--correct-contrast': '#ffffff',
            '--present': '#555555',
            '--present-contrast': '#ffffff',
            '--absent': '#aaaaaa',
            '--absent-contrast': '#000000',
            '--kb-key-absent-bg': '#999999',
            '--kb-key-absent-text': '#000000'
        }
    },
    {
        id: 'pastel',
        name: 'Pastel',
        colors: {
            '--correct': '#86efac',
            '--correct-contrast': '#14532d',
            '--present': '#fdba74',
            '--present-contrast': '#7c2d12',
            '--absent': '#c4b5fd',
            '--absent-contrast': '#3b0764',
            '--kb-key-absent-bg': '#ddd6fe',
            '--kb-key-absent-text': '#4c1d95'
        }
    },
    {
        id: 'contraste',
        name: 'Contraste élevé',
        colors: {
            '--correct': '#00ff00',
            '--correct-contrast': '#000000',
            '--present': '#ffff00',
            '--present-contrast': '#000000',
            '--absent': '#ffffff',
            '--absent-contrast': '#000000',
            '--kb-key-absent-bg': '#ffffff',
            '--kb-key-absent-text': '#000000'
        }
    },
    {
        id: 'automne',
        name: 'Automne',
        colors: {
            '--correct': '#b45309',
            '--correct-contrast': '#fffbeb',
            '--present': '#d97706',
            '--present-contrast': '#1c1917',
            '--absent': '#57534e',
            '--absent-contrast': '#fafaf9',
            '--kb-key-absent-bg': '#292524',
            '--kb-key-absent-text': '#a8a29e'
        }
    },
    {
        id: 'tusmo',
        name: 'Tusmo',
        colors: {
            '--correct': '#DB3A34',
            '--correct-contrast': '#ffffff',
            '--present': '#f7b735',
            '--present-contrast': '#1e293b',
            '--absent': '#6b7280',
            '--absent-contrast': '#ffffff',
            '--kb-key-absent-bg': '#374151',
            '--kb-key-absent-text': '#9ca3af'
        }
    }
];

function getStoredPaletteId() {
    return localStorage.getItem('bottesmo-letter-palette') || 'original';
}

function setStoredPaletteId(id) {
    localStorage.setItem('bottesmo-letter-palette', id);
}

function applyPalette(id) {
    const palette = LETTER_PALETTES.find(p => p.id === id) || LETTER_PALETTES[0];
    for (const [prop, value] of Object.entries(palette.colors)) {
        document.documentElement.style.setProperty(prop, value);
    }
}

// Theme management
function getPreferredTheme() {
    const saved = localStorage.getItem('bottesmo-theme');
    if (saved) return saved;
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function applyTheme(theme) {
    document.documentElement.dataset.theme = theme === 'light' ? 'light' : '';
    const btn = document.getElementById('theme-toggle');
    if (btn) btn.textContent = theme === 'light' ? '🌙' : '☀️';
}

function toggleTheme() {
    const next = document.documentElement.dataset.theme === 'light' ? 'dark' : 'light';
    localStorage.setItem('bottesmo-theme', next);
    applyTheme(next);
}

document.addEventListener('DOMContentLoaded', () => {
    const theme = getPreferredTheme();
    applyTheme(theme);
    applyPalette(getStoredPaletteId());
    document.getElementById('theme-toggle')?.addEventListener('click', toggleTheme);

    const gameDiv = document.getElementById('game');
    if (gameDiv) {
        const mode = gameDiv.dataset.mode;
        startGame(mode);
        return;
    }

    const homeDiv = document.getElementById('home');
    if (homeDiv) {
        document.querySelectorAll('.mode-btn').forEach(btn => {
            const mode = btn.dataset.mode;
            if (!mode) return;
            btn.addEventListener('click', () => {
                window.location.href = '/game?mode=' + mode;
            });
        });
    }

    const paletteGrid = document.getElementById('palette-grid');
    if (paletteGrid) {
        const currentId = getStoredPaletteId();
        LETTER_PALETTES.forEach(palette => {
            const card = document.createElement('div');
            card.className = 'palette-card' + (palette.id === currentId ? ' selected' : '');
            card.dataset.paletteId = palette.id;

            const name = document.createElement('div');
            name.className = 'palette-name';
            name.textContent = palette.name;

            const preview = document.createElement('div');
            preview.className = 'palette-preview';
            [
                ['A', '--correct', '--correct-contrast'],
                ['B', '--present', '--present-contrast'],
                ['C', '--absent', '--absent-contrast']
            ].forEach(([label, bg, fg]) => {
                const tile = document.createElement('span');
                tile.className = 'palette-tile';
                tile.textContent = label;
                tile.style.backgroundColor = palette.colors[bg];
                tile.style.color = palette.colors[fg];
                preview.appendChild(tile);
            });

            card.appendChild(preview);
            card.appendChild(name);
            card.addEventListener('click', () => {
                setStoredPaletteId(palette.id);
                applyPalette(palette.id);
                document.querySelectorAll('.palette-card').forEach(c => c.classList.remove('selected'));
                card.classList.add('selected');
            });
            paletteGrid.appendChild(card);
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

function restoreTile(row, col) {
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

function handleBackspace() {
    const col = gameState.currentCol;
    if (col <= 0) return;

    const newCol = col - 1;
    restoreTile(gameState.currentRow, newCol);
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
    for (let col = 0; col < gameState.wordLength; col++) {
        restoreTile(row, col);
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

        const statusClass = STATUSES[r.Status] || 'absent';

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

    document.querySelectorAll('.kb-key').forEach(key => {
        const letter = key.dataset.key;
        const status = gameState.letterStatuses[letter];
        key.classList.remove('correct', 'present', 'absent');
        if (status !== undefined) {
            key.classList.add(STATUSES[status]);
        }
    });
}

function renderKeyboard() {
    const container = document.getElementById('keyboard');
    container.innerHTML = '';

    for (const row of KEYBOARD_ROWS) {
        const rowDiv = document.createElement('div');
        rowDiv.className = 'kb-row';

        for (const key of row) {
            const btn = document.createElement('button');
            btn.className = 'kb-key';
            btn.dataset.key = key;
            if (key === 'Enter' || key === 'Backspace') {
                btn.classList.add('special');
                btn.textContent = key === 'Enter' ? 'Entrée' : 'Suppr';
            } else {
                btn.textContent = key;
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
    const stats = JSON.parse(localStorage.getItem('bottesmo-stats') || DEFAULT_STATS);
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
    localStorage.setItem('bottesmo-stats', JSON.stringify(stats));
    displayStats(stats);
}

function displayStats(stats) {
    if (!stats) {
        stats = JSON.parse(localStorage.getItem('bottesmo-stats') || DEFAULT_STATS);
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
