package jieba

import (
	"bufio"
	"cmp"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
)

func NewTfidf(rootPath string, segHandler *SegmentHandler) (Tfidf, error) {
	idfMap, err := loadTfidfDict(rootPath)
	if err != nil {
		return nil, err
	}
	stopWords, err := loadStopWord(rootPath)
	if err != nil {
		return nil, err
	}
	return &tfIdfImpl{
		idfMap:      idfMap,
		mediumValue: calMedium(idfMap),
		stopWords:   stopWords,
		segHandler:  segHandler,
	}, nil
}

func loadTfidfDict(rootPath string) (map[string]float64, error) {
	fileFullName := path.Join(rootPath, "idf_dict.txt")
	f, err := os.Open(fileFullName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stopMap := map[string]float64{}
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := scan.Text()
		line = strings.TrimSpace(line)
		if err != nil {
			return nil, err
		}

		items := strings.Fields(line)

		freq, _ := strconv.ParseFloat(items[1], 64)

		stopMap[items[1]] = freq
	}

	return stopMap, scan.Err()
}

func calMedium(idMap map[string]float64) float64 {
	l := len(idMap)
	list := make([]float64, 0, l)
	for _, freq := range idMap {
		list = append(list, freq)
	}
	slices.Sort(list)
	return list[l/2]
}

func loadStopWord(rootPath string) (map[string]struct{}, error) {
	fileFullName := path.Join(rootPath, "stop_words.txt")

	f, err := os.Open(fileFullName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stopMap := map[string]struct{}{}
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := scan.Text()
		line = strings.TrimSpace(line)
		if err != nil {
			return nil, err
		}

		stopMap[line] = struct{}{}
	}

	return stopMap, scan.Err()
}

type Keyword struct {
	Word       string
	TfidfValue float64
}

type Tfidf interface {
	TopN(input []rune, n int) []*Keyword
	TopNByString(input string, n int) []*Keyword
}

type tfIdfImpl struct {
	idfMap      map[string]float64
	mediumValue float64
	stopWords   map[string]struct{}
	segHandler  *SegmentHandler
}

func (tf *tfIdfImpl) TopN(input []rune, n int) []*Keyword {
	tfTokenMap := tf.getTf(input)
	l := len(tfTokenMap)
	keyWords := make([]*Keyword, 0, l)
	var ok bool
	var idfValue float64
	for token, tfValue := range tfTokenMap {
		if idfValue, ok = tf.idfMap[token]; !ok {
			idfValue = tf.mediumValue
		}
		keyWords = append(keyWords, &Keyword{
			Word:       token,
			TfidfValue: tfValue * idfValue,
		})
	}
	slices.SortFunc(keyWords, func(a, b *Keyword) int {
		return cmp.Compare(a.TfidfValue, b.TfidfValue) * -1
	})
	if l <= n {
		return keyWords
	}
	if n <= 10000 && l >= 100000 {
		result := make([]*Keyword, n)
		copy(result, keyWords)
		return result
	}
	return keyWords[:n]
}

func (tf *tfIdfImpl) TopNByString(input string, n int) []*Keyword {
	return tf.TopN([]rune(input), n)
}

func (tf *tfIdfImpl) getTf(input []rune) map[string]float64 {
	tokens := tf.segHandler.segSentence(input)

	freqMap := map[string]int{}

	wordCount := 0
	for _, token := range tokens {
		if len(token) <= 1 {
			continue
		}
		if _, ok := tf.stopWords[token]; ok {
			continue
		}
		wordCount++

		freq := freqMap[token]
		freq++
		freqMap[token] = freq
	}

	tfMap := make(map[string]float64, len(freqMap))
	for token, freq := range freqMap {
		tfMap[token] = float64(freq) / float64(wordCount)
	}
	return tfMap
}
