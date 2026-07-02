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

## Notes

- C'est la seule variable d'environnement lue par l'application.
- Les autres paramètres de configuration (chemins des dictionnaires, motifs des templates, timeout HTTP) sont codés en dur dans `cmd/server/main.go`.
