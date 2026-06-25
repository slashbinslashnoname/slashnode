// Package registry queries container registries for the tags available for an
// image, so the UI can offer real, up-to-date versions instead of a hardcoded
// list. Docker Hub (the default registry) is supported; images on other
// registries return an empty list (best-effort).
package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Tags returns up to ~100 tags for a docker image, sorted newest-first using a
// version-aware ordering (so v31 precedes v28, and stable releases precede
// pre-releases/variants). Returns an empty slice (no error) for images hosted
// outside Docker Hub.
func Tags(image string) ([]string, error) {
	repo, ok := dockerHubRepo(image)
	if !ok {
		return nil, nil
	}
	url := fmt.Sprintf(
		"https://hub.docker.com/v2/repositories/%s/tags/?page_size=100&ordering=last_updated",
		repo,
	)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned %s for %s", resp.Status, repo)
	}
	var body struct {
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	tags := make([]string, 0, len(body.Results))
	for _, r := range body.Results {
		if r.Name != "" {
			tags = append(tags, r.Name)
		}
	}
	SortVersions(tags)
	return tags, nil
}

// SortedTags is retained for callers that want an explicit sorted call; Tags
// already returns version-sorted results.
func SortedTags(image string) ([]string, error) { return Tags(image) }

// LatestStable returns the best default tag from a (version-sorted) list: the
// newest tag that carries a real numeric version, is not a pre-release (rc /
// alpha / beta) and has no variant suffix (e.g. "-alpine"). Falls back to the
// first tag, or "" for an empty list. Use it to default a version picker to the
// latest production release.
func LatestStable(tags []string) string {
	for _, t := range tags {
		v := parseVersion(t)
		if v.hasNum && !v.pre && v.variant == "" {
			return t
		}
	}
	if len(tags) > 0 {
		return tags[0]
	}
	return ""
}

// SortVersions sorts tags in place, newest-first.
func SortVersions(tags []string) {
	sort.SliceStable(tags, func(i, j int) bool {
		return lessVersion(parseVersion(tags[i]), parseVersion(tags[j]))
	})
}

// version is a parsed image tag used for ordering.
type version struct {
	nums    []int  // numeric dotted components (e.g. 31.1 → [31 1])
	hasNum  bool   // the tag begins with a numeric version
	pre     bool   // pre-release (rc/alpha/beta/...)
	variant string // trailing variant after '-' (e.g. "alpine")
	raw     string
}

var preMarkers = []string{"rc", "alpha", "beta", "pre", "dev", "snapshot", "nightly"}

// parseVersion extracts a comparable version from a tag like "31.1rc1-alpine".
func parseVersion(tag string) version {
	v := version{raw: tag}
	s := strings.TrimPrefix(tag, "v")

	// Split off a variant suffix ("-alpine", "-slim", …) so "31" sorts before
	// "31-alpine" but both group under the same release.
	if dash := strings.IndexByte(s, '-'); dash >= 0 {
		v.variant = s[dash+1:]
		s = s[:dash]
	}

	// Read the leading dotted numeric version.
	rest := s
	for {
		i := 0
		for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
			i++
		}
		if i == 0 {
			break
		}
		n, _ := strconv.Atoi(rest[:i])
		v.nums = append(v.nums, n)
		v.hasNum = true
		rest = rest[i:]
		if len(rest) > 0 && rest[0] == '.' {
			rest = rest[1:]
			continue
		}
		break
	}
	// Whatever trails the numeric prefix (e.g. "rc1") marks a pre-release.
	lower := strings.ToLower(rest + "-" + v.variant)
	for _, m := range preMarkers {
		if strings.Contains(lower, m) {
			v.pre = true
			break
		}
	}
	return v
}

// lessVersion reports whether a should sort before b in newest-first order.
func lessVersion(a, b version) bool {
	// Numeric versions come before non-numeric tags (master, latest, …).
	if a.hasNum != b.hasNum {
		return a.hasNum
	}
	if a.hasNum && b.hasNum {
		for k := 0; k < len(a.nums) && k < len(b.nums); k++ {
			if a.nums[k] != b.nums[k] {
				return a.nums[k] > b.nums[k] // higher first
			}
		}
		if len(a.nums) != len(b.nums) {
			return len(a.nums) > len(b.nums) // more specific first (31.1 before 31)
		}
	}
	// Same base version: stable before pre-release.
	if a.pre != b.pre {
		return !a.pre
	}
	// Base image before variant (-alpine, -slim).
	if (a.variant == "") != (b.variant == "") {
		return a.variant == ""
	}
	return a.raw > b.raw
}

// dockerHubRepo extracts the Docker Hub repository ("namespace/name", or
// "library/name" for official images) from an image ref, returning false when
// the image is hosted on another registry (a host with a dot/port).
func dockerHubRepo(image string) (string, bool) {
	name := image
	if at := strings.Index(name, "@"); at >= 0 {
		name = name[:at]
	}
	// Strip the tag (the ':' in the final path segment).
	if slash := strings.LastIndex(name, "/"); slash >= 0 {
		if colon := strings.Index(name[slash+1:], ":"); colon >= 0 {
			name = name[:slash+1+colon]
		}
	} else if colon := strings.Index(name, ":"); colon >= 0 {
		name = name[:colon]
	}

	parts := strings.Split(name, "/")
	if len(parts) > 1 {
		if first := parts[0]; strings.ContainsAny(first, ".:") || first == "localhost" {
			return "", false // explicit non-Docker-Hub registry
		}
	}
	switch len(parts) {
	case 1:
		return "library/" + parts[0], true
	case 2:
		return name, true
	default:
		return "", false
	}
}
