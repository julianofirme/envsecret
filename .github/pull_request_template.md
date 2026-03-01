## Summary

<!-- What does this PR do and why? -->

## Changes

<!-- List the main changes -->
-
-

## Type of change

- [ ] Bug fix
- [ ] New feature
- [ ] Refactor
- [ ] Documentation
- [ ] CI/tooling

## Security checklist

<!-- Required for changes touching vault, keychain, or the run command -->
- [ ] No secrets are written to disk, logs, or stdout
- [ ] Vault writes go through `vault.Save` (atomic + 0600)
- [ ] `run` still uses `syscall.Exec` (not `exec.Command`)
- [ ] New keys are validated with `validateKey`

## Testing

<!-- How was this tested? -->

## Related issues

<!-- Closes #... -->
