package parser

import (
	"bufio"
	"errors"
	"strings"

	"github.com/tal-tech/go-zero/tools/goctl/api/spec"
)

var emptyType spec.Type

type ApiStruct struct {
	Info             string
	StructBody       string
	Service          string
	Imports          string
	serviceBeginLine int
}

func GetType(api *spec.ApiSpec, t string) spec.Type {
	for _, tp := range api.Types {
		if tp.Name == t {
			return tp
		}
	}

	return emptyType
}

func isLetterDigit(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || ('0' <= r && r <= '9')
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isNewline(r rune) bool {
	return r == '\n' || r == '\r'
}

func read(r *bufio.Reader) (rune, error) {
	ch, _, err := r.ReadRune()
	return ch, err
}

func readLine(r *bufio.Reader) (string, error) {
	line, _, err := r.ReadLine()
	if err != nil {
		return "", err
	} else {
		return string(line), nil
	}
}

func skipSpaces(r *bufio.Reader) error {
	for {
		next, err := read(r)
		if err != nil {
			return err
		}
		if !isSpace(next) {
			return unread(r)
		}
	}
}

func unread(r *bufio.Reader) error {
	return r.UnreadRune()
}

// removeComments remove comments in template
func removeComments(api string) string {

	var out []string
	var lo []byte
	b := true

	scanner := bufio.NewScanner(strings.NewReader(api))
	for scanner.Scan() {
		l := scanner.Text()
		bs := []byte(l)
		i := 0
		for ; i < len(bs)-1; i++ {
			if b {
				if bs[i] == '/' && bs[i+1] == '*' {
					b = false
					i++
				} else if bs[i] == '/' && bs[i+1] == '/' {
					break
				} else {
					lo = append(lo, bs[i])
				}
			} else if bs[i] == '*' && bs[i+1] == '/' {
				b = true
				i++
			}
		}
		if b && i == len(bs)-1 {
			lo = append(lo, bs[len(bs)-1])
		}
		if len(lo) > 0 && b {
			out = append(out, string(lo))
			lo = []byte{}
		}
	}
	return strings.Join(out, "\n")
}

func ParseApi(api string) (*ApiStruct, error) {
	// delete comment in template
	api = removeComments(api)
	var result ApiStruct
	scanner := bufio.NewScanner(strings.NewReader(api))
	var parseInfo = false
	var parseImport = false
	var parseType = false
	var parseService = false
	var segment string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "info(" {
			parseInfo = true
		}
		if line == ")" && parseInfo {
			parseInfo = false
			result.Info = segment + ")"
			segment = ""
			continue
		}

		if isImportBeginLine(line) {
			parseImport = true
		}
		if parseImport && (isTypeBeginLine(line) || isServiceBeginLine(line)) {
			parseImport = false
			result.Imports = segment
			segment = line + "\n"
			continue
		}

		if isTypeBeginLine(line) {
			parseType = true
		}
		if isServiceBeginLine(line) {
			parseService = true
			if parseType {
				parseType = false
				result.StructBody = segment
				segment = line + "\n"
				continue
			}
		}
		segment += scanner.Text() + "\n"
	}

	if !parseService {
		return nil, errors.New("no service defined")
	}
	result.Service = segment
	result.serviceBeginLine = lineBeginOfService(api)
	return &result, nil
}

func isImportBeginLine(line string) bool {
	return strings.HasPrefix(line, "import") && strings.HasSuffix(line, ".api")
}

func isTypeBeginLine(line string) bool {
	return strings.HasPrefix(line, "type")
}

func isServiceBeginLine(line string) bool {
	return strings.HasPrefix(line, "@server") || (strings.HasPrefix(line, "service") && strings.HasSuffix(line, "{"))
}

func lineBeginOfService(api string) int {
	scanner := bufio.NewScanner(strings.NewReader(api))
	var number = 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if isServiceBeginLine(line) {
			break
		}
		number++
	}
	return number
}
