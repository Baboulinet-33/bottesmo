# Dictionnaire

## Fichiers

Le jeu utilise deux fichiers de mots :

| Fichier | Rôle | Taille |
|---------|------|--------|
| `words.txt` | Mots cibles (candidats pour daily/solo) | 50 mots |
| `words_full.txt` | Dictionnaire complet de validation | ~280 000 mots |

Les mots cibles ont une longueur de **6 à 10 lettres**.

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

La validation orthographique des mots saisis par le joueur est effectuée côté client via **Typo.js** (dictionnaire Hunspell français). Le serveur ne valide que la longueur, la première lettre et la correspondance avec les mots déjà existants dans le dictionnaire.
