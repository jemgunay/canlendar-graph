package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jemgunay/canlendar-graph/api"
	"github.com/jemgunay/canlendar-graph/config"
	"github.com/jemgunay/canlendar-graph/storage"
)

func main() {
	conf := config.New()
	apiHandlers := api.New(&demoStore{}, nil)

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/query", apiHandlers.Query).Methods(http.MethodGet)

	// HTTP file server
	staticFileHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("../../static/")))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/static/"
		staticFileHandler.ServeHTTP(w, r)
	})
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// start HTTP server
	log.Printf("starting demo HTTP server on port %d", conf.Port)
	err := http.ListenAndServe(":"+strconv.Itoa(conf.Port), router)
	log.Printf("demo HTTP server shut down: %s", err)
}

var _ storage.Storer = (*demoStore)(nil)

type demoStore struct{}

func (d demoStore) Query(_ context.Context, options ...storage.QueryOption) ([]storage.Plot, error) {
	queryOpts, err := storage.NewQuery(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	switch queryOpts.Aggregation {
	case storage.Day:
		return dayPlotData(), nil
	case storage.Week:
		return weekPlotData(), nil
	case storage.Month:
		return monthPlotData(), nil
	case storage.Year:
		return yearPlotData(), nil
	default:
		return nil, errors.New("unsupported aggregation provided")
	}
}

func (d demoStore) ReadLastTimestamp(_ context.Context) (time.Time, error) {
	return time.Now().UTC(), nil
}

func (d demoStore) ReadFirstTimestamp(_ context.Context) (time.Time, error) {
	return time.Time{}, nil
}

func (d demoStore) Store(_ context.Context, _ ...storage.Record) error {
	return nil
}

func dayPlotData() []storage.Plot {
	return []storage.Plot{{X: 1652140800000, Y: 0}, {X: 1652227200000, Y: 0}, {X: 1652400000000, Y: 0}, {X: 1652486400000, Y: 2}, {X: 1652659200000, Y: 0}, {X: 1652745600000, Y: 0}, {X: 1652918400000, Y: 0}, {X: 1653004800000, Y: 0}, {X: 1653177600000, Y: 7}, {X: 1653264000000, Y: 1}, {X: 1653436800000, Y: 0}, {X: 1653523200000, Y: 0}, {X: 1653696000000, Y: 5}, {X: 1653782400000, Y: 7}, {X: 1653955200000, Y: 0}, {X: 1654041600000, Y: 1}, {X: 1654214400000, Y: 7}, {X: 1654300800000, Y: 5}, {X: 1654473600000, Y: 0}, {X: 1654560000000, Y: 6}, {X: 1654732800000, Y: 0}, {X: 1654819200000, Y: 3}, {X: 1654992000000, Y: 0}, {X: 1655078400000, Y: 0}, {X: 1655251200000, Y: 0}, {X: 1655337600000, Y: 0}, {X: 1655510400000, Y: 7}, {X: 1655596800000, Y: 6}, {X: 1655769600000, Y: 1}, {X: 1655856000000, Y: 2}, {X: 1656028800000, Y: 0}, {X: 1656115200000, Y: 2}, {X: 1656288000000, Y: 0}, {X: 1656374400000, Y: 0}, {X: 1656547200000, Y: 6}, {X: 1656633600000, Y: 6}, {X: 1656806400000, Y: 0}, {X: 1656892800000, Y: 0}, {X: 1657065600000, Y: 0}, {X: 1657152000000, Y: 0}, {X: 1657324800000, Y: 6}, {X: 1657411200000, Y: 5}, {X: 1657584000000, Y: 0}, {X: 1657670400000, Y: 0}, {X: 1657843200000, Y: 0}, {X: 1657929600000, Y: 5}, {X: 1658102400000, Y: 1}, {X: 1658188800000, Y: 0}, {X: 1658361600000, Y: 0}, {X: 1658448000000, Y: 0}, {X: 1658620800000, Y: 5}, {X: 1658707200000, Y: 0}, {X: 1658880000000, Y: 0}, {X: 1658966400000, Y: 0}, {X: 1659139200000, Y: 3}, {X: 1659225600000, Y: 0}, {X: 1659398400000, Y: 0}, {X: 1659484800000, Y: 3}, {X: 1659657600000, Y: 1}, {X: 1659744000000, Y: 5}, {X: 1659916800000, Y: 0}, {X: 1660003200000, Y: 0}, {X: 1660176000000, Y: 0}, {X: 1660262400000, Y: 0}, {X: 1660435200000, Y: 4}, {X: 1660521600000, Y: 5}, {X: 1660694400000, Y: 0}, {X: 1660780800000, Y: 0}, {X: 1660953600000, Y: 5}, {X: 1661040000000, Y: 6}, {X: 1661212800000, Y: 0}, {X: 1661299200000, Y: 0}, {X: 1661472000000, Y: 0}, {X: 1661558400000, Y: 6}, {X: 1661731200000, Y: 2}, {X: 1661817600000, Y: 0}, {X: 1661990400000, Y: 0}, {X: 1662076800000, Y: 0}, {X: 1662249600000, Y: 7}, {X: 1662336000000, Y: 0}, {X: 1662508800000, Y: 0}, {X: 1662595200000, Y: 0}, {X: 1662768000000, Y: 2}, {X: 1662854400000, Y: 2}, {X: 1663027200000, Y: 0}, {X: 1663113600000, Y: 0}, {X: 1663286400000, Y: 7}, {X: 1663372800000, Y: 0}, {X: 1663545600000, Y: 7}, {X: 1663632000000, Y: 0}, {X: 1663804800000, Y: 0}, {X: 1663891200000, Y: 4}, {X: 1664064000000, Y: 6}, {X: 1664150400000, Y: 1}, {X: 1664323200000, Y: 0}, {X: 1664409600000, Y: 0}, {X: 1664582400000, Y: 3}, {X: 1664668800000, Y: 3}, {X: 1664841600000, Y: 0}, {X: 1664928000000, Y: 2}, {X: 1665093322000, Y: 0}}
}

func weekPlotData() []storage.Plot {
	return []storage.Plot{{X: 1598227200000, Y: 15}, {X: 1598832000000, Y: 14}, {X: 1600041600000, Y: 6}, {X: 1600646400000, Y: 6}, {X: 1601856000000, Y: 11}, {X: 1602460800000, Y: 11}, {X: 1603670400000, Y: 13}, {X: 1604275200000, Y: 10}, {X: 1605484800000, Y: 15}, {X: 1606089600000, Y: 15}, {X: 1607299200000, Y: 13}, {X: 1607904000000, Y: 10}, {X: 1609113600000, Y: 15}, {X: 1609718400000, Y: 12}, {X: 1610928000000, Y: 12}, {X: 1611532800000, Y: 4}, {X: 1612742400000, Y: 2}, {X: 1613347200000, Y: 10}, {X: 1614556800000, Y: 10}, {X: 1615161600000, Y: 10}, {X: 1616371200000, Y: 8}, {X: 1616976000000, Y: 13}, {X: 1618185600000, Y: 8}, {X: 1618790400000, Y: 16}, {X: 1620000000000, Y: 8}, {X: 1620604800000, Y: 11}, {X: 1621814400000, Y: 8}, {X: 1622419200000, Y: 10}, {X: 1623628800000, Y: 11}, {X: 1624233600000, Y: 12}, {X: 1625443200000, Y: 12}, {X: 1626048000000, Y: 15}, {X: 1627257600000, Y: 15}, {X: 1627862400000, Y: 4}, {X: 1629072000000, Y: 10}, {X: 1629676800000, Y: 16}, {X: 1630886400000, Y: 13}, {X: 1631491200000, Y: 12}, {X: 1632700800000, Y: 6}, {X: 1633305600000, Y: 14}, {X: 1634515200000, Y: 7}, {X: 1635120000000, Y: 11}, {X: 1636329600000, Y: 6}, {X: 1636934400000, Y: 3}, {X: 1638144000000, Y: 8}, {X: 1638748800000, Y: 11}, {X: 1639958400000, Y: 14}, {X: 1640563200000, Y: 13}, {X: 1641772800000, Y: 0}, {X: 1642377600000, Y: 0}, {X: 1643587200000, Y: 8}, {X: 1644192000000, Y: 9}, {X: 1645401600000, Y: 15}, {X: 1646006400000, Y: 17}, {X: 1647216000000, Y: 10}, {X: 1647820800000, Y: 17}, {X: 1649030400000, Y: 0}, {X: 1649635200000, Y: 20}, {X: 1650844800000, Y: 10}, {X: 1651449600000, Y: 8}, {X: 1652659200000, Y: 9}, {X: 1653264000000, Y: 12}, {X: 1654473600000, Y: 27}, {X: 1655078400000, Y: 9}, {X: 1656288000000, Y: 14}, {X: 1656892800000, Y: 12}, {X: 1658102400000, Y: 8}, {X: 1658707200000, Y: 11}, {X: 1659916800000, Y: 15}, {X: 1660521600000, Y: 12}, {X: 1661731200000, Y: 10}, {X: 1662336000000, Y: 14}, {X: 1663545600000, Y: 16}, {X: 1664150400000, Y: 16}, {X: 1665093123000, Y: 5}}
}

func monthPlotData() []storage.Plot {
	return []storage.Plot{{X: 1601510400000, Y: 28}, {X: 1604188800000, Y: 57}, {X: 1609459200000, Y: 58}, {X: 1612137600000, Y: 26}, {X: 1617235200000, Y: 43}, {X: 1619827200000, Y: 51}, {X: 1625097600000, Y: 35}, {X: 1627776000000, Y: 51}, {X: 1633046400000, Y: 49}, {X: 1635724800000, Y: 53}, {X: 1640995200000, Y: 66}, {X: 1643673600000, Y: 8}, {X: 1648771200000, Y: 61}, {X: 1651363200000, Y: 54}, {X: 1656633600000, Y: 75}, {X: 1659312000000, Y: 38}, {X: 1664582400000, Y: 61}, {X: 1665093275000, Y: 8}}
}

func yearPlotData() []storage.Plot {
	return []storage.Plot{{X: 1640995200000, Y: 526}, {X: 1665093310000, Y: 464}}
}
