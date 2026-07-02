# Multi-joueur

## Vue d'ensemble

Le mode multi-joueur permet à plusieurs joueurs de s'affronter en temps réel sur une séquence de mots. Chaque joueur progresse à son rythme sur les mêmes mots, et un classement est établi à la fin.

## Architecture

- **`internal/game/multiplayer.go`** : Logique métier (rooms, joueurs, séquence de mots, classement).
- **`internal/handlers/multiplayer.go`** : Endpoints HTTP, SSE (Server-Sent Events), gestion des tokens.
- **`web/static/multiplayer.js`** : Client JavaScript multi-joueur.

## Room

Une room est un espace de jeu partagé identifié par un code unique à 6 caractères (sans voyelles pour éviter les mots indésirables).

```go
type MultiplayerRoom struct {
    Code         string
    Mode         string              // "progressif" | "aleatoire"
    WordCount    int
    CreatorID    string
    Players      map[string]*MultiplayerPlayer
    State        string              // "lobby" | "playing"
    MaxPlayers   int                 // 20
    WordSequence []string
}
```

### États d'une room

| État | Description |
|------|-------------|
| `lobby` | En attente de joueurs. Le créateur peut lancer la partie. |
| `playing` | Partie en cours. Les joueurs peuvent rejoindre (late-join). |

## Modes de jeu multi-joueur

| Mode | Description |
|------|-------------|
| **progressif** | Les mots augmentent en longueur : 6, 7, 8, 9, 10 lettres (répété si nécessaire). |
| **aleatoire** | Chaque mot a une longueur aléatoire entre 6 et 10 lettres. |

## Joueur

```go
type MultiplayerPlayer struct {
    ID             string
    Nickname       string
    WordGames      []*Game          // Un jeu par mot
    CurrentWordIdx int
    Failed         bool              // true si épuisé les tentatives sur un mot
    Finished       bool              // true si tous les mots complétés ou échec
}
```

Chaque joueur dispose d'une instance de `Game` par mot dans la séquence. Les joueurs progressent indépendamment sur leur propre copie des mots.

## Déroulement d'une partie

1. Un joueur **crée une room** et définit le mode et le nombre de mots (1-10).
2. Les autres joueurs **rejoignent** via le code ou le lien de partage.
3. Le **créateur** lance la partie.
4. Le serveur génère une séquence de mots et initialise les jeux de chaque joueur.
5. Chaque joueur **soumet ses propositions** indépendamment.
6. Quand un joueur termine un mot (trouvé ou échec), il passe au mot suivant.
7. La partie se termine quand **tous les joueurs** ont fini tous leurs mots.
8. Un **classement** est affiché (temps de complétion).

### Late-join

Un joueur peut rejoindre une room même après le début de la partie. Il reçoit l'état complet de la partie (mots déjà joués, tentatives, etc.) et commence au mot en cours.

## Classement

Le classement trie les joueurs par :
1. Joueurs finis avant joueurs non finis.
2. Joueurs non éliminés avant joueurs éliminés (Failed).
3. Temps de complétion croissant.

Les résultats détaillés (mots, tentatives, statuts) sont inclus pour les joueurs finis.

## SSE (Server-Sent Events)

Les événements temps réel sont diffusés via SSE :

| Événement | Déclencheur | Données |
|-----------|-------------|---------|
| `player-joined` | Un joueur rejoint la room | `{ players }` |
| `player-left` | Un joueur quitte la room | `{ playerID, players, newCreatorID }` |
| `game-started` | La partie commence | `{ players }` |
| `progress` | Un joueur soumet une proposition | `{ players }` |
| `player-finished` | Un joueur termine tous ses mots | `{ playerID, rankings }` |
| `game-over` | Tous les joueurs ont fini | `{ rankings }` |
| `game-restarted` | Le créateur redémarre la partie | `{ players }` |

## Séquence de mots

```go
func GenerateWordSequence(mode string, count int) []string
```

- Mode `progressif` : alterne les longueurs 6, 7, 8, 9, 10 dans l'ordre.
- Mode `aleatoire` : longueur aléatoire entre 6 et 10 pour chaque mot.

## Redémarrage

Le créateur peut redémarrer une partie terminée. La room repasse en état `lobby`, les joueurs sont conservés mais leurs jeux sont réinitialisés. Un token d'authentification est requis pour cette opération.

## Départ du créateur

Si le créateur quitte la room, un nouveau créateur est désigné parmi les joueurs restants. Si la room devient vide, elle est supprimée.
