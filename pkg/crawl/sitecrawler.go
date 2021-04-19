package crawl

import (
	"bytes"
	"crypto/tls"
	"github.com/gocolly/colly"
	"github.com/ip-rw/rank/pkg/sources"
	cregex "github.com/mingrammer/commonregex"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
	"jaytaylor.com/html2text"
	"mime"
	"path"
	_ "regexp"
	"time"

	"net/http"
	"net/url"
	"strings"
	"sync"
)

type SiteCrawler struct {
	*colly.Collector
	Results *CrawlResult
	Errors int
}

type CrawlResult struct {
	sync.Mutex
	Scraped []*url.URL
	text    []string
	Email   sync.Map
}

func (cr *CrawlResult)Emails() []string {
	var emails []string
	cr.Email.Range(func(key, value interface{}) bool {
		emails = append(emails, key.(string))
		return true
	})
	return emails
}

func (cr *CrawlResult)Text() string {
	return strings.Join(cr.text, "\n")
}

func NewCrawlResults() *CrawlResult {
	return &CrawlResult{
		Mutex:   sync.Mutex{},
		Scraped: []*url.URL{},
		text:    []string{},
		Email:   sync.Map{},
	}
}

func ParseAhref(e *colly.HTMLElement) {
	link := e.Attr("href")
	abs := e.Request.AbsoluteURL(link)
	if len(abs) > 1 {
		e.Request.Visit(abs)
	}
}

func ParseResponse(response *colly.Response, c *CrawlResult) {
	l := logrus.WithField("url", response.Request.URL)
	if strings.Index(http.DetectContentType(response.Body), "text/") != 0 {
		l.Debug("not html, skipping")
		return
	}

	text, err := html2text.FromReader(bytes.NewReader(response.Body), html2text.Options{OmitLinks: true, PrettyTables: false})
	if err != nil {
		l.WithError(err).Error("error stripping html")
		return
	}

	// Add page to history
	c.Scraped = append(c.Scraped, response.Request.URL)

	// Find e-mails
	emails := cregex.Emails(text)
	if len(emails) > 0 {
		for _, e := range emails {
			c.Email.Store(e, 0)
		}
	}

	for _, w := range strings.Split(sources.CleanCompanyName(text), " ") {
		c.text = append(c.text, w)
		//c.text = util.AppendUniq(c.text, w)
	}
	l.Debug("finished")
}

func NewSiteCrawler(depth int) *SiteCrawler {
	c := &SiteCrawler{
		Collector: colly.NewCollector(
			colly.MaxDepth(depth),
			colly.Async(true),
			colly.UserAgent("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"),
		),
	}
	c.WithTransport(&http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout:   3 * time.Second,
//		MaxIdleConns:          100,
//		MaxIdleConnsPerHost:   5,
//		MaxConnsPerHost:       10,
		IdleConnTimeout:       3 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
//		ForceAttemptHTTP2:     false,
	})
	c.IgnoreRobotsTxt = true
	c.CheckHead = false
	//c.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
	//	logrus.Println(via[len(via)-1].URL.String(), " redirected to ", req.URL.String())
	//	return nil
	//})
	c.OnError(func(response *colly.Response, e error) {
		c.Errors += 1
		logrus.WithError(e).WithField("url", response.Request.URL).Debug("request error")
	})
	c.OnRequest(func(request *colly.Request) {
		if len(c.Results.Scraped) > 50 && c.Errors > 10 {
			request.Abort()
		}
		if t := mime.TypeByExtension(path.Ext(request.URL.Path)); t != "" && strings.Index(t, "text/") != 0 {
			logrus.WithField("url", request.URL).Debug("mime looks binary")
			request.Abort()
		}
	})
	c.OnScraped(func(response *colly.Response) {
		ParseResponse(response, c.Results)
	})
	c.OnHTML("a[href]", func(element *colly.HTMLElement) {
		ParseAhref(element)
	})

	return c
}

func Crawl(uri string, concurrent int, depth int) (*CrawlResult, error) {
	c := NewSiteCrawler(depth)
	c.Results = NewCrawlResults()
	u, err := url.Parse(uri)
	if err != nil {
		return c.Results, err
	}
	c.AllowSubdomains(u, concurrent)
	c.Visit(u.String())
	if err != nil {
		return c.Results, err
	}
	c.Wait()
	return c.Results, nil
}

func (c *SiteCrawler) AllowSubdomains(u *url.URL, concurrent int) {
	if domain, err := publicsuffix.EffectiveTLDPlusOne(u.String()); err == nil {
//		c.URLFilters = append(c.URLFilters, regexp.MustCompile(`(?i)^http(s)://[a-zA-Z0-9\-_\.]*?`+regexp.QuoteMeta(domain)))
//		c.URLFilters = append(c.URLFilters, regexp.MustCompile(`(?i)^http(s)://` + regexp.QuoteMeta(domain)))

		c.Limit(&colly.LimitRule{
			DomainRegexp: `.*` + domain,
			Parallelism:  concurrent,
		})
	}
}
