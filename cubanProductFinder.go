package storeClient

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	httpClient "github.com/stdevHsequeda/CubanProductFinder/client"
	html "github.com/zlepper/encoding-html"
	"net/http"
	"runtime"
	"strings"
	"time"
)

var sectionList []TuEnvioSection

func NewStoreClient() *StoreClient {
	// Init all attribs
	return &StoreClient{client: httpClient.NewClient(), pool: NewPool(runtime.NumCPU()), cache: NewCache()}
}

func (sc *StoreClient) Start() {
	logrus.Info("Starting client...")

	defer sc.pool.Shutdown()

	storeList, err := sc.getStoreList()
	if err != nil {
		logrus.Fatal(err)
	}

	sectionList = make([]TuEnvioSection, 0)
	for i := range storeList {
		list, err := sc.getSectionsFromStore(storeList[i])
		if err != nil {
			logrus.Fatal(err)
		}

		sectionList = append(sectionList, list...)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(sectionListInternal []TuEnvioSection, sc *StoreClient, ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for i, section := range sectionListInternal {
					if i == 15 {
						break
					}

					if sectionListInternal[i].ReadyTime.Before(time.Now()) {
						sectionListInternal[i].ReadyTime = time.Now().Add(1 * time.Minute)
						sc.pool.Run(
							&W{
								ctx: context.WithValue(context.WithValue(ctx, "sc", sc), "section", section),
							},
						)
					}
				}
			}
		}
	}(sectionList, sc, ctx)
}

func (sc *StoreClient) SearchProduct(pattern string) ([]TuEnvioProduct, error) {
	_, list, err := sc.cache.SearchProducts(pattern)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (sc *StoreClient) getStoreList() ([]Store, error) {
	logrus.Info("Getting list of stores")

	req, err := http.NewRequest(http.MethodGet, "https://www.tuenvio.cu/stores.json", nil)
	if err != nil {
		logrus.Warn(err)
		return nil, err

	}

	resp, err := sc.client.CallRetryable(req)
	if err != nil {
		logrus.Warn(err)
		return nil, err
	}

	var storeList = make([]Store, 0)
	err = json.NewDecoder(resp).Decode(&storeList)
	if err != nil {
		logrus.Warn(err)
		return nil, err
	}
	return storeList, nil
}

func (sc *StoreClient) getSectionsFromStore(store Store) ([]TuEnvioSection, error) {
	logrus.Infof("Getting sections from store: %s", store.Name)

	req, err := http.NewRequest("GET", store.Url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := sc.client.CallRetryable(req)
	if err != nil {
		return nil, err
	}

	var htmlContent = struct {
		Content []TuEnvioSection `css:".nav li"`
	}{}

	err = html.NewDecoder(resp).Decode(&htmlContent)
	if err != nil {
		return nil, err
	}

	var result = make([]TuEnvioSection, 0)

	var currentParent string
	for _, section := range htmlContent.Content {
		switch section.Url {
		case "default.aspx":
			continue
		case "#":
			currentParent = section.Name
			continue
		default:
			result = append(result, TuEnvioSection{
				Name:      section.Name,
				Url:       fmt.Sprintf("%s/%s", strings.TrimSpace(store.Url), strings.TrimSpace(section.Url)),
				Parent:    currentParent,
				Store:     &store,
				ReadyTime: time.Now(),
			})
		}
	}

	return result, nil
}

func (sc *StoreClient) getProductsFromSection(section TuEnvioSection) ([]TuEnvioProduct, error) {
	logrus.Infof("Getting products from %s in %s", section.Name, section.Store.Name)
	req, err := http.NewRequest("GET", section.Url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := sc.client.CallRetryable(req)
	if err != nil {
		return nil, err
	}

	var list = struct {
		Content []TuEnvioProduct `css:".hProductItems .clearfix"`
	}{}

	err = html.NewDecoder(resp).Decode(&list)

	var result = make([]TuEnvioProduct, 0)
	for _, product := range list.Content {
		result = append(result, TuEnvioProduct{
			Name:    strings.TrimSpace(product.Name),
			Price:   strings.TrimSpace(product.Price),
			Link:    fmt.Sprintf("%s/%s", section.Store.Url, strings.TrimSpace(product.Link)),
			Section: &section,
		})
	}

	return result, nil
}
