# Dictionnaire

## Fichiers

Le jeu utilise deux fichiers de mots (attention fichiers de démo) :

| Fichier | Rôle | Taille |
|---------|------|--------|
| `words.txt` | Mots cibles (candidats pour daily/solo) | 11 mots |
| `words_full.txt` | Dictionnaire complet de validation | ~280 000 mots |

Les mots cibles ont une longueur de **6 à 10 lettres**.

Il est possible de paramétrer les dictionnaires avec les variables d'environnements suivante (voir [Variables d'environnement](specs/env.md)):
- DICT_WORDS_SOURCE
- DICT_WORDS_FULL_SOURCE

## Indexation par longueur

Le dictionnaire indexe les mots par longueur au chargement pour permettre une sélection rapide par taille de mot.

## Sélection quotidienne (mode Daily)

```go
func DailyWord(length int, date string) (string, error) {
    key := "daily:" + length + ":" + date
    h := sha256.Sum256([]byte(key))
    idx := int(h[0]) | (int(h[1]) << 8)
    return words[idx % len(words)], nil
}
```

## Sélection aléatoire (mode Solo)

```go
func RandomWord() (string, error)
```

Longueur choisie uniformément parmi les longueurs disponibles. Mot choisi aléatoirement parmi les mots de cette longueur.

## Validation des mots

La validation orthographique des mots saisis par le joueur est effectuée côté serveur via la dictionnaire `words_full.txt` ou pointé par la variable d'environnement **DICT_WORDS_FULL_SOURCE**. Le serveur valide la longueur, la première lettre et la correspondance avec les mots déjà existants dans le dictionnaire.
