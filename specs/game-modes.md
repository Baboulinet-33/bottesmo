# Modes de jeu

## Modes disponibles

| Mode | Description |
|------|-------------|
| **Daily** (Mot du jour) | Mot identique pour tous les joueurs pendant une journée. La longueur du mot est aussi déterministe pour la journée. |
| **Solo** | Mot aléatoire, différent à chaque partie. La longueur est choisie aléatoirement parmi les longueurs disponibles. |

## Sélection quotidienne (mode Daily)

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

## Sélection aléatoire (mode Solo)

```go
// Longueur choisie uniformément parmi les longueurs disponibles
// Mot choisi aléatoirement parmi les mots de cette longueur
func RandomWord() (string, error)
```
