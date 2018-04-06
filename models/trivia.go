package models

import (
	"sync"
	"net/http"
	"github.com/sirupsen/logrus"
	"github.com/PuerkitoBio/goquery"
	"fmt"
	"strings"
	"strconv"
	"net/url"
)

const (
	//SEARCHBASE = "https://google.co.uk/search"
	SEARCHBASE = "https://www.bing.com/search"
)

type Question struct {
	Text    string
	Options []*Option
}

type Option struct {
	Text        string
	ResultCount int
	Results     []*GoogleResult
}

type SearchResult struct {
	ResultCount int
	Results     []*GoogleResult
}

type GoogleResult struct {
	ResultRank  int
	ResultURL   string
	ResultTitle string
	ResultDesc  string
}

func (q *Question) Solve() *Option {
	var correctAnswer *Option
	var wg sync.WaitGroup

	for _, o := range q.Options {
		wg.Add(1)
		go func(o *Option) {
			defer wg.Done()
			o.Query(q.Text)
		}(o)
	}

	wg.Wait() //Wait for all to be checked

	//Loop through each result and check the result count, highest wins
	for _, o := range q.Options {
		if correctAnswer == nil {
			correctAnswer = o
			continue
		}

		if o.ResultCount > correctAnswer.ResultCount {
			correctAnswer = o
		}
	}

	return correctAnswer
}

//Calculates the probabiltiy of the answer being correct
func (q *Question) Probability(a *Option) float32 {
	totalCount := 0

	for _, o := range q.Options {
		totalCount += o.ResultCount
	}

	logrus.WithField("total_count", totalCount).Debug("Total Results")

	if totalCount <= 0 {
		return 0
	}

	return (float32(a.ResultCount) / float32(totalCount)) * 100
}

//Checks google and returns the amount of results
func (o *Option) Query(question string) {
	lctx := logrus.WithFields(logrus.Fields{
		"option": o.Text,
	})

	lctx.Debug("Checking Option")

	client := newClient()
	req, _ := http.NewRequest("GET", SEARCHBASE, nil)

	q := req.URL.Query()
	q.Add("q", o.QueryString(question))
	q.Add("num", "5") //Limit to 5 results
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		lctx.WithError(err).Error("Unable to get results")
		return
	}

	lctx.WithFields(logrus.Fields{
		"status_code": resp.Status,
		"url":         req.URL.String(),
	}).Debug("Response Code")

	defer resp.Body.Close()
	results, err := parseResp(resp)
	if err != nil {
		logrus.WithError(err).Error("Unable to parse search results")
	}

	//Assign the results
	o.ResultCount = results.ResultCount
	o.Results = results.Results
	lctx.WithFields(logrus.Fields{
		"count": o.ResultCount,
	}).Debug("Option Result")

	for _, r := range results.Results {
		logrus.WithField("title", r.ResultTitle).Debug("Result Title")
	}

}

//Builds the qurery string used to calculate the answer
func (o *Option) QueryString(q string) string {
	return fmt.Sprintf("%s %s%s", q, "intitle:", o.Text)
}

func parseResp(response *http.Response) (*SearchResult, error) {
	res := SearchResult{} //Our container

	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		return nil, err
	}

	countSel := doc.Find("span.sb_count")
	countI := strings.Split(strings.Replace(countSel.Text(), "[", "", -1), " ")

	if len(countI) > 1 {
		res.ResultCount, err = strconv.Atoi(strings.Replace(countI[0], ",", "", -1)) //Store the count of results
		if err != nil {
			logrus.WithError(err).Error("Unable to get result count")
		}
	}

	//Find our top 5 results
	sel := doc.Find("div.g")
	rank := 1
	for i := range sel.Nodes {
		item := sel.Eq(i)
		linkTag := item.Find("a")
		link, _ := linkTag.Attr("href")
		titleTag := item.Find("h3.r")
		descTag := item.Find("span.st")
		desc := descTag.Text()
		title := titleTag.Text()
		link = strings.Trim(link, " ")
		if link != "" && link != "#" {
			res.Results = append(res.Results, &GoogleResult{
				rank,
				link,
				title,
				desc,
			})
			rank += 1
		}
	}

	return &res, err
}

var proxyIndex = 0
var proxySlice = []string{
	"http://217.79.54.68:8080",
	"http://158.232.15.83:80",
	"http://213.136.89.121:80",
}

func newClient() *http.Client {

	return &http.Client{}

	tr := &http.Transport{
		Proxy: http.ProxyURL(nextProxy()),
	}

	return &http.Client{Transport: tr}
}

func nextProxy() *url.URL {

	if proxyIndex > len(proxySlice)-1 {
		proxyIndex = 0
	}

	logrus.WithFields(logrus.Fields{"proxy_index": proxyIndex, "list_len": len(proxySlice)}).Debug("Proxy Setup")

	proxy, err := url.Parse(proxySlice[proxyIndex])
	proxyIndex++

	if err != nil {
		logrus.WithError(err).Error("Unable to get proxy, trying next")
		nextProxy()
	}

	return proxy
}
