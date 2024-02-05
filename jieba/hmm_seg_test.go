package jieba

import (
	"fmt"
	"path"
	"path/filepath"
	"testing"
)

func TestSegHmm(t *testing.T) {
	rootDict, err := filepath.Abs("../dict")
	if err != nil {
		t.Fatal(err)
	}
	hmm, err := newHmmSeg(path.Join(rootDict, BaseProbName))
	if err != nil {
		t.Fatal(err)
	}
	sentence := "三黄鸡99元和太子奶18.90元"
	//sentence := "元和太子奶"
	tokens := hmm.Cut([]rune(sentence))
	t.Log(tokens)
}

func TestReSkip(t *testing.T) {
	other := "我的0.997和abc"
	tokenIndexes := reSkip.FindAllStringIndex(other, -1)
	offset := 0

	var tokens []string

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
	fmt.Println(tokens)
}
