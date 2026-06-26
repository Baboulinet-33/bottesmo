# Tusmo

## Architecture

```
tusmo/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── dictionary/
│   │   ├── dictionary.go
│   │   └── dictionary_test.go
│   ├── game/
│   │   ├── game.go
│   │   ├── game_test.go
│   │   ├── vs.go              # VS multiplayer types and game logic
│   │   └── vs_test.go
│   └── handlers/
│       ├── game.go
│       └── vs.go              # SSE broker, VS HTTP handlers
├── web/
│   ├── templates/
│   │   ├── layout.html
│   │   ├── game.html
│   │   └── vs.html            # VS page template (create, lobby, game, results)
│   └── static/
│       ├── lib/
│       │   └── typo/          # Typo.js + fr_FR Hunspell dictionary
│       ├── style.css
│       ├── app.js
│       └── vs.js              # VS frontend logic (SSE, state, grid, progress, results)
├── words.txt
└── go.mod
```

## Components

### `internal/dictionary`

Charge les mots depuis `words.txt`, permet de les filtrer par longueur, de tirer un mot aléatoire ou déterministe. Le dictionnaire sert uniquement pour choisir le mot cible (modes daily et solo) et pour valider la correspondance des lettres — la validation orthographique est déléguée au client.

### `internal/game/vs.go`

Types et logique du mode VS multijoueur : `VSPlayer`, `VSRoom`, `VSWordLengthMode`, génération de séquence de mots, ajout de joueurs, démarrage de partie, traitement des essais, classement.

### `internal/handlers/vs.go`

SSE broker (per-room event broadcasting via channels), handlers HTTP pour créer/rejoindre/démarrer/jouer/rematch VS, nettoyage des rooms inactives. Utilise SSE (Server-Sent Events) pour les mises à jour temps réel.

### `web/templates/vs.html`

Template VS avec 4 états : création (formulaire), salon d'attente (lien partageable, liste joueurs), partie (grille + barre latérale progression), résultats (classement + rejouer).

### `web/static/vs.js`

Frontend VS : connexion SSE, gestion d'état, rendu de la grille de jeu, barre de progression par joueur (points colorés), tableau des résultats, rematch.

### Validation orthographique

La validation orthographique des mots saisis par le joueur est effectuée côté client via Typo.js (dictionnaire Hunspell français). Le serveur n'effectue plus de validation dictionnaire.

### SSE-based real-time event system

Le mode VS utilise Server-Sent Events (SSE) pour les mises à jour temps réel :
- `player-joined` : diffusion lors de l'arrivée d'un joueur dans le salon
- `game-started` : notification du début de partie (mots, liste joueurs)
- `progress` : mise à jour de la progression d'un joueur après chaque essai
- `game-over` : classement final quand tous les joueurs ont terminé
- `rematch` : nouvelle room créée pour une revanche
