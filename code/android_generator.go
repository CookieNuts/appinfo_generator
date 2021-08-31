package code

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
)

// AndroidGenerator struct
type AndroidGenerator struct {
}

// GetAppInfo return appInfo
func (s *AndroidGenerator) GetAppInfo(bundle string, appID int, publisherID int, adunitID int) (*AppInfo, *Publisher, []*AdUnit, error) {
	baseURL := "https://play.google.com/store/apps/details?id=%s"
	getURL := fmt.Sprintf(baseURL, bundle)
	appInfo := new(AppInfo)

	resp, err := http.Get(getURL)
	if err != nil {
		return nil, nil, nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	appName := doc.Find("h1[itemprop=name]>span").First().Text()
	// replace all Punctuation on appName
	re := regexp.MustCompile("\\p{P}")
	appName = re.ReplaceAllString(appName, "")

	// extract description from html
	descriptionContent, err := doc.Find("div[itemprop=description]>span>div").First().Html()
	if err != nil {
		return nil, nil, nil, err
	}
	descriptionArr := strings.Split(descriptionContent, "\u003cbr/\u003e")
	var description string
	if descriptionArr != nil && len(descriptionArr) > 0 {
		r, _ := regexp.Compile("[.|!|?]")
		arr := r.FindStringIndex(descriptionArr[0])
		if len(arr) > 0 {
			description = descriptionArr[0][:arr[1]]
		}
	}

	publisherName := doc.Find("div:contains('Offered By')+span>div>span").First().Text()
	sellerURL, sellerURLExists := doc.Find("a:contains('Visit website')").First().Attr("href")

	appInfo.AppKey = strings.ReplaceAll(uuid.NewString(), "-", "")
	appInfo.AppSecret = strings.ReplaceAll(uuid.NewString(), "-", "")
	appInfo.Bundle = bundle
	appInfo.Cat = "IAB9"
	appInfo.Description = description
	if sellerURLExists {
		appInfo.Domain = getDomain(sellerURL)
	}
	appInfo.ID = appID
	appInfo.IntegrationType = "OPENRTB"
	appInfo.Name = appName
	appInfo.OctopusAppType = "ORDINARY"
	appInfo.Os = "ANDROID"
	appInfo.PublisherID = publisherID
	appInfo.StoreURL = getURL

	publisher := getPublisher(publisherName, publisherID)

	adUnitList := getAdUnit(appName, appID, adunitID)

	return appInfo, publisher, adUnitList, nil
}
