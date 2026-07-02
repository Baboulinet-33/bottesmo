# Modèle de données

## Côté serveur (Go)

### Game (solo/daily)

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

### Session store

Les sessions sont stockées dans une map mémoire : `map[string]*session`. Chaque session contient une partie et un timestamp. Les sessions de plus d'une heure sont nettoyées toutes les 5 minutes. Capacité max : 100 000 sessions simultanées.

### MultiplayerRoom

```go
type MultiplayerRoom struct {
    Code         string
    Mode         string                    // "progressif" | "aleatoire"
    WordCount    int
    CreatorID    string
    Players      map[string]*MultiplayerPlayer
    Started      bool
    StartTime    time.Time
    Finished     bool
    State        string                    // "lobby" | "playing"
    CreatedAt    time.Time
    MaxPlayers   int                       // 20
    WordSequence []string
}

type MultiplayerPlayer struct {
    ID             string
    Nickname       string
    WordGames      []*Game
    CurrentWordIdx int
    StartTime      time.Time
    CompletedTime  time.Time
    Failed         bool
    Finished       bool
    JoinedAt       time.Time
}
```

### MultiplayerManager

```go
type MultiplayerManager struct {
    rooms        map[string]*MultiplayerRoom
    hub          *MultiplayerHub   // Gestion SSE
    playerTokens map[string]string // Tokens d'authentification
}
```

Les rooms multi-joueur sont nettoyées après 1 heure (vérification toutes les 5 minutes).

## Côté client (JavaScript)

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

### Statistiques (localStorage)

```json
{ "played": 42, "won": 35, "streak": 12, "maxStreak": 18, "lastResult": "won" }
```

Clé : `bottesmo-stats`. Affichage en pied de page : `Parties: 42 | Victoires: 35 | Séries: 12`.
