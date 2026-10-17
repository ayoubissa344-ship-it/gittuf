# gittuf Fork: Correctness Fixes

This fork documents two targeted correctness fixes developed
for the gittuf codebase.

## 1. Idempotent Upstream Path Propagation

Pull request:
https://github.com/ayoubissa344-ship-it/gittuf/pull/1

- Addresses unnecessary Reference State Log (RSL) entries
  during repeated propagation when no additional change is needed.
- Adds a regression test for repeated propagation.
- Focuses on idempotency and state consistency.

## 2. RSL Ref-Tip Synchronization

Pull request:
https://github.com/ayoubissa344-ship-it/gittuf/pull/2

- Updates ref-tip reconstruction to account for propagation entries.
- Adds regression tests covering propagation-related reference
  synchronization in push and pull flows.
- Focuses on reference-state consistency.

## Validation Status

Regression tests have been added, but they have not yet been
executed in the development environment because the installed
Go toolchain does not meet the repository's required version.

Runtime validation remains pending. No claim of passing tests
is made here.

## Contribution Status

Both pull requests were merged into the author's public fork.
This does not imply acceptance into the upstream gittuf repository.

## Scope

These notes describe the changes and their current validation
status. They do not replace upstream review or independent testing.
