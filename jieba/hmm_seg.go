package jieba

import (
	"bufio"
	"os"
	"slices"
	"strconv"
	"strings"
)

const minFloat = -3.14e100

var (
	states     = []rune{'B', 'M', 'E', 'S'}
	prevStatus = map[rune][]rune{}
	start      = map[rune]float64{}
	trans      = map[rune]map[rune]float64{}
)

func init() {
	prevStatus['B'] = []rune{'E', 'S'}
	prevStatus['M'] = []rune{'M', 'B'}
	prevStatus['S'] = []rune{'S', 'E'}
	prevStatus['E'] = []rune{'B', 'M'}

	start['B'] = -0.26268660809250016
	start['E'] = -3.14e+100
	start['M'] = -3.14e+100
	start['S'] = -1.4652633398537678

	transB := map[rune]float64{}
	transB['E'] = -0.510825623765990
	transB['M'] = -0.916290731874155
	trans['B'] = transB
	transE := map[rune]float64{}
	transE['B'] = -0.5897149736854513
	transE['S'] = -0.8085250474669937
	trans['E'] = transE
	transM := map[rune]float64{}
	transM['E'] = -0.33344856811948514
	transM['M'] = -1.2603623820268226
	trans['M'] = transM

	transS := map[rune]float64{}
	transS['B'] = -0.7211965654669841
	transS['S'] = -0.6658631448798212
	trans['S'] = transS
}

type HmmSeg interface {
	Cut(statement []rune) []string
}

func newHmmSeg(dictPath string) (HmmSeg, error) {
	hmm := &hmmSegImpl{
		emits: map[rune]map[rune]float64{},
	}

	if err := hmm.loadModel(dictPath); err != nil {
		return nil, err
	}

	return hmm, nil
}

type hmmSegImpl struct {
	emits map[rune]map[rune]float64
}

func (hmm *hmmSegImpl) loadModel(fp string) error {
	f, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	var values map[rune]float64
	for scan.Scan() {
		line := scan.Text()
		items := strings.Fields(line)

		if len(items) == 1 {
			values = map[rune]float64{}
			rvs := []rune(items[0])
			hmm.emits[rvs[0]] = values
		} else {
			rvs := []rune(items[0])
			if values[rvs[0]], err = strconv.ParseFloat(items[1], 64); err != nil {
				return err
			}
		}
	}

	return scan.Err()
}

func resetChinese(chinese []rune) []rune {
	l := len(chinese)
	if l == 0 || l > 4096 {
		return make([]rune, 0, 128)
	}
	return chinese[:0]
}

func (hmm *hmmSegImpl) Cut(statement []rune) []string {
	var tokens []string

	chinese := resetChinese(nil)
	var other strings.Builder

	for _, r := range statement {
		if isCjkNormal(r) {
			if other.Len() > 0 {
				tokens = hmm.processOtherUnknownWords(other.String(), tokens)
				other.Reset()
			}
			chinese = append(chinese, r)
			continue
		}
		if len(chinese) > 0 {
			tokens = hmm.viterbi(chinese, tokens)
			chinese = resetChinese(chinese)
		}
		other.WriteRune(r)
	}

	if len(chinese) > 0 {
		tokens = hmm.viterbi(chinese, tokens)
	}
	if other.Len() > 0 {
		tokens = hmm.processOtherUnknownWords(other.String(), tokens)
	}
	return tokens
}

type vNode struct {
	v rune
	p *vNode
}

type candidate struct {
	v    rune
	freq float64
}

func (hmm *hmmSegImpl) viterbi(chinese []rune, tokens []string) []string {
	l := len(chinese)
	v := make([]map[rune]float64, l)
	for i := 0; i < l; i++ {
		v[i] = map[rune]float64{}
	}
	path := map[rune]*vNode{}

	for _, state := range states {
		emP, ok := hmm.emits[state][chinese[0]]
		if !ok {
			emP = minFloat
		}
		v[0][state] = start[state] + emP
		path[state] = &vNode{
			v: state,
		}
	}
	for i, r := range chinese {
		if i == 0 {
			continue
		}
		newPath := map[rune]*vNode{}
		var emP float64
		var tranP float64
		var ok bool
		for _, y := range states {
			if emP, ok = hmm.emits[y][r]; !ok {
				emP = minFloat
			}
			var candi *candidate
			for _, y0 := range prevStatus[y] {
				if tranP, ok = trans[y0][y]; !ok {
					tranP = minFloat
				}
				tranP += emP + v[i-1][y0]
				if candi == nil {
					candi = &candidate{
						v:    y0,
						freq: tranP,
					}
				} else if candi.freq <= tranP {
					candi.v = y0
					candi.freq = tranP
				}
			}
			v[i][y] = candi.freq
			newPath[y] = &vNode{
				v: y,
				p: path[candi.v],
			}
		}
		path = newPath
	}

	probE := v[l-1]['E']
	probS := v[l-1]['S']
	postList := make([]rune, 0, l)
	win := path['E']
	if probE < probS {
		win = path['S']
	}
	for win != nil {
		postList = append(postList, win.v)
		win = win.p
	}
	slices.Reverse(postList)

	begin := 0
	next := 0
	for i := 0; i < l; i++ {
		postR := postList[i]
		if postR == 'B' {
			begin = i
		} else if postR == 'E' {
			next = i + 1
			tokens = append(tokens, string(chinese[begin:next]))
		} else if postR == 'S' {
			next = i + 1
			tokens = append(tokens, string(chinese[i:next]))
		}
	}

	if next < l {
		tokens = append(tokens, string(chinese[next:]))
	}

	return tokens
}

func (hmm *hmmSegImpl) processOtherUnknownWords(other string, tokens []string) []string {
	tokenIndexes := reSkip.FindAllStringIndex(other, -1)
	offset := 0

	for _, loc := range tokenIndexes {
		if loc[0] > offset {
			tokens = append(tokens, other[offset:loc[0]])
		}
		tokens = append(tokens, other[loc[0]:loc[1]])
		offset = loc[1]
	}
	if offset < len(other) {
		tokens = append(tokens, other[offset:len(other)])
	}
	return tokens
}
