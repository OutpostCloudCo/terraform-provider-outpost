#!/usr/bin/env python3
"""List Terrakube orgs/providers/versions/implementations and probe registry URLs."""

from __future__ import annotations

import json
import os
import sys
import urllib.error
import urllib.request

ENDPOINT = os.environ.get(
    "TERRAKUBE_ENDPOINT",
    os.environ.get("TERRAKUBE_API_URL", "https://terrakube-api.increbesco.net"),
).rstrip("/")
TOKEN = os.environ.get("TERRAKUBE_TOKEN", "")
REGISTRY = os.environ.get(
    "TERRAKUBE_REGISTRY",
    "https://terrakube-registry.increbesco.net",
).rstrip("/")
PROVIDER = os.environ.get("PROVIDER_NAME", "outpost")
JSONAPI = "application/vnd.api+json"


def api_get(path: str) -> dict:
    req = urllib.request.Request(
        f"{ENDPOINT}{path}",
        headers={
            "Authorization": f"Bearer {TOKEN}",
            "Accept": JSONAPI,
        },
    )
    with urllib.request.urlopen(req, timeout=60) as resp:
        return json.loads(resp.read().decode())


def registry_get(path: str) -> tuple[int, str]:
    req = urllib.request.Request(
        f"{REGISTRY}{path}",
        headers={"Authorization": f"Bearer {TOKEN}"},
    )
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            return resp.status, resp.read().decode()
    except urllib.error.HTTPError as exc:
        return exc.code, exc.read().decode(errors="replace")


def main() -> None:
    if not TOKEN:
        print("Set TERRAKUBE_TOKEN", file=sys.stderr)
        sys.exit(1)

    orgs = api_get("/api/v1/organization").get("data", [])
    print("=== Organizations ===")
    for org in orgs:
        name = org.get("attributes", {}).get("name")
        org_id = org["id"]
        print(f"  {name!r}  id={org_id}")

        providers = api_get(f"/api/v1/organization/{org_id}/provider").get("data", [])
        for prov in providers:
            pname = prov.get("attributes", {}).get("name")
            if pname != PROVIDER:
                continue
            pid = prov["id"]
            print(f"    provider {pname!r}  id={pid}")

            versions = api_get(
                f"/api/v1/organization/{org_id}/provider/{pid}/version"
            ).get("data", [])
            for ver in versions:
                vnum = ver.get("attributes", {}).get("versionNumber")
                vid = ver["id"]
                impls = api_get(
                    f"/api/v1/organization/{org_id}/provider/{pid}/version/{vid}/implementation"
                ).get("data", [])
                platforms = [
                    f"{i.get('attributes', {}).get('os')}/{i.get('attributes', {}).get('arch')}"
                    for i in impls
                ]
                print(f"      version {vnum}  implementations: {platforms or 'NONE'}")

    print("\n=== Registry version lists (needs linux/amd64 implementation) ===")
    for org_name in {o.get("attributes", {}).get("name") for o in orgs}:
        if not org_name:
            continue
        code, body = registry_get(
            f"/terraform/providers/v1/{org_name}/{PROVIDER}/versions"
        )
        print(f"  {org_name}/{PROVIDER} -> HTTP {code}: {body[:200]}")


if __name__ == "__main__":
    main()
