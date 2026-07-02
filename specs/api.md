# API (endpoints HTTP)

## Endpoints

| Méthode | Chemin | Description |
|---------|--------|-------------|
| GET | `/` | Page d'accueil avec sélection du mode de jeu. |
| GET | `/game?mode=` | Page de jeu (`daily` ou `solo`). |
| POST | `/api/game/new` | Crée une nouvelle partie. Body : `{ mode }`. Réponse : `{ id, wordLength, firstLetter, maxTries, mode }`. |
| POST | `/api/game/guess` | Soumet une proposition. Body : `{ gameId, word }`. Réponse : `{ results: [{Letter, Status}], won, gameOver, attempts }`. |
| GET | `/static/*` | Fichiers statiques (`app.js`, `style.css`). |

### Détail des validations côté serveur (`POST /api/game/guess`)

- Vérification que la partie n'est pas déjà terminée.
- Vérification de la longueur du mot proposé.
- Vérification que la première lettre correspond à la lettre dévoilée.
- Validation du mot proposé contre le dictionnaire complet (`words_full.txt`).
- Calcul des statuts via l'algorithme Correct/Present/Absent.

## API multi-joueur

| Méthode | Chemin | Description |
|---------|--------|-------------|
| GET | `/multiplayer` | Page multi-joueur. |
| POST | `/api/multiplayer/room/create` | Crée une room. Body : `{ mode, wordCount, nickname }`. Réponse : `{ roomCode, shareURL, playerID, token }`. |
| POST | `/api/multiplayer/room/join` | Rejoint une room. Body : `{ roomCode, nickname, playerID? }`. Réponse : `{ playerID, token, roomCode, mode, wordCount, state, creatorID, players }`. |
| POST | `/api/multiplayer/room/start` | Démarre la partie (créateur seulement). Body : `{ roomCode, playerID }`. |
| POST | `/api/multiplayer/guess` | Soumet une proposition multi-joueur. Body : `{ roomCode, playerID, word }`. |
| POST | `/api/multiplayer/room/leave` | Quitte une room. Body : `{ roomCode, playerID }`. |
| POST | `/api/multiplayer/room/restart` | Redémarre la partie (créateur seulement, avec token). Body : `{ roomCode, playerID, token }`. |
| GET | `/api/multiplayer/events?room=&player=` | SSE (Server-Sent Events) pour les mises à jour temps réel. |
