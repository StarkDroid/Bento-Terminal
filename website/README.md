# Website

The project site, built with SvelteKit and prerendered to static HTML for GitHub
Pages. It is a separate Node project; the Go build does not touch it.

```sh
cd website
npm install
npm run dev      # local preview
npm run build    # writes build/
npm run preview  # serve the build
```

`BASE_PATH` sets the path prefix. Locally it is empty; the Pages workflow passes
the repository's base path (`/Bento-Terminal`). The deploy workflow is
`.github/workflows/pages.yml` and runs when `website/` changes on `main`.

Pages: `/` (overview and install), `/docs/` (install, shortcuts, editing,
changes, terminal, builds, layout, releases, about), and `/privacy/`. The docs page keeps every
section in the prerendered HTML for search engines and switches panels with
tabs, so it stays short.

The canonical links, `static/robots.txt`, and `static/sitemap.xml` hardcode
`https://starkdroid.github.io/Bento-Terminal/`. Change all three together if the
site moves to a custom domain.

## Analytics

The website uses [GoatCounter](https://www.goatcounter.com/) for aggregate,
privacy-friendly page views and install-copy events. It does not use cookies or
track people across sites. The app itself has no telemetry.

Set the GoatCounter site code, the part before `.goatcounter.com`:

```sh
cp .env.example .env
# edit VITE_GOATCOUNTER_CODE=yourcode
npm run build
```

The Pages workflow reads the repository variable `GOATCOUNTER_CODE`. When it is
empty, the analytics script is not emitted, so forks and local builds stay
analytics-free.
