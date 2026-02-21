# Task Specification: Nutmeg Bundled Executable (NBE) Format

## 1. Objective
Implement a custom SQLite Virtual File System (VFS) and a post-compilation "bundler" to allow the Nutmeg interpreter to execute code and access assets embedded directly within its own binary. This enables "zero-dependency" standalone distribution of Nutmeg applications.

## 2. The NBE File Structure
The compiler must concatenate segments into a single file. Alignment to page boundaries is mandatory to ensure performance parity with native binaries through OS demand-paging.

| Segment | Description | Alignment |
| :--- | :--- | :--- |
| **Interpreter** | The compiled C++ `nutmeg` engine binary. | N/A |
| **Padding** | Null bytes (`0x00`) to reach the next page boundary. | 4096-byte boundary |
| **SQLite DB** | The payload containing bytecode, symbols, and assets. | Start at `0x...000` |
| **Footer** | 16-byte metadata block for discovery. | End of file |

### Footer Specification (16 bytes)
The last 16 bytes of the file are reserved for the "Map":
* **Offset (8 bytes):** `uint64_t` (Little Endian). The absolute byte position where the SQLite header (`SQLite format 3`) begins.
* **Magic Tag (8 bytes):** ASCII string `NUTMEGDB`.



---

## 3. The `NutmegVFS` Implementation (C++)
To allow SQLite to read a subset of the host binary, implement a "Shim" VFS that redirects all I/O calls based on the footer offset.

### A. Core Structures
* **`NutmegFile`**: A wrapper struct inheriting from `sqlite3_file`. It must store a pointer to the actual OS file handle (`pReal`) and the `szOffset` discovered in the footer.
* **`sqlite3_io_methods`**: A custom table of function pointers where `xRead` and `xFileSize` are the primary overrides.

### B. Logic Overrides
* **`xRead`**: Given a request for `amt` bytes at `iOffset`, the shim executes:
    `pReal->pMethods->xRead(pReal, pBuf, iAmt, iOffset + szOffset)`
* **`xFileSize`**: Returns the effective size of the database:
    `*pSize = (Actual_File_Size - szOffset - 16)`
* **`xWrite` / `xTruncate`**: Explicitly return `SQLITE_READONLY`.

---

## 4. Interpreter Startup Sequence
The `nutmeg` C++ entry point must determine if it is running as a "Standalone Bundle" or a "Generic Interpreter."

1.  **Path Resolution**: Obtain the path to the current executable (e.g., `/proc/self/exe` on Linux).
2.  **Footer Check**: Seek to `EOF - 16`. Read the Magic Tag.
3.  **Branching Logic**:
    * **IF `NUTMEGDB` is found**: 
        1. Read the `uint64_t` offset.
        2. Initialize `NutmegVFS` with this offset.
        3. Open the executable itself as a database via `sqlite3_open_v2(path, &db, SQLITE_OPEN_READONLY, "nutmeg-vfs")`.
        4. Begin execution from the `main` table.
    * **ELSE**: Proceed to standard CLI mode (REPL or script execution).



---

## 5. Build-Toolchain Requirements
The Nutmeg compiler must include a "dist" or "bundle" command that:
1.  Locates the pre-compiled `nutmeg` interpreter binary.
2.  Calculates the padding required: `padding = 4096 - (interpreter_size % 4096)`.
3.  Writes the Interpreter + Padding + SQLite DB to a new file.
4.  Appends the 8-byte offset and the `NUTMEGDB` tag.
5.  Sets the file permissions to executable (`chmod +x`).

## 6. Constraints & Optimization
* **WAL Mode**: Databases must be compiled in `journal_mode = DELETE` or `OFF` as the binary is read-only.
* **Memory Mapping**: If `PRAGMA mmap_size` is desired, the VFS must also shim `xFetch` and `xUnfetch`.
* **Page Alignment**: Ensure the `szOffset` is always a multiple of 4096 to prevent unaligned disk reads and CPU cache misses.`