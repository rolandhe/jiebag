package jieba

import (
	"bufio"
	"errors"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Segment struct {
	Start int
	End   int
}

type segTokenInternal struct {
	*Segment
	weight float64
}

func newDictTrie(baseDict string, userDictDir string) (Trie, error) {
	root := &trieNodeHolder{
		trieNode:  &trieNode{},
		minFreq:   0x1.fffffffffffffp+1023,
		shortWord: map[string]int8{},
	}

	if err := loadBase(root, baseDict); err != nil {
		return nil, err
	}

	if err := loadByDir(root, userDictDir, func(nd *trieNode) {
		nd.freq = math.Log(nd.freq / root.total)
	}); err != nil {
		return nil, err
	}
	return root, nil
}

func loadBase(root *trieNodeHolder, baseDict string) error {
	var collect []*trieNode
	if err := root.loadDict(baseDict, func(nd *trieNode) {
		root.total += nd.freq
		collect = append(collect, nd)
	}); err != nil {
		return err
	}
	for _, nd := range collect {
		nd.freq = math.Log(nd.freq / root.total)
		root.minFreq = math.Min(nd.freq, root.minFreq)
	}
	return nil
}

func loadByDir(root *trieNodeHolder, dirPath string, afterWord func(nd *trieNode)) error {
	if len(dirPath) == 0 {
		return nil
	}

	fileSystem := os.DirFS(dirPath)

	return fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		return root.loadDict(filepath.Join(dirPath, path), afterWord)
	})
}

func (seg *Segment) ToString(sentence []rune) string {
	return string(sentence[seg.Start:seg.End])
}
func (seg *Segment) len() int {
	return seg.End - seg.Start
}

type Trie interface {
	Match(sentence []rune) []*Segment
	ExistShortWord(word string) bool
}

type trieNode struct {
	children map[rune]*trieNode
	value    rune

	wordEnd bool
	freq    float64
}

type trieNodeHolder struct {
	*trieNode
	total     float64
	minFreq   float64
	shortWord map[string]int8
}

func (node *trieNode) hasNext() bool {
	l := len(node.children)
	return l > 0
}

func (root *trieNodeHolder) Match(sentence []rune) []*Segment {
	l := len(sentence)
	if l == 0 {
		return nil
	}
	matchSegTokens := make([][]*segTokenInternal, l)
	for i := 0; i < l; i++ {
		matchSegTokens[i] = root.matchForward(i, sentence[i:])
	}
	stat := make([]*segTokenInternal, l)

	for i := l - 1; i >= 0; i-- {
		calc(stat, matchSegTokens[i], i)
	}

	last := stat[0]
	var result []*Segment
	for last != nil {
		result = append(result, last.Segment)
		if last.End > l-1 {
			break
		}
		last = stat[last.End]
	}

	return result
}

func calc(stat []*segTokenInternal, thisSegTokens []*segTokenInternal, pos int) {
	l := len(stat)
	//if len(thisSegTokens) == 0 {
	//	if pos != l-1 {
	//		stat[pos] = stat[pos+1]
	//	}
	//	return
	//}
	var maxWeight = math.Inf(-1)
	var maxIndex int
	for i, seg := range thisSegTokens {
		weight := seg.weight
		if seg.End < l {
			next := stat[seg.End]
			if next != nil {
				weight += next.weight
			}
		}
		if maxWeight < weight {
			maxWeight = weight
			maxIndex = i
		}
	}

	stat[pos] = &segTokenInternal{
		Segment: thisSegTokens[maxIndex].Segment,
		weight:  maxWeight,
	}
}

func (root *trieNodeHolder) matchForward(from int, statement []rune) []*segTokenInternal {
	var ret []*segTokenInternal
	p := root.trieNode

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
			ret = append(ret, &segTokenInternal{
				Segment: &Segment{
					Start: from,
					End:   from + size,
				},
				weight: found.freq,
			})
		}
		p = found
	}
	if ret == nil {
		ret = append(ret, &segTokenInternal{
			Segment: &Segment{
				Start: from,
				End:   from + 1,
			},
			weight: root.minFreq,
		})
	}

	return ret
}

func (root *trieNodeHolder) loadDict(fp string, afterWord func(nd *trieNode)) error {
	f, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	preventRepeat := map[string]int8{}

	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := scan.Text()
		word, freq, err := splitDictLine(line)
		if err != nil {
			return err
		}

		if _, ok := preventRepeat[word]; ok {
			log.Printf("%s is repeat in %s\n", line, fp)
			continue
		}
		preventRepeat[word] = 1
		root.addWord([]rune(word), freq, afterWord)
	}

	return scan.Err()
}

func (root *trieNodeHolder) addWord(runes []rune, freq float64, afterWord func(nd *trieNode)) {
	l := len(runes)
	if l == 0 {
		return
	}
	p := root.trieNode
	for i, v := range runes {
		if p.children == nil {
			p.children = map[rune]*trieNode{}
		}
		isEnd := i == l-1
		curNode := p.children[v]
		if curNode == nil {
			curNode = &trieNode{
				value: v,
			}
			p.children[v] = curNode
		}
		if isEnd {
			curNode.wordEnd = true
			curNode.freq = freq
			if afterWord != nil {
				afterWord(curNode)
			}
		}
		p = curNode
	}
	root.shortWord[string(runes)] = 1
}

func (root *trieNodeHolder) ExistShortWord(word string) bool {
	_, ok := root.shortWord[word]
	return ok
}

func splitDictLine(line string) (string, float64, error) {
	items := strings.Fields(line)
	if len(items) < 2 {
		return "", 0.0, errors.New("bad items:" + line)
	}
	freq, err := strconv.ParseFloat(items[1], 64)
	word := strings.ToLower(items[0])
	return word, freq, err
}
