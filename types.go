package storeClient

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	httpClient "github.com/stdevHsequeda/CubanProductFinder/client"
	"time"
)

type StoreClient struct {
	pool   *Pool // Pool of workers
	stores []Store
	client *httpClient.Client // Client
	cache  *Cache             // Cache data
}

type Store struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	// Address        string `json:"address"`
	Province string `json:"province"`
	Online   bool   `json:"online"`
	// PickUpOnStore  bool   `json:"pickUpOnStore"`
	// HomeDelivery   bool   `json:"homeDelivery"`
	// FreezeDelivery string `json:"freezeDelivery"`
	// DeliveryTime   string `json:"deliveryTime"`
	// Cost           string `json:"cost"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Url   string `json:"url"`
	// Cadena         string `json:"cadena"`
}

type W struct {
	ctx  context.Context
	task func(ctx context.Context)
}

func (w *W) GetArgs() context.Context {
	return w.ctx
}

func (w *W) Task(ctx context.Context) {
	section, ok := ctx.Value("section").(TuEnvioSection)
	if !ok {
		fmt.Println("ERROR")
		return
	}

	sc, ok := ctx.Value("sc").(*StoreClient)
	if !ok {
		fmt.Println("ERROR")
		return
	}
	list, err := sc.getProductsFromSection(section)

	for i := range list {
		err = sc.cache.AddProduct(&list[i])
		if err != nil {
			logrus.Warn(err)
			continue
		}
	}

	if err != nil {
		logrus.Warn(err)
		return
	}
}

type TuEnvioSection struct {
	Name      string `css:"div ul li a"`
	Url       string `css:"div ul li a" extract:"attr" attr:"href"`
	Parent    string
	Store     *Store
	Priority  int
	ReadyTime time.Time
}

type TuEnvioProduct struct {
	Name    string `css:".thumbTitle",redis:"name"`
	Price   string `css:".thumbPrice",redis:"price"`
	Link    string `css:".thumbnail a" extract:"attr" attr:"href",redis:"link"`
	Section *TuEnvioSection
}
