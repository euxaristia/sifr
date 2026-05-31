package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	shiftVal    int
	bruteMode   bool
	autoMode    bool
	patternVal  string
	cleanMode   bool
	versionMode bool
)

func init() {
	flag.IntVar(&shiftVal, "s", 0, "Shift the input by a specific number of positions (1-25)")
	flag.IntVar(&shiftVal, "shift", 0, "Shift the input by a specific number of positions (1-25)")

	flag.BoolVar(&bruteMode, "b", false, "Brute-force and print all 25 possible shifts")
	flag.BoolVar(&bruteMode, "brute", false, "Brute-force and print all 25 possible shifts")

	flag.BoolVar(&autoMode, "a", false, "Auto-solve mode: check all shifts for pattern")
	flag.BoolVar(&autoMode, "auto", false, "Auto-solve mode: check all shifts for pattern")

	flag.StringVar(&patternVal, "p", "picoCTF", "Custom pattern to look for in auto-solve mode")
	flag.StringVar(&patternVal, "pattern", "picoCTF", "Custom pattern to look for in auto-solve mode")

	flag.BoolVar(&cleanMode, "c", false, "Only output the raw plaintext, omitting metadata and color codes")
	flag.BoolVar(&cleanMode, "clean", false, "Only output the raw plaintext, omitting metadata and color codes")

	flag.BoolVar(&versionMode, "v", false, "Print version information")
	flag.BoolVar(&versionMode, "version", false, "Print version information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "sifr - A fast, POSIX-compliant Caesar cipher solver for picoCTF\n")
		fmt.Fprintf(os.Stderr, "Created by euxaristia (https://github.com/euxaristia)\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  sifr [flags] [ciphertext_or_filepath]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fmt.Fprintf(os.Stderr, "  -s, --shift <int>     Shift the input by a specific number of positions (1-25)\n")
		fmt.Fprintf(os.Stderr, "  -b, --brute           Brute-force and print all 25 possible shifts\n")
		fmt.Fprintf(os.Stderr, "  -a, --auto            Auto-solve mode: check all shifts for pattern (default pattern: picoCTF)\n")
		fmt.Fprintf(os.Stderr, "  -p, --pattern <str>   Custom pattern to look for in auto-solve mode (default: picoCTF)\n")
		fmt.Fprintf(os.Stderr, "  -c, --clean           Only output the raw plaintext, omitting metadata and color codes\n")
		fmt.Fprintf(os.Stderr, "  -v, --version         Show version information\n")
		fmt.Fprintf(os.Stderr, "  -h, --help            Show this help message\n")
	}
}

func decryptCaesar(text string, shift int) string {
	shift = (shift%26 + 26) % 26
	var sb strings.Builder
	sb.Grow(len(text))
	for i := 0; i < len(text); i++ {
		c := text[i]
		if c >= 'a' && c <= 'z' {
			sb.WriteByte('a' + (c-'a'+byte(shift))%26)
		} else if c >= 'A' && c <= 'Z' {
			sb.WriteByte('A' + (c-'A'+byte(shift))%26)
		} else {
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

func parseInputFormat(text, pattern string) (prefix, inner, suffix string, innerOnly bool) {
	firstCurly := strings.Index(text, "{")
	lastCurly := strings.LastIndex(text, "}")
	if firstCurly == -1 || lastCurly == -1 || firstCurly >= lastCurly {
		return "", "", "", false
	}

	prefix = text[:firstCurly+1]
	suffix = text[lastCurly:]
	inner = text[firstCurly+1 : lastCurly]

	// Heuristic: if the prefix (excluding '{') contains the pattern (case-insensitive),
	// we should only shift the inner part.
	prefixName := text[:firstCurly]
	if strings.Contains(strings.ToLower(prefixName), strings.ToLower(pattern)) {
		return prefix, inner, suffix, true
	}

	return "", "", "", false
}

func processShift(text string, shift int, pattern string) string {
	prefix, inner, suffix, innerOnly := parseInputFormat(text, pattern)
	if innerOnly {
		return prefix + decryptCaesar(inner, shift) + suffix
	}
	return decryptCaesar(text, shift)
}

func isTTY() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func containsPattern(text, pattern string) bool {
	if pattern == "" {
		return false
	}
	return strings.Contains(strings.ToLower(text), strings.ToLower(pattern))
}

func highlight(text, pattern string, color bool) string {
	if !color || pattern == "" {
		return text
	}
	lowerText := strings.ToLower(text)
	lowerPattern := strings.ToLower(pattern)

	var sb strings.Builder
	start := 0
	for {
		idx := strings.Index(lowerText[start:], lowerPattern)
		if idx == -1 {
			sb.WriteString(text[start:])
			break
		}
		matchIdx := start + idx
		sb.WriteString(text[start:matchIdx])
		sb.WriteString("\033[1;32m") // Bold Green
		sb.WriteString(text[matchIdx : matchIdx+len(pattern)])
		sb.WriteString("\033[0m")
		start = matchIdx + len(pattern)
	}
	return sb.String()
}

func main() {
	flag.Parse()

	if versionMode {
		fmt.Println("sifr v1.0.0")
		fmt.Println("Created by euxaristia (https://github.com/euxaristia)")
		return
	}

	// Gather inputs
	var input string
	var err error

	args := flag.Args()
	if len(args) == 0 {
		// Read from stdin
		input, err = readStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
	} else {
		arg := args[0]
		if arg == "-" {
			input, err = readStdin()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
				os.Exit(1)
			}
		} else if fileExists(arg) {
			data, err := os.ReadFile(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", arg, err)
				os.Exit(1)
			}
			input = string(data)
		} else {
			input = arg
		}
	}

	// Trim trailing newlines for presentation but keep intermediate spacing
	trimmedInput := strings.TrimSuffix(input, "\n")
	trimmedInput = strings.TrimSuffix(trimmedInput, "\r")

	useColor := isTTY() && !cleanMode

	// Determine modes
	hasShift := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "s" || f.Name == "shift" {
			hasShift = true
		}
	})

	if hasShift {
		// Single shift mode
		decrypted := processShift(trimmedInput, shiftVal, patternVal)
		if cleanMode {
			fmt.Println(decrypted)
		} else {
			if useColor {
				fmt.Printf("\033[36mShift %2d:\033[0m %s\n", shiftVal, highlight(decrypted, patternVal, useColor))
			} else {
				fmt.Printf("Shift %2d: %s\n", shiftVal, decrypted)
			}
		}
		return
	}

	// Brute-force explicitly requested
	if bruteMode {
		runBrute(trimmedInput, useColor)
		return
	}

	// Auto-solve explicitly requested
	if autoMode {
		runAuto(trimmedInput, useColor, true)
		return
	}

	// Default behavior (neither -s, -b, nor -a was specified)
	// Try auto-solve first. If matches are found, display them.
	// Otherwise, fall back to brute-forcing all shifts.
	matched := runAuto(trimmedInput, useColor, false)
	if !matched {
		if !cleanMode {
			fmt.Fprintln(os.Stderr, "No shifts matched pattern. Brute-forcing all shifts:")
		}
		runBrute(trimmedInput, useColor)
	}
}

func runBrute(input string, useColor bool) {
	for s := 1; s <= 25; s++ {
		decrypted := processShift(input, s, patternVal)
		if cleanMode {
			fmt.Println(decrypted)
		} else {
			if containsPattern(decrypted, patternVal) {
				if useColor {
					fmt.Printf("\033[1;32mShift %2d:\033[0m %s\n", s, highlight(decrypted, patternVal, useColor))
				} else {
					fmt.Printf("Shift %2d: %s (MATCH)\n", s, decrypted)
				}
			} else {
				if useColor {
					fmt.Printf("\033[36mShift %2d:\033[0m %s\n", s, decrypted)
				} else {
					fmt.Printf("Shift %2d: %s\n", s, decrypted)
				}
			}
		}
	}
}

func runAuto(input string, useColor bool, explicit bool) bool {
	matched := false
	for s := 1; s <= 25; s++ {
		decrypted := processShift(input, s, patternVal)
		if containsPattern(decrypted, patternVal) {
			matched = true
			if cleanMode {
				fmt.Println(decrypted)
			} else {
				if useColor {
					fmt.Printf("\033[1;32mShift %2d (MATCH):\033[0m %s\n", s, highlight(decrypted, patternVal, useColor))
				} else {
					fmt.Printf("Shift %2d (MATCH): %s\n", s, decrypted)
				}
			}
		}
	}

	if !matched && explicit && !cleanMode {
		fmt.Fprintf(os.Stderr, "No shifts matched the pattern: %q\n", patternVal)
	}
	return matched
}

func readStdin() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
