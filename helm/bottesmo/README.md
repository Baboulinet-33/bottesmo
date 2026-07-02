# Helm Chart — Bottesmo

Chart Helm pour déployer **Bottesmo** (jeu Wordle-like en français avec mode multijoueur SSE) sur Kubernetes.

## Prérequis

- Cluster Kubernetes (≥ 1.23)
- [ingress-nginx](https://kubernetes.github.io/ingress-nginx/) installé dans le cluster
- Image Docker Hub disponible : `bnoleau/bottesmo:latest` (le build/push est hors scope de ce chart)
- `helm` ≥ 3.x installé localement

## Installation

```bash
helm install bottesmo ./helm/bottesmo \
  --namespace bottesmo --create-namespace \
  --set ingress.hosts[0].host=bottesmo.example.com
```

Pour vérifier le déploiement :

```bash
kubectl -n bottesmo get pods,svc,ingress
```

## Mise à jour

```bash
helm upgrade bottesmo ./helm/bottesmo \
  --namespace bottesmo \
  --set ingress.hosts[0].host=bottesmo.example.com
```

Comme l'image utilise le tag `latest` avec `pullPolicy: Always`, un `helm upgrade` force le pull de la dernière image. Pour forcer un redémarrage du pod sans changer les values :

```bash
kubectl rollout restart deployment/bottesmo -n bottesmo
```

## Surcharge des variables d'environnement

Les variables d'application sont exposées via `values.yaml` → `env` :

```bash
# Changer le port interne (rare, le Service s'adapte)
helm install bottesmo ./helm/bottesmo --set env.PORT=8080

# Utiliser des dictionnaires personnalisés
helm install bottesmo ./helm/bottesmo \
  --set env.DICT_WORDS_SOURCE=https://exemple.com/words.txt \
  --set env.DICT_WORDS_FULL_SOURCE=https://exemple.com/words_full.txt
```

## Activer TLS

```bash
helm install bottesmo ./helm/bottesmo \
  --set ingress.hosts[0].host=bottesmo.example.com \
  --set ingress.tls[0].secretName=bottesmo-tls \
  --set ingress.tls[0].hosts[0]=bottesmo.example.com
```

Le secret TLS doit exister dans le namespace (via cert-manager ou manuellement).

## Contrainte 1 réplica — IMPORTANT

L'état multijoueur (rooms, joueurs connectés, flux SSE) est stocké **en mémoire dans le pod**. Scaler à N > 1 réplicas sans backend partagé (Redis, etc.) causerait des **rooms fantômes** : les joueurs arrivant sur des pods différents ne se verraient pas.

La stratégie de déploiement est `Recreate` (plutôt que `RollingUpdate`) pour éviter deux pods actifs simultanément.

**Ne pas activer l'autoscaling (HPA) sans avoir d'abord externalisé l'état multijoueur.**

## Valeurs par défaut notables

| Clé | Valeur | Description |
|-----|--------|-------------|
| `replicaCount` | `1` | Voir contrainte ci-dessus |
| `image.repository` | `bnoleau/bottesmo` | Image Docker Hub |
| `image.tag` | `latest` | Tag mutable |
| `image.pullPolicy` | `Always` | Garantit le repull à chaque déploiement |
| `service.targetPort` | `3102` | Port interne de l'app |
| `ingress.enabled` | `true` | Ingress nginx activé par défaut |
| `strategy.type` | `Recreate` | Évite deux pods actifs simultanément |

## Architecture de sécurité

Le pod s'exécute avec :
- `runAsNonRoot: true` + `runAsUser: 1001` (utilisateur `appuser` du Dockerfile)
- `readOnlyRootFilesystem: true` (aucune écriture disque nécessaire)
- `allowPrivilegeEscalation: false`
- `capabilities.drop: [ALL]`

## Désinstallation

```bash
helm uninstall bottesmo -n bottesmo
kubectl delete namespace bottesmo
```
