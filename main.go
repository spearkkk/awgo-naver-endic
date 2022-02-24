package main

import (
	"github.com/deanishe/awgo"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var re = regexp.MustCompile(`[^a-zA-Zㄱ-ㅎ가-힣,() ]`)
var re2 = regexp.MustCompile(`[<>]`)

var wf *aw.Workflow

func init() {
	wf = aw.New()
}

func search(query string) []byte {
	if !wf.Cache.Expired(query, time.Minute) {
		cached, _ := wf.Cache.Load(query)
		return cached
	}

	baseUrl := "https://en.dict.naver.com/api3/enko/search"

	response, _ := http.PostForm(baseUrl, url.Values{"range": {"word"}, "query": {query}})
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	body, _ := ioutil.ReadAll(response.Body)
	_ = wf.Cache.Store(query, body)
	return body
}

func run() {
	query := wf.Args()[0]
	meanings := gjson.Get(string(search(query)), "searchResultMap.searchResultListMap.WORD.items.#.meansCollector.#.means.#.value")

	for _, meaning := range meanings.Array() {
		result := strings.ReplaceAll(strings.ReplaceAll(meaning.String(), "[", ""), "]", "")

		if re2.MatchString(result) {
			continue
		}

		title := strings.TrimSpace(re.ReplaceAllString(result, ""))

		item := wf.NewItem(title)
		item.Valid(true)
		item.Arg(query)
		item.Var("result", title)

		item.Cmd().Subtitle("네이버에서 검색하기...").Valid(true)

		item.Opt().Subtitle("크게 보기...").Valid(true)
	}

	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
