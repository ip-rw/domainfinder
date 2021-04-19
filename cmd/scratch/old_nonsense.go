package main

import (
	"bufio"
	"github.com/ip-rw/rank/pkg/crawl"
	"github.com/ip-rw/rank/pkg/sources"
	"github.com/ip-rw/rank/pkg/util"
	"github.com/sirupsen/logrus"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var waitGroup = &sync.WaitGroup{}

func CrawlUrl(uri string, concurrent, depth int) *crawl.CrawlResult {
	results, err := crawl.Crawl(uri, concurrent, depth)
	if err != nil {
		logrus.WithError(err).Error("crawl aborted")
	}
	//logrus.Println(results.Scraped)
	//logrus.Println(results.Email)
	return results
}

func FindCompanyDomain(cno string) {
	var (
		concurrent = 8
		depth      = 2
		wg         = &sync.WaitGroup{}
		top_score  = 0
		page_count = 0
		top string
	)

	company, err := util.GetCompanyKeywords(cno)
	if err != nil {
		logrus.WithError(err).Error("failed to process documents")
		return
	}
	uris := sources.FindPossibleDomains(company)
	weights := map[string]int{
		sources.CleanCompanyName(company.Name):                         10,
		sources.CleanCompanyName(company.RegisteredAddress.PostalCode): 50,
		company.CompanyNumber: 100,
	}
	for _, val := range strings.Split(company.Bag, "\n") {
		//println(strings.TrimSpace(strings.ToLower(val)))
		weights[strings.TrimSpace(strings.ToLower(val))] = 1
	}
	//if len(company.IndustryCodes) > 0 {
	//	fmt.Println("SEC", company.IndustryCodes[0].IndustryCode.Description)
	//}
	for _, u := range uris {
		if strings.Index(u, "http") == 0 {
			wg.Add(1)

			go func(comp util.Company, uri string) {
				defer wg.Done()
				//fmt.Println("Crawling", uri)
				results := CrawlUrl(uri, concurrent, depth)
				if err != nil {
					logrus.WithError(err).Error("crawl aborted")
					return
				}
				page_count += len(results.Scraped)

				//text := strings.Join(results.text, " ")
				text := results.Text()
				score := 0
				for term, weight := range weights {
					ma := CountMatches(strings.ToLower(text), term)
					if ma > 0 {
						score += weight
						//fmt.Println(uri, term, score)
					}
				}
				if score > top_score {
					top_score = score
					top = uri
					//logrus.Println(uri, comp.Name, score)
				}

			}(*company, u)
		}
	}
	wg.Wait()
	if top_score > 0 {
		logrus.Infof("%s %s (%d) (crawled %d pages)", company.Name, top, top_score, page_count)
	} else {
		logrus.Warnf("%s failed (crawled %d pages from %d domains)", company.Name, page_count, len(uris))
	}
}

func CountMatches(haystack, needle string) int {
	//logrus.Println(needle)
	p := regexp.MustCompile(needle)
	//p := regexp.MustCompile(`(?i).{0,16}\b` + needle + `\b.{0,16}`)
	r := p.FindAllString(haystack, -1)
	//if len(r) > 0 {
		//for i := range r {
			//logrus.Println(r[i])
		//}
	//}
	return len(r)
}

func process(cnos chan string) {
	for cno := range cnos {
		FindCompanyDomain(cno)
	}
	waitGroup.Done()
}

func main() {
	var (
		scanner = bufio.NewScanner(os.Stdin)
		cnoChan = make(chan string, 6000)
	)
	scanner.Split(bufio.ScanLines)
	logrus.SetLevel(logrus.InfoLevel)
	concurrent, _ := strconv.Atoi(os.Args[1])
	waitGroup.Add(concurrent)
	for i := 0; i < concurrent; i++ {
		go process(cnoChan)
	}

	for scanner.Scan() {
		t := scanner.Text()
		n := strings.Trim(t, "\n ")
		if len(n) > 1 {
			cnoChan <- n
		}
	}
	close(cnoChan)
	waitGroup.Wait()
}
