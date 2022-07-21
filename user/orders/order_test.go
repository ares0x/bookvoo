package orders

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/yzimhao/bookvoo/core/base"

	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"xorm.io/xorm"
)

var (
	test_symbol = "ethusd"
)

func init() {
	driver := "mysql"
	dsn := "root:root@tcp(localhost:13306)/test?charset=utf8&loc=Local"

	logrus.Infof("dsn: %s", dsn)

	conn, err := xorm.NewEngine(driver, dsn)
	if err != nil {
		logrus.Panic(err)
	}
	db_engine = conn
	db_engine.ShowSQL(true)

	table := TradeOrder{TradeSymbol: test_symbol}
	db_engine.DropTables(
		new(UnfinishedOrder),
		table,
	)

	SetDbEngine(db_engine)
	base.SetDbEngine(db_engine)
}

func Test_main(t *testing.T) {
	db := db_engine.NewSession()
	defer db.Close()

	Convey("限价买单", t, func() {
		order, err := limit_order(1, "ethusd", OrderSideBid, "1.00", "2")
		So(err, ShouldBeNil)
		So(order.OrderId, ShouldStartWith, "B")
	})

	Convey("限价卖单", t, func() {
		order, err := limit_order(1, "ethusd", OrderSideAsk, "1.00", "2")
		So(err, ShouldBeNil)
		So(order.OrderId, ShouldStartWith, "A")
	})
}
