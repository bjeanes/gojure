package edn

import "fmt"
import "strings"
import "unicode"

// A lot of this is based on: http://cuddle.googlecode.com/hg/talk/lex.html

type tokenType int

const (
	tEOF tokenType = 1 << iota
	tOpenParen
	tCloseParen
	tOpenBrace
	tCloseBrace
	tOpenBracket
	tCloseBracket
	tQuoteNextForm
	tOpenSet
	tString
	tSymbol
	tKeyword
	tNumber
	tMetadata
	tIgnoreNextForm
	tWhitespace // needed?
	tComment    // needed?

	tError
)

type token struct {
	kind  tokenType
	value string
}

func (t *token) String() string {
	switch t.kind {
	case tEOF:
		return "EOF"
	case tError:
		return t.value
	default:
		return fmt.Sprintf("%q", t.value)
	}

	panic("unreachable")
}

func isWhitespace(ch rune) bool {
	return unicode.IsSpace(ch) || ch == ','
}

func lexEDN(l *lexer) {
	for {
		ch, _, err := l.read()

		if err != nil {
			return
		}

		switch ch {
		case '(':
			l.emit(tOpenParen)
		case '[':
			l.emit(tOpenBracket)
		case '{':
			l.emit(tOpenBrace)
		case ')':
			l.emit(tCloseParen)
		case ']':
			l.emit(tCloseBracket)
		case '}':
			l.emit(tCloseBrace)
		case ':': // keyword
			// TODO: Read keyword value
			l.emit(tKeyword)
		case '\'': l.emit(tQuoteNextForm)
		case '^': l.emit(tMetadata)
		case '#': // ...
			ch, _, _ := l.read()

			switch ch {
				case '_':
					l.emit(tIgnoreNextForm)
				case '{':
					l.emit(tOpenSet)
				case '^':
					l.emit(tMetadata)
			}
		case '"':
			for {
				// FIXME: strings with \"
				ch, _, _ := l.read()
				if ch == '"' {
					break
				}
			}

			l.emit(tString)
		case ';': // comment
		case '\t', ',', ' ': // whitespace
			for {
				ch, size, _ := l.read()
				if !isWhitespace(ch) {
					l.unread(size)
					break
				}
			}
			// l.emit(tWhitespace)
			l.start = l.position // Temporarily ignore whitespace
		default:
			switch {
				case unicode.IsNumber(ch):
					// TODO: non integer numbers...
					for {
						ch, size, _ := l.read()
						if !unicode.IsNumber(ch) {
							l.unread(size)
							break
						}
					}
					l.emit(tNumber)
			default:
				// TODO: proper error handling
				panic("Unexpected character " + fmt.Sprintf("'%c'", ch))
			}

		}
	}
}

type lexer struct {
	input    string
	reader   *strings.Reader
	start    int
	position int
	tokens   chan token
}

func (l *lexer) peek() (ch rune, size int, err error) {
	ch, size, err = l.read()
	l.unread(size)
	return
}

func (l *lexer) unread(size int) {
	l.reader.UnreadRune()
	l.position -= size
}

func (l *lexer) read() (ch rune, size int, err error) {
	ch, size, err = l.reader.ReadRune()
	l.position += size
	return
}

func (l *lexer) run() {
	lexEDN(l)
	l.emit(tEOF)
	close(l.tokens)
}

func (l *lexer) emit(tt tokenType) {
	value := l.input[l.start:l.position]
	l.start = l.position
	l.tokens <- token{kind: tt, value: value}
}

func (l *lexer) Next() (t token, more bool) {
	t, more = <-l.tokens
	return
}

func Lex(input string) *lexer {
	l := &lexer{
		input:  input,
		reader: strings.NewReader(input),
		tokens: make(chan token, 100),
	}

	go l.run()
	return l
}
