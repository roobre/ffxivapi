package ffxivapi

import (
	"github.com/PuerkitoBio/goquery"
	"strconv"
	"strings"
)

type SearchResult struct {
	ID     int
	Level  int
	Avatar string
	Lang   string
	Name   string
	World  string
}

func (api *FFXIVAPI) Search(characterName string, world string) ([]SearchResult, error) {
	doc, err := api.lodestone("/lodestone/character/",
		URLParam{"q", characterName},
		URLParam{"worldname", strings.Title(strings.ToLower(world))},
	)
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	doc.Find("a.entry__link").Each(func(i int, sel *goquery.Selection) {
		result := SearchResult{}
		idlink, found := sel.Attr("href")
		if !found {
			return
		}

		parts := strings.Split(strings.Trim(idlink, "/"), "/")
		if len(parts) == 0 {
			return
		}

		result.ID, _ = strconv.Atoi(parts[len(parts)-1])

		image := sel.Find("img").First()
		imagesrc, found := image.Attr("src")
		if !found {
			return
		}

		result.Avatar = imagesrc

		result.Name = sel.Find("p.entry__name").First().Text()
		result.World = sel.Find("p.entry__world").First().Text()
		result.Lang = sel.Find(".entry__chara__lang").First().Text()

		levelText := sel.Find(".entry__chara_info").First().Find("span").First().Text()
		if levelText == "" {
			return
		}
		result.Level, _ = strconv.Atoi(levelText)

		results = append(results, result)
	})

	return results, nil
}
