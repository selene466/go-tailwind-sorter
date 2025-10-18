# Tailwind Sorter

This is fork from [https://github.com/Dexter2389/go-tailwind-sorter](https://github.com/Dexter2389/go-tailwind-sorter).

Added [daisyUI](https://daisyui.com/) classes.

---

[![Go Version](https://img.shields.io/badge/go-1.18+-blue.svg)](https://go.dev/dl/)
[![Go Report Card](https://goreportcard.com/badge/github.com/Dexter2389/go-tailwind-sorter)](https://goreportcard.com/report/github.com/Dexter2389/go-tailwind-sorter)
![license](https://img.shields.io/badge/license-BSD--2--Clause-brightgreen)

A fast, standalone, and configurable tool for sorting [Tailwind CSS](https://tailwindcss.com/) classes in your project files without needing Prettier. It provides a single, dependency-free binary that you can drop into any project or CI/CD pipeline.

---

## ✨ Features

- **Extremely Fast:** Built in Go, it processes files concurrently to sort your entire codebase in milliseconds.
- **Easy Interface:** By default, it checks for violations and reports them clearly. Use the `--fix` flag to write changes to disk.
- **Zero Dependencies:** Distributes as a single binary. No need for Node.js, npm, or any other runtime.
- **Configurable:** Customize file patterns and class-like attributes via a `tailwind-sorter.toml` file.
- **`pre-commit` Integration:** Automatically sort your classes before you commit your code.
- **CI/CD Friendly:** Drop it into any GitHub Action or other CI pipeline to enforce a consistent class order.

## Motivation

Automatic code formatting has always been an invaluable part of my development process. It ensures that code is always clean, consistent, and easy to read.

For projects using Tailwind CSS, the official `prettier-plugin-tailwindcss` is the gold standard. However, my workflow is often centered around backend technologies like Go and Python. I frequently build dashboards and user interfaces using libraries like [Aether](https://github.com/pyaether) combined with [HTMX](https://htmx.org/), [AlpineJS](https://alpinejs.dev), and of course—Tailwind CSS.

This means I often work in environments without a Node.js toolchain. Setting up a whole Node.js toolchain just for one plugin felt like a step backward. I needed a solution that felt native to my stack. That's why I built `tailwind-sorter`, a single, fast, dependency-free binary that brings the power of the official sorter to any project, regardless of the tech stack.

## 🚀 Installation

### From GitHub Releases (Recommended for most users)

Download the pre-compiled binary for your operating system from the [**Latest Release**](https://github.com/Dexter2389/go-tailwind-sorter/releases/latest) page. Unpack the archive and place the `tailwind-sorter` binary in a directory included in your system's `PATH`.

### With `go install`

If you have Go installed, you can install `go-tailwind-sorter` globally:

```bash
go install github.com/dexter2389/go-tailwind-sorter@latest
```

### From Source

```bash
git clone https://github.com/Dexter2389/go-tailwind-sorter.git
cd go-tailwind-sorter

# Build with proper version info
VERSION=$(git describe --tags --abbrev=0)
go build -ldflags="-X 'github.com/dexter2389/go-tailwind-sorter/cmd.version=${VERSION}'" -o tailwind-sorter .

# Then move the binary to a location in your PATH
# sudo mv ./tailwind-sorter /usr/local/bin/
```

## Usage

The CLI is designed to be simple and intuitive, following the modern "check by default" pattern.

### Global Flags

- `--fix`: Apply fixes to files instead of just checking.
- `--config <path>`: Path to a custom TOML config file.
- `--version`: Show the application version.

### Checking for Unsorted Classes (Default)

To check one or more files or directories, simply provide them as arguments.

```bash
tailwind-sorter ./src/
```

The output will list any files with unsorted classes, along with their locations:

```
src/components/Card.jsx:12:45: TWS001 [*] Unsorted Tailwind classes
  11 |
  12 | const Card = () => (
  13 |   <div className="p-4 m-2 flex items-center">
     |                 ^^^^^^^^^^^^^^^^^^^^^^^^^^^^ TWS001
  14 |     A regular card component
     |
  = help: Sort the Tailwind CSS classes in the attribute

Found 1 violations..
[*] 1 potentially fixable with the --fix option.
```

### Applying Fixes

To automatically sort the classes and write the changes to the files, use the `--fix` flag.

```bash
tailwind-sorter --fix ./src/
```

The output will confirm that the fixes have been applied:

```
Found 1 violations (1 fixed, 0 remaining).
```

## ⚙️ Configuration

`tailwind-sorter` is configured via a TOML file. The tool can be configured in two ways:

1. Create a `tailwind-sorter.toml` file in your project's root directory, which will be discovered automatically.
2. Use a custom path with the `--config` flag: `tailwind-sorter --config /path/to/my-config.toml .`

#### Example config

The configuration must be nested under a `[tool.tailwind_sorter]` table.

```toml
# tailwind-sorter.toml

[tool.tailwind_sorter]
# Override the default file patterns to search.
# This is useful for projects with different file extensions.
file_patterns = [".py", ".templ"]

# Override the default attributes to search for class strings.
# This is useful for frameworks like Alpine.js or Aether or Templ.
class_attributes = ["_class", "class", "x-bind:class", ":class"]
```

## Git `pre-commit` Hook

Automate class sorting by integrating `tailwind-sorter` with [`pre-commit`](https://pre-commit.com/).

1. Install `pre-commit` if you haven't already.
2. Add the following to your `.pre-commit-config.yaml` file:

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: tailwind-sorter
        name: Go Tailwind Sorter
        # The entry point should match the name of the executable.
        entry: tailwind-sorter
        # Use the 'golang' language to have pre-commit build the tool from source.
        language: golang
        # The 'files' key will only run the hook on staged files
        # whose paths start with "src/".
        files: ^src/.*
        # Specify the file types that should trigger this hook.
        types_or: [py, templ]
        # Recommend using --fix to have the hook automatically apply changes.
        # The commit will fail the first time with a "files were modified"
        # message. Simply stage the changes and commit again.
        args: [--fix]
```

3. Install the hook:

   ```bash
   pre-commit install
   ```

   Now, your Tailwind classes will be sorted automatically every time you commit!

## CI/CD Integration

You can easily use `tailwind-sorter` to enforce a consistent class order in your CI pipeline.

#### Example GitHub Action

This workflow step checks for any unsorted classes and fails the build if any are found.

```yaml
- name: Check Tailwind Class Order
  run: |
    go install github.com/dexter2389/go-tailwind-sorter@latest
    tailwind-sorter .
```

## Contributing

Contributions are welcome! Whether it's a bug report, a feature request, or a pull request, we appreciate your help.

Please feel free to open an issue or submit a PR.

### Development

1. Clone the repository.
2. Install Go (see `go.mod` for the required version).
3. Run `go build .` to build the binary.

## Acknowledgements

- **Tailwind Labs** for creating `prettier-plugin-tailwindcss`, which serves as the reference for the class sorting order.
- **Astral** for creating `Ruff`, which is the inspiration for the CLI design and user experience.

## License

This project is licensed under the [BSD-2-Clause License](./LICENSE.md)
