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
	"strings"
	"time"
)

// Tags returns up to ~100 recent tags for a docker image. Returns an empty
// slice (no error) for images hosted outside Docker Hub.
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
	return tags, nil
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

// SortedTags is Tags with the results sorted descending (newest-looking first),
// kept as a convenience for callers that want stable ordering.
func SortedTags(image string) ([]string, error) {
	tags, err := Tags(image)
	if err != nil {
		return nil, err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(tags)))
	return tags, nil
}
