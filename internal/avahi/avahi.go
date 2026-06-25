// Package avahi génère le service Avahi (mDNS) qui annonce SlashNode sur le
// réseau local en _http._tcp, permettant la résolution slashnode.local.
package avahi

import (
	"fmt"
	"os"
	"path/filepath"
)

// ServiceContent rend le fichier de service Avahi pour le port donné.
func ServiceContent(port int) string {
	return fmt.Sprintf(`<?xml version="1.0" standalone='no'?>
<!DOCTYPE service-group SYSTEM "avahi-service.dtd">
<service-group>
  <name replace-wildcards="yes">SlashNode sur %%h</name>
  <service>
    <type>_http._tcp</type>
    <port>%d</port>
  </service>
</service-group>
`, port)
}

// WriteService écrit le fichier de service Avahi à path (mode 0644).
func WriteService(path string, port int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(ServiceContent(port)), 0o644)
}
