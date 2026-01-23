# Ansible vs Network-Builder Comparitive Analysis Report

This document records the results of the comparative analysis of network-builder and Ansible configurations, as well as any found deviations and fixes.

## Directory Structure Comparison

### Confirmed to be Consistent

- ✅ Directory Naming Convention: `orderer-{type}-{index}/config/` and `committer-{type}/config/`
- ✅ MSP Directory Stucture: `msp/admincerts/`, `msp/cacerts/`, `msp/keystore/`, `msp/signcerts/`, `msp/tlscacerts/`
- ✅ TLS Directory Structure: `tls/` 包含 `ca.crt`, `server.crt`, `server.key`
- ✅ Store Directory: orderer node contains `store/` directory

### Difference Description

- **out_origin** Contains directories generated at run time (e.g.`store/chains/`,`store/batchDB/`,`pgdata/`) - these are generated after the container runs
- **out** Only include the initially generated directory, this is normal

## Configuration File Comparison

### 1. File Permissions

#### Question

- **Ansible**: Configuration file usage `0o640` (rw-r-----), directory usage `0o750` (rwxr-x---)
- **Builder (before repair)**: Use default permissions `0o644` (rw-r--r--) and `0o755` (rwxr-xr-x)

#### Repair

- ✅ Modify `template.go` in `executeTemplate()` function that sets the permissions after writing the file to`0o640`
- ✅ Modify all `MkdirAll()` calls to change permissions from `0755` to `0750`
- ✅ Modify `copyFile()` function to preserve original permissions of sensitive files

**Document**: `network-builder/internal/template/template.go`

### 2. Committer Sidecar Configuration

#### Question

- **Ansible**: `genesis-block-file-path:` empty, do not copy `genesis.block` file
- **Builder (before repair)**: set `genesis-block-file-path: /config/genesis.block` and copied the file

#### Repair

- ✅ Removed copying for sidecar `genesis.block` code
- ✅ Modify the template and add `genesis-block-file-path` set to null

**Document**:

- `network-builder/internal/template/template.go` (lines 150-152)
- `network-builder/internal/template/committer.go` (line 198)

### 3. Committer DB Configuration

#### Question

- **Ansible**: Do not generate configuration files for db type
- **Builder (before repair)**: Generated for db type `config-db.yml` file

#### Repair

- ✅ In `generateCommitterConfigs()`, skip db type configuration file generation
- ✅ Create data directory only for db type

**File**: `network-builder/internal/template/template.go` (lines 135-143)

## Docker Compose Configuration Comparison

### 1. Container User (container_run_as_host_user)

#### Question

- **Ansible**: Use `container_run_as_host_user: true`, setting `user: "{{ ansible_facts.user_uid ~ ':' ~ ansible_facts.user_gid }}"`
- **Builder (before repair)**: No setting `user` field

#### Repair

- ✅ Add `getCurrentUserUIDGID()` function to get the UID:GID of the current user
- ✅ In `buildOrdererService()` and `buildCommitterService()` medium settings `User` field
- ✅ PostgreSQL container does not set user (uses default postgres user)

**Document**: `network-builder/internal/compose/compose.go`

### 2. Mounting the PostgreSQL Data Directory

#### Question

- **Ansible**: Mount to `/var/lib/postgresql/data:Z`, settings `PGDATA=/var/lib/postgresql/data/pgdata`
- **Builder (before repair)**: Mount to `/config` (error)

#### Repair

- ✅ Modify the data directory to be mounted as `/var/lib/postgresql/data:Z`
- ✅ Add `PGDATA` environment variables
- ✅ Use separate `data` directory (instead of `config` directory)
- ✅ Remove PostgreSQL container `working_dir` settings

**File**: `network-builder/internal/compose/compose.go` (lines 343-375)

### 3. Volume Mount Flag

#### Confirmed

- ✅ PostgreSQL data directory usage `:Z` Flags (SELinux context)
- ✅ Other volumes are mounted correctly

### 4. Health Checkup

#### Confirmed

- ✅ Orderer component: use `nc -z localhost {port}` check
- ✅ Committer component: use `nc -z localhost {port}` check
- ✅ PostgreSQL：use `pg_isready` check
- ✅ Timeout and retry settings consistent with Ansible

### 5. Working Directory

#### Confirmed

- ✅ Orderer and Committer components: `working_dir: /config`
- ✅ PostgreSQL: Not Setting `working_dir` (Fixed)

## Environment Variable Comparison

### Confirmed to be Consistent

- ✅ PostgreSQL environment variables: `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `PGDATA`
- ✅ HSM environment variables: `SOFTHSM2_CONF` (when enabled)

## Summary of Fixed Deviations

| Deviation Item                         | Status   | Repair Location                                               |
| -------------------------------------- | -------- | ------------------------------------------------------------- |
| File Permissions (Configuration Files) | ✅ Fixed | `template.go: executeTemplate()`                              |
| Directory Permissions                  | ✅ Fixed | `template.go: MkdirAll()`call                                 |
| Sidecar genesis-block-file-path        | ✅ Fixed | `template.go`,`committer.go`                                  |
| Committer DB configuration file        | ✅ Fixed | `template.go: generateCommitterConfigs()`                     |
| Container user (run_as_host_user)      | ✅ Fixed | `compose.go: buildOrdererService()`,`buildCommitterService()` |
| PostgreSQL data directory mount        | ✅ Fixed | `compose.go: buildCommitterService()`                         |
| PostgreSQL PGDATA environment variable | ✅ Fixed | `compose.go: buildCommitterService()`                         |
| PostgreSQL working_dir                 | ✅ Fixed | `compose.go: buildCommitterService()`                         |

## Verified Consistent Configuration

- ✅ Directory structure and naming conventions
- ✅ MSP and TLS directory structure
- ✅ Store catalog creation
- ✅ Health check configuration
- ✅ Network configuration
- ✅ Port mapping
- ✅ Service dependencies

## Things to Note

1.  **Certificate Differences**: It is normal for the certificate content generated each time to be different, as long as the certificate structure and paths are consistent.
2.  **Runtime File**: `out_origin` contains files generated at runtime (such as database data, store subdirectories) that are not part of the initial configuration.
3.  **Number of Files**: `out_origin` has 21 files (post-run state), `out` has 38 files (initial build state), differences are normal

## Next Steps

1. Rebuild the configuration to apply all fixes: `make setup`
2. Verify file permissions are correct
3. Verify that docker-compose configuration is correct
4. Test container up and running

## Reference

- Ansible role location: `.ansible/collections/ansible_collections/hyperledger/fabricx/roles/`
- Network-builder code location: `network-builder/internal/`
