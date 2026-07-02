# Configuration serveur

## Paramètres

| Paramètre | Valeur défaut | Source |
|-----------|---------------|--------|
| Port | 3102 | Variable d'env `PORT` |
| ReadTimeout | 5s | Code (`main.go`) |
| WriteTimeout | 10s | Code (`main.go`) |
| IdleTimeout | 60s | Code (`main.go`) |
| Max sessions | 100 000 | Code (`main.go`) |
| TTL session | 1 heure | Code (`handlers/game.go`) |
| Nettoyage sessions | 5 minutes | Code (`handlers/game.go`) |
| Taille max body | 1 Mo | Code (`main.go`) |
| Tentatives max | 6 | Code (`game/game.go`) |
| Max joueurs par room | 20 | Code (`game/multiplayer.go`) |
| Nettoyage rooms | 1 heure / 5 minutes | Code (`handlers/multiplayer.go`) |
