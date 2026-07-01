package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf16"
)

var (
	shiftVal      int
	bruteMode     bool
	autoMode      bool
	patternVal    string
	cleanMode     bool
	versionMode   bool
	newCaesarMode bool
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

	flag.BoolVar(&newCaesarMode, "n", false, "Use New Caesar cipher (custom Base16 / modulo-16 solver)")
	flag.BoolVar(&newCaesarMode, "new", false, "Use New Caesar cipher (custom Base16 / modulo-16 solver)")
	flag.BoolVar(&newCaesarMode, "new-caesar", false, "Use New Caesar cipher (custom Base16 / modulo-16 solver)")

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
		fmt.Fprintf(os.Stderr, "  -n, --new-caesar      Use New Caesar cipher (custom Base16 / modulo-16 solver)\n")
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
	if newCaesarMode {
		if innerOnly {
			dec, _ := decryptNewCaesar(inner, shift)
			return prefix + dec + suffix
		}
		dec, _ := decryptNewCaesar(text, shift)
		return dec
	}

	if innerOnly {
		return prefix + decryptCaesar(inner, shift) + suffix
	}
	return decryptCaesar(text, shift)
}

func decryptNewCaesar(text string, shift int) (string, error) {
	shift = (shift%16 + 16) % 16
	if len(text)%2 != 0 {
		return "", fmt.Errorf("odd length for New Caesar decoding")
	}
	var sb strings.Builder
	sb.Grow(len(text) / 2)
	for i := 0; i < len(text); i += 2 {
		c1 := text[i]
		c2 := text[i+1]
		if c1 < 'a' || c1 > 'p' || c2 < 'a' || c2 > 'p' {
			return "", fmt.Errorf("invalid character for New Caesar: %c or %c", c1, c2)
		}
		t1 := int(c1 - 'a')
		t2 := int(c2 - 'a')

		u1 := (t1 - shift) % 16
		if u1 < 0 {
			u1 += 16
		}
		u2 := (t2 - shift) % 16
		if u2 < 0 {
			u2 += 16
		}

		val := (u1 << 4) | u2
		sb.WriteByte(byte(val))
	}
	return sb.String(), nil
}

func isValidNewCaesar(text string, pattern string) error {
	_, inner, _, innerOnly := parseInputFormat(text, pattern)
	target := text
	if innerOnly {
		target = inner
	}
	if len(target)%2 != 0 {
		return fmt.Errorf("ciphertext length (%d) is odd, New Caesar requires an even number of characters", len(target))
	}
	for i := 0; i < len(target); i++ {
		c := target[i]
		if c < 'a' || c > 'p' {
			return fmt.Errorf("ciphertext contains invalid character %q (New Caesar only allows 'a' through 'p')", c)
		}
	}
	return nil
}

func isPrintable(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < 32 || c > 126) && c != 9 && c != 10 && c != 13 {
			return false
		}
	}
	return true
}

func isMatch(text, pattern string) bool {
	if !containsPattern(text, pattern) {
		return false
	}
	if newCaesarMode {
		prefix, inner, _, innerOnly := parseInputFormat(text, pattern)
		target := text
		if innerOnly {
			target = inner
		}
		_ = prefix
		return isPrintable(target)
	}
	return true
}

func getShiftLabel(s int) string {
	if newCaesarMode {
		return fmt.Sprintf("Shift %2d (key %c)", s, 'a'+s)
	}
	return fmt.Sprintf("Shift %2d", s)
}

func parsePowerShellEncodedCommand(posArg string) (string, bool) {
	// Look for -encodedCommand in os.Args
	var encodedCmd string
	for i := 0; i < len(os.Args)-1; i++ {
		if os.Args[i] == "-encodedCommand" {
			encodedCmd = os.Args[i+1]
			break
		}
	}
	if encodedCmd == "" {
		return "", false
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(encodedCmd)
	if err != nil {
		return "", false
	}

	// PowerShell encodes using UTF-16LE
	if len(decodedBytes)%2 != 0 {
		return "", false
	}
	u16s := make([]uint16, len(decodedBytes)/2)
	for i := 0; i < len(u16s); i++ {
		u16s[i] = uint16(decodedBytes[2*i]) | (uint16(decodedBytes[2*i+1]) << 8)
	}
	decodedStr := string(utf16.Decode(u16s))

	// Reconstruct the full string
	if posArg != "" {
		return posArg + "{" + decodedStr + "}", true
	}
	return decodedStr, true
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
		if pwshInput, ok := parsePowerShellEncodedCommand(arg); ok {
			input = pwshInput
		} else if arg == "-" {
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

	if newCaesarMode {
		if err := isValidNewCaesar(trimmedInput, patternVal); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	useColor := isTTY() && !cleanMode

	// Determine modes
	hasShift := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "s" || f.Name == "shift" {
			hasShift = true
		}
	})

	if hasShift {
		if newCaesarMode {
			if shiftVal < 0 || shiftVal > 15 {
				fmt.Fprintf(os.Stderr, "Error: New Caesar shift must be between 0 and 15\n")
				os.Exit(1)
			}
		}
		// Single shift mode
		decrypted := processShift(trimmedInput, shiftVal, patternVal)
		if cleanMode {
			fmt.Println(decrypted)
		} else {
			if useColor {
				fmt.Printf("\033[36m%s:\033[0m %s\n", getShiftLabel(shiftVal), highlight(decrypted, patternVal, useColor))
			} else {
				fmt.Printf("%s: %s\n", getShiftLabel(shiftVal), decrypted)
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
	// Try auto-solve first. If it's a New Caesar candidate, try New Caesar auto-solve first.
	// Otherwise, fallback to standard Caesar auto-solve.
	// If no matches are found, brute-force standard Caesar.
	var matched bool
	if !newCaesarMode && isValidNewCaesar(trimmedInput, patternVal) == nil {
		newCaesarMode = true
		matched = runAuto(trimmedInput, useColor, false)
		if !matched {
			newCaesarMode = false
		}
	}

	if !matched {
		matched = runAuto(trimmedInput, useColor, false)
	}

	if !matched {
		if !cleanMode {
			fmt.Fprintln(os.Stderr, "No shifts matched pattern. Brute-forcing all shifts:")
		}
		runBrute(trimmedInput, useColor)
	}
}

func runBrute(input string, useColor bool) {
	startShift := 1
	maxShift := 25
	if newCaesarMode {
		startShift = 0
		maxShift = 15
	}
	for s := startShift; s <= maxShift; s++ {
		decrypted := processShift(input, s, patternVal)
		if cleanMode {
			fmt.Println(decrypted)
		} else {
			if isMatch(decrypted, patternVal) {
				if useColor {
					fmt.Printf("\033[1;32m%s:\033[0m %s\n", getShiftLabel(s), highlight(decrypted, patternVal, useColor))
				} else {
					fmt.Printf("%s: %s (MATCH)\n", getShiftLabel(s), decrypted)
				}
			} else {
				if useColor {
					fmt.Printf("\033[36m%s:\033[0m %s\n", getShiftLabel(s), decrypted)
				} else {
					fmt.Printf("%s: %s\n", getShiftLabel(s), decrypted)
				}
			}
		}
	}
}

func runAuto(input string, useColor bool, explicit bool) bool {
	matched := false
	startShift := 1
	maxShift := 25
	if newCaesarMode {
		startShift = 0
		maxShift = 15
	}
	for s := startShift; s <= maxShift; s++ {
		decrypted := processShift(input, s, patternVal)
		if isMatch(decrypted, patternVal) {
			matched = true
			if cleanMode {
				fmt.Println(decrypted)
			} else {
				if useColor {
					fmt.Printf("\033[1;32m%s (MATCH):\033[0m %s\n", getShiftLabel(s), highlight(decrypted, patternVal, useColor))
				} else {
					fmt.Printf("%s (MATCH): %s\n", getShiftLabel(s), decrypted)
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
