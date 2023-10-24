package main

import (
	"fmt"
	"encoding/json"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
	"github.com/gocolly/colly"
	"sync"
)

type CachedData struct {
	Urls []string
	TimeStamp int64
}
type URLQueue struct {
    mutex sync.Mutex
    queue []string
}

func (q *URLQueue) Enqueue(url string) {
    q.mutex.Lock()
    defer q.mutex.Unlock()
    q.queue = append(q.queue, url)
}

func (q *URLQueue) Dequeue() string {
    q.mutex.Lock()
    defer q.mutex.Unlock()

    if len(q.queue) == 0 {
        return ""
    }

    url := q.queue[0]
    q.queue = q.queue[1:]

    return url
}

func (q *URLQueue) close() {
	q.mutex.Lock()
	q.queue= nil
	defer q.mutex.Unlock()
}

var mu sync.Mutex
var cache = make(map[string]CachedData)
var visitedURLs = make(map[string]bool)
var UserAgents = []string {
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36 OPR/38.0.2220.41",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 13_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.1 Mobile/15E148 Safari/604.1",
}

// The function `renderToUser` takes a `http.ResponseWriter` and a slice of strings as input, marshals
// the data into JSON format, sets the response header to indicate JSON content, and writes the JSON
// response to the `http.ResponseWriter`.
func renderToUser(w http.ResponseWriter, data []string) {
	renderData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Failed to fetch rendering data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.Write(renderData)
}

// The function "crawling" performs web crawling on a given URL and returns a list of crawled URLs.
func crawling(url string) ([]string ,error) {	
	c := colly.NewCollector(
		colly.MaxDepth(2),
	)
	c.UserAgent = UserAgents[rand.Int() % len(UserAgents)]
	crawledData := []string{}
	var mutex sync.Mutex

	c.OnError(func(_ *colly.Response, err error) { 
		log.Println("Something went wrong:", err) 
	}) 
	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Vistied this url", r.Request.URL)
	})
    c.OnHTML("a[href]", func(e *colly.HTMLElement) { 
        link := e.Request.AbsoluteURL(e.Attr("href"))
		if !visitedURLs[link] {
			mutex.Lock()
			crawledData = append(crawledData, link)
			visitedURLs[link] = true
			mutex.Unlock()
			e.Request.Visit(link)
		}
	})

	fmt.Println("Starting crawl at: ", url) 
		
	if err := c.Visit(url); err != nil { 
	    fmt.Println("Error on start of crawl: ", err) 
		return nil,err
	} 
	c.Wait()
	return crawledData, nil 
}

// The function retryCrawlWebsite attempts to crawl a website multiple times, with a specified number
// of retries, and returns the crawled data if successful.
func retryCrawlWebsite(url string, retries int) ([]string) {
    for i := 0; i < retries; i++ {
        crawledData,err := crawling(url)
        if err != nil {
			fmt.Println("Try again...",)
			continue
		}
		return crawledData;
    }

    return nil
}

// The `crawlingMain` function crawls a website, caches the crawled data, renders it to the user, and
// prints the number of crawled sites.
func crawlingMain(w http.ResponseWriter, url string) {
	mu.Lock()
	crawledData := retryCrawlWebsite(url,3)
	if crawledData == nil {
		fmt.Println("Error occured")
		return
	} else {
		cache[url] = CachedData{Urls: crawledData,TimeStamp: time.Now().Unix()}
	    renderToUser(w,crawledData)
		// Print the number of crawled sites.
		fmt.Println("Total crawled sites:", len(crawledData))
	}
	
	defer mu.Unlock()
}

// The function `checkInCache` checks if a URL is stored in the cache and if it is not expired.
func checkInCache (w http.ResponseWriter,url string) bool {
	currentTime := time.Now().Unix()
	currentTimeStamp := time.Unix(currentTime,0)
	resp, ok := cache[url]
	if ok {
		storedTime := resp.TimeStamp 
		storedTimeStamp := time.Unix(storedTime,0)
		difference := currentTimeStamp.Sub(storedTimeStamp)
		minutes := difference.Minutes()
		if minutes <= 60 {
			fmt.Println("it is stored in cache")
			renderToUser(w,resp.Urls)
			return true
		}
		return false
	}
	return false
}

// The function `crawlurl` is a concurrent web crawler that crawls URLs based on whether the customer
// is paid or non-paid.
func crawlurl(w http.ResponseWriter, r *http.Request){
	url := r.FormValue("url")

	if checkInCache(w,url) {
		return
	}

    isPaidCustomer := r.FormValue("url")
	isPaid := false
	if isPaidCustomer == "true" {
		isPaid = true
	}

	payingCustomerQueue := URLQueue{}
    nonPayingCustomerQueue := URLQueue{}

    var wg1 sync.WaitGroup
	var wg2 sync.WaitGroup

    // Add the initial URL to the queue.

    if isPaid {
        payingCustomerQueue.Enqueue(url)
    } else {
        nonPayingCustomerQueue.Enqueue(url)
    }

	for i := 0; i < 5; i++ {
		wg1.Add(1)
		go func() {
			defer wg1.Done()

			for {
				if len(payingCustomerQueue.queue) > 0 {
					url := payingCustomerQueue.Dequeue()
					crawlingMain(w,url)
				} else {
					payingCustomerQueue.close()
					return
				}
			}
		}()
	}

	for i := 0; i < 2; i++ {
		wg1.Add(1)
		go func() {
			defer wg2.Done()

			for {
				// First check if there are any URLs in the paying customer queue.
				if len(payingCustomerQueue.queue) > 0 {
					continue
				}
				// Pick a URL from the non-paying customer queue.
				if len(nonPayingCustomerQueue.queue) > 0 {
					url := nonPayingCustomerQueue.Dequeue()
					crawlingMain(w,url)
				} else {
					nonPayingCustomerQueue.close()
					return
				}
			}
			
		}()
	}

    wg1.Wait()
    wg2.Wait()
}

// The function "home" parses and executes an HTML template file named "index.html" and writes the
// output to the http.ResponseWriter.
func home(w http.ResponseWriter, r *http.Request){
	var fileName = "index.html"
	t,err := template.ParseFiles(fileName)
	if err != nil {
		fmt.Println("Error occured in parsing file", err)
		return
	}
	err = t.ExecuteTemplate(w,fileName,nil)
	if err != nil {
		fmt.Println("Error occured during execution of file",err)
		return
	}
}

// The handler function routes incoming HTTP requests to different 
// functions based on the requested URL path.
func handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/home":
		home(w,r)
	case "/crawl-url":
		crawlurl(w,r)
	}
}

// The main function sets up a basic HTTP server that listens on port 8080 and handles requests by
// calling the "handler" function.
func main(){
	http.HandleFunc("/",handler)
	http.ListenAndServe(":8080",nil)
}