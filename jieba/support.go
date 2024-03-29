package jieba

import (
	"regexp"
	"slices"
)

const (
	cjkNormalStart = 0x4E00
	cjkNormalEnd   = 0x9FFF
)

var connectors = []rune{'+', '#', '&', '.', '_', '-'}
var reSkip *regexp.Regexp

func init() {
	slices.Sort(connectors)
	var err error
	reSkip, err = regexp.Compile("(\\d+\\.\\d+|[a-zA-Z0-9]+)")
	if err != nil {
		panic(err.Error())
	}
}

func isCjkNormal(r rune) bool {
	return r >= cjkNormalStart && r <= cjkNormalEnd
}

func isEnglish(r rune) bool {
	return (r >= 0x0041 && r <= 0x005A) || (r >= 0x0061 && r <= 0x007A)
}

func isDigit(r rune) bool {
	return r >= 0x0030 && r <= 0x0039
}

func isConnector(r rune) bool {
	_, found := slices.BinarySearch(connectors, r)
	return found
}

func regularize(input rune) rune {
	// 全角空格
	if input == 12288 {
		return 32
	}
	// 繁体->简体
	if input > 65280 && input < 65375 {
		return input - 65248
	}
	// 大写转小写
	if input >= 'A' && input <= 'Z' {
		return input + 32
	}
	return input
}

func couldTrieSegSupport(r rune) bool {
	return isCjkNormal(r) || isEnglish(r) || isDigit(r) || isConnector(r)
}
