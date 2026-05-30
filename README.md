# sifr (صفر)

`sifr` (Arabic for *zero* or *cipher*) is a fast, lightweight, POSIX-compliant Caesar cipher solver written in Go, specifically optimized for **picoCTF** and general CTF challenges.

Designed to be highly compatible (compiled for macOS and Linux), dependency-free, and shell-pipeline friendly.

---

## Features

- ⚡ **Zero-dependency, Single Binary**: Built entirely using the Go standard library for instant startup.
- 🤖 **Zero-Config Intelligent Auto-Solve**: Run `sifr <ciphertext>` without any flags, and it will automatically search all 25 shifts for a standard flag prefix (default: `picoCTF`). If no matches are found, it gracefully falls back to brute-forcing all shifts.
- 🎨 **Beautiful ANSI Color Terminal Output**: Highlights matching substrings in bold green and formats shifts clearly in cyan. Automatically disables color codes when output is piped or redirected to non-TTY endpoints.
- 🧼 **Clean Output Mode (`-c`, `--clean`)**: Emits only the raw decrypted text—perfect for Unix pipeline automation (e.g. `sifr -a -c | pbcopy`).
- 📁 **Flexible Input Targets**: Seamlessly reads ciphertext directly from arguments, files, or standard input (pipes).
- 🧭 **Short & Long Flag Configurations**: Supports standard POSIX single-dash and GNU double-dash style parameters.

---

## Installation

Ensure you have [Go](https://go.dev/) installed, then run:

```bash
# Build locally
go build -o sifr

# Or install it to your $GOPATH/bin
go install
```

---

## Usage Guide

```text
sifr - A fast, POSIX-compliant Caesar cipher solver for picoCTF

Usage:
  sifr [flags] [ciphertext_or_filepath]

Flags:
  -s, --shift <int>     Shift the input by a specific number of positions (1-25)
  -b, --brute           Brute-force and print all 25 possible shifts
  -a, --auto            Auto-solve mode: check all shifts for pattern (default pattern: picoCTF)
  -p, --pattern <str>   Custom pattern to look for in auto-solve mode (default: picoCTF)
  -c, --clean           Only output the raw plaintext, omitting metadata and color codes
  -h, --help            Show this help message
```

### Examples

#### 1. Zero-Config Flag Hunting (Automatic Auto-Solve)
Pass the ciphertext as an argument. `sifr` automatically scans all shifts for `picoCTF` and highlights the match:
```bash
$ sifr "cvpbPGS{guvf_vf_n_grfg}"
Shift 13 (MATCH): picoCTF{this_is_a_test}
```

#### 2. Brute-Forcing Unfamiliar Ciphertext
If no flag matches are found, it shows all 25 possibilities:
```bash
$ sifr "uau"
No shifts matched pattern. Brute-forcing all shifts:
Shift  1: vbv
...
Shift 12: gmg
...
Shift 25: tzt
```

#### 3. Custom Flag Format Patterns
For other CTFs, use the `-p` / `--pattern` flag:
```bash
$ sifr -p "flag" "synt{guvf_vf_n_flzcyr_flag}"
Shift 13 (MATCH): flag{this_is_a_symple_synt}
```

#### 4. Clean Output & UNIX Piping
Extract and copy the decrypted flag straight to your clipboard:
```bash
$ echo "cvpbPGS{guvf_vf_n_grfg}" | sifr -c | pbcopy
```

#### 5. File Inputs
Read from an encrypted text file directly:
```bash
$ sifr flag.enc
```

## License

This software is released into the public domain under [The Unlicense](LICENSE). Feel free to copy, modify, publish, sell, or distribute it however you wish. See the `LICENSE` file for more details.
