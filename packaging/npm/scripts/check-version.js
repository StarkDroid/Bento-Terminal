'use strict';

// The release workflow stamps the real version before publishing. A checkout
// that still says 0.0.0 must not reach the registry.
const { version } = require('../package.json');

if (version === '0.0.0') {
  console.error(
    'bento: set package.json version to the matching GitHub release before publishing',
  );
  process.exit(1);
}
