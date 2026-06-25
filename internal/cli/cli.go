// Package cli implémente les sous-commandes de slashnoded.
package cli

import (
	"fmt"
	"os"
)

// Version est renseignée depuis main au démarrage.
var Version = "dev"

// ANSI : rouge gras / reset. Désactivé si NO_COLOR est défini ou si la sortie
// n'est pas un terminal interactif.
const (
	ansiRed   = "\x1b[1;31m"
	ansiDim   = "\x1b[2m"
	ansiReset = "\x1b[0m"
)

func colorize(s, color string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return color + s + ansiReset
}

// asciiBonhomme est le bonhomme rouge emblématique de SlashNode.
const asciiBonhomme = `   ___
  / o o\     ___ _         _ _  _         _
  \  ^  /   / __| |__ _ __| | \| |___  __| |___
   |||||    \__ \ / _` + "`" + ` (_-< | .` + "`" + ` / _ \/ _` + "`" + ` / -_)
  / ||| \   |___/_\__,_/__/_|_|\_\___/\__,_\___|
   /   \`

// Banner imprime le logo + bonhomme rouge sur stdout.
func Banner() {
	fmt.Println(colorize(asciiBonhomme, ansiRed))
	fmt.Println(colorize("        // your node, your rules\n", ansiDim))
}

// Usage affiche l'aide générale.
func Usage() {
	Banner()
	fmt.Print(`Usage : slashnoded <commande> [options]

Commandes :
  init         Génère config + secrets + unit systemd + service Avahi (idempotent)
  serve        Démarre le démon (API Go + front Next.js supervisé)
  status       Affiche l'état du nœud (--post-install pour l'URL + identifiants)
  update       Applique la dernière mise à jour du binaire (--to <tag>)
  check-update Vérifie la disponibilité d'une mise à jour (notify-only)
  uninstall    Retire le service et le binaire (--purge supprime aussi les données)
  version     Affiche la version
  help        Affiche cette aide

Variables d'environnement :
  SLASHNODE_ROOT   Préfixe racine pour tous les chemins (tests sans root)
  NO_COLOR         Désactive la sortie colorée
`)
}
