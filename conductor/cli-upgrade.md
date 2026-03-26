# Plan: CarrotDB CLI Upgrade (Simplicity + Power)

This plan outlines the transformation of the `carrotdb` CLI from a basic TCP client into a professional-grade tool that is intuitive for beginners and powerful for advanced users.

## Objective
Implement a "dual-mode" CLI:
1.  **Simple Mode (Default):** Interactive REPL with sensible defaults, colors, and a helpful prompt.
2.  **Advanced Mode:** Flag-based connection settings and "one-shot" command execution for scripting and power users.

## Changes

### 1. `cmd/carrotdb/main.go` (CLI Core)
-   **Flag Parsing:** Add `-host` (default `localhost`) and `-port` (default `8000`).
-   **Mode Selection:** 
    -   If no positional arguments are provided, start the **Interactive REPL**.
    -   If positional arguments are provided (e.g., `carrotdb get key`), execute the command once and exit (**One-Shot Mode**).
-   **Visual Enhancements:**
    -   Use `github.com/fatih/color` for status messages and data.
    -   Implement a rich prompt: `carrot [localhost:8000]> `.
    -   Format `KEYS` output as a list instead of a single line.
-   **Error Handling:** Provide clear, color-coded error messages when connection fails or commands are invalid.

### 2. `internal/router/router.go` (Protocol Support)
-   Add a `STATUS` or `CLUSTER` command to the TCP protocol.
-   This will return a summary of shards and node health, allowing the CLI to show cluster state without hitting the HTTP API.

### 3. Documentation & Help
-   Implement a built-in `help` command in the REPL.
-   Add a `--help` flag for the CLI itself.

## Implementation Steps

### Phase 1: Structural Upgrade
1.  Modify `cmd/carrotdb/main.go` to use the `flag` package.
2.  Separate the connection logic into a helper function.
3.  Implement the logic to detect `flag.Args()` and choose between REPL and One-Shot.

### Phase 2: REPL & Visual Polish
1.  Implement the `repl()` function with color-coded input/output.
2.  Add logic to handle the `EXIT`, `QUIT`, and `HELP` commands locally within the REPL.
3.  Add special formatting for common command responses (e.g., listing keys line-by-line).

### Phase 3: Cluster Insights (Optional but recommended)
1.  Add `CLUSTER` command to `internal/router/router.go`.
2.  Update the CLI to support and display this command.

## Verification & Testing
1.  **Interactive Test:** Run `./carrotdb` and verify colors, prompt, and `HELP` command.
2.  **One-Shot Test:** Run `./carrotdb set name "Akshai"` followed by `./carrotdb get name` and verify direct output.
3.  **Remote Test:** Run `./carrotdb -host 127.0.0.1 -port 8000` to ensure flags work.
4.  **Error Test:** Try connecting to a non-existent port and verify the red error message.
