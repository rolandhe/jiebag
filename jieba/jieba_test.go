package jieba

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
)

func TestMatch(t *testing.T) {
	rootDict, err := filepath.Abs("../dict")
	if err != nil {
		t.Fatal(err)
	}

	handler, err := MewSegmentHandler(rootDict)

	if err != nil {
		t.Fatal(err)
	}

	tokens := handler.SegParagraph("三黄鸡99元和太子奶18.90元。", ModeSearch)
	rjson, _ := json.Marshal(tokens)
	fmt.Printf("%s\n", string(rjson))
}
