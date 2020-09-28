package ffxivapi

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	FeatureClassJob     = 1 << 1
	FeatureAchievements = 1 << 2
)

// Character models FFXIV character data
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

// ClassJob stores the progress of a character in a given class or job
type ClassJob struct {
	Name  string
	Level int

	Exp     int64
	ExpNext int64
}

// Achievement contains name, ID and unlocking time for a achievements
type Achievement struct {
	ID       int
	Name     string
	Obtained time.Time
}

// urlIdRegex is used to extract IDs from lodestone urls, such as 31688528 in https://eu.finalfantasyxiv.com/lodestone/character/31688528/
var urlIdRegex = regexp.MustCompile(`/(\d+)/?$`)

// Character returns character data given its ID
// Achievements and non-active classes and jobs will be returned if features bitmask contains the respective bits
func (api *FFXIVAPI) Character(id int, features uint) (*Character, error) {
	doc, err := api.lodestone(fmt.Sprintf("/lodestone/character/%d/", id), nil)
	if err != nil {
		return nil, err
	}

	wg := &sync.WaitGroup{}

	character := &Character{ID: id, ParsedAt: time.Now()}

	// Features (achievements and secondary classes and jobs) are queried in parallel
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

	// Query first page of achievements
	doc, err := api.lodestone(fmt.Sprintf("/lodestone/character/%d/achievement/", c.ID), nil)
	if err != nil {
		return err
	}

	// Find index of last page
	lastPageUrl, public := doc.Find(".btn__pager__next--all").First().Attr("href")
	if !public {
		return errors.New(fmt.Sprintf("could not find achievements for %d", c.ID))
	}

	lp := silentAtoi(lastPageUrl[len(lastPageUrl)-1:])

	achvChan := make(chan []Achievement, 8)

	// Parse first page asynchronously
	go parseAchievementPage(doc, achvChan)

	// For next pages, if any, parse asyncrhonously as well
	for i := 2; i <= lp; i++ {
		page := i
		go func() {
			doc, err := api.lodestone(fmt.Sprintf("/lodestone/character/%d/achievement", c.ID), map[string]string{
				"page": fmt.Sprint(page),
			})
			if err != nil {
				log.Fatal(err)
			}
			parseAchievementPage(doc, achvChan)
		}()
	}

	// Collect updates from each page and append them to the character's achievement list
	for i := 1; i <= lp; i++ {
		c.Achievements = append(c.Achievements, <-achvChan...)
	}

	return nil
}

// achNameRegex obtains the achievement name from the flavour text
var achNameRegex = regexp.MustCompile(`achievement "(.+)" earned`)

// achDatetimeRegex obtains the unix timestamp from the js code used by the lodestone to display dates
var achDatetimeRegex = regexp.MustCompile(`ldst_strftime\((\d+), 'YMD'\)`)

// parseAchievementPage pushes to a channel the list of achievements found in a goquery.Document
func parseAchievementPage(doc *goquery.Document, achvChan chan []Achievement) {
	// Preallocate list for 50 achievements (50 per page)
	achievements := make([]Achievement, 0, 50)
	doc.Find(".entry__achievement").Each(func(i int, sel *goquery.Selection) {
		// Find the achievement details link and extract ID from it
		aurl, found := sel.Attr("href")
		if !found {
			return
		}

		matches := urlIdRegex.FindStringSubmatch(aurl)
		if len(matches) <= 1 {
			return
		}

		id, err := strconv.Atoi(matches[1])
		if err != nil {
			return
		}

		a := Achievement{ID: id}

		// Obtain name from flavour text
		name := sel.Find(".entry__activity__txt").Text()
		matches = achNameRegex.FindStringSubmatch(name)
		if len(matches) >= 2 {
			a.Name = matches[1]
		}

		// Decode unlock time from js snippet
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

// classImgMap holds the class name for each class icon used in the Lodestone
// This is hacky and might stop working at any moment, but I have not found any better way to obtain it
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
