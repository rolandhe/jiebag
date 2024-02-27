移植[jieba分词java](!https://github.com/huaban/jieba-analysis)，并实现了[nlp-lang](https://github.com/NLPchina/nlp-lang)的pinyin功能。
jieba分词的原理请参jieba分词原作者[fxsjy](https://github.com/fxsjy) 的代码。


# 使用说明
## copy 整个dict目录到你的项目

jiebag库没有把词库内置到go代码文件中，而是可以独立的放到一个目录中，所以使用之前需要把词库文件copy到你的项目中。

## 分词使用示例

### 初始化 SegmentHandler

```

func initHandler() *SegmentHandler {
	rootDict, err := filepath.Abs("你的dict目录") // e.g "../dict"
	if err != nil {
		log.Fatal(err)
	}

	handler, err := MewSegmentHandler(rootDict)

	if err != nil {
		log.Fatal(err)
	}
	return handler
}

```

### 使用分词

```

tokens := handler.SegParagraph("中华人民共和国站起来了", jiebag.ModeIndex) // 索引模式


tokens1 := handler.SegParagraph("我是中国人", jiebag.ModeSearch) // 搜索模式

```


## tfidf使用

tfidf依赖 SegmentHandler，因此需要先调用 initHandler函数，然后NewTfidf函数初始化。

### 初始化

```
    handler := initHandler()
    rootDict, _ := filepath.Abs("../dict")
    tf, err := NewTfidf(rootDict, handler)
    if err != nil {
        fmt.Println(err)
	}
```

### 使用

```

    content := "太阳照在桑干河上，太阳每天升起，每天有落下，当太阳落山后，月亮升了起来，桑干河静静地流淌着，月光洒落在河面上，月亮慢慢落下，黎明前一片漆黑，伸手不见五指，桑干河安静的等待着明天的太阳再升起。"
	all := tf.TopNByString(content, 100)
```


## 拼音使用

同样需要初始库。

## 初始化

```

    rootPath := "../dict/pinyin" // 指定你自己的词库目录
	node, err := LoadDict(rootPath)
	if err != nil {
		fmt.Println(err)
	}

```

## 调用

```
    
    pinyinList1 := node.ConvertString("a河北乐亭核心目标a与，，，，，，,@#$%^&*(发展战略都市绿", WithoutTone)
	fmt.Println(pinyinList1)
```

### 支持的转换方式

*  WithoutTone, 不带音调， biao
*  ToneTail, 音调在后面， biao3, 音调支持1、2、3、4、5，5是轻声
*  UnicodeWithTone， biăo