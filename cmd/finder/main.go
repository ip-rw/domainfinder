package main

import (
	"github.com/ip-rw/rank/pkg/crawl"
	"github.com/ip-rw/rank/pkg/sources"
	"github.com/ip-rw/rank/pkg/util"
	"os"
	"strings"
	"sync"

	"github.com/james-bowman/nlp"
	"github.com/james-bowman/nlp/measures/pairwise"
	"github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/mat"
)

func CrawlUrl(uri string, concurrent, depth int) *crawl.CrawlResult {
	results, err := crawl.Crawl(uri, concurrent, depth)
	if err != nil {
		logrus.WithError(err).Error("crawl aborted")
	}
	//logrus.Println(results.Scraped)
	//logrus.Println(results.Email)
	return results
}

var stopWords = []string{"a", "about", "above", "above", "across", "after", "afterwards", "again", "against", "all", "almost", "alone", "along", "already", "also", "although", "always", "am", "among", "amongst", "amoungst", "amount", "an", "and", "another", "any", "anyhow", "anyone", "anything", "anyway", "anywhere", "are", "around", "as", "at", "back", "be", "became", "because", "become", "becomes", "becoming", "been", "before", "beforehand", "behind", "being", "below", "beside", "besides", "between", "beyond", "bill", "both", "bottom", "but", "by", "call", "can", "cannot", "cant", "co", "con", "could", "couldnt", "cry", "de", "describe", "detail", "do", "done", "down", "due", "during", "each", "eg", "eight", "either", "eleven", "else", "elsewhere", "empty", "enough", "etc", "even", "ever", "every", "everyone", "everything", "everywhere", "except", "few", "fifteen", "fify", "fill", "find", "fire", "first", "five", "for", "former", "formerly", "forty", "found", "four", "from", "front", "full", "further", "get", "give", "go", "had", "has", "hasnt", "have", "he", "hence", "her", "here", "hereafter", "hereby", "herein", "hereupon", "hers", "herself", "him", "himself", "his", "how", "however", "hundred", "ie", "if", "in", "inc", "indeed", "interest", "into", "is", "it", "its", "itself", "keep", "last", "latter", "latterly", "least", "less", "ltd", "made", "many", "may", "me", "meanwhile", "might", "mill", "mine", "more", "moreover", "most", "mostly", "move", "much", "must", "my", "myself", "name", "namely", "neither", "never", "nevertheless", "next", "nine", "no", "nobody", "none", "noone", "nor", "not", "nothing", "now", "nowhere", "of", "off", "often", "on", "once", "one", "only", "onto", "or", "other", "others", "otherwise", "our", "ours", "ourselves", "out", "over", "own", "part", "per", "perhaps", "please", "put", "rather", "re", "same", "see", "seem", "seemed", "seeming", "seems", "serious", "several", "she", "should", "show", "side", "since", "sincere", "six", "sixty", "so", "some", "somehow", "someone", "something", "sometime", "sometimes", "somewhere", "still", "such", "system", "take", "ten", "than", "that", "the", "their", "them", "themselves", "then", "thence", "there", "thereafter", "thereby", "therefore", "therein", "thereupon", "these", "they", "thickv", "thin", "third", "this", "those", "though", "three", "through", "throughout", "thru", "thus", "to", "together", "too", "top", "toward", "towards", "twelve", "twenty", "two", "un", "under", "until", "up", "upon", "us", "very", "via", "was", "we", "well", "were", "what", "whatever", "when", "whence", "whenever", "where", "whereafter", "whereas", "whereby", "wherein", "whereupon", "wherever", "whether", "which", "while", "whither", "who", "whoever", "whole", "whom", "whose", "why", "will", "with", "within", "without", "would", "yet", "you", "your", "yours", "yourself", "yourselves"}

func FindCompanyDomain(cno string) bool {
	var (
		urls          = []string{}
		crawl_results = []*crawl.CrawlResult{}
		corpus        = []string{}
		concurrent    = 15
		depth         = 1
		wg            = sync.WaitGroup{}
		lock          = sync.Mutex{}
		vectoriser    = nlp.NewCountVectoriser(stopWords...)
		transformer   = nlp.NewTfidfTransformer()
		reducer       = nlp.NewTruncatedSVD(260)
		lsiPipeline   = nlp.NewPipeline(vectoriser, transformer, reducer)
	)
	//println(cno)
	company, err := util.GetCompanyKeywords(cno)
	if err != nil {
		logrus.WithError(err).Error("failed to process documents")
		return false
	}
	//fmt.Println(company)
	uris := sources.FindPossibleDomains(company)
	for _, u := range uris {
		if strings.Index(u, "http") == 0 {
			wg.Add(1)
			go func(uri string) {
				defer wg.Done()
				results := CrawlUrl(uri, concurrent, depth)
				if err != nil {
					logrus.WithError(err).Error("crawl aborted")
				}
				lock.Lock()
				urls = append(urls, uri)
				crawl_results = append(crawl_results, results)
				corpus = append(corpus, results.Text())
				lock.Unlock()
			}(u)
		}
	}
	wg.Wait()
	if len(corpus) < 1 {
		return false
	}
	//println(len(corpus))
	valid := false
	for _, c := range corpus {
		if len(c) > 0 {
			valid = true
		}
	}
	if !valid {
		return false
	}
	lsi, err := lsiPipeline.FitTransform(corpus...)
	if err != nil {
		logrus.WithError(err).Error("failed to process documents")
		return false
	}

	queryVector, err := lsiPipeline.Transform(company.Bag)
	if err != nil {
		logrus.WithError(err).Error("failed to process documents")
		return false
	}

	highestSimilarity := -1.0
	var matched int
	_, docs := lsi.Dims()
	for i := 0; i < docs; i++ {
		similarity := pairwise.CosineSimilarity(queryVector.(mat.ColViewer).ColView(0), lsi.(mat.ColViewer).ColView(i))
		logrus.WithField("match", urls[i]).WithField("cosine", similarity).Debug("cosine")
		if similarity > highestSimilarity {
			matched = i
			highestSimilarity = similarity
		}
	}
	logrus.WithField("match", urls[matched]).WithField("emails", crawl_results[matched].Emails()).WithField("cosine", highestSimilarity).WithField("company", company.Name).Infof("found result")
	return true
}

func main() {
	logrus.SetLevel(logrus.InfoLevel)
	if !FindCompanyDomain(os.Args[1]) {
		logrus.WithField("company_number", os.Args[1]).Infof("failed to find result")
	}
}
