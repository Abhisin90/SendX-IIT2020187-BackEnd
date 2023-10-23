package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
	"github.com/gocolly/colly"
)

type CachedData struct {
	Urls []string
	TimeStamp int64
}

var cache = make(map[string]CachedData)

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

//  By using different user agent strings, the crawler can mimic different browsers and
// potentially avoid being blocked or rate-limited by websites.
var UserAgents = []string {
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36 OPR/38.0.2220.41",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 13_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.1 Mobile/15E148 Safari/604.1",
}

// The `crawling` function is responsible for performing web crawling on a given URL. It uses the
// `colly` package, which is a powerful web scraping framework for Go.
func crawling(url string) ([]string ,error) {	
	c := colly.NewCollector()
	c.UserAgent = UserAgents[rand.Int() % len(UserAgents)]
	caching := []string{}
	fmt.Println(c.UserAgent)
	c.OnRequest(func(r *colly.Request) { 
		fmt.Println("Visiting", r.URL) 	
	}) 
	c.OnError(func(_ *colly.Response, err error) { 
		log.Println("Something went wrong:", err) 
	}) 
	c.OnResponse(func(r *colly.Response) { 
		fmt.Println("Visited", r.Request.URL) 
		caching = append(caching, r.Request.URL.String())
	}) 	
    c.OnHTML("a[href]", func(e *colly.HTMLElement) { 
        if len(caching) > 100 {
			return 
		}
		e.Request.Visit(e.Attr("href"))
	})

	fmt.Println("Starting crawl at: ", url) 
		
	if err := c.Visit(url); err != nil { 
		if len(caching) > 100 {
			return caching, nil
		}
	    fmt.Println("Error on start of crawl: ", err) 
		return nil,err
	} 
	c.Wait()
	return caching, nil 
}


// The function `crawlurl` is a handler function for an HTTP request that crawls a given URL, checks if
// it is stored in cache, and if not, retries crawling the URL until it succeeds or reaches the maximum
// number of retries.
func crawlurl(w http.ResponseWriter, r *http.Request){
	url := r.FormValue("url")
	retries := 3
	currentTime := time.Now().Unix()
	currentTimeStamp := time.Unix(currentTime,0)
	fmt.Println(currentTime)
	resp, ok := cache[url]
	if ok {
		storedTime := resp.TimeStamp 
		storedTimeStamp := time.Unix(storedTime,0)
		difference := currentTimeStamp.Sub(storedTimeStamp)
		minutes := difference.Minutes()
		if minutes <= 60 {
			fmt.Println("it is stored in cache")
			fmt.Println(resp.Urls)
			return
		}
	}
    // Set the delay between retries.
    delay := 100 * time.Millisecond

    // Retry the function until it succeeds.
    for i := 0; i < retries; i++ {
        caching, err := crawling(url)
		if len(caching) > 100 {
			fmt.Println(caching)
			cache[url] = CachedData{Urls:caching,TimeStamp: time.Now().Unix()}
            break
		}
        if err != nil {
            fmt.Println("Error:", err)
            time.Sleep(delay)
            continue
        }
        // The function succeeded
		break
    }

	for _,data := range cache {
		fmt.Println(data.TimeStamp)
		fmt.Println(data.Urls)
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

