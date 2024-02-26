package pinyin

import (
	"bufio"
	"log"
	"math"
	"os"
	"path"
	"strings"
)

type segment struct {
	start int
	end   int
}

type segTokenInternal struct {
	*segment
	count int
}

type LatinGuessNode struct {
	children map[uint8]*LatinGuessNode
	wordEnd  bool
}

func (node *LatinGuessNode) Guess(stmt string) []string {
	all := []uint8(stmt)
	l := len(all)
	if l == 0 {
		return nil
	}
	matchSegTokens := make([][]*segment, l)
	for i := 0; i < l; i++ {
		matchSegTokens[i] = node.matchForward(i, all[i:])
	}
	stat := make([]*segTokenInternal, l)

	for i := l - 1; i >= 0; i-- {
		calc(stat, matchSegTokens[i], i)
	}

	last := stat[0]
	var result []string
	for last != nil {
		result = append(result, string(all[last.segment.start:last.segment.end]))
		if last.end > l-1 {
			break
		}
		last = stat[last.end]
	}

	return result

}

func calc(stat []*segTokenInternal, thisSegTokens []*segment, pos int) {
	l := len(stat)
	var minCount = math.MaxInt
	var minIndex int
	for i, seg := range thisSegTokens {
		count := 1
		if seg.end < l {
			next := stat[seg.end]
			if next != nil {
				count += next.count
			}
		}
		if minCount > count {
			minCount = count
			minIndex = i
		}
	}

	stat[pos] = &segTokenInternal{
		segment: thisSegTokens[minIndex],
		count:   minCount,
	}
}

func (node *LatinGuessNode) matchForward(from int, statement []uint8) []*segment {
	var ret []*segment
	p := node

	fCreate := func(wl int) *segment {
		return &segment{
			start: from,
			end:   from + wl,
		}
	}

	size := 0
	for _, v := range statement {
		if p.children == nil {
			break
		}
		found := p.children[v]
		if found == nil {
			break
		}
		size++
		if found.wordEnd {
			ret = append(ret, fCreate(size))
		}
		p = found
	}
	if ret == nil {
		ret = append(ret, fCreate(1))
	}

	return ret
}

func (node *LatinGuessNode) addWord(word string) {
	all := []uint8(word)
	l := len(all)
	p := node
	for i, c := range all {
		if !p.hasChild() {
			p.children = make(map[uint8]*LatinGuessNode)
		}
		child, ok := p.children[c]
		if !ok {
			child = &LatinGuessNode{}
			p.children[c] = child
		}
		if !child.wordEnd && (i == 0 || i == l-1) {
			child.wordEnd = true
		}
		p = child
	}
}

func (node *LatinGuessNode) hasChild() bool {
	return len(node.children) > 0
}

func LoadGuess(rootPath string) (*LatinGuessNode, error) {
	fp := path.Join(rootPath, "pinyin_alphabet.txt")
	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	root := &LatinGuessNode{}

	preventRepeat := map[string]struct{}{}

	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := scan.Text()

		word := strings.TrimSpace(line)

		if _, ok := preventRepeat[word]; ok {
			log.Printf("%s is repeat in %s\n", line, fp)
			continue
		}
		preventRepeat[word] = struct{}{}
		root.addWord(word)
	}

	return root, scan.Err()
}
