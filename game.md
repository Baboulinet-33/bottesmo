# Bottesmo — Spécification du Jeu

## Vue d'ensemble

Bottesmo est un jeu de devinettes de mots en français, inspiré de Wordle. Le joueur dispose de **6 tentatives** pour trouver un mot cible. Une lettre de départ est dévoilée pour amorcer la réflexion. Après chaque proposition, un code couleur indique la pertinence de chaque lettre. Le jeu se joue dans un navigateur web, sans inscription.

---

## Règles du jeu

1. Le serveur choisit un **mot cible** (en français, en majuscules, de 6 à 10 lettres).
2. Le joueur voit uniquement la **première lettre** du mot cible (verrouillée, pré-remplie).
3. Le joueur a **6 tentatives** pour trouver le mot.
4. Chaque proposition doit :
   - Avoir la **même longueur** que le mot cible.
   - **Commencer par la première lettre** dévoilée.
   - Être un **mot français valide** (vérifié côté serveur).
5. Après chaque proposition, les lettres sont colorées selon leur statut.
6. Les lettres bien placées sont **verrouillées** et pré-remplies dans la ligne suivante.
7. La partie se termine quand le joueur trouve le mot ou épuise ses 6 tentatives.

---

## Modes de jeu

| Mode     | Description                                                                 |
|----------|-----------------------------------------------------------------------------|
| **Daily** (Mot du jour) | Mot identique pour tous les joueurs pendant une journée. La longueur du mot est aussi déterministe pour la journée. |
| **Solo**  | Mot aléatoire, différent à chaque partie. La longueur est choisie aléatoirement parmi les longueurs disponibles. |

---

## Système de feedback (code couleur)

Après chaque proposition, chaque lettre reçoit un statut :

| Couleur | Signification                  | Classe CSS   | Contexte du jeu          |
|---------|--------------------------------|--------------|--------------------------|
| 🟩 Vert   | Lettre **bonne** (bonne lettre, bonne position) | `.correct`   | La lettre est correcte et bien placée. |
| 🟠 Ambre | Lettre **mal placée** (bonne lettre, mauvaise position) | `.present`   | La lettre est dans le mot mais pas à cette position. |
| ⚫ Grise  | Lettre **absente** (lettre pas dans le mot) | `.absent`    | La lettre n'apparaît pas dans le mot cible. |

### Gestion des lettres doubles

L'algorithme suit la logique standard de Wordle pour les lettres dupliquées :
1. Les lettres **bien placées** (StatusCorrect) sont comptabilisées en premier.
2. Les occurrences restantes de chaque lettre sont ensuite attribuées aux positions **mal placées** (StatusPresent).
3. Les excédents sont marqués **absents**.

---

## Mécanique clé : première lettre verrouillée

C'est la fonctionnalité signature qui distingue Bottesmo de Wordle standard :

1. La **première lettre** du mot cible est toujours pré-remplie et verrouillée dans **chaque ligne** de la grille.
2. Après chaque proposition, les lettres marquées `Correct` (bien placées) deviennent des lettres **"trouvées"**.
3. Ces lettres trouvées sont automatiquement pré-remplies et verrouillées dans la **ligne suivante**.
4. Les tuiles verrouillées ne peuvent pas être modifiées (la touche Retour arrière les ignore).

Ce mécanisme réduit progressivement l'espace de recherche et guide le joueur vers la solution.

---

## Conditions de victoire / défaite

| Issue     | Condition                                      | Message affiché                              |
|-----------|------------------------------------------------|----------------------------------------------|
| ✅ Victoire | La proposition correspond exactement au mot cible | "Bravo ! Vous avez trouvé le mot !" (vert)   |
| ❌ Défaite  | Après 6 propositions incorrectes               | "Perdu ! Le mot était : {mot}" (rouge)       |

Dans les deux cas, le mot cible est révélé et une option "Rejouer" est proposée.

---

## Stack technique

| Couche    | Technologie          | Détails                                                    |
|-----------|----------------------|------------------------------------------------------------|
| Backend   | Go 1.25              | Serveur HTTP standard (stdlib), aucun framework externe.   |
| Frontend  | JavaScript vanilla   | Aucun framework ni « build step ». CSS vanilla.             |
| Templating| `html/template` (Go) | Pages rendues côté serveur.                                |
| Tests unitaires | `testing` (Go) | Tests pour la logique de jeu et le dictionnaire.            |
| Tests E2E | Playwright 1.61      | Tests cross-navigateur.                                    |
| Stockage  | Mémoire live + `localStorage` | Sessions en mémoire côté serveur, stats en localStorage côté client. |
| Port      | 3102 (défaut)        | Configurable via la variable d'environnement `PORT`.       |

---

## API (endpoints HTTP)

| Méthode | Chemin                | Description                                                    |
|---------|----------------------|----------------------------------------------------------------|
| GET     | `/`                  | Page d'accueil avec sélection du mode de jeu.                  |
| GET     | `/game?mode=`        | Page de jeu (`daily` ou `solo`).                               |
| POST    | `/api/game/new`      | Crée une nouvelle partie. Body : `{ mode }`. Réponse : `{ id, wordLength, firstLetter, maxTries, mode }`. |
| POST    | `/api/game/guess`    | Soumet une proposition. Body : `{ gameId, word }`. Réponse : `{ results: [{Letter, Status}], won, gameOver, attempts }`. |
| GET     | `/static/*`          | Fichiers statiques (`app.js`, `style.css`).                    |

### Détail des validations côté serveur (`POST /api/game/guess`)

- Vérification que la partie n'est pas déjà terminée.
- Vérification de la longueur du mot proposé.
- Vérification que la première lettre correspond à la lettre dévoilée.
- Validation du mot proposé contre le dictionnaire complet (`words_full.txt`).
- Calcul des statuts via l'algorithme Correct/Present/Absent.

---

## Modèle de données

### Côté serveur (Go)

```go
type Game struct {
    ID         string
    Target     string
    Attempts   []string
    MaxTries   int          // Toujours 6
    Mode       GameMode     // "daily" | "solo"
    Won        bool
    GameOver   bool
    WordLength int
}

type LetterStatus int     // 0=Correct, 1=Present, 2=Absent

type LetterResult struct {
    Letter rune
    Status LetterStatus
}
```

Les sessions sont stockées dans une map mémoire : `map[string]*session`. Chaque session contient une partie et un timestamp. Les sessions de plus d'une heure sont nettoyées toutes les 5 minutes. Capacité max : 100 000 sessions simultanées.

### Côté client (JavaScript)

```javascript
gameState = {
    id: string,
    wordLength: number,
    firstLetter: string,
    maxTries: number,
    mode: string,
    attempts: [],              // Mots déjà proposés
    won: boolean,
    gameOver: boolean,
    currentRow: number,
    currentCol: number,
    foundLetters: [],          // {position, letter}
    lockedPositions: Set,      // Indices de colonnes verrouillées
    letterStatuses: {}         // {A: 0, B: 1, C: 2, ...}
}
```

---

## Dictionnaire

Le jeu utilise deux fichiers de mots :

| Fichier          | Rôle                                  | Taille       |
|------------------|---------------------------------------|--------------|
| `words.txt`      | Mots cibles (candidats pour daily/solo) | 50 mots      |
| `words_full.txt` | Dictionnaire complet de validation     | ~280 000 mots|

Les mots cibles ont une longueur de **6 à 10 lettres**.

### Sélection quotidienne (mode Daily)

```go
// La longueur du mot est déterministe (basée sur la date)
// Le mot est sélectionné via SHA-256 de "daily:" + longueur + date
func DailyWord(length int, date string) (string, error) {
    key := "daily:" + length + ":" + date
    h := sha256.Sum256([]byte(key))
    idx := int(h[0]) | (int(h[1]) << 8)
    return words[idx % len(words)], nil
}
```

### Sélection aléatoire (mode Solo)

```go
// Longueur choisie uniformément parmi les longueurs disponibles
// Mot choisi aléatoirement parmi les mots de cette longueur
func RandomWord() (string, error)
```

---

## UI / UX

### Thème et style

- **Double thème** : sombre (défaut) et clair, basculable via un bouton dans l'en-tête.
- **Bouton de thème** : intégré dans l'en-tête (flexbox `space-between`), il est stylisé sans fond ni bordure (`background: transparent; border: none`) avec un hover couleur accent (`var(--accent)`).
- **Persistance** : le choix du thème est sauvegardé dans `localStorage` (clé `bottesmo-theme`). À la première visite, le thème respecte la préférence système (`prefers-color-scheme`).
- **Mécanisme** : les couleurs sont définies via des **CSS custom properties** (`var(--…)`) dans `:root` (thème sombre) et `[data-theme="light"]` (surcharges claires). Voir `web/static/style.css` pour la palette complète.
- **Palette** : accent ambre `#de802b`, correct émeraude `#10b981`, présent ambre `#f59e0b`, absent gris `#6b7280`.
- **Police** : système (Segoe UI, Roboto, sans-serif).
- **Disposition** : centrée, largeur max 600px, flexbox.
- **Responsive** : point de rupture à 480px (taille réduite des tuiles et du clavier).

### Grille de jeu

- 6 rangées × N colonnes, espacement de 6px entre les tuiles.
- Tuiles de 52×52px (desktop) / 42×42px (mobile).
- Tuiles verrouillées : fond vert émeraude (`--correct`) avec bordure verte foncée.
- Curseur : bordure blanche clignotante (animation).
- Animation de soumission : `flip` (rotationX 90° puis retour) avec décalage de 100ms entre chaque tuile.

### Clavier virtuel

- Disposition **AZERTY** (française) :
  - Rangée 1 : `A Z E R T Y U I O P`
  - Rangée 2 : `Q S D F G H J K L M`
  - Rangée 3 : `[Entrée] W X C V B N [Suppr]`
- Touches de 32×48px (desktop) / 28×42px (mobile).
- Touches Entrée et Suppr plus larges.
- Après chaque proposition, les touches du clavier se colorent (priorité : Correct > Present > Absent).
- Interaction possible au clavier physique et au clic souris/tactile.

### Messages

- Erreur (mot invalide, problème réseau) : ambre (`var(--error)`).
- Victoire : vert émeraude (`var(--win)`).
- Défaite : ambre (`var(--accent)`).

---

## Statistiques (localStorage)

Les statistiques sont stockées dans le navigateur du joueur sous la clé `bottesmo-stats` :

```json
{ "played": 42, "won": 35, "streak": 12, "maxStreak": 18, "lastResult": "won" }
```

Affichage en pied de page : `Parties: 42 | Victoires: 35 | Séries: 12`

---

## Tests

### Tests unitaires (Go)

- `internal/game/game_test.go` : Tests de la logique de jeu (validation des statuts de lettres, gestion des parties, etc.).
- `internal/dictionary/dictionary_test.go` : Tests de chargement du dictionnaire, sélection quotidienne et aléatoire.

### Tests E2E (Playwright)

- `bottesmo.spec.js` : 8 tests couvrant :
  1. Saisie au clavier et soumission.
  2. Comportement de la touche Retour arrière.
  3. Pré-remplissage des lettres trouvées à la ligne suivante.
  4. Saisie via clavier virtuel (clic).
  5. Effacement via le bouton Suppr du clavier virtuel.
  6. Coloration des touches du clavier après soumission.
  7. Parcours complet d'une partie gagnante.
  8. Affichage responsive (viewport mobile 375×667).
  9. Mode Daily : mot déterministe avec première lettre verrouillée.

---

## Configuration serveur

| Paramètre          | Valeur défaut | Source                  |
|--------------------|---------------|-------------------------|
| Port               | 3102          | Variable d'env `PORT`   |
| ReadTimeout        | 5s            | Code (`main.go`)        |
| WriteTimeout       | 10s           | Code (`main.go`)        |
| IdleTimeout        | 60s           | Code (`main.go`)        |
| Max sessions       | 100 000       | Code (`main.go`)        |
| TTL session        | 1 heure       | Code (`handlers/game.go`) |
| Nettoyage sessions | 5 minutes     | Code (`handlers/game.go`) |
| Taille max body    | 1 Mo          | Code (`main.go`)        |
| Tentatives max     | 6             | Code (`game/game.go`)   |

---

## Déroulement complet d'une partie

1. Le joueur atterrit sur la **page d'accueil** (`/`) et choisit un mode.
2. Le navigateur est redirigé vers `/game?mode=daily` ou `/game?mode=solo`.
3. Le frontend appelle `POST /api/game/new` pour créer une partie.
4. Le serveur génère un **mot cible**, crée une session et retourne les métadonnées (longueur, première lettre).
5. La **grille** est affichée : 6 rangées × N colonnes, première tuile verrouillée.
6. Le joueur **tape une proposition** (clavier physique ou AZERTY virtuel).
7. La proposition est envoyée au serveur via `POST /api/game/guess`.
8. Le serveur valide le mot et retourne les **statuts** de chaque lettre.
9. Les tuiles s'animent (flip) et se colorent (vert/ambre/grise).
10. Les touches du clavier virtuel se colorent.
11. Les lettres correctes sont **verrouillées** dans la ligne suivante.
12. Les étapes 6 à 11 se répètent jusqu'à victoire ou défaite.
13. Le message de fin s'affiche, les **statistiques** sont mises à jour.
14. Le joueur peut cliquer sur "Rejouer" pour recommencer.
