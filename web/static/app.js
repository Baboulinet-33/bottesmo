const spellchecker = new Typo("fr_FR", null, null, {
    dictionaryPath: "/static/lib/typo/"
});

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
    input.focus();

    const submit = () => {
        const word = input.value.trim().toUpperCase();
        if (word.length !== wordLength) {
            showMessage('Le mot doit faire ' + wordLength + ' lettres', 'error');
            return;
        }
        submitGuess(word);
    };

    btn.disabled = false;
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

function submitGuess(word) {
    const input = document.getElementById('guess-input');
    const btn = document.getElementById('submit-btn');
    input.disabled = true;
    btn.disabled = true;

    if (!spellchecker.check(word.toLowerCase())) {
        showMessage('Mot invalide', 'error');
        input.disabled = false;
        btn.disabled = false;
        input.focus();
        input.value = '';
        return;
    }

    fetch('/api/game/guess', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ gameId: gameState.id, word: word })
    })
    .then(res => res.json())
    .then(data => {
        if (!data.results) {
            showMessage('Mot invalide', 'error');
            input.disabled = false;
            btn.disabled = false;
            input.focus();
            input.value = '';
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
            input.disabled = false;
            btn.disabled = false;
            input.value = '';
            input.focus();
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

        let statusClass = '';
        if (r.Status === 0) statusClass = 'correct';
        else if (r.Status === 1) statusClass = 'present';
        else statusClass = 'absent';

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
