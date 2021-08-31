package code

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// IOSGenerator struct
type IOSGenerator struct {
}

// IOSResult struct
type iOSResult struct {
	ResultCount int           `json:"resultCount,omitempty"`
	Results     []*iOSAppInfo `json:"results,omitempty"`
}

// IOSAppInfo struct
type iOSAppInfo struct {
	TrackID      string `json:"trackId,omitempty"`
	TrackName    string `json:"trackName,omitempty"`
	TrackViewURL string `json:"trackViewUrl,omitempty"`
	SellerURL    string `json:"sellerUrl,omitempty"`
	SellerName   string `json:"sellerName,omitempty"`
	Description  string `json:"description,omitempty"`
}

// GetAppInfo return appInfo
func (g *IOSGenerator) GetAppInfo(bundle string, appID int, publisherID int, adunitID int) (*AppInfo, *Publisher, []*AdUnit, error) {
	baseURL := "https://itunes.apple.com/lookup?id=%s"
	getURL := fmt.Sprintf(baseURL, bundle)
	var result iOSResult
	appInfo := new(AppInfo)

	resp, err := http.Get(getURL)
	if err != nil {
		return nil, nil, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, nil, err
	}
	json.Unmarshal(body, &result)
	iosApp := result.Results[0]

	appInfo.AppKey = strings.ReplaceAll(uuid.NewString(), "-", "")
	appInfo.AppSecret = strings.ReplaceAll(uuid.NewString(), "-", "")
	appInfo.Bundle = bundle
	appInfo.Cat = "IAB9"

	description := ""
	if iosApp.Description != "" {
		r, _ := regexp.Compile("[.|!|?]")
		arr := r.FindStringIndex(iosApp.Description)
		if len(arr) > 0 {
			description = iosApp.Description[:arr[1]]
		}
	}
	appInfo.Description = description
	appInfo.Domain = getDomain(iosApp.SellerURL)
	appInfo.ID = appID
	appInfo.IntegrationType = "OPENRTB"
	appInfo.Name = iosApp.TrackName
	appInfo.OctopusAppType = "ORDINARY"
	appInfo.Os = "IOS"
	appInfo.PublisherID = publisherID
	appInfo.StoreURL = iosApp.TrackViewURL

	publisher := getPublisher(iosApp.SellerName, publisherID)

	adUnitList := getAdUnit(iosApp.TrackName, appID, adunitID)

	return appInfo, publisher, adUnitList, nil
}
