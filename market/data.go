package market

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yzimhao/bookvoo/market/models"
	te "github.com/yzimhao/trading_engine"
	"github.com/yzimhao/utilgo"
)

var (
	rdc    *redis.Client
	config *viper.Viper
	kdh    *kdataHandler

	ChNewKline chan Klinetips
)

type Klinetips struct {
	Symbol string `json:"symbol"`
	Period string `json:"period"`

	OpenAt  int64  `json:"open_at"`  //开盘时间
	Open    string `json:"open"`     //开盘价
	High    string `json:"high"`     // 最高价
	Low     string `json:"low"`      //最低价
	Close   string `json:"close"`    //收盘价(当前K线未结束的即为最新价)
	Volume  string `json:"volume"`   //成交量
	CloseAt int64  `json:"close_at"` // 收盘时间
	Amount  string `json:"amount"`   //成交额

}

func Run(config_path string) {
	initConfig(config_path)
	go handleKLDataService()
	startWeb()
}

func RunWithGinRouter(config_path string, router *gin.Engine) {
	initConfig(config_path)
	setupRouter(router)
	handleKLDataService()
}

func initConfig(config_path string) {
	config = utilgo.ViperInit(config_path)
	rdc = redis.NewClient(&redis.Options{
		Addr:     config.GetString("kline.redis.host"),
		DB:       config.GetInt("kline.redis.db"),
		Password: config.GetString("kline.redis.password"),
	})
	models.InitDbEngine(config)
	ChNewKline = make(chan Klinetips, 1000)
}

func handleKLDataService() {
	config.SetDefault("kline.redis.trade_log_subscribe_key", "list:trade_log")
	kdh = NewKdataHandler(rdc, config.GetStringSlice("kline.interval"))
	//todo 初始化kline的最新缓存
	kdh.RebuildCache()

	// 通过pub/sub方式获取成交记录
	// subTradeLog()
	//通过list获取成交记录
	popTradeLog()
}

func popTradeLog() {
	subcribeKey := config.GetString("kline.redis.trade_log_subscribe_key")
	ctx := context.Background()
	logrus.Infof("subscribe key=%s...", subcribeKey)

	for {
		msg := rdc.BRPop(ctx, time.Duration(30)*time.Second, subcribeKey).Val()
		if len(msg) > 1 {
			kdh.WaitGroupAdd(len(kdh.NeedPeriods()) * 1)
			handleData(msg[1])
		}
	}
}

func handleData(msg string) error {
	var tr te.TradeResult
	err := json.Unmarshal([]byte(msg), &tr)
	if err != nil {
		logrus.Errorf("%s, err: %s", msg, err)
		logrus.Errorf("数据解析出错，请参考 %s", "[]")
		return err
	}

	tl := models.TradeLog{
		Symbol:   strings.ToLower(tr.Symbol),
		At:       time.Unix(tr.TradeTime/1e9, 0),
		Price:    tr.TradePrice.String(),
		Quantity: tr.TradeQuantity.String(),
		Amount:   tr.TradeAmount.String(),
		AskId:    tr.AskOrderId,
		BidId:    tr.BidOrderId,
	}

	kdh.InputTradeLog <- tl
	tl.Save()
	return nil
}