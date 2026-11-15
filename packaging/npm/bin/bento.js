#!/usr/bin/env node
'use strict';

const { spawn } = require('node:child_process');
const { ensure } = require('../lib/install');

async function main() {
  let binary;
  try {
    binary = await ensure();
  } catch (error) {
    console.error(`bento: ${error.message}`);
    process.exit(1);
  }

  const child = spawn(binary, process.argv.slice(2), {
    stdio: 'inherit',
    env: process.env,
  });

  child.on('error', (error) => {
    console.error(`bento: ${error.message}`);
    process.exit(1);
  });

  child.on('exit', (code, signal) => {
    if (signal) {
      try {
        process.kill(process.pid, signal);
      } catch {
        process.exit(1);
      }
      return;
    }
    process.exit(code ?? 0);
  });
}

main();
