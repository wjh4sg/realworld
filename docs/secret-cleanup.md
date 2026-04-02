# Secret Cleanup Guide

This repository previously contained real or directly usable local credentials in tracked files. The current tip has been sanitized, but any value that was committed before should still be treated as compromised.

## Immediate Actions

1. Rotate the MySQL application user password.
2. Rotate the MySQL root password if it was ever used outside an isolated local machine.
3. Rotate the JWT signing secret.
4. Revoke any refresh tokens or active sessions signed with the old JWT secret.
5. Check whether the same password or secret was reused in test, staging, cloud, or production environments.

## What Was Sanitized

- `config.yaml`
- `apiserver/config.yaml`
- `compose.yaml`
- `.env.example`
- `gen_gorm_model.go`

These files now contain placeholders only. Replace them locally through `.env`, private config files, or CI/CD secret managers.

## What This Change Does Not Fix

Updating files on the latest commit does not remove old secrets from Git history. Anyone who can read earlier commits may still recover them.

## Recommended History Rewrite

If you want the repository history cleaned, use `git filter-repo` or BFG Repo-Cleaner, then force-push the rewritten branch.

### Option A: `git filter-repo`

1. Install `git-filter-repo`.
2. Create a local replacement rules file, for example `replacements.txt`. Do not commit this file.
3. Rewrite history.

Example replacement rules:

```text
literal:<old-mysql-app-password>==>***REMOVED***
literal:<old-mysql-root-password>==>***REMOVED***
literal:<old-jwt-secret>==>***REMOVED***
```

Example command:

```bash
git filter-repo --replace-text replacements.txt
```

### Option B: BFG Repo-Cleaner

If you prefer BFG:

```bash
bfg --replace-text replacements.txt
git reflog expire --expire=now --all
git gc --prune=now --aggressive
```

## Publish the Clean History

After rewriting history:

```bash
git push --force-with-lease origin main
```

If there are tags that also contain the secrets, rewrite or delete and republish those tags as well.

## Coordinate With Collaborators

1. Tell collaborators that history changed.
2. Ask them to re-clone the repository, or hard-reset only after they have backed up local work they need.
3. Ask them to delete any old local `.env` files or copied secrets they no longer need.

## GitHub Follow-up

- Review open pull requests, issue comments, Actions logs, and pasted snippets that may still contain the old values.
- If the repository was public, assume the old secrets may already have been indexed or copied.
- If needed, open a GitHub support request for cache removal only after the secrets have been rotated.

## Safer Ongoing Practice

- Keep real secrets in `.env`, local-only config files, or a secret manager.
- Commit templates only, such as `.env.example`.
- Avoid reusing local passwords in any shared environment.
- Rotate JWT secrets and passwords immediately whenever they are committed by mistake.
