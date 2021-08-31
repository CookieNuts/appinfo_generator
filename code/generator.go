package code

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TopListDomainMap key -> domain level, value -> domain list
var TopListDomainMap map[int][]string

// GeneralAdSizeList general ad size List
var GeneralAdSizeList []string

func init() {
	TopListDomainMap = make(map[int][]string)
	GeneralAdSizeList = []string{"300x250", "320x50", "320x480", "480x320", "728x90", "768x1024", "1024x768"}
	file, err := os.Open("data/public_suffix_list.dat")
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer file.Close()

	br := bufio.NewReader(file)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}

		row := strings.Trim(string(a), " ")
		if row == "" || strings.HasPrefix(row, "//") {
			continue
		}

		if strings.HasPrefix(row, "*.") {
			row = strings.Replace(row, "*.", "", 1)
		}

		domainLevel := strings.Count(row, ".") + 1
		var domainList []string
		if TopListDomainMap[domainLevel] == nil {
			domainList = make([]string, 0)
			TopListDomainMap[domainLevel] = domainList
		} else {
			domainList = TopListDomainMap[domainLevel]
		}

		domainList = append(domainList, row)
		TopListDomainMap[domainLevel] = domainList
	}
}

// Generator interface
type Generator interface {
	// GetAppInfo return appInfo
	GetAppInfo(bundle string, appID int, publisherID int, adunitID int) (*AppInfo, *Publisher, []*AdUnit, error)
}

// GenerateAppInfo result struct
type GenerateAppInfo struct {
	AppInfo   *AppInfo   `json:"appInfo,omitempty"`
	Publisher *Publisher `json:"publisher,omitempty"`
	AdUnit    []*AdUnit  `json:"adUnit,omitempty"`
}

// AppInfo struct
type AppInfo struct {
	ID              int    `json:"id"`
	PublisherID     int    `json:"publisherId"`
	AppKey          string `json:"appKey"`
	AppSecret       string `json:"appSecret"`
	Os              string `json:"os"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Bundle          string `json:"bundle"`
	Domain          string `json:"domain"`
	StoreURL        string `json:"storeUrl"`
	Cat             string `json:"cat"`
	OctopusAppType  string `json:"type"`
	IntegrationType string `json:"integrationType"`
}

// AdUnit struct
type AdUnit struct {
	ID         int     `json:"id"`
	AppID      int     `json:"appId"`
	Name       string  `json:"name"`
	AdType     string  `json:"type"`
	Size       string  `json:"size"`
	FloorPrice float32 `json:"floorPrice"`
}

// Publisher struct
type Publisher struct {
	ID            int    `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	InternalName  string `json:"internalName,omitempty"`
	AppKey        string `json:"appKey,omitempty"`
	PublisherType string `json:"type,omitempty"`
}

// GeneratAppInfo generate all appinfo from Platform
func GeneratAppInfo(bundle string, appID int, publisherID int, adunitID int) string {
	generateAppInfo := new(GenerateAppInfo)
	var generator Generator
	r, _ := regexp.Compile(`^[0-9]+`)

	if bundle == "" {
		log.Fatalln(errors.New("Bundle is null error"))
		return ""
	} else if r.MatchString(bundle) {
		generator = new(IOSGenerator)
	} else {
		generator = new(AndroidGenerator)
	}

	appInfo, publisher, adunit, err := generator.GetAppInfo(bundle, appID, publisherID, adunitID)
	if err != nil {
		log.Fatalln(err)
		return ""
	}

	generateAppInfo.AdUnit = adunit
	generateAppInfo.AppInfo = appInfo
	generateAppInfo.Publisher = publisher

	resultByte, err := json.Marshal(generateAppInfo)
	if err != nil {
		log.Fatalln(err)
		return ""
	}

	return string(resultByte)
}

// GetDomain from URL
func getDomain(domainURL string) string {
	start := time.Now()

	result := ""
	u, err := url.Parse(domainURL)
	if err != nil {
		log.Fatal(err)
		return result
	}

	hostname := u.Hostname()
	domainLevel := strings.Count(hostname, ".")
	for i := domainLevel; i > 0; i-- {
		tmpList := TopListDomainMap[i]
		if tmpList == nil || len(tmpList) == 0 {
			continue
		} else {
			for _, v := range tmpList {
				if strings.HasSuffix(hostname, v) {
					tmpArr := strings.Split(hostname, ".")
					result = strings.Join(tmpArr[len(tmpArr)-(i+1):], ".")
					break
				}
			}

			if result != "" {
				break
			}
		}
	}

	elapsed := time.Since(start)
	log.Println("method: [GetDomain] cost", elapsed)
	log.Printf("Url:[%s] Format Domain result:[%s]", domainURL, result)

	return result
}

// getAdUnit return adunit slice
func getAdUnit(appName string, appID int, adunitID int) []*AdUnit {
	result := make([]*AdUnit, 0)

	for i, v := range GeneralAdSizeList {
		adunit := new(AdUnit)
		adunit.AdType = "BANNER"
		adunit.AppID = appID
		adunit.FloorPrice = 0.1
		adunit.ID = adunitID + i
		adunit.Size = v
		adunitName := strings.ReplaceAll(strings.Title(appName), " ", "") + "_" + adunit.AdType + "_" + adunit.Size
		adunit.Name = adunitName

		result = append(result, adunit)
	}

	return result
}

// getPublisher return publisher
func getPublisher(publisherName string, publisherID int) *Publisher {
	publisher := new(Publisher)
	publisher.AppKey = strings.ReplaceAll(uuid.NewString(), "-", "")
	publisher.ID = publisherID
	publisher.Name = strings.Title(publisherName)
	publisher.InternalName = strings.ReplaceAll(strings.ToUpper(publisherName), " ", "_")
	publisher.PublisherType = "DEV"

	return publisher
}
