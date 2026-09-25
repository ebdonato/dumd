#!/usr/bin/env node
// Points core.hooksPath at the repo's tracked .githooks/ directory so the
// pre-commit resgen safeguard (.githooks/pre-commit) is active on every
// clone. Runs automatically as the postinstall step of `npm install` inside
// frontend/ — every dev/build workflow (AGENTS.md) already runs that.
"use strict";

const { execFileSync } = require("child_process");

try {
  const root = execFileSync("git", ["rev-parse", "--show-toplevel"], {
    encoding: "utf8",
  }).trim();
  execFileSync("git", ["config", "core.hooksPath", `${root}/.githooks`]);
} catch {
  // Not inside a git checkout (e.g. frontend/ installed standalone) — skip.
}
