package jieba

import "path"

type ModeStyle int

const ModeSearch ModeStyle = 0
const ModeIndex ModeStyle = 1

const (
	BaseDictName    = "dict.txt"
	UserDictDirName = "user"
	BaseProbName    = "prob_emit.txt"
)

func MewSegmentHandler(dictRootPath string) (*SegmentHandler, error) {
	trie, err := newDictTrie(path.Join(dictRootPath, BaseDictName), path.Join(dictRootPath, UserDictDirName))
	if err != nil {
		return nil, err
	}

	hmm, err := newHmmSeg(path.Join(dictRootPath, BaseProbName))
	if err != nil {
		return nil, err
	}
	return &SegmentHandler{
		dict: trie,
		hmm:  hmm,
	}, nil
}

type SegToken struct {
	Word  string
	Start int
	End   int
}

type sentenceTrace struct {
	from   int
	to     int
	offset int
}

func (st *sentenceTrace) length() int {
	return st.to - st.from
}

type SegmentHandler struct {
	dict Trie
	hmm  HmmSeg
}

func (h *SegmentHandler) SegParagraph(s string, mode ModeStyle) []*SegToken {
	paragraph := []rune(s)

	var st sentenceTrace
	var segTokens []*SegToken

	for i, r := range paragraph {
		nr := regularize(r)
		paragraph[i] = nr
		if couldTrieSegSupport(nr) {
			st.to++
			continue
		}
		if st.length() > 0 {
			tokens := segSentence(h.dict, h.hmm, paragraph[st.from:st.to])
			segTokens = h.accept(segTokens, tokens, st.offset, mode)
			st.from = i + 1
			st.to = i + 1
		}
		segTokens = append(segTokens, &SegToken{
			Word:  string([]rune{nr}),
			Start: st.offset,
			End:   st.offset + 1,
		})
		st.offset++
	}

	if st.length() > 0 {
		tokens := segSentence(h.dict, h.hmm, paragraph[st.from:st.to])
		segTokens = h.accept(segTokens, tokens, st.offset, mode)
	}
	return segTokens
}

func (h *SegmentHandler) accept(segTokens []*SegToken, tokens []string, offset int, mode ModeStyle) []*SegToken {
	for _, token := range tokens {
		tokenArray := []rune(token)
		if mode == ModeIndex {
			segTokens = h.acceptShort(segTokens, tokenArray, offset)
		}
		end := offset + len(tokenArray)
		segTokens = append(segTokens, &SegToken{
			Word:  token,
			Start: offset,
			End:   end,
		})
		offset += end
	}

	return segTokens
}
func (h *SegmentHandler) acceptShort(segTokens []*SegToken, token []rune, offset int) []*SegToken {
	l := len(token)
	if l <= 2 {
		return segTokens
	}
	f := func(step int) {
		if l > step {
			for j := 0; j < l-step+1; j++ {
				gram := string(token[j : j+step])
				if h.dict.ExistShortWord(gram) {
					segTokens = append(segTokens, &SegToken{
						Word:  gram,
						Start: offset + j,
						End:   offset + j + step,
					})
				}
			}
		}
	}
	f(2)
	f(3)
	return segTokens
}

func segSentence(dict Trie, hmm HmmSeg, sentence []rune) []string {
	segments := dict.Match(sentence)

	var tokens []string

	f := func(needHmmStat []rune) {
		word := string(needHmmStat)
		if dict.ExistShortWord(word) {
			tokens = append(tokens, word)
		} else {
			// call hmm
			tokens = append(tokens, hmm.Cut(needHmmStat)...)
		}
	}

	var st sentenceTrace
	for _, seg := range segments {
		if seg.len() == 1 {
			st.to = seg.End
			continue
		}
		if st.length() > 0 {
			needHmmStat := sentence[st.from:st.to]
			f(needHmmStat)
		}

		tokens = append(tokens, string(sentence[seg.Start:seg.End]))
		st.from = seg.End
		st.to = seg.End
	}

	l := len(sentence)

	if st.from < l {
		f(sentence[st.from:])
	}

	return tokens
}
