// Package avahi generates the Avahi service (mDNS) that announces SlashNode on
// the local network as _http._tcp, enabling slashnode.local resolution.
package avahi

import (
	"fmt"
	"os"
	"path/filepath"
)

// ServiceContent renders the Avahi service file for the given port.
func ServiceContent(port int) string {
	return fmt.Sprintf(`<?xml version="1.0" standalone='no'?>
<!DOCTYPE service-group SYSTEM "avahi-service.dtd">
<service-group>
  <name replace-wildcards="yes">SlashNode on %%h</name>
  <service>
    <type>_http._tcp</type>
    <port>%d</port>
  </service>
</service-group>
`, port)
}

// WriteService writes the Avahi service file to path (mode 0644).
func WriteService(path string, port int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(ServiceContent(port)), 0o644)
}
