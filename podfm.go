package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

const podcastsFileList = "podcasts.txt"

func main() {
	start := time.Now()

	os.Remove(podcastsFileList)

	ch := make(chan string)
	pages := 33

	for i := 1; i < pages; i++ {
		url := fmt.Sprint("http://brand.podfm.ru/?page=", i)

		go fetch(url, ch)
	}

	for i := 1; i < pages; i++ {
		fmt.Println(<-ch)
	}

	fmt.Printf("%.2fs total elapse\n", time.Since(start).Seconds())
}

func fetch(url string, ch chan<- string) {
	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}

	if resp.StatusCode != 200 {
		ch <- fmt.Sprintf("Response status code: %d", resp.StatusCode)
	} else {
		bodyBytes, err2 := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)

		if err2 != nil {
			ch <- fmt.Sprint(err2)
			return
		}

		var pageIds []int64

		var re = regexp.MustCompile(`(?mi)<a href=\"http:\/\/brand.podfm.ru\/(.*)\/\"><img`)
		for _, match := range re.FindAllStringSubmatch(bodyString, -1) {
			if pageId, err := strconv.ParseInt(match[1], 10, 32); err == nil {
				pageIds = append(pageIds, pageId)
			}
		}

		fmt.Printf("%s: %d\n", url, len(pageIds))

		getLinks(pageIds)
	}

	defer resp.Body.Close()

	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}

	secs := time.Since(start).Seconds()
	ch <- fmt.Sprintf("<- %.2fs %d %s", secs, resp.StatusCode, url)
}

func getLinks(pageIds []int64) {
	fmt.Sprintf("Found podcasts: %d \t", len(pageIds))

	file, err3 := os.OpenFile(podcastsFileList, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err3 != nil {
		panic(err3)
	}
	defer file.Close()

	for _, pageId := range pageIds {
		podcastUrl := "http://brand.podfm.ru/" + fmt.Sprintf("%d", pageId) + "/"
		resp, err := http.Get(podcastUrl)
		if err != nil {
			fmt.Println(err)
			return
		}

		bodyBytes, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			fmt.Println(err2)
			return
		}

		bodyString := string(bodyBytes)

		var re = regexp.MustCompile(`(?im)<a class=\"box_download\" rel=\'no-follow\' href=\"(.*)\" download`)
		for _, match := range re.FindAllStringSubmatch(bodyString, -1) {
			fmt.Println("\t" + podcastUrl + "\t" + match[1])

			if _, err = file.WriteString(match[1] + "\n"); err != nil {
				panic(err)
			}
		}
	}
}
