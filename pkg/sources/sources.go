package sources

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/ip-rw/rank/pkg/util"
	"github.com/levigross/grequests"
	"net/url"
	"regexp"
	"strings"
)

type Company struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Logo   string `json:"logo"`
}

var (
	stop        = regexp.MustCompile("(?i)limited|ltd|plc|inc|incorporated|the")
	strip       = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	ws          = regexp.MustCompile(`\s+`)
)
type DuckDuckGo struct {
}
func (c DuckDuckGo) Name() string {
	return "DuckDuckGo"
}

func (g DuckDuckGo) Lookup(q string) ([]string, error) {
	r, err := grequests.Get("http://duckduckgo.com/html?kh=-1&kp=-2&kl=uk-en&q=" + url.QueryEscape(CleanCompanyName(q)), util.GetProxyRequestOptions())
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r.Bytes()))
	if err != nil {
		return nil, err
	}
	domains := make([]string, 0)
	doc.Find("a.result__a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if !exists {
			return
		}
		u, _ := url.Parse(link)
		link = u.Query().Get("uddg")
		if link == "" {
			return
		}
		u, _ = url.Parse(link)
		if u.Host != "" && u.Path == "/" {
			link = u.Scheme + "://" + u.Hostname()
			domains = util.AppendUniq(domains, link)
		}
	})

	return domains, nil
}
type TLD struct{}

func (c TLD) Name() string {
	return "TLD"
}
func (c TLD) Lookup(company string) ([]string, error) {
	out := []string{}
	cc := CleanCompanyName(company)
	for _, tld := range []string{".co.uk", ".com"} {
		out = append(out, "http://"+strip.ReplaceAllString(cc, "") + tld)
		out = append(out, "http://www."+strip.ReplaceAllString(cc, "") + tld)
	}
	return out, nil
}

type Clearbit struct{}

func (c Clearbit) Name() string {
	return "Clearbit"
}

func CleanCompanyName(company string) string {
	cleaned := strip.ReplaceAllString(strings.ToLower(company), " ")
	cleaned = ws.ReplaceAllString(cleaned, " ")
	return strings.TrimSpace(stop.ReplaceAllString(cleaned, ""))
}

func (c Clearbit) Lookup(company string) ([]string, error) {
	rawUrl := "https://autocomplete.clearbit.com/v1/companies/suggest?query=" + url.QueryEscape(CleanCompanyName(company))
	r, err := grequests.Get(rawUrl, util.GetProxyRequestOptions())
	if err != nil {
		return nil, err
	}
	var results []*Company
	if err = r.JSON(&results); err != nil {
		return nil, err
	}
	domains := make([]string, 0, len(results))
	for _, c := range results {
		domains = append(domains, "http://"+c.Domain)
		domains = append(domains, "https://"+c.Domain)

	}
	return domains, nil
}
