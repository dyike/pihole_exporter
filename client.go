package pihole_exporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"pihole_exporter/metrics"
)

var (
	statsUrl = "http://%s/admin/api.php?summaryRaw&overTimeData&topItems&recentItems&getQueryTypes&getForwardDestinations&getQuerySources&jsonForceObject"
)

type Client struct {
	EndPoint string
}

func NewClient(endpoint string) (*Client, error) {
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	return &Client{
		EndPoint: url.String(),
	}, nil
}

func (c *Client) GetStats() (*Stats, error) {
	resp, err := htt.Get(fmt.Sprintf(statsUrl, c.EndPoint))
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
	if stats.Status == enabledStatus {
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
