# Chocolatey

The package id is `bento-terminal`. `packaging/render.py` writes the nuspec and
the install scripts from a release's `SHA256SUMS`.

The release workflow packs and pushes this package on Windows when the
repository secret `CHOCO_API_KEY` is set. To push one version by hand:

```sh
gh workflow run publish-chocolatey.yml -f version=0.9.0
```

The package does not embed a binary: `tools/chocolateyinstall.ps1` downloads the
release asset and verifies it against the checksum in `SHA256SUMS`. A
`legal/VERIFICATION.txt` is therefore not needed and must not be added back;
Chocolatey moderation flags it when it lists stale release URLs and hashes
(that happened on the 0.9.0 submission).

To pack locally on Windows, or on any machine with Chocolatey:

```sh
python3 packaging/render.py --version <version> --sums SHA256SUMS --out packages
cd packages/chocolatey
choco pack
choco push "bento-terminal.<version>.nupkg" `
  --source https://push.chocolatey.org/ `
  --api-key "$env:CHOCO_API_KEY"
```

Packages are published with `choco push`. Uploading a `.nupkg` through the
community site is not supported. Get the API key from
<https://community.chocolatey.org/account>. The first submission of a version
goes through moderation before it appears in the community repository.
