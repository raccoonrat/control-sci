package sanitize

import (
	"strings"
	"unicode"
)

type Normalizer struct {
	obfuscators  map[rune]struct{}
	replacements map[rune]rune
}

func NewNormalizer() *Normalizer {
	return &Normalizer{
		obfuscators: map[rune]struct{}{
			'.': {}, '-': {}, '_': {}, '~': {}, '*': {}, '^': {},
			' ': {}, '\t': {}, '\n': {}, '\r': {},
			'|': {}, '/': {}, '\\': {},
			'｜': {}, '·': {}, '。': {}, '，': {}, '、': {},
			'！': {}, '？': {}, '；': {}, '：': {},
			'「': {}, '」': {}, '『': {}, '』': {},
			'（': {}, '）': {}, '《': {}, '》': {},
		},
		replacements: map[rune]rune{
			'臺': '台',
			'檯': '台',
			'後': '后',
			'裏': '里',
			'裡': '里',
			'祕': '秘',
		},
	}
}

func (n *Normalizer) Normalize(input string) string {
	if input == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(input))

	for _, r := range input {
		if _, ok := n.obfuscators[r]; ok {
			continue
		}

		r = toHalfWidth(r)
		if unicode.IsSpace(r) {
			continue
		}
		if unicode.IsLetter(r) {
			r = unicode.ToLower(r)
		}
		if replacement, ok := n.replacements[r]; ok {
			r = replacement
		}

		builder.WriteRune(r)
	}

	return builder.String()
}

func toHalfWidth(r rune) rune {
	if r >= 0xff01 && r <= 0xff5e {
		return r - 0xfee0
	}
	if r == 0x3000 {
		return ' '
	}

	return r
}
