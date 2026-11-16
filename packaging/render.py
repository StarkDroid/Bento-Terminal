#!/usr/bin/env python3
"""Render the package-manager manifests for one Bento terminal release.

The release workflow downloads the published SHA256SUMS and runs this script;
the same command works locally:

    python3 packaging/render.py --version 0.7.0 --sums SHA256SUMS --out /tmp/bento-packages

The rendered tree is what the Homebrew, Scoop, WinGet, Chocolatey, and nfpm
publishers copy into their own repositories.
"""

from __future__ import annotations

import argparse
import re
import shutil
import sys
from pathlib import Path

VERSION_RE = re.compile(r"^[0-9]+\.[0-9]+\.[0-9]+$")
PLACEHOLDER_RE = re.compile(r"\{\{([A-Z0-9_]+)\}\}")

ROOT = Path(__file__).resolve().parent

ASSETS = {
    "DARWIN_ARM64": "bento_darwin_arm64",
    "DARWIN_AMD64": "bento_darwin_amd64",
    "LINUX_ARM64": "bento_linux_arm64",
    "LINUX_AMD64": "bento_linux_amd64",
    "WINDOWS_AMD64": "bento_windows_amd64.exe",
    "WINDOWS_ARM64": "bento_windows_arm64.exe",
}


def parse_sums(path: Path) -> dict[str, str]:
    sums: dict[str, str] = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        match = re.match(r"^([a-fA-F0-9]{64})\s+\*?(.+?)\s*$", line)
        if match:
            sums[match.group(2)] = match.group(1).lower()
    return sums


def values_for(version: str, sums: dict[str, str]) -> dict[str, str]:
    values = {"VERSION": version}
    for key, asset in ASSETS.items():
        digest = sums.get(asset)
        if not digest:
            raise SystemExit(f"render: SHA256SUMS does not list {asset}")
        values[f"SHA256_{key}"] = digest
        values[f"SHA256_{key}_UPPER"] = digest.upper()
    return values


def render(text: str, values: dict[str, str]) -> str:
    def replace(match: re.Match[str]) -> str:
        key = match.group(1)
        if key not in values:
            raise SystemExit(f"render: no value for {{{{{key}}}}}")
        return values[key]

    return PLACEHOLDER_RE.sub(replace, text)


def write(out: Path, relative: str, text: str) -> None:
    path = out / relative
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text, encoding="utf-8")
    print(path)


def copy(out: Path, relative: str, source: Path) -> None:
    path = out / relative
    path.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(source, path)
    print(path)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--version", required=True, help="release version, without v")
    parser.add_argument("--sums", required=True, type=Path, help="published SHA256SUMS")
    parser.add_argument("--out", required=True, type=Path, help="directory to write")
    args = parser.parse_args()

    version = args.version.removeprefix("v")
    if not VERSION_RE.match(version):
        raise SystemExit(f"render: {version} is not a version like 0.7.0")
    sums = parse_sums(args.sums)
    values = values_for(version, sums)
    out = args.out.resolve()

    templates = {
        "homebrew/bento.rb.tmpl": "homebrew/bento.rb",
        "scoop/bento.json.tmpl": "scoop/bento.json",
        "chocolatey/bento-terminal.nuspec.tmpl": "chocolatey/bento-terminal.nuspec",
        "chocolatey/tools/chocolateyinstall.ps1.tmpl": "chocolatey/tools/chocolateyinstall.ps1",
    }
    for source, target in templates.items():
        write(out, target, render((ROOT / source).read_text(encoding="utf-8"), values))

    winget = f"winget/manifests/s/StarkDroid/Bento/{version}"
    for name in (
        "StarkDroid.Bento.yaml",
        "StarkDroid.Bento.installer.yaml",
        "StarkDroid.Bento.locale.en-US.yaml",
    ):
        write(
            out,
            f"{winget}/{name}",
            render((ROOT / "winget" / f"{name}.tmpl").read_text(encoding="utf-8"), values),
        )

    for arch in ("amd64", "arm64"):
        nfpm = (ROOT / "nfpm/nfpm.yaml.tmpl").read_text(encoding="utf-8")
        nfpm = nfpm.replace("{{ARCH}}", arch).replace(
            "{{SRC}}", f"dist/bento_linux_{arch}"
        )
        write(out, f"nfpm/{arch}.yaml", render(nfpm, values))

    copy(
        out,
        "chocolatey/tools/chocolateyuninstall.ps1",
        ROOT / "chocolatey/tools/chocolateyuninstall.ps1",
    )
    legal = ROOT / "chocolatey/legal"
    if legal.is_dir():
        for source in sorted(legal.iterdir()):
            if source.is_file():
                copy(out, f"chocolatey/legal/{source.name}", source)
    return 0


if __name__ == "__main__":
    sys.exit(main())
