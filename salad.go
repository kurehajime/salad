package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"

	"github.com/ikawaha/kagome/tokenizer"
)

const betterLen = 60  //どのくらいの長さの文章にしたいか(その長さにするとは言ってない)
const retryCount = 13 //文章の長さを揃えるのに何回試行錯誤するか

//Salad マルコフ連鎖してランダムな文章を作る。
type Salad struct {
	dict map[string][]string //基本的に辞書は使い回し
}

func init() {
	tokenizer.SysDic()
}

//NewSalad 初期化
func NewSalad(s string) *Salad {
	s = strings.NewReplacer(" ", "", "\n", "。", "「", "。", "」", "。", "(", "。", ")", "。", "?", "。", "!", "。").Replace(s)
	ham := Salad{}
	t := tokenizer.New()
	dict := make(map[string][]string)
	morphs := t.Tokenize(s)
	now, next := "", ""
	EOSKEYWORD := ".\n。"

	//辞書を作る
	//dict[前の単語]=["次の単語","次の単語","次の単語","次の単語"]
	for i := range morphs {
		now = morphs[i].Surface
		if i+1 < len(morphs) {
			next = morphs[i+1].Surface
		} else {
			next = "EOS"
		}

		//該当キーワードが最終文字だったらEOSに置き換え、BOSに次のキーワードを追加。
		if strings.Index(EOSKEYWORD, now) != -1 {
			if dict["BOS"] == nil {
				dict["BOS"] = make([]string, 0)
			}
			if strings.Index(EOSKEYWORD, next) == -1 {
				dict["BOS"] = append(dict["BOS"], next)
			}
			continue
		}
		//次の文字が最終文字ならEOSに置き換え
		if strings.Index(EOSKEYWORD, next) != -1 {
			next = "EOS"
		}
		//辞書に追加
		if dict[now] == nil {
			dict[now] = make([]string, 0)
		}
		dict[now] = append(dict[now], next)
	}
	ham.dict = dict
	return &ham
}

//辞書を元にマルコフ連鎖して文章を作る（内部用メソッド）。
func (h Salad) makeWord(s int64) string {
	rtn := ""
	rand.Seed(s)
	now := h.dict["BOS"][rand.Intn(len(h.dict["BOS"]))]
	for i := 0; i < 100; i++ {
		if now == "EOS" {
			break
		}
		rtn += now
		if len(h.dict[now]) != 0 {
			now = h.dict[now][rand.Intn(len(h.dict[now]))]
		}

	}
	return strings.Replace(rtn, "EOS", "", -1)
}

//MakeWord 辞書を元にマルコフ連鎖して文章を作る。
func (h Salad) MakeWord() string {
	rtn := ""
	line := ""
	for i := 0; i <= retryCount; i++ {
		s := int64(time.Now().Nanosecond())
		line = h.makeWord(s)
		if math.Abs(float64(len(rtn)-betterLen)) > math.Abs(float64(len(line)-betterLen)) {
			rtn = line
		}
	}
	return rtn
}

// エンコード変換
func transEnc(text string, encode string) (string, error) {
	body := []byte(text)
	var f []byte

	encodings := []string{"sjis", "utf-8"}
	if encode != "" {
		encodings = append([]string{encode}, encodings...)
	}
	for _, enc := range encodings {
		if enc != "" {
			ee, _ := charset.Lookup(enc)
			if ee == nil {
				continue
			}
			var buf bytes.Buffer
			ic := transform.NewWriter(&buf, ee.NewDecoder())
			_, err := ic.Write(body)
			if err != nil {
				continue
			}
			err = ic.Close()
			if err != nil {
				continue
			}
			f = buf.Bytes()
			break
		}
	}
	return string(f), nil
}
func readPipe() (string, error) {
	stats, _ := os.Stdin.Stat()
	if stats != nil && (stats.Mode()&os.ModeCharDevice) == 0 {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}
	return "", nil
}
func readStdin() (string, error) {
	var text string
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		if s.Text() == "" {
			break
		}
		text += s.Text() + "\n"
	}
	if s.Err() != nil {
		return "", s.Err()
	}
	return text, nil
}

func readFileByArg(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func main() {
	var err error
	var count int
	var defaultEncoding string
	var encode string
	text := ""
	if runtime.GOOS == "windows" {
		defaultEncoding = "sjis"
	} else {
		defaultEncoding = "utf-8"
	}

	flag.IntVar(&count, "c", 1, "count")
	flag.StringVar(&encode, "e", defaultEncoding, "encoding")
	flag.Parse()

	//get text
	if len(flag.Args()) == 0 {
		text, err = readPipe()
	} else if flag.Arg(0) == "-" {
		text, err = readStdin()
	} else {
		text, err = readFileByArg(flag.Arg(0))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	text, err = transEnc(text, encode)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	//salad
	h := NewSalad(text)
	for i := 0; i < count; i++ {
		fmt.Println(h.MakeWord())
	}
}
