package ffxivapi

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	FeatureClassJob     = 1 << 1
	FeatureAchievements = 1 << 2
)

type Character struct {
	ParsedAt time.Time

	ID    int
	World string

	Avatar   string
	Portrait string

	Name    string
	Nameday string
	City    string

	GC struct {
		Name string
		Rank string
	}

	FC struct {
		ID   string
		Name string
	}

	Achievements []Achievement
	ClassJobs    []ClassJob
}

type ClassJob struct {
	Name  string
	Level int

	Exp     int64
	ExpNext int64
}

type Achievement struct {
	ID       int
	Name     string
	Obtained time.Time
}

var urlIdRegex = regexp.MustCompile(`/(\d+)/?$`)

func (api *FFXIVAPI) Character(id int, features uint) (*Character, error) {
	doc, err := api.lodestone(fmt.Sprintf("/lodestone/character/%d/", id))
	if err != nil {
		return nil, err
	}

	wg := &sync.WaitGroup{}

	character := &Character{ID: id, ParsedAt: time.Now()}

	if features&FeatureClassJob != 0 {
		wg.Add(1)
		go api.parseClassJob(character, wg)
	}
	if features&FeatureAchievements != 0 {
		wg.Add(1)
		go api.parseAchievements(character, wg)
	}

	character.Name = doc.Find(".frame__chara__name").First().Text()
	character.World = doc.Find(".frame__chara__world").First().Text()

	character.ClassJobs = append(character.ClassJobs, ClassJob{
		Name:  classImgMap[doc.Find(".character__class_icon > img").First().AttrOr("src", "")],
		Level: silentAtoi(strings.ReplaceAll(strings.TrimSpace(doc.Find(".character__class__data > p").Text()), "LEVEL ", "")),
	})

	character.Avatar = doc.Find(".frame__chara__face > img").First().AttrOr("src", "")
	character.Portrait = doc.Find(".character__detail__image > a > img").First().AttrOr("src", "")

	character.Nameday = doc.Find(".character-block__birth").First().Text()

	details := doc.Find(".character__profile__data__detail").Children()
	character.City = details.Eq(2).Find(".character-block__name").Text()

	gc := strings.Split(details.Eq(3).Find(".character-block__name").Text(), " / ")
	if len(gc) >= 2 {
		character.GC.Name = gc[0]
		character.GC.Rank = gc[1]
	}

	fc := doc.Find(".character__freecompany__name").Find("a").First()
	matches := urlIdRegex.FindStringSubmatch(fc.AttrOr("href", ""))
	if len(matches) >= 2 {
		character.FC.Name = fc.Text()
		character.FC.ID = matches[1]
	}

	wg.Wait()
	return character, nil
}

func (api *FFXIVAPI) parseClassJob(c *Character, wg *sync.WaitGroup) error {
	defer wg.Done()
	return nil
}

func (api *FFXIVAPI) parseAchievements(c *Character, wg *sync.WaitGroup) error {
	defer wg.Done()

	doc, err := api.lodestone(fmt.Sprintf("/lodestone/character/%d/achievement/", c.ID))
	if err != nil {
		return err
	}

	lastPageUrl, public := doc.Find(".btn__pager__next--all").First().Attr("href")
	if !public {
		return errors.New(fmt.Sprintf("could not find achievements for %d", c.ID))
	}

	lp := silentAtoi(lastPageUrl[len(lastPageUrl)-1:])

	achvChan := make(chan []Achievement, 8)

	go parseAchievementPage(doc, achvChan)

	for i := 2; i <= lp; i++ {
		page := i
		go func() {
			doc, err := api.lodestone(fmt.Sprintf("/lodestone/character/%d/achievement/?page=%d", c.ID, page))
			if err != nil {
				log.Fatal(err)
			}
			parseAchievementPage(doc, achvChan)
		}()
	}

	for i := 1; i <= lp; i++ {
		c.Achievements = append(c.Achievements, <-achvChan...)
	}

	return nil
}

var achNameRegex = regexp.MustCompile(`achievement "(.+)" earned`)
var achDatetimeRegex = regexp.MustCompile(`ldst_strftime\((\d+), 'YMD'\)`)

func parseAchievementPage(doc *goquery.Document, achvChan chan []Achievement) {
	achievements := make([]Achievement, 0, 50)
	doc.Find(".entry__achievement").Each(func(i int, sel *goquery.Selection) {
		aurl, found := sel.Attr("href")
		if !found {
			return
		}

		matches := urlIdRegex.FindStringSubmatch(aurl)
		if len(matches) <= 1 {
			return
		}
		id := silentAtoi(matches[1])

		a := Achievement{ID: id}
		name := sel.Find(".entry__activity__txt").Text()
		matches = achNameRegex.FindStringSubmatch(name)
		if len(matches) >= 2 {
			a.Name = matches[1]
		}

		datescript := sel.Find("script").First().Text()
		matches = achDatetimeRegex.FindStringSubmatch(datescript)
		if len(matches) >= 2 {
			ts := silentAtoi(matches[1])
			a.Obtained = time.Unix(int64(ts), 0)
		}

		achievements = append(achievements, a)
	})

	achvChan <- achievements
}

var classImgMap = map[string]string{
	"https://img.finalfantasyxiv.com/lds/h/U/F5JzG9RPIKFSogtaKNBk455aYA.png": "Gladiator",
	"https://img.finalfantasyxiv.com/lds/h/E/d0Tx-vhnsMYfYpGe9MvslemEfg.png": "Paladin",
	"https://img.finalfantasyxiv.com/lds/h/y/A3UhbjZvDeN3tf_6nJ85VP0RY0.png": "Warrior",
	"https://img.finalfantasyxiv.com/lds/h/N/St9rjDJB3xNKGYg-vwooZ4j6CM.png": "Marauder",
	"https://img.finalfantasyxiv.com/lds/h/l/5CZEvDOMYMyVn2td9LZigsgw9s.png": "Dark Knight",
	"https://img.finalfantasyxiv.com/lds/h/8/hg8ofSSOKzqng290No55trV4mI.png": "Gunbreaker",
	"https://img.finalfantasyxiv.com/lds/h/V/iW7IBKQ7oglB9jmbn6LwdZXkWw.png": "Pugilist",
	"https://img.finalfantasyxiv.com/lds/h/K/HW6tKOg4SOJbL8Z20GnsAWNjjM.png": "Monk",
	"https://img.finalfantasyxiv.com/lds/h/k/tYTpoSwFLuGYGDJMff8GEFuDQs.png": "Lancer",
	"https://img.finalfantasyxiv.com/lds/h/m/gX4OgBIHw68UcMU79P7LYCpldA.png": "Dragoon",
	"https://img.finalfantasyxiv.com/lds/h/y/wdwVVcptybfgSruoh8R344y_GA.png": "Rogue",
	"https://img.finalfantasyxiv.com/lds/h/0/Fso5hanZVEEAaZ7OGWJsXpf3jw.png": "Ninja",
	"https://img.finalfantasyxiv.com/lds/h/m/KndG72XtCFwaq1I1iqwcmO_0zc.png": "Samurai",
	"https://img.finalfantasyxiv.com/lds/h/s/gl62VOTBJrm7D_BmAZITngUEM8.png": "Conjurer",
	"https://img.finalfantasyxiv.com/lds/h/7/i20QvSPcSQTybykLZDbQCgPwMw.png": "White Mage",
	"https://img.finalfantasyxiv.com/lds/h/7/WdFey0jyHn9Nnt1Qnm-J3yTg5s.png": "Scholar",
	"https://img.finalfantasyxiv.com/lds/h/1/erCgjnMSiab4LiHpWxVc-tXAqk.png": "Astrologian",
	"https://img.finalfantasyxiv.com/lds/h/Q/ZpqEJWYHj9SvHGuV9cIyRNnIkk.png": "Archer",
	"https://img.finalfantasyxiv.com/lds/h/F/KWI-9P3RX_Ojjn_mwCS2N0-3TI.png": "Bard",
	"https://img.finalfantasyxiv.com/lds/h/E/vmtbIlf6Uv8rVp2YFCWA25X0dc.png": "Machinist",
	"https://img.finalfantasyxiv.com/lds/h/t/HK0jQ1y7YV9qm30cxGOVev6Cck.png": "Dancer",
	"https://img.finalfantasyxiv.com/lds/h/4/IM3PoP6p06GqEyReygdhZNh7fU.png": "Thaumaturge",
	"https://img.finalfantasyxiv.com/lds/h/P/V01m8YRBYcIs5vgbRtpDiqltSE.png": "Black Mage",
	"https://img.finalfantasyxiv.com/lds/h/e/VYP1LKTDpt8uJVvUT7OKrXNL9E.png": "Arcanist",
	"https://img.finalfantasyxiv.com/lds/h/h/4ghjpyyuNelzw1Bl0sM_PBA_FE.png": "Summoner",
	"https://img.finalfantasyxiv.com/lds/h/q/s3MlLUKmRAHy0pH57PnFStHmIw.png": "Red Mage",
	"https://img.finalfantasyxiv.com/lds/h/p/jdV3RRKtWzgo226CC09vjen5sk.png": "Blue Mage",
	"https://img.finalfantasyxiv.com/lds/h/v/YCN6F-xiXf03Ts3pXoBihh2OBk.png": "Carpenter",
	"https://img.finalfantasyxiv.com/lds/h/5/EEHVV5cIPkOZ6v5ALaoN5XSVRU.png": "Blacksmith",
	"https://img.finalfantasyxiv.com/lds/h/G/Rq5wcK3IPEaAB8N-T9l6tBPxCY.png": "Armorer",
	"https://img.finalfantasyxiv.com/lds/h/L/LbEjgw0cwO_2gQSmhta9z03pjM.png": "Goldsmith",
	"https://img.finalfantasyxiv.com/lds/h/b/ACAcQe3hWFxbWRVPqxKj_MzDiY.png": "Leatherworker",
	"https://img.finalfantasyxiv.com/lds/h/X/E69jrsOMGFvFpCX87F5wqgT_Vo.png": "Weaver",
	"https://img.finalfantasyxiv.com/lds/h/C/bBVQ9IFeXqjEdpuIxmKvSkqalE.png": "Alchemist",
	"https://img.finalfantasyxiv.com/lds/h/m/1kMI2v_KEVgo30RFvdFCyySkFo.png": "Culinarian",
	"https://img.finalfantasyxiv.com/lds/h/A/aM2Dd6Vo4HW_UGasK7tLuZ6fu4.png": "Miner",
	"https://img.finalfantasyxiv.com/lds/h/I/jGRnjIlwWridqM-mIPNew6bhHM.png": "Botanist",
	"https://img.finalfantasyxiv.com/lds/h/x/B4Azydbn7Prubxt7OL9p1LZXZ0.png": "Fisher",
}
