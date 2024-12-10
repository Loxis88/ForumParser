package crawler

import (
	"github.com/gocolly/colly/v2"
	"os"
	"strings"
)

func CrawlTopics(fileName string) error {
	// Создание нового коллектора
	c := colly.NewCollector(
		colly.AllowedDomains("mirea.org"),
	)

	// Используем map для хранения уникальных ссылок на темы
	temaLinks := make(map[string]struct{})

	// Нормализуем URL, убирая все после первого тире
	normalizeURL := func(url string) string {
		// Убираем протокол и хост
		trimmed := strings.TrimPrefix(url, "http://mirea.org")
		trimmed = strings.TrimPrefix(trimmed, "/")

		// Находим первое вхождение тире и обрезаем до него
		if hyphenIndex := strings.Index(trimmed, "-"); hyphenIndex != -1 {
			trimmed = trimmed[:hyphenIndex]
		}

		return "http://mirea.org/" + trimmed
	}

	// Находим все ссылки на странице, исключая те, которые находятся в div с классом "ipsFilterbar maintitle"
	c.OnHTML("a[href]:not(.ipsFilterbar.maintitle *)", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))

		// Проверяем, начинается ли ссылка с vmirea для дальнейшего краулинга
		if strings.HasPrefix(link, "http://mirea.org/vmirea/") {
			// fmt.Printf("Link found: %q\n", link)
			e.Request.Visit(link)
		}

		// Нормализуем URL
		normalizedLink := normalizeURL(link)

		// Добавляем ссылку на тему в map, что автоматически удаляет дубликаты
		if strings.Contains(normalizedLink, "http://mirea.org/tema/") {
			temaLinks[normalizedLink] = struct{}{}
			// fmt.Printf("Tema link found: %q\n", normalizedLink)
		}
	})

	// Выводим информацию о посещенных URL
	c.OnRequest(func(r *colly.Request) {
		// fmt.Println("Visiting", r.URL.String())
	})

	// Начинаем с главной страницы
	err := c.Visit("http://mirea.org/")
	if err != nil {
		return err
	}

	// Вывод всех найденных уникальных ссылок на темы после завершения краулинга
	// fmt. Println("Tema links found:")

	file, err := os.Create("temaLinks.txt")
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	for link := range temaLinks {
		_, err := file.WriteString(link + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}
