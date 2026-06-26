package cli

import (
	"fmt"
	"strings"

	"github.com/slashbinslashnoname/slashnode/internal/paths"
	"github.com/slashbinslashnoname/slashnode/internal/secrets"
)

// Passwd sets the admin password: `slashnoded passwd [new-password]`. With no
// argument a strong random password is generated and printed. Use it to recover
// access when the login password is lost (the web UI always requires a login).
func Passwd(args []string) error {
	sec, err := secrets.Load(paths.SecretsFile())
	if err != nil {
		return fmt.Errorf("secrets not found (run `slashnoded init`): %w", err)
	}
	var pw string
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		pw = args[0]
		if len(pw) < 8 {
			return fmt.Errorf("password must be at least 8 characters")
		}
	} else if pw, err = secrets.NewPassword(); err != nil {
		return err
	}
	if err := sec.SetPassword(pw); err != nil {
		return err
	}
	if err := sec.Save(paths.SecretsFile()); err != nil {
		return err
	}
	fmt.Println(colorize("✓ admin password updated", ansiRed))
	fmt.Printf("  password: %s\n", pw)
	return nil
}
