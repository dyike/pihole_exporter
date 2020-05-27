package exporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	statsUrl = "http://%s/admin/api.php?summaryRaw&overTimeData&topItems&recentItems&getQueryTypes&getForwardDestinations&getQuerySources&jsonForceObject&auth=%s"
)

type Client struct {
	hostname string
	token    string
	interval int
}

func NewClient(endpoint string, token string, interval int) *Client {

	return &Client{
		hostname: endpoint,
		token:    token,
		interval: interval,
	}
}

func (c *Client) Collect() {
	for range time.Tick(time.Duration(c.interval)) {
		stats, err := c.GetStats()
		if err != nil {
			continue
		}
		c.GetMetrics(stats)
	}
}

func (c *Client) GetStats() (*Stats, error) {
	resp, err := http.Get(fmt.Sprintf(statsUrl, c.hostname, c.token))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var stats Stats
	dec := json.NewDecoder(bytes.NewBuffer(body))
	if err := dec.Decode(&stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

func (c *Client) GetMetrics(stats *Stats) {
	DomainsBlocked.WithLabelValues(c.hostname).Set(float64(stats.DomainsBeingBlocked))
	DNSQueriesToday.WithLabelValues(c.hostname).Set(float64(stats.DNSQueriesToday))
	AdsBlockedToday.WithLabelValues(c.hostname).Set(float64(stats.AdsBlockedToday))
	AdsPercentageToday.WithLabelValues(c.hostname).Set(float64(stats.AdsPercentageToday))
	UniqueDomains.WithLabelValues(c.hostname).Set(float64(stats.UniqueDomains))
	QueriesForwarded.WithLabelValues(c.hostname).Set(float64(stats.QueriesForwarded))
	QueriesCached.WithLabelValues(c.hostname).Set(float64(stats.QueriesCached))
	ClientsEverSeen.WithLabelValues(c.hostname).Set(float64(stats.ClientsEverSeen))
	UniqueClients.WithLabelValues(c.hostname).Set(float64(stats.UniqueClients))
	DNSQueriesAllTypes.WithLabelValues(c.hostname).Set(float64(stats.DNSQueriesAllTypes))

	Reply.WithLabelValues(c.hostname, "no_data").Set(float64(stats.ReplyNoData))
	Reply.WithLabelValues(c.hostname, "nx_domain").Set(float64(stats.ReplyNxDomain))
	Reply.WithLabelValues(c.hostname, "cname").Set(float64(stats.ReplyCname))
	Reply.WithLabelValues(c.hostname, "ip").Set(float64(stats.ReplyIP))

	var isEnabled int = 0
	if stats.Status == "enable" {
		isEnabled = 1
	}
	Status.WithLabelValues(c.hostname).Set(float64(isEnabled))
	for domain, value := range stats.TopQueries {
		TopQueries.WithLabelValues(c.hostname, domain).Set(float64(value))
	}

	for domain, value := range stats.TopAds {
		TopAds.WithLabelValues(c.hostname, domain).Set(float64(value))
	}

	for source, value := range stats.TopSources {
		TopSources.WithLabelValues(c.hostname, source).Set(float64(value))
	}

	for destination, value := range stats.ForwardDestinations {
		ForwardDestinations.WithLabelValues(c.hostname, destination).Set(value)
	}

	for queryType, value := range stats.QueryTypes {
		QueryTypes.WithLabelValues(c.hostname, queryType).Set(value)
	}
}
