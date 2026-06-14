#!/usr/bin/env python3
"""Register this provider release in the Terrakube private registry.

Expects a public GitHub repo so release asset URLs work for Terrakube executors
without authentication.
"""

from __future__ import annotations

import hashlib
import json
import os
import re
import sys
import urllib.error
import urllib.request

TERRAKUBE_ENDPOINT = os.environ.get(
    "TERRAKUBE_ENDPOINT",
    os.environ.get("TERRAKUBE_API_URL", "https://terrakube-api.increbesco.net"),
)
TERRAKUBE_TOKEN = os.environ.get("TERRAKUBE_TOKEN", "")
TERRAKUBE_ORGANIZATION = os.environ.get("TERRAKUBE_ORGANIZATION", "outpost")
RELEASE_TAG = os.environ.get("RELEASE_TAG", "")
GITHUB_REPOSITORY = os.environ.get("GITHUB_REPOSITORY", "OutpostCloudCo/terraform-provider-outpost")
GITHUB_TOKEN = os.environ.get("GITHUB_TOKEN", "")

PROVIDER_NAME = "outpost"
PROTOCOLS = "5.0"
JSONAPI = "application/vnd.api+json"
USER_AGENT = "outpost-terraform-provider-register/1.0"


class RegisterError(Exception):
    pass


def github_headers() -> dict[str, str]:
    headers = {
        "Accept": "application/vnd.github+json",
        "User-Agent": USER_AGENT,
    }
    if GITHUB_TOKEN:
        headers["Authorization"] = f"Bearer {GITHUB_TOKEN}"
    return headers


def http_get(url: str) -> bytes:
    headers = {"User-Agent": USER_AGENT}
    req = urllib.request.Request(url, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=120) as resp:
            return resp.read()
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode(errors="replace")
        raise RegisterError(f"GET {url} -> {exc.code}: {detail}") from exc
    except urllib.error.URLError as exc:
        raise RegisterError(f"GET {url} failed: {exc}") from exc


def api_request(method: str, url: str, token: str, body: dict | None = None) -> dict:
    headers = {
        "Authorization": f"Bearer {token}",
        "Accept": JSONAPI,
        "Content-Type": JSONAPI,
        "User-Agent": USER_AGENT,
    }
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            raw = resp.read().decode()
            return json.loads(raw) if raw else {}
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode(errors="replace")
        raise RegisterError(f"{method} {url} -> {exc.code}: {detail}") from exc


def sha256_url(url: str) -> str:
    digest = hashlib.sha256()
    digest.update(http_get(url))
    return digest.hexdigest()


def parse_shasum(shasums_text: str, filename: str) -> str | None:
    for line in shasums_text.splitlines():
        parts = line.strip().split()
        if len(parts) >= 2 and parts[1] == filename:
            return parts[0]
    return None


def release_assets(tag: str) -> dict[str, str]:
    """Return release asset name -> public browser_download_url."""
    url = f"https://api.github.com/repos/{GITHUB_REPOSITORY}/releases/tags/{tag}"
    req = urllib.request.Request(url, headers=github_headers())
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            payload = json.loads(resp.read().decode())
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode(errors="replace")
        raise RegisterError(f"GET {url} -> {exc.code}: {detail}") from exc

    assets = {
        asset["name"]: asset["browser_download_url"]
        for asset in payload.get("assets", [])
    }
    if not assets:
        raise RegisterError(f"no release assets found for {tag}")
    return assets


def pick_linux_amd64_zip(assets: dict[str, str]) -> tuple[str, str]:
    for name, url in assets.items():
        if name.endswith("_linux_amd64.zip"):
            return name, url
    raise RegisterError(f"no linux_amd64.zip asset in release: {list(assets)}")


def pick_shasums_file(assets: dict[str, str]) -> tuple[str, str] | None:
    for name, url in assets.items():
        if "SHA256SUMS" in name or name == "checksums.txt":
            return name, url
    return None


def parse_version(tag: str) -> str:
    match = re.fullmatch(r"v(.+)", tag.strip())
    if not match:
        raise RegisterError(f"RELEASE_TAG must look like v0.1.0, got {tag!r}")
    return match.group(1)


def main() -> None:
    if not TERRAKUBE_TOKEN:
        raise RegisterError("TERRAKUBE_TOKEN is required")
    if not RELEASE_TAG:
        raise RegisterError("RELEASE_TAG is required")

    version = parse_version(RELEASE_TAG)
    endpoint = TERRAKUBE_ENDPOINT.rstrip("/")

    orgs = api_request("GET", f"{endpoint}/api/v1/organization", TERRAKUBE_TOKEN)
    org_id = next(
        (o["id"] for o in orgs.get("data", []) if o.get("attributes", {}).get("name") == TERRAKUBE_ORGANIZATION),
        None,
    )
    if not org_id:
        raise RegisterError(f"organization not found: {TERRAKUBE_ORGANIZATION}")

    providers = api_request("GET", f"{endpoint}/api/v1/organization/{org_id}/provider", TERRAKUBE_TOKEN)
    provider_id = next(
        (p["id"] for p in providers.get("data", []) if p.get("attributes", {}).get("name") == PROVIDER_NAME),
        None,
    )
    if not provider_id:
        created = api_request(
            "POST",
            f"{endpoint}/api/v1/organization/{org_id}/provider",
            TERRAKUBE_TOKEN,
            {
                "data": {
                    "type": "provider",
                    "attributes": {
                        "name": PROVIDER_NAME,
                        "description": "Outpost utility provider",
                    },
                }
            },
        )
        provider_id = created.get("data", {}).get("id")
        print(f"created provider {PROVIDER_NAME} ({provider_id})")
    else:
        print(f"provider {PROVIDER_NAME} exists ({provider_id})")

    versions = api_request(
        "GET",
        f"{endpoint}/api/v1/organization/{org_id}/provider/{provider_id}/version",
        TERRAKUBE_TOKEN,
    )
    version_id = next(
        (
            v["id"]
            for v in versions.get("data", [])
            if v.get("attributes", {}).get("versionNumber") == version
        ),
        None,
    )
    if not version_id:
        created = api_request(
            "POST",
            f"{endpoint}/api/v1/organization/{org_id}/provider/{provider_id}/version",
            TERRAKUBE_TOKEN,
            {
                "data": {
                    "type": "version",
                    "attributes": {"versionNumber": version, "protocols": PROTOCOLS},
                }
            },
        )
        version_id = created.get("data", {}).get("id")
        print(f"created version {version} ({version_id})")
    else:
        print(f"version {version} exists ({version_id})")

    implementations = api_request(
        "GET",
        f"{endpoint}/api/v1/organization/{org_id}/provider/{provider_id}/version/{version_id}/implementation",
        TERRAKUBE_TOKEN,
    )
    for item in implementations.get("data", []):
        attrs = item.get("attributes", {})
        if attrs.get("os") == "linux" and attrs.get("arch") == "amd64":
            print("linux/amd64 implementation already registered")
            return

    assets = release_assets(RELEASE_TAG)
    filename, download_url = pick_linux_amd64_zip(assets)
    shasums_url: str | None = None
    shasum: str | None = None

    shasums = pick_shasums_file(assets)
    if shasums:
        shasums_name, shasums_url = shasums
        try:
            shasum = parse_shasum(http_get(shasums_url).decode(), filename)
        except RegisterError as exc:
            print(f"warning: could not read {shasums_name}: {exc}")

    if not shasum:
        print(f"hashing {filename} from GitHub release")
        shasum = sha256_url(download_url)

    # Terrakube DB columns are NOT NULL; unsigned providers use empty placeholders.
    attrs: dict[str, str] = {
        "os": "linux",
        "arch": "amd64",
        "filename": filename,
        "downloadUrl": download_url,
        "shasum": shasum,
        "shasumsUrl": shasums_url or "",
        "shasumsSignatureUrl": "",
        "keyId": "",
        "asciiArmor": "",
        "trustSignature": "",
        "source": "",
        "sourceUrl": "",
    }

    api_request(
        "POST",
        f"{endpoint}/api/v1/organization/{org_id}/provider/{provider_id}/version/{version_id}/implementation",
        TERRAKUBE_TOKEN,
        {"data": {"type": "implementation", "attributes": attrs}},
    )
    print(f"registered {PROVIDER_NAME} {version} linux/amd64")


if __name__ == "__main__":
    try:
        main()
    except RegisterError as exc:
        print(f"error: {exc}", file=sys.stderr)
        sys.exit(1)
