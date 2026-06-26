import type { App } from "./api";

// tagOf extracts the tag from a docker image ref (repo[:tag], registry host with
// a port preserved). Defaults to "latest" when no tag is present.
function tagOf(image: string): string {
  const slash = image.lastIndexOf("/");
  const last = slash >= 0 ? image.slice(slash + 1) : image;
  const colon = last.indexOf(":");
  return colon < 0 ? "latest" : last.slice(colon + 1);
}

// appVersion returns the version to display for an app: the docker image tag it
// actually runs (so a bitcoind pinned to :31 shows "31", not the manifest's
// fixed "28.1"). For multi-image apps whose tags differ there's no single docker
// version, so it falls back to the installed/manifest version.
export function appVersion(app: App): string {
  const tags = app.images ? [...new Set(Object.values(app.images).map(tagOf))] : [];
  if (tags.length === 1) return tags[0];
  return app.installed_version || app.version;
}
