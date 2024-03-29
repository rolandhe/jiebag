package pinyin

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

var unicodeToneMap = map[uint8][]rune{}

func init() {
	unicodeToneMap[uint8('a')] = []rune("āáăàa")
	unicodeToneMap[uint8('e')] = []rune("ēéĕèe")
	unicodeToneMap[uint8('i')] = []rune("īíĭìi")
	unicodeToneMap[uint8('o')] = []rune("ōóŏòo")
	unicodeToneMap[uint8('u')] = []rune("ūúŭùu")
	unicodeToneMap[uint8('v')] = []rune("ǖǘǚǜü")
}

type PinyinFmt int

const (
	WithoutTone     PinyinFmt = 0
	ToneTail        PinyinFmt = 1
	UnicodeWithTone PinyinFmt = 2
)

type singlePinyin struct {
	word string
	tone uint8
}

func (sp *singlePinyin) toString(formatter PinyinFmt) string {
	if formatter == WithoutTone {
		return sp.word
	}
	if formatter == ToneTail {
		return fmt.Sprintf("%s%d", sp.word, sp.tone)
	}
	if formatter == UnicodeWithTone {
		return unicodeTone(sp.word, sp.tone)
	}
	return ""
}

type wordPinyin struct {
	group [][]*singlePinyin
}

func (wp *wordPinyin) add(mulPinyin []string) {
	l := len(mulPinyin)
	if l == 0 {
		return
	}
	wp.group = make([][]*singlePinyin, 0, l)
	for _, py := range mulPinyin {
		py = strings.TrimSpace(py)
		l := len(py)
		tone, _ := strconv.Atoi(py[l-1:])
		one := []*singlePinyin{{
			word: py[:l-1],
			tone: uint8(tone),
		}}
		wp.group = append(wp.group, one)
	}
}

func (wp *wordPinyin) addWordGroup(pyGroup []string) {
	l := len(pyGroup)
	if l == 0 {
		return
	}
	mul := make([]*singlePinyin, 0, l)
	for _, py := range pyGroup {
		py = strings.TrimSpace(py)
		l := len(py)
		tone, _ := strconv.Atoi(py[l-1:])
		mul = append(mul, &singlePinyin{
			word: py[:l-1],
			tone: uint8(tone),
		})
	}

	wp.group = append(wp.group, mul)
}
func (wp *wordPinyin) getFirst() []*singlePinyin {
	return wp.group[0]
}

func (wp *wordPinyin) getFirstString(formatter PinyinFmt) []string {
	firstPinyin := wp.getFirst()
	var ret []string

	for _, word := range firstPinyin {
		ret = append(ret, word.toString(formatter))
	}

	return ret
}

type DictNode struct {
	children map[rune]*DictNode
	isWord   bool
	pinyin   *wordPinyin
}

func (root *DictNode) addSingle(r rune, pys []string) *DictNode {
	wp := &wordPinyin{}
	wp.add(pys)

	child := &DictNode{
		isWord: true,
		pinyin: wp,
	}
	return root.addChild(r, child)
}

func (root *DictNode) addChild(r rune, child *DictNode) *DictNode {
	if root.children == nil {
		root.children = map[rune]*DictNode{}
	}
	root.children[r] = child
	return child
}

func (root *DictNode) addWord(word string, pys []string) {
	runes := []rune(word)
	l := len(runes)
	if l != len(pys) {
		return
	}

	af := func(p *DictNode, r rune, group []string, isWord bool) *DictNode {
		var wp *wordPinyin
		if isWord {
			wp = &wordPinyin{}
			wp.addWordGroup(group)
		}
		child := &DictNode{
			isWord: isWord,
			pinyin: wp,
		}
		return p.addChild(r, child)
	}

	var isWord bool
	var wordsPys []string
	p := root
	for i, r := range runes {
		if l-1 == i {
			isWord = true
			wordsPys = pys
		}
		if !p.hasChildren() {
			p = af(p, r, wordsPys, isWord)
			continue
		}
		child, ok := p.children[r]
		if !ok {
			p = af(p, r, wordsPys, isWord)
			continue
		}
		if isWord && !child.isWord {
			child.isWord = isWord
			wp := &wordPinyin{}
			wp.addWordGroup(pys)
			child.pinyin = wp
		}
		p = child
	}
}

func (root *DictNode) ConvertString(next string, formatter PinyinFmt) []string {
	return root.Convert([]rune(next), formatter)
}
func (root *DictNode) Convert(next []rune, formatter PinyinFmt) []string {
	dataLen := len(next)
	allLen := dataLen

	var ret []string

	for dataLen > 0 {
		var wp *wordPinyin
		var startIndex int
		wp, next, startIndex = root.matchFirst(next)
		if wp != nil {
			for startIndex > 0 {
				startIndex--
				ret = append(ret, "")
			}
			ret = append(ret, wp.getFirstString(formatter)...)
		}
		dataLen = len(next)
	}

	allLen -= len(ret)
	for allLen > 0 {
		allLen--
		ret = append(ret, "")
	}

	return ret
}

func (root *DictNode) hasChildren() bool {
	return len(root.children) > 0
}

func (root *DictNode) matchFirst(next []rune) (*wordPinyin, []rune, int) {
	p := root
	var candidate *wordPinyin
	var nextIndex int
	var startIndex int
	for i, r := range next {
		child := p.matchChild(r)
		if child == nil {
			if candidate == nil {
				nextIndex = i + 1
				startIndex = i + 1
				continue
			}
			break
		}

		if child.isWord {
			candidate = child.pinyin
			nextIndex = i + 1
		}
		p = child
	}
	if nextIndex == len(next) {
		next = nil
	} else {
		next = next[nextIndex:]
	}
	return candidate, next, startIndex
}

func (root *DictNode) matchChild(r rune) *DictNode {
	if !root.hasChildren() {
		return nil
	}
	return root.children[r]
}

func LoadDict(rootPath string) (*DictNode, error) {
	node := &DictNode{}

	if err := loadFile(path.Join(rootPath, "pinyin.txt"), func(line string) {
		items := strings.Split(line, "=")
		pinyins := strings.Split(strings.TrimSpace(items[1]), ",")
		w := strings.TrimSpace(items[0])
		node.addSingle([]rune(w)[0], pinyins)
	}); err != nil {
		return nil, err
	}

	if err := loadFile(path.Join(rootPath, "polyphone.txt"), func(line string) {
		items := strings.Split(line, "=")
		pinyins := strings.Fields(items[1])
		node.addWord(strings.TrimSpace(items[0]), pinyins)
	}); err != nil {
		return nil, err
	}

	return node, nil
}

func loadFile(filePath string, acceptor func(line string)) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := scan.Text()
		acceptor(line)
	}

	return scan.Err()
}

// unicodeTone Algorithm from nlp-lang java project, see pinyin.PinyinFormatter class
//
// Algorithm for determining location of tone mark,
// A simple algorithm for determining the vowel on which the tone mark
// appears is as follows:
//
// 1. First, look for an "a" or an "e". If either vowel appears, it takes
// the tone mark. There are no possible pinyin syllables that contain both
// an "a" and an "e".
// 2. If there is no "a" or "e", look for an "ou". If "ou" appears, then
// the "o" takes the tone mark.
// 3.If none of the above cases hold, then the last vowel in the syllable
// takes the tone mark.
func unicodeTone(py string, tone uint8) string {
	data := []rune(py)
	l := len(data)
	var bak []int
	ouIndex := -1
	finish := false
	for i := l - 1; i >= 0; i-- {
		c := data[i]
		if c == 'a' || c == 'e' {
			data[i] = unicodeToneMap[uint8(c)][tone-1]
			finish = true
			break
		}
		if c == 'o' || c == 'i' || c == 'v' {
			bak = append(bak, i)
			continue
		}
		if c == 'u' {
			if i > 0 && data[i-1] == 'o' {
				i--
				ouIndex = i
				continue
			}
			bak = append(bak, i)
			continue
		}
	}
	if !finish && ouIndex != -1 {
		data[ouIndex] = unicodeToneMap[uint8(data[ouIndex])][tone-1]
		finish = true
	}
	if !finish {
		if len(bak) == 0 {
			panic(py + " is not valid pinyin")
		}
		c := data[bak[0]]
		data[bak[0]] = unicodeToneMap[uint8(c)][tone-1]
	}
	return string(data)
}
