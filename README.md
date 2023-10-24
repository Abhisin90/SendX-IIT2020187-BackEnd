# Web Crawler

## It is a backend of an web application which accepts users input for URL and fetches crawled URLs


### Tasks done: 
1. Required
2. Good

## Functionality Overview:

- Accepts a URL to be crawled alongwith isUserPaid status
- Checks for URL in cached URL data
- If available, return result
- Schedules queues depending upon user Paid status
- Retrives crawled data
- Displays a JSON object to user consisting of URLs
- Uses Colly Web Framework for crawling logic
- Uses goroutines alongwith waitGroups
- Uses five goroutiens for paidUsers retrieving faster response
- Two goroutines are used for nonPaidUsers
- Uses mutex lock during writing in crawled data, queue,


# Setup & Usage
Make sure you have Go installed in your local environment by running go in your terminal
Install the Colly framework by using the command go get github.com/gocolly/colly
Install necesssary dependencies by go install
Run the server by go run main.go

# Demo

https://drive.google.com/file/d/14zyUx4h_32uYs6t2zjMycnibYQBSIom7/view
