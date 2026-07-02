# Variables d'environnement

## `PORT`

Port HTTP sur lequel le serveur écoute.

| | |
|---|---|
| **Obligatoire** | Non |
| **Valeur par défaut** | `3102` |
| **Utilisé dans** | `cmd/server/main.go` |

## Définition dans le code

```go
port := os.Getenv("PORT")
if port == "" {
    port = "3102"
}
```

## Utilisation dans le Dockerfile

Le port est exposé dans le Dockerfile via `EXPOSE ${PORT:-3102}`, ce qui permet de le surcharger au moment du `docker run` avec `-e PORT=XXXX`.

## `DICT_WORDS_SOURCE`

Chemin (fichier local) ou URL (HTTP/HTTPS) du dictionnaire principal utilisé pour les mots du jeu.

| | |
|---|---|
| **Obligatoire** | Non |
| **Valeur par défaut** | `internal/dictionary/words.txt` (chemin relatif depuis la racine du projet) |
| **Utilisé dans** | `cmd/server/main.go` (ligne 21), `internal/dictionary/dictionary.go` (fonction `LoadFromSource`) |

La source peut être :
- un chemin local (ex: `/data/dict/words.txt`)
- une URL HTTP/HTTPS (ex: `https://example.com/dictionary.txt`)

Le dictionnaire est chargé au démarrage via `LoadFromSource` qui appelle `resolveSource` dans `internal/dictionary/dictionary.go`.

## `DICT_WORDS_FULL_SOURCE`

Chemin (fichier local) ou URL (HTTP/HTTPS) du dictionnaire complet utilisé pour la validation des mots (tentatives des joueurs).

| | |
|---|---|
| **Obligatoire** | Non |
| **Valeur par défaut** | `internal/dictionary/words_full.txt` (chemin relatif depuis la racine du projet) |
| **Utilisé dans** | `cmd/server/main.go` (ligne 29), `internal/dictionary/dictionary.go` (fonction `LoadFullFromSource`) |

La source supporte les mêmes formats que `DICT_WORDS_SOURCE` (chemin local ou URL).

## Notes

- Les trois variables d'environnement (`PORT`, `DICT_WORDS_SOURCE`, `DICT_WORDS_FULL_SOURCE`) sont optionnelles et ont des valeurs par défaut.
- Les autres paramètres de configuration (motif des templates, timeout HTTP) sont codés en dur dans `cmd/server/main.go`.
