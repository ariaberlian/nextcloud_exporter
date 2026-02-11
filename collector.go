package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// NextcloudCollector implements prometheus.Collector
type NextcloudCollector struct {
	config  *Config
	client  *http.Client
	metrics *MetricDescriptors

	// Caching for rate limiting
	cacheMu         sync.RWMutex
	cachedStatus    *StatusResponse
	cachedData      *OCSResponse
	lastFetchTime   time.Time
	lastStatusFetch time.Time
}

// NewNextcloudCollector creates a new collector with the given configuration
func NewNextcloudCollector(config *Config) *NextcloudCollector {
	return &NextcloudCollector{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		metrics: NewMetricDescriptors(),
	}
}

// Describe implements prometheus.Collector
func (c *NextcloudCollector) Describe(ch chan<- *prometheus.Desc) {
	c.metrics.DescribeAll(ch)
}

// Collect implements prometheus.Collector
func (c *NextcloudCollector) Collect(ch chan<- prometheus.Metric) {
	// Fetch status data (with caching)
	status, statusErr := c.fetchStatusCached()
	if statusErr != nil {
		log.Printf("Error fetching status: %v", statusErr)
	} else {
		c.collectStatusMetrics(ch, status)
	}

	// Fetch serverinfo data (with caching)
	data, dataErr := c.fetchDataCached()
	if dataErr != nil {
		log.Printf("Error fetching data: %v", dataErr)
		ch <- prometheus.MustNewConstMetric(c.metrics.ScrapeSuccess, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.metrics.ScrapeSuccess, prometheus.GaugeValue, 1)
	c.collectAllMetrics(ch, data)
}

func (c *NextcloudCollector) collectStatusMetrics(ch chan<- prometheus.Metric, status *StatusResponse) {
	ch <- prometheus.MustNewConstMetric(c.metrics.StatusInfo, prometheus.GaugeValue, 1,
		status.Version, status.VersionString, status.ProductName, status.Edition)
	ch <- prometheus.MustNewConstMetric(c.metrics.StatusInstalled, prometheus.GaugeValue, boolToFloat(status.Installed))
	ch <- prometheus.MustNewConstMetric(c.metrics.StatusMaintenance, prometheus.GaugeValue, boolToFloat(status.Maintenance))
	ch <- prometheus.MustNewConstMetric(c.metrics.StatusNeedsDbUpgrade, prometheus.GaugeValue, boolToFloat(status.NeedsDbUpgrade))
	ch <- prometheus.MustNewConstMetric(c.metrics.StatusExtendedSupport, prometheus.GaugeValue, boolToFloat(status.ExtendedSupport))
}

func (c *NextcloudCollector) collectAllMetrics(ch chan<- prometheus.Metric, data *OCSResponse) {
	nc := data.OCS.Data.Nextcloud
	srv := data.OCS.Data.Server
	users := data.OCS.Data.ActiveUsers

	// System metrics
	ch <- prometheus.MustNewConstMetric(c.metrics.SystemInfo, prometheus.GaugeValue, 1, nc.System.Version)
	ch <- prometheus.MustNewConstMetric(c.metrics.FreeSpace, prometheus.GaugeValue, float64(nc.System.FreeSpace))

	if len(nc.System.CPULoad) >= 3 {
		ch <- prometheus.MustNewConstMetric(c.metrics.CPULoad, prometheus.GaugeValue, nc.System.CPULoad[0], "1m")
		ch <- prometheus.MustNewConstMetric(c.metrics.CPULoad, prometheus.GaugeValue, nc.System.CPULoad[1], "5m")
		ch <- prometheus.MustNewConstMetric(c.metrics.CPULoad, prometheus.GaugeValue, nc.System.CPULoad[2], "15m")
	}

	ch <- prometheus.MustNewConstMetric(c.metrics.CPUCount, prometheus.GaugeValue, float64(nc.System.CPUNum))
	// Memory values from API are in KB, convert to bytes
	ch <- prometheus.MustNewConstMetric(c.metrics.MemTotal, prometheus.GaugeValue, float64(nc.System.MemTotal)*1024)
	ch <- prometheus.MustNewConstMetric(c.metrics.MemFree, prometheus.GaugeValue, float64(nc.System.MemFree)*1024)
	ch <- prometheus.MustNewConstMetric(c.metrics.SwapTotal, prometheus.GaugeValue, float64(nc.System.SwapTotal)*1024)
	ch <- prometheus.MustNewConstMetric(c.metrics.SwapFree, prometheus.GaugeValue, float64(nc.System.SwapFree)*1024)

	// Apps metrics
	ch <- prometheus.MustNewConstMetric(c.metrics.AppsInstalled, prometheus.GaugeValue, float64(nc.System.Apps.NumInstalled))
	ch <- prometheus.MustNewConstMetric(c.metrics.AppsUpdatesAvailable, prometheus.GaugeValue, float64(nc.System.Apps.NumUpdatesAvailable))

	// Update metrics
	updateVal := 0.0
	if nc.System.Update.Available {
		updateVal = 1.0
	}
	ch <- prometheus.MustNewConstMetric(c.metrics.UpdateAvailable, prometheus.GaugeValue, updateVal, nc.System.Update.AvailableVersion)

	// Storage metrics
	ch <- prometheus.MustNewConstMetric(c.metrics.UsersTotal, prometheus.GaugeValue, float64(nc.Storage.NumUsers))
	ch <- prometheus.MustNewConstMetric(c.metrics.FilesTotal, prometheus.GaugeValue, float64(nc.Storage.NumFiles))
	ch <- prometheus.MustNewConstMetric(c.metrics.StoragesTotal, prometheus.GaugeValue, float64(nc.Storage.NumStorages))
	ch <- prometheus.MustNewConstMetric(c.metrics.StoragesLocalTotal, prometheus.GaugeValue, float64(nc.Storage.NumStoragesLocal))
	ch <- prometheus.MustNewConstMetric(c.metrics.StoragesHomeTotal, prometheus.GaugeValue, float64(nc.Storage.NumStoragesHome))
	ch <- prometheus.MustNewConstMetric(c.metrics.StoragesOtherTotal, prometheus.GaugeValue, float64(nc.Storage.NumStoragesOther))

	// Shares metrics
	ch <- prometheus.MustNewConstMetric(c.metrics.SharesTotal, prometheus.GaugeValue, float64(nc.Shares.NumShares))
	ch <- prometheus.MustNewConstMetric(c.metrics.SharesUserTotal, prometheus.GaugeValue, float64(nc.Shares.NumSharesUser))
	ch <- prometheus.MustNewConstMetric(c.metrics.SharesGroupsTotal, prometheus.GaugeValue, float64(nc.Shares.NumSharesGroups))
	ch <- prometheus.MustNewConstMetric(c.metrics.SharesLinkTotal, prometheus.GaugeValue, float64(nc.Shares.NumSharesLink))
	ch <- prometheus.MustNewConstMetric(c.metrics.SharesMailTotal, prometheus.GaugeValue, float64(nc.Shares.NumSharesMail))
	ch <- prometheus.MustNewConstMetric(c.metrics.SharesRoomTotal, prometheus.GaugeValue, float64(nc.Shares.NumSharesRoom))
	ch <- prometheus.MustNewConstMetric(c.metrics.SharesLinkNoPasswordTotal, prometheus.GaugeValue, float64(nc.Shares.NumSharesLinkNoPassword))
	ch <- prometheus.MustNewConstMetric(c.metrics.SharesFederatedSentTotal, prometheus.GaugeValue, float64(nc.Shares.NumFedSharesSent))
	ch <- prometheus.MustNewConstMetric(c.metrics.SharesFederatedReceivedTotal, prometheus.GaugeValue, float64(nc.Shares.NumFedSharesReceived))

	// Server metrics
	ch <- prometheus.MustNewConstMetric(c.metrics.PHPMemoryLimit, prometheus.GaugeValue, float64(srv.PHP.MemoryLimit))
	ch <- prometheus.MustNewConstMetric(c.metrics.PHPUploadMaxFilesize, prometheus.GaugeValue, float64(srv.PHP.UploadMaxFilesize))
	ch <- prometheus.MustNewConstMetric(c.metrics.PHPOpcacheMemoryUsed, prometheus.GaugeValue, float64(srv.PHP.OPcache.MemoryUsage.UsedMemory))
	ch <- prometheus.MustNewConstMetric(c.metrics.PHPOpcacheMemoryFree, prometheus.GaugeValue, float64(srv.PHP.OPcache.MemoryUsage.FreeMemory))
	ch <- prometheus.MustNewConstMetric(c.metrics.PHPOpcacheHitRate, prometheus.GaugeValue, srv.PHP.OPcache.OPcacheStatistics.OPcacheHitRate)

	// Database size (parse string to int)
	if dbSize, err := strconv.ParseInt(srv.Database.Size, 10, 64); err == nil {
		ch <- prometheus.MustNewConstMetric(c.metrics.DatabaseSize, prometheus.GaugeValue, float64(dbSize))
	}

	// Active users metrics
	ch <- prometheus.MustNewConstMetric(c.metrics.ActiveUsers, prometheus.GaugeValue, float64(users.Last5Minutes), "5min")
	ch <- prometheus.MustNewConstMetric(c.metrics.ActiveUsers, prometheus.GaugeValue, float64(users.Last1Hour), "1hour")
	ch <- prometheus.MustNewConstMetric(c.metrics.ActiveUsers, prometheus.GaugeValue, float64(users.Last24Hours), "24hours")
	ch <- prometheus.MustNewConstMetric(c.metrics.ActiveUsers, prometheus.GaugeValue, float64(users.Last7Days), "7days")
	ch <- prometheus.MustNewConstMetric(c.metrics.ActiveUsers, prometheus.GaugeValue, float64(users.Last1Month), "1month")
	ch <- prometheus.MustNewConstMetric(c.metrics.ActiveUsers, prometheus.GaugeValue, float64(users.Last3Months), "3months")
	ch <- prometheus.MustNewConstMetric(c.metrics.ActiveUsers, prometheus.GaugeValue, float64(users.Last6Months), "6months")
	ch <- prometheus.MustNewConstMetric(c.metrics.ActiveUsers, prometheus.GaugeValue, float64(users.LastYear), "1year")
}

// fetchStatusCached returns cached status if within fetch interval, otherwise fetches fresh data
func (c *NextcloudCollector) fetchStatusCached() (*StatusResponse, error) {
	c.cacheMu.RLock()
	if c.cachedStatus != nil && time.Since(c.lastStatusFetch) < c.config.FetchInterval {
		status := c.cachedStatus
		c.cacheMu.RUnlock()
		return status, nil
	}
	c.cacheMu.RUnlock()

	// Need to fetch fresh data
	status, err := c.fetchStatus()
	if err != nil {
		// If fetch fails but we have cached data, return cached data
		c.cacheMu.RLock()
		if c.cachedStatus != nil {
			cachedStatus := c.cachedStatus
			c.cacheMu.RUnlock()
			log.Printf("Using cached status data due to fetch error: %v", err)
			return cachedStatus, nil
		}
		c.cacheMu.RUnlock()
		return nil, err
	}

	c.cacheMu.Lock()
	c.cachedStatus = status
	c.lastStatusFetch = time.Now()
	c.cacheMu.Unlock()

	return status, nil
}

// fetchDataCached returns cached data if within fetch interval, otherwise fetches fresh data
func (c *NextcloudCollector) fetchDataCached() (*OCSResponse, error) {
	c.cacheMu.RLock()
	if c.cachedData != nil && time.Since(c.lastFetchTime) < c.config.FetchInterval {
		data := c.cachedData
		c.cacheMu.RUnlock()
		return data, nil
	}
	c.cacheMu.RUnlock()

	// Need to fetch fresh data
	data, err := c.fetchData()
	if err != nil {
		// If fetch fails but we have cached data, return cached data
		c.cacheMu.RLock()
		if c.cachedData != nil {
			cachedData := c.cachedData
			c.cacheMu.RUnlock()
			log.Printf("Using cached serverinfo data due to fetch error: %v", err)
			return cachedData, nil
		}
		c.cacheMu.RUnlock()
		return nil, err
	}

	c.cacheMu.Lock()
	c.cachedData = data
	c.lastFetchTime = time.Now()
	c.cacheMu.Unlock()

	return data, nil
}

func (c *NextcloudCollector) fetchStatus() (*StatusResponse, error) {
	url := c.config.BaseURL + "/status.php"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited (429): too many requests")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var data StatusResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return &data, nil
}

func (c *NextcloudCollector) fetchData() (*OCSResponse, error) {
	url := c.config.BaseURL + "/ocs/v2.php/apps/serverinfo/api/v1/info?format=json&skipApps=false&skipUpdate=false"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("NC-Token", c.config.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited (429): too many requests")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var data OCSResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return &data, nil
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
