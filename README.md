<pre>
   ___
  / o o\     ___ _         _ _  _         _
  \  ^  /   / __| |__ _ __| | \| |___  __| |___
   |||||    \__ \ / _` (_-< | .` / _ \/ _` / -_)
  / ||| \   |___/_\__,_/__/_|_|\_\___/\__,_\___|
   /   \
</pre>

# SlashNode

**Orchestrateur de services auto-hébergés piloté par manifestes JSON, posé sur
Docker.** Pensé pour les nœuds Bitcoin/Lightning et le self-hosting :
plug-and-play (`exports`/`wiring`), App Store, thème clair/sombre.

> Ce n'est pas un OS custom. C'est un démon Go (`slashnoded`) qui s'installe sur
> n'importe quel Debian/Ubuntu existant via une ligne.

## Installation

```bash
curl -fsSL https://get.slashnode.sh | bash
```

Public crypto-conscient ? Auditez avant d'exécuter :

```bash
curl -fsSL https://get.slashnode.sh -o slashnode.sh
less slashnode.sh
bash slashnode.sh
```

Le bootstrap reste minimal : il installe Docker + Node, pose le binaire
`slashnoded` (checksum vérifié), déploie le front et délègue toute l'init à
`slashnoded init`. Une fois fini, le nœud est joignable sur
**http://slashnode.local:8080** (mDNS/Avahi).

## Architecture

| Composant       | Rôle                                                            |
|-----------------|-----------------------------------------------------------------|
| `slashnoded`    | Démon Go, binaire unique. Lance l'API locale **et** le front.   |
| Front Next.js   | UI (thème clair/sombre, primary rouge), lancé par le démon.     |
| API Go locale   | `127.0.0.1`, auth par token Bearer. Consommée par le front.     |
| systemd         | `slashnoded.service` + `slashnoded-update.timer` (vérif. MAJ).  |
| Avahi           | mDNS → `slashnode.local`.                                       |
| Docker          | Runtime des apps (manifestes JSON, voir `docs/app-manifest.md`).|

```
slashnoded serve
 ├─ API Go        127.0.0.1:8081   (/api/v1/status, /api/v1/update[/apply])
 └─ next start    0.0.0.0:8080     (front, proxifie l'API via le token)
```

## Commandes

```
slashnoded init           # config + secrets + systemd + Avahi (idempotent)
slashnoded serve          # API Go + front Next.js supervisé
slashnoded status         # état du nœud (--post-install : URL + identifiants)
slashnoded update         # applique la dernière MAJ du binaire (--to <tag>)
slashnoded check-update   # vérifie une MAJ (appelé par le timer ; notify-only)
slashnoded uninstall      # retire service + binaire (--purge : aussi les données)
```

## Mises à jour

Politique **notify-only** par défaut : le timer systemd vérifie chaque jour la
dernière release et persiste l'état (`/var/lib/slashnode/update.json`). L'UI
affiche alors une bannière avec un bouton **« Appliquer »** qui déclenche
`slashnoded update` (téléchargement + vérification checksum + remplacement
atomique du binaire + redémarrage). Configurable dans `config.json`
(`update.policy`, `update.channel`).

## Développement

```bash
# Front
cd web && npm install && npm run build

# Démon (tests sans root : SLASHNODE_ROOT préfixe tous les chemins)
export SLASHNODE_ROOT=/tmp/sn
go run ./cmd/slashnoded init
go run ./cmd/slashnoded serve --web-dir web   # front sur :8080, API sur :8081
```

Build des artefacts de release (binaires amd64/arm64 + macOS, bundle web) :

```bash
./scripts/build.sh v0.1.0   # → dist/
```

## Apps (manifestes)

Chaque app = un manifeste JSON (services Docker, `inputs` saisis par
l'utilisateur → variables d'env, `exports`/`wiring` pour le branchement
automatique). Voir **[docs/app-manifest.md](docs/app-manifest.md)** et les
exemples dans `examples/`.

## Licence

À définir.
