package FilesCrawler

import (
	"encoding/csv"
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"os"
	"strconv"
	"strings"
)

type File struct {
	url    string
	author string
	date   string
	size   string
}

func GetLastId() string {
	c := colly.NewCollector()
	var link string
	c.OnHTML("div#board_stats a.value:not(.url)", func(e *colly.HTMLElement) {
		link = e.Attr("href")
	})
	err := c.Visit("http://mirea.org/files/")
	if err != nil || link == "" {
		log.Fatalf("%v", err)
	}
	startStr := "http://mirea.org/files/file/"
	endStr := "-"

	startIndex := strings.Index(link, startStr) + len(startStr)
	endIndex := strings.Index(link, endStr)

	number := link[startIndex:endIndex]

	return number
}

func urlBuilder(id string) string {
	return "http://mirea.org/files/file/" + id
}

func GetFiles(lastId int, fileName string) map[int]File {
	file, error := os.Create(fileName)
	defer file.Close()
	if error != nil {
		log.Fatal(error)
	}

	writer := csv.NewWriter(file)
	writer.Comma = '\t' // разделитель
	defer writer.Flush()

	errWr := writer.Write([]string{"id", "url", "author", "date", "size"})
	if errWr != nil {
		log.Fatalf("Error writing record to csv: %v", errWr)
		return nil
	}

	var fileArray = make(map[int]File)

	for i := 1; i <= lastId; i++ {
		c := colly.NewCollector()
		processed := false
		link := urlBuilder(strconv.Itoa(i))
		c.OnHTML("div.ipsBox_container.ipsPad", func(e *colly.HTMLElement) {
			if processed != true {
				ipsBox := e.DOM.First()

				newFile := File{
					url:    link,
					author: ipsBox.Find("div.ipsPad#submitter_info span[itemprop='name']").Text(),
					date:   ipsBox.Find("div.ipsPad.ipsSideBlock li:nth-child(1)").Contents().Not("*:not(html)").Text(),
					size:   ipsBox.Find("div.ipsPad.ipsSideBlock li:nth-child(3)").Contents().Not("*:not(html)").Text(),
				}

				// Добавляем новый файл в массив
				newFile.url = strings.Split(newFile.url, "\n")[0]

				fileArray[i] = newFile
				fmt.Println(fmt.Sprintf("Url: %s", fileArray[i].url))
				processed = true
				errWr := writer.Write([]string{strconv.Itoa(i), fileArray[i].url, fileArray[i].author, fileArray[i].date, newFile.size})
				if errWr != nil {
					log.Fatalf("Error writing record to csv: %v", errWr)
				}
			}

		})

		c.Visit(link)
	}
	return fileArray
}
