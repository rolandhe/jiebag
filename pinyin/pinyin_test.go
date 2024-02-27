package pinyin

import (
	"fmt"
	"testing"
)

func TestConvert(t *testing.T) {
	rootPath := "../dict/pinyin"
	node, err := LoadDict(rootPath)
	if err != nil {
		fmt.Println(err)
	}
	pinyinList := node.ConvertString("正品行货 正品行货 码完代码，他起身关上电脑，用滚烫的开水为自己泡制一碗腾着热气的老坛酸菜面。中国的程序员更偏爱拉上窗帘，在黑暗中享受这独特的美食。这是现代工业给一天辛苦劳作的人最好的馈赠。南方一带生长的程序员虽然在京城多年，但仍口味清淡，他们往往不加料包，由脸颊自然淌下的热泪补充恰当的盐分。他们相信，用这种方式，能够抹平思考着现在是不是过去想要的未来而带来的大部分忧伤…小李的父亲在年轻的时候也是从爷爷手里接收了祖传的代码，不过令人惊讶的是，到了小李这一代，很多东西都遗失了，但是程序员苦逼的味道保存的是如此的完整。 就在24小时之前，最新的需求从PM处传来，为了得到这份自然的馈赠，码农们开机、写码、调试、重构，四季轮回的等待换来这难得的丰收时刻。码农知道，需求的保鲜期只有短短的两天，码农们要以最快的速度对代码进行精致的加工，任何一个需求都可能在24小时之后失去原本的活力，变成一文不值的垃圾创意。", WithoutTone)
	fmt.Println(pinyinList)

	pinyinList1 := node.ConvertString("a河北乐亭核心目标a与，，，，，，,@#$%^&*(发展战略都市绿", WithoutTone)
	fmt.Println(pinyinList1)
}

func TestCloser(t *testing.T) {
	pos := 100

	var ret []int
	f := func() {
		ret = append(ret, pos)
		fmt.Println(pos)
	}

	f()
	pos = 123
	f()

	fmt.Println(ret)
}

func TestUnicodeTone(t *testing.T) {
	utone := unicodeTone("er", 3)
	fmt.Println(utone)
}
