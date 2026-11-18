'use strict';

// Download the binary during npm install so the first `bento` run is instant.
// A failure is a warning, not an install failure: the launcher downloads again
// on first use and reports the real error there. `npm install --ignore-scripts`
// therefore still works.
const { ensure } = require('../lib/install');

if (process.env.BENTO_SKIP_DOWNLOAD) {
  process.exit(0);
}

ensure()
  .then(() => process.exit(0))
  .catch((error) => {
    console.warn(`bento: could not download the binary during install: ${error.message}`);
    console.warn('bento: it will try again on first run.');
    process.exit(0);
  });
