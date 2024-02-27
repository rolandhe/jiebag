package jieba

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestTf(t *testing.T) {
	handler := loadHandler()
	rootDict, _ := filepath.Abs("../dict")
	tf, err := NewTfidf(rootDict, handler)
	if err != nil {
		fmt.Println(err)
	}
	content := "太阳照在桑干河上，太阳每天升起，每天有落下，当太阳落山后，月亮升了起来，桑干河静静地流淌着，月光洒落在河面上，月亮慢慢落下，黎明前一片漆黑，伸手不见五指，桑干河安静的等待着明天的太阳再升起。"
	all := tf.TopNByString(content, 100)
	fmt.Print("+v\n", all)
}
