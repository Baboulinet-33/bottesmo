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
        document.querySelectorAll('.mode-btn[data-mode]').forEach(btn => {
            btn.addEventListener('click', () => {
                const mode = btn.dataset.mode;
                window.location.href = '/game?mode=' + mode;
            });
        });
        const vsBtn = document.getElementById('vs-mode-btn');
        if (vsBtn) {
            vsBtn.addEventListener('click', () => {
                window.location.href = '/vs';
            });
        }
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
        initGrid(data.wordLength, data.firstLetter);
        setupInput(data.wordLength);
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
            if (row === 0 && col === 0) {
                tile.textContent = firstLetter;
                tile.classList.add('first');
            }
            rowDiv.appendChild(tile);
        }
        grid.appendChild(rowDiv);
    }
}

function setupInput(wordLength) {
    const input = document.getElementById('guess-input');
    const btn = document.getElementById('submit-btn');

    if (!input) return;

    input.disabled = false;
    btn.disabled = false;
    input.focus();

    const submit = () => {
        const word = input.value.trim().toUpperCase();
        if (word.length !== wordLength) {
            showMessage('Le mot doit faire ' + wordLength + ' lettres', 'error');
            return;
        }
        submitGuess(word);
    };

    btn.onclick = submit;

    input.onkeydown = (e) => {
        if (e.key === 'Enter') {
            submit();
        }
    };

    input.oninput = () => {
        input.value = input.value.replace(/[^a-zA-Z]/g, '').toUpperCase();
    };
}

function enableInput() {
    const el = document.getElementById('guess-input');
    const btn = document.getElementById('submit-btn');
    el.disabled = false;
    btn.disabled = false;
    el.value = '';
    el.focus();
}

function submitGuess(word) {
    const firstLetter = gameState.firstLetter;
    if (word[0] !== firstLetter) {
        showMessage('Le mot doit commencer par ' + firstLetter, 'error');
        enableInput();
        return;
    }

    const input = document.getElementById('guess-input');
    const btn = document.getElementById('submit-btn');
    input.disabled = true;
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
            enableInput();
            return;
        }

        const row = gameState.attempts.length;
        updateGrid(row, data.results);
        gameState.attempts.push(word);

        if (data.gameOver) {
            if (data.won) {
                showMessage('Bravo ! Vous avez trouvé le mot !', 'win');
            } else {
                showMessage('Perdu ! Le mot était : ' + data.word, 'lose');
            }
            input.disabled = true;
            btn.disabled = true;
            updateStats(data.won);
            addReplayButton();
        } else {
            enableInput();
        }
    })
    .catch(err => {
        showMessage('Erreur de connexion', 'error');
        input.disabled = false;
        btn.disabled = false;
    });
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
