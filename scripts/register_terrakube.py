#!/usr/bin/env python3
"""Register this provider release in the Terrakube private registry."""

from __future__ import annotations

import hashlib
import json
import os
import re
import sys
import urllib.error
import urllib.request

TERRAKUBE_ENDPOINT = os.environ.get("TERRAKUBE_ENDPOINT", "https://terrakube-api.increbesco.net")
TERRAKUBE_TOKEN = os.environ.get("TERRAKUBE_TOKEN", "")
TERRAKUBE_ORGANIZATION = os.environ.get("TERRAKUBE_ORGANIZATION", "Outpost")
RELEASE_TAG = os.environ.get("RELEASE_TAG", "")
GITHUB_REPOSITORY = os.environ.get("GITHUB_REPOSITORY", "OutpostCloudCo/terraform-provider-outpost")

PROVIDER_NAME = "outpost"
PROTOCOLS = "5.0"
JSONAPI = "application/vnd.api+json"


class RegisterError(Exception):
    pass


def api_request(method: str, url: str, token: str, body: dict | None = None) -> dict:
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(
        url,
        data=data,
        headers={
            "Authorization": f"Bearer {token}",
            "Accept": JSONAPI,
            "Content-Type": JSONAPI,
        },
        method=method,
    )
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            raw = resp.read().decode()
            return json.loads(raw) if raw else {}
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode(errors="replace")
        raise RegisterError(f"{method} {url} -> {exc.code}: {detail}") from exc


def http_text(url: str) -> str:
    req = urllib.request.Request(url, headers={"User-Agent": "outpost-provider-register/1.0"})
    with urllib.request.urlopen(req, timeout=120) as resp:
        return resp.read().decode()


def http_sha256(url: str) -> str:
    req = urllib.request.Request(url, headers={"User-Agent": "outpost-provider-register/1.0"})
    digest = hashlib.sha256()
    with urllib.request.urlopen(req, timeout=120) as resp:
        while True:
            chunk = resp.read(65536)
            if not chunk:
                break
            digest.update(chunk)
    return digest.hexdigest()


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
    repo_name = GITHUB_REPOSITORY.split("/", 1)[-1]
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

    release_base = f"https://github.com/{GITHUB_REPOSITORY}/releases/download/{RELEASE_TAG}"
    filename = f"{repo_name}_{version}_linux_amd64.zip"
    download_url = f"{release_base}/{filename}"
    shasums_url = f"{release_base}/{repo_name}_{version}_SHA256SUMS"

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

    shasum = None
    try:
        for line in http_text(shasums_url).splitlines():
            parts = line.strip().split()
            if len(parts) >= 2 and parts[1] == filename:
                shasum = parts[0]
                break
    except RegisterError:
        pass
    if not shasum:
        shasum = http_sha256(download_url)

    api_request(
        "POST",
        f"{endpoint}/api/v1/organization/{org_id}/provider/{provider_id}/version/{version_id}/implementation",
        TERRAKUBE_TOKEN,
        {
            "data": {
                "type": "implementation",
                "attributes": {
                    "os": "linux",
                    "arch": "amd64",
                    "filename": filename,
                    "downloadUrl": download_url,
                    "shasum": shasum,
                    "shasumsUrl": shasums_url,
                },
            }
        },
    )
    print(f"registered {PROVIDER_NAME} {version} linux/amd64")


if __name__ == "__main__":
    try:
        main()
    except RegisterError as exc:
        print(f"error: {exc}", file=sys.stderr)
        sys.exit(1)
