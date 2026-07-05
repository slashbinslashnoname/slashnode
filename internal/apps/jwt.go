package apps

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// signSupabaseJWT builds the kind of HS256 JWT Supabase's API keys carry — a
// token whose only meaningful claim is `role` (anon / service_role) — signed
// with the operator's own generated JWT secret. This replaces Supabase's public
// demo anon/service keys (which anyone can forge because their signing secret is
// published) with keys unique to this install. iat/exp span ten years so the
// keys don't silently expire on a long-lived self-hosted node.
func signSupabaseJWT(secret, role string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("jwt: empty signing secret")
	}
	now := time.Now()
	return signHS256(secret, map[string]any{
		"role": role,
		"iss":  "supabase",
		"iat":  now.Unix(),
		"exp":  now.AddDate(10, 0, 0).Unix(),
	})
}

// signHS256 encodes claims as a compact HS256 JWT signed with secret.
func signHS256(secret string, claims map[string]any) (string, error) {
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	enc := base64.RawURLEncoding
	signingInput := enc.EncodeToString(header) + "." + enc.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	return signingInput + "." + enc.EncodeToString(mac.Sum(nil)), nil
}
