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

Pages: `/` (overview and install) and `/docs/` (install options, releases,
shortcuts, layout, about). The docs page keeps every section in the prerendered
HTML for search engines and switches panels with tabs, so it stays short.
