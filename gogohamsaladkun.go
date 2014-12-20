package main

import (
	"fmt"
	kagome "github.com/ikawaha/kagome"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const BETTER_STR_LEN = 60 //どのくらいの長さの文章にしたいか(その長さにするとは言ってない)
const RETRY = 13          //文章の長さを揃えるのに何回試行錯誤するか

//マルコフ連鎖してランダムな文章を作る。
type Hamsalad struct {
	dict map[string][]string //基本的に辞書は使い回し
}

//初期化
func NewHamsalad() *Hamsalad {
	ham := Hamsalad{}
	s := ham.readData()
	t := kagome.NewTokenizer() //形態素解析
	dict := make(map[string][]string)
	morphs := t.Tokenize(s)
	now, next := "", ""
	EOSKEYWORD := ".\n。"

	//辞書を作る
	//dict[前の単語]=["次の単語","次の単語","次の単語","次の単語"]
	for i, _ := range morphs {
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
func (this *Hamsalad) makeWord() string {
	rtn := ""
	rand.Seed(int64(time.Now().Nanosecond()))
	now := this.dict["BOS"][rand.Intn(len(this.dict["BOS"]))]
	for i := 0; i < 100; i++ {
		if now == "EOS" {
			break
		}
		rtn += now
		if len(this.dict[now]) != 0 {
			now = this.dict[now][rand.Intn(len(this.dict[now]))]
		}

	}
	return rtn
}

//辞書を元にマルコフ連鎖して文章を作る。
func (this *Hamsalad) MakeWord() string {
	rtn := ""
	line := ""
	for i := 0; i <= RETRY; i++ {
		line = this.makeWord()
		if math.Abs(float64(len(rtn)-BETTER_STR_LEN)) > math.Abs(float64(len(line)-BETTER_STR_LEN)) {
			rtn = line
		}
	}
	return rtn
}

//カレントディレクトリの***.txtを全部読み込む
func (this *Hamsalad) readData() string {
	s := ""
	pwd, _ := filepath.Abs(".")
	list, err := ioutil.ReadDir(pwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	for _, finfo := range list {
		if strings.Index(finfo.Name(), ".txt") != -1 {
			contents, err := ioutil.ReadFile(finfo.Name())
			if err != nil {
				fmt.Println(contents, err)
				return ""
			}
			s += string(contents)
		}
	}
	r := strings.NewReplacer(" ", "", "\n", "。", "「", "。", "」", "。", "(", "。", ")", "。", "?", "。", "!", "。")
	s = r.Replace(s)
	return s
}
func main() {
	count := 1
	if len(os.Args) >= 2 {
		i, err := strconv.Atoi(os.Args[1])
		if err == nil {
			count = i
		}
	}
	h := NewHamsalad()
	for i := 0; i < count; i++ {
		fmt.Println(h.MakeWord())
	}
}
