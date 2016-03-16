// Package shell supports splitting and joining of shell command strings.
//
// The Split function divides a string into whitespace-separated fields,
// respecting single and double quotation marks as defined by the Shell Command
// Language section of IEEE Std 1003.1 2013.  The Quote function quotes
// characters that would otherwise be subject to shell evaluation, and the Join
// function concatenates quoted strings with spaces between them.
//
// The relationship between Split and Join is that given
//
//     fields, ok := Split(Join(ss))
//
// the following relationship will hold:
//
//     fields == ss && ok
//
package shell

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// These characters must be quoted to escape special meaning.  This list
// doesn't include the single quote.
const mustQuote = "|^;<>()$\\\"\t\n`"

// These characters should be quoted to escape special meaning, since in some
// contexts they are special (e.g., "x=y" in command position, "*" for globs).
const shouldQuote = `*?[#~=%`

// These are the separator characters in unquoted text.
const spaces = " \t\n"

const allQuote = mustQuote + shouldQuote + spaces

type state int

const (
	stNone state = iota
	stBreak
	stBreakQ
	stWord
	stWordQ
	stSingle
	stDouble
	stDoubleQ
)

type class int

const (
	clBreak class = iota
	clNewline
	clQuote
	clSingle
	clDouble
	clOther
)

type action int

const (
	drop action = iota
	push
	xpush
	emit
)

var update = map[state]map[class]struct {
	state
	action
}{
	stBreak: {
		clBreak:   {stBreak, drop},
		clNewline: {stBreak, drop},
		clQuote:   {stBreakQ, drop},
		clSingle:  {stSingle, drop},
		clDouble:  {stDouble, drop},
		clOther:   {stWord, push},
	},
	stBreakQ: {
		clBreak:   {stWord, push},
		clNewline: {stBreak, drop},
		clQuote:   {stWord, push},
		clSingle:  {stWord, push},
		clDouble:  {stWord, push},
		clOther:   {stWord, push},
	},
	stWord: {
		clBreak:   {stBreak, emit},
		clNewline: {stBreak, emit},
		clQuote:   {stWordQ, drop},
		clSingle:  {stSingle, drop},
		clDouble:  {stDouble, drop},
		clOther:   {stWord, push},
	},
	stWordQ: {
		clBreak:   {stWord, push},
		clNewline: {stWord, drop},
		clQuote:   {stWord, push},
		clSingle:  {stWord, push},
		clDouble:  {stWord, push},
		clOther:   {stWord, push},
	},
	stSingle: {
		clBreak:   {stSingle, push},
		clNewline: {stSingle, push},
		clQuote:   {stSingle, push},
		clSingle:  {stWord, drop},
		clDouble:  {stSingle, push},
		clOther:   {stSingle, push},
	},
	stDouble: {
		clBreak:   {stDouble, push},
		clNewline: {stDouble, push},
		clQuote:   {stDoubleQ, drop},
		clSingle:  {stDouble, push},
		clDouble:  {stWord, drop},
		clOther:   {stDouble, push},
	},
	stDoubleQ: {
		clBreak:   {stDouble, xpush},
		clNewline: {stDouble, drop},
		clQuote:   {stDouble, push},
		clSingle:  {stDouble, xpush},
		clDouble:  {stDouble, push},
		clOther:   {stDouble, xpush},
	},
}

var byteClass = map[byte]class{
	' ':  clBreak,
	'\t': clBreak,
	'\n': clNewline,
	'\\': clQuote,
	'\'': clSingle,
	'"':  clDouble,
}

func classOf(b byte) class {
	if c, ok := byteClass[b]; ok {
		return c
	}
	return clOther
}

// A Scanner partitions input from a reader into tokens divided on space, tab,
// and newline characters.  Single and double quotation marks are handled as
// described in http://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html#tag_18_02.
type Scanner struct {
	buf *bufio.Reader
	cur bytes.Buffer
	st  state
	err error
}

// NewScanner returns a Scanner that reads input from r.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		buf: bufio.NewReader(r),
		st:  stBreak,
	}
}

// Next advances the scanner and reports whether there are any further tokens
// to be consumed.
func (s *Scanner) Next() bool {
	if s.err != nil {
		return false
	}
	s.cur.Reset()
	for {
		c, err := s.buf.ReadByte()
		s.err = err
		if err == io.EOF {
			break
		} else if err != nil {
			return false
		}
		next := update[s.st][classOf(c)]
		s.st = next.state
		switch next.action {
		case push:
			s.cur.WriteByte(c)
		case xpush:
			s.cur.Write([]byte{'\\', c})
		case emit:
			return true // s.cur has a complete token
		case drop:
			break
		default:
			panic("unknown action")
		}
	}
	return s.st != stBreak
}

// Text returns the text of the current token, or "" if there is none.
func (s *Scanner) Text() string { return s.cur.String() }

// Err returns the error, if any, that resulted from the most recent action.
func (s *Scanner) Err() error { return s.err }

// Complete reports whether the current token is complete, meaning that it is
// unquoted or its quotes were balanced.
func (s *Scanner) Complete() bool { return s.st == stBreak || s.st == stWord }

// Rest returns an io.Reader for the remainder of the unconsumed input in s.
// After calling this method, Next will always return false.  The remainder
// does not include the text of the current token at the time Rest is called.
func (s *Scanner) Rest() io.Reader {
	s.st = stNone
	s.cur.Reset()
	s.err = io.EOF
	return s.buf
}

// Each calls f for each token in the scanner until the input is exhausted, f
// returns false, or an error occurs.
func (s *Scanner) Each(f func(tok string) bool) error {
	for s.Next() {
		if !f(s.Text()) {
			return nil
		}
	}
	if err := s.Err(); err != io.EOF {
		return err
	}
	return nil
}

// Split partitions s into tokens divided on space, tab, and newline characters
// using a *Scanner.  Leading and trailing whitespace are ignored.
//
// The Boolean flag reports whether the final token is "valid", meaning there
// were no unclosed quotations in the string.
func Split(s string) ([]string, bool) {
	var ss []string
	sc := NewScanner(strings.NewReader(s))
	for sc.Next() {
		ss = append(ss, sc.Text())
	}
	return ss, sc.Complete()
}

func quotable(s string) (hasQ, hasOther bool) {
	for i := 0; i < len(s); i++ {
		hasQ = hasQ || s[i] == '\''
		hasOther = hasOther || strings.IndexByte(allQuote, s[i]) >= 0
	}
	return
}

// Quote returns a copy of s in which shell metacharacters are quoted to
// protect them from evaluation.
func Quote(s string) string {
	if s == "" {
		return "''"
	}
	hasQ, hasOther := quotable(s)
	if !hasQ && !hasOther {
		return s // fast path: nothing needs quotation
	}

	var buf bytes.Buffer
	inq := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '\'' {
			if inq {
				buf.WriteByte('\'')
				inq = false
			}
			buf.WriteByte('\\')
		} else if !inq && hasOther {
			buf.WriteByte('\'')
			inq = true
		}
		buf.WriteByte(ch)
	}
	if inq {
		buf.WriteByte('\'')
	}
	return buf.String()
}

// Join quotes each element of ss with Quote and concatenates the resulting
// strings separated by spaces.
func Join(ss []string) string {
	quoted := make([]string, len(ss))
	for i, s := range ss {
		quoted[i] = Quote(s)
	}
	return strings.Join(quoted, " ")
}