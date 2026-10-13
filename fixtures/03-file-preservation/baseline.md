# Task: Safe Atomic File Update

## Objective
Implement `SafeAtomicUpdate(targetPath string, updateFn func(old []byte) ([]byte, error)) error` in Go.

## Requirements
1. If the file at `targetPath` exists, read its contents and pass them to `updateFn`. If the file does not exist, pass `nil` to `updateFn`.
2. If `updateFn` returns an error, the function must immediately return that error:
   - The file at `targetPath` must remain completely untouched.
   - Any temporary scratch file created must be deleted.
3. If `updateFn` succeeds:
   - Write new bytes to a temporary sibling file in the same directory (`filepath.Dir(targetPath)`).
   - Sync file contents to disk.
   - Atomically rename the temporary file over `targetPath`.
   - Preserve original file permissions (or default to 0644 if file was newly created).

## Interfaces
```go
package filepreservation

func SafeAtomicUpdate(targetPath string, updateFn func(old []byte) ([]byte, error)) error
```

## Constraints
- Pure Go standard library only (`os`, `path/filepath`, `fmt`).
- Under no circumstances may an existing file be corrupted or left empty when `updateFn` errors.
