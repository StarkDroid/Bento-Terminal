# Chocolatey

The package id is `bento-terminal`. `packaging/render.py` writes the nuspec and
the install scripts from a release's `SHA256SUMS`.

The release workflow packs and pushes this package on Windows when the
repository secret `CHOCO_API_KEY` is set.

To push a package built locally:

```sh
choco pack
choco push bento-terminal.0.7.0.nupkg \
  --source https://push.chocolatey.org/ \
  --api-key "$CHOCO_API_KEY"
```

On macOS or Linux, push the same `.nupkg` through the Chocolatey API:

```sh
curl -X PUT \
  -H "X-NuGet-ApiKey: $CHOCO_API_KEY" \
  --upload-file packaging/chocolatey/dist/bento-terminal.0.7.0.nupkg \
  https://push.chocolatey.org/api/v2/package/
```

Get the API key from <https://community.chocolatey.org/account>. The first
submission goes through Chocolatey moderation before it appears in the
community repository.
