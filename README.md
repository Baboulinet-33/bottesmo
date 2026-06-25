# Tusmo

Jeu de mots inspiré de Wordle, en français.

Deux modes de jeu :
- **Mot du jour** : un mot identique pour tous les joueurs, changé chaque jour.
- **Solo** : un mot aléatoire à chaque partie.

## Lancer le serveur

```bash
go run ./cmd/server
```

Le serveur écoute sur le port `3102` par défaut. Pour utiliser un autre port :

```bash
PORT=3103 go run ./cmd/server
```

Ouvrir http://localhost:3102 dans un navigateur.

## Lancer les tests

```bash
go test ./... -v
```

## Modifier le dictionnaire

Le fichier `words.txt` à la racine du projet est un lien symbolique vers `internal/dictionary/words.txt`. Vous pouvez modifier l'un ou l'autre.

Le dictionnaire est un fichier texte avec un mot par ligne, en majuscules :

```
ABRITE
ACCORD
ACTION
```

Après modification, redémarrer le serveur pour prendre en compte les changements.

Les mots sont automatiquement convertis en majuscules au chargement. Les lignes vides sont ignorées.
