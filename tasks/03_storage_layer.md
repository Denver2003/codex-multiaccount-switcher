# Task 03 - Metadata, Profile, and Backup Storage

## Goal

Build the storage layer for metadata, profile records, backups, and atomic writes.

## Checklist

- [ ] Create the config directory layout.
- [ ] Create the `profiles/` and `backups/` directories.
- [ ] Implement atomic JSON file writes.
- [ ] Enforce `0600` permissions for managed secret files.
- [ ] Enforce `0700` permissions for managed directories.
- [ ] Implement `metadata.json` read/write.
- [ ] Implement `profile.json` read/write.
- [ ] Implement backup file creation.
- [ ] Implement profile ID generation with the fixed MVP format.
- [ ] Implement metadata visibility rules using `metadata.json` as the commit point.
- [ ] Add tests for write ordering and orphaned file handling.

## Done When

- [ ] Profiles persist correctly on disk.
- [ ] Backups are created before destructive operations.
- [ ] Managed files use the expected permissions.
- [ ] Normal reads ignore orphaned profile directories.

