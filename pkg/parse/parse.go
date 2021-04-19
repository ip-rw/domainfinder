package parse

import (
	"github.com/go-gota/gota/dataframe"
	"github.com/jdkato/prose/summarize"
	"strconv"
)

func GetKeywordDataFrame(kw map[string]int) dataframe.DataFrame {
	kwi := make([][]string, len(kw)+1)
	i := 0
	kwi[i] = []string{"keyword", "freq"}
	for k, v := range kw {
		i++
		kwi[i] = []string{ k, strconv.Itoa(v) }
	}
	df := dataframe.LoadRecords(
		kwi,
	)
	return df
}

func GetKeywords(text string) map[string]int {
	d := summarize.NewDocument(text)
	return d.Keywords()
}
