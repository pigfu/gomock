package regen

import (
	"fmt"
	"math/rand"
	"regexp/syntax"
	"strings"
)

type generator func(*syntax.Regexp, *strings.Builder) error

var (
	generatorMap map[syntax.Op]generator
	noneWord     = []rune{32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 58, 59, 60, 61, 62, 63, 64, 91, 92, 93, 94, 96, 123,
		124, 125, 126}
	word = []rune{49, 50, 51, 52, 53, 54, 55, 56, 57, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86,
		87, 88, 89, 90, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120,
		121, 122, 95}
)

func init() {
	generatorMap = map[syntax.Op]generator{
		syntax.OpNoMatch:        regEmpty,
		syntax.OpEmptyMatch:     regEmpty,
		syntax.OpBeginLine:      regEmpty,          //^
		syntax.OpEndLine:        regEmpty,          //$
		syntax.OpBeginText:      regEmpty,          //\A
		syntax.OpEndText:        regEmpty,          //\z
		syntax.OpWordBoundary:   regWordBoundary,   //\b
		syntax.OpNoWordBoundary: regNoWordBoundary, //\B
		syntax.OpLiteral:        regLiteral,        //eg: abc
		syntax.OpStar:           regStar,           //eg: a*
		syntax.OpQuest:          regQuest,          //eg: a?
		syntax.OpPlus:           regPlus,           //eg: a+
		syntax.OpRepeat:         regRepeat,         //eg: a{1,9}
		syntax.OpCharClass:      regCharClass,      //eg: [a-z0-9A-Z_],
		syntax.OpConcat:         regConcat,         //eg: [a-z]+[0-9]?
		syntax.OpAlternate:      regAlternate,      //eg:(ac)|(xy+)
		syntax.OpAnyCharNotNL:   regAnyCharNotNL,
		syntax.OpAnyChar:        regAnyChar,
		syntax.OpCapture:        regCapture, //support naming group
	}
}

const (
	defaultMaxTimes       = 9
	asciiMinPrintableChar = 32
	asciiMaxPrintableChar = 126
)

func Generate(pattern string) (string, error) {
	reg, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return "", err
	}
	output := &strings.Builder{}
	err = generate(reg, output)
	return output.String(), err
}

func generate(regexp *syntax.Regexp, b *strings.Builder) error {
	if f, ok := generatorMap[regexp.Op]; ok {
		return f(regexp, b)
	}
	return fmt.Errorf("not support the type %s", regexp.Op)
}

func regEmpty(_ *syntax.Regexp, _ *strings.Builder) error {
	return nil
}

func regLiteral(regexp *syntax.Regexp, b *strings.Builder) error {
	b.WriteString(string(regexp.Rune))
	return nil
}

func regWordBoundary(_ *syntax.Regexp, b *strings.Builder) error {
	b.WriteRune(noneWord[rand.Intn(len(noneWord))])
	return nil
}

func regNoWordBoundary(_ *syntax.Regexp, b *strings.Builder) error {
	b.WriteRune(word[rand.Intn(len(word))])
	return nil
}

// *  0 or more of previous expression
func regStar(regexp *syntax.Regexp, b *strings.Builder) (err error) {
	return genRepeat(regexp.Sub[0], b, rand.Intn(defaultMaxTimes))
}

// ? 0 or 1 of previous expression
func regQuest(regexp *syntax.Regexp, b *strings.Builder) (err error) {
	return genRepeat(regexp.Sub[0], b, rand.Intn(2))
}

// + 1 or more of previous expression
func regPlus(regexp *syntax.Regexp, b *strings.Builder) (err error) {
	return genRepeat(regexp.Sub[0], b, 1+rand.Intn(defaultMaxTimes))
}

// {m,n} number of previous expression
func regRepeat(regexp *syntax.Regexp, b *strings.Builder) (err error) {
	max := regexp.Max
	if max == -1 {
		max = regexp.Min + defaultMaxTimes
	}
	return genRepeat(regexp.Sub[0], b, regexp.Min+rand.Intn(max-regexp.Min+1))
}

func genRepeat(regexp *syntax.Regexp, b *strings.Builder, number int) (err error) {
	for i := 0; i < number; i++ {
		if err = generate(regexp, b); err != nil {
			return err
		}
	}
	return
}

func regCharClass(regexp *syntax.Regexp, b *strings.Builder) error {
	if len(regexp.Rune)&1 != 0 || len(regexp.Rune) == 0 {
		return nil
	}
	//[0,1,2,3,4,5]
	randIndex := rand.Intn(len(regexp.Rune) / 2)
	l, r := regexp.Rune[randIndex*2], regexp.Rune[randIndex*2+1]
	b.WriteRune(rune(rand.Int63n(int64(r-l)+1) + int64(l))) //[m,n]
	return nil
}

func regConcat(regexp *syntax.Regexp, b *strings.Builder) error {
	var err error
	for _, sub := range regexp.Sub {
		if err = generate(sub, b); err != nil {
			return err
		}
	}
	return nil
}

// Alternation
func regAlternate(regexp *syntax.Regexp, b *strings.Builder) error {
	if len(regexp.Sub) == 0 {
		return nil
	}
	return generate(regexp.Sub[rand.Intn(len(regexp.Sub))], b)
}

// Any character (except \n newline)
func regAnyCharNotNL(_ *syntax.Regexp, b *strings.Builder) error {
	genPrintableChar(b)
	return nil
}

// Any character
func regAnyChar(_ *syntax.Regexp, b *strings.Builder) error {
	genPrintableChar(b)
	return nil
}

func genPrintableChar(b *strings.Builder) {
	b.WriteRune(rune(rand.Intn(asciiMaxPrintableChar-asciiMinPrintableChar+1) + asciiMinPrintableChar))
}

func regCapture(regexp *syntax.Regexp, b *strings.Builder) error {
	return generate(regexp.Sub[0], b)
}
