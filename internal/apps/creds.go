package apps

// CredField is one stored input/secret of an installed app, for display in the
// UI (rpc user, passwords, …). Secret marks values that should be masked.
type CredField struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Value  string `json:"value"`
	Secret bool   `json:"secret"`
}

// AppExports returns an installed app's resolved exports (its exposed
// endpoints: rpc host/port/user/password, electrum host/port, …).
func AppExports(id string) map[string]string {
	return loadRegistry()[id]
}

// Credentials returns the stored input/secret values of an installed app,
// labelled from its manifest, so the frontend can display and reuse them.
func Credentials(dir, id string) ([]CredField, error) {
	man, err := Find(dir, id)
	if err != nil {
		return nil, err
	}
	inst := LoadState().Installed[id]
	secs := loadAppSecrets(id)

	out := make([]CredField, 0, len(man.Inputs))
	for _, in := range man.Inputs {
		v := inst.Inputs[in.Key]
		if v == "" {
			v = secs[in.Key]
		}
		out = append(out, CredField{
			Key:    in.Key,
			Label:  in.Label,
			Value:  v,
			Secret: in.Secret || in.Type == "password",
		})
	}
	return out, nil
}
