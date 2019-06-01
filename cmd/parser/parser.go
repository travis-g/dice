// +build ignore

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Token constants are lexer/parser tokens.
type Token int

// Token constants
const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	WS

	// Literals
	IDENT
	DIE
	NUM

	// Misc
	PLUS
	MINUS
	ASTERISK
	SLASH
	EQ
	LT
	GT
	LPAREN
	RPAREN
)

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

var eof = rune(0)

// Scanner represents a lexical scanner.
type Scanner struct {
	r *bufio.Reader
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() { _ = s.r.UnreadRune() }

// Scan returns the next token and literal value.
func (s *Scanner) Scan() (tok Token, lit string) {
	// Read the next rune.
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter then consume as an ident or reserved word.
	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) {
		s.unread()
		return s.scanLetters()
	} else if isDigit(ch) {
		s.unread()
		return s.scanDigit()
	}

	// Otherwise read the individual character.
	switch ch {
	case eof:
		return EOF, ""
	case '-':
		return MINUS, string(ch)
	case '+':
		return PLUS, string(ch)
	case '<':
		return LT, string(ch)
	case '>':
		return GT, string(ch)
	case '=':
		return EQ, string(ch)
	}

	return ILLEGAL, string(ch)
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return WS, buf.String()
}

// scanIdent consumes the current rune and all contiguous ident runes.
func (s *Scanner) scanIdent() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	// If the string matches a keyword then return that keyword.
	upper := strings.ToUpper(buf.String())
	switch {
	case strings.HasPrefix(upper, "D"):
		return DIE, buf.String()
	}

	// Otherwise return as a regular identifier.
	return IDENT, buf.String()
}

func (s *Scanner) scanLetters() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	// If the string matches a keyword then return that keyword.
	upper := strings.ToUpper(buf.String())
	switch {
	case strings.HasPrefix(upper, "D"):
		return DIE, buf.String()
	}

	// Otherwise return as a regular identifier.
	return IDENT, buf.String()
}

// scanDigit consumes the current rune and all contiguous digit runes.
func (s *Scanner) scanDigit() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isDigit(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	// Otherwise return as number.
	return NUM, buf.String()
}

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

type Props struct {
	Count uint
	Size  uint
}

func (p *Parser) Parse() (*Props, error) {
	props := &Props{}
	tok, lit := p.scanIgnoreWhitespace()
	if tok == DIE {
		// no prefixed count found, default to 1
		props.Count = 1
	} else if tok == NUM {
		icount, err := strconv.Atoi(lit)
		if err != nil {
			return nil, err
		}
		props.Count = uint(icount)
		// scan next, which should be a notation string
		p.scanIgnoreWhitespace()
	} else {
		return nil, fmt.Errorf("found unexpected %q", lit)
	}
	tok, lit = p.scanIgnoreWhitespace()
	if tok != NUM {
		return nil, fmt.Errorf("found %q, expected NUM", lit)
	}
	isize, err := strconv.Atoi(lit)
	if err != nil {
		return nil, err
	}
	props.Size = uint(isize)
	return props, nil
}

func main() {
	parser := NewParser(strings.NewReader("3d20+1"))
	props, err := parser.Parse()
	fmt.Println(props, err)
}
