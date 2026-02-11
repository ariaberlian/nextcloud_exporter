package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// JSON response structures
type OCSResponse struct {
	OCS struct {
		Meta struct {
			Status     string `json:"status"`
			StatusCode int    `json:"statuscode"`
			Message    string `json:"message"`
		} `json:"meta"`
		Data struct {
			Nextcloud   NextcloudData   `json:"nextcloud"`
			Server      ServerData      `json:"server"`
			ActiveUsers ActiveUsersData `json:"activeUsers"`
		} `json:"data"`
	} `json:"ocs"`
}

type NextcloudData struct {
	System  SystemData  `json:"system"`
	Storage StorageData `json:"storage"`
	Shares  SharesData  `json:"shares"`
}

type SystemData struct {
	Version   string    `json:"version"`
	FreeSpace int64     `json:"freespace"`
	CPULoad   []float64 `json:"cpuload"`
	CPUNum    int       `json:"cpunum"`
	MemTotal  int64     `json:"mem_total"`
	MemFree   int64     `json:"mem_free"`
	SwapTotal int64     `json:"swap_total"`
	SwapFree  int64     `json:"swap_free"`
	Apps      struct {
		NumInstalled          int `json:"num_installed"`
		NumUpdatesAvailable   int `json:"num_updates_available"`
	} `json:"apps"`
	Update struct {
		Available        bool   `json:"available"`
		AvailableVersion string `json:"available_version"`
	} `json:"update"`
}

type StorageData struct {
	NumUsers         int `json:"num_users"`
	NumFiles         int `json:"num_files"`
	NumStorages      int `json:"num_storages"`
	NumStoragesLocal int `json:"num_storages_local"`
	NumStoragesHome  int `json:"num_storages_home"`
	NumStoragesOther int `json:"num_storages_other"`
}

type SharesData struct {
	NumShares               int `json:"num_shares"`
	NumSharesUser           int `json:"num_shares_user"`
	NumSharesGroups         int `json:"num_shares_groups"`
	NumSharesLink           int `json:"num_shares_link"`
	NumSharesMail           int `json:"num_shares_mail"`
	NumSharesRoom           int `json:"num_shares_room"`
	NumSharesLinkNoPassword int `json:"num_shares_link_no_password"`
	NumFedSharesSent        int `json:"num_fed_shares_sent"`
	NumFedSharesReceived    int `json:"num_fed_shares_received"`
}

type ServerData struct {
	Webserver string `json:"webserver"`
	PHP       struct {
		Version             string `json:"version"`
		MemoryLimit         int64  `json:"memory_limit"`
		MaxExecutionTime    int    `json:"max_execution_time"`
		UploadMaxFilesize   int64  `json:"upload_max_filesize"`
		OPcache             struct {
			OPcacheEnabled bool `json:"opcache_enabled"`
			MemoryUsage    struct {
				UsedMemory   int64 `json:"used_memory"`
				FreeMemory   int64 `json:"free_memory"`
				WastedMemory int64 `json:"wasted_memory"`
			} `json:"memory_usage"`
			OPcacheStatistics struct {
				Hits           int64   `json:"hits"`
				Misses         int64   `json:"misses"`
				OPcacheHitRate float64 `json:"opcache_hit_rate"`
			} `json:"opcache_statistics"`
		} `json:"opcache"`
	} `json:"php"`
	Database struct {
		Type    string `json:"type"`
		Version string `json:"version"`
		Size    string `json:"size"`
	} `json:"database"`
}

type ActiveUsersData struct {
	Last5Minutes int `json:"last5minutes"`
	Last1Hour    int `json:"last1hour"`
	Last24Hours  int `json:"last24hours"`
	Last7Days    int `json:"last7days"`
	Last1Month   int `json:"last1month"`
	Last3Months  int `json:"last3months"`
	Last6Months  int `json:"last6months"`
	LastYear     int `json:"lastyear"`
}

// Status response from /status.php
type StatusResponse struct {
	Installed       bool   `json:"installed"`
	Maintenance     bool   `json:"maintenance"`
	NeedsDbUpgrade  bool   `json:"needsDbUpgrade"`
	Version         string `json:"version"`
	VersionString   string `json:"versionstring"`
	Edition         string `json:"edition"`
	ProductName     string `json:"productname"`
	ExtendedSupport bool   `json:"extendedSupport"`
}

// Collector implements prometheus.Collector
type NextcloudCollector struct {
	baseURL string
	token   string
	client  *http.Client

	// Status metrics (from /status.php)
	statusInfo        *prometheus.Desc
	statusInstalled   *prometheus.Desc
	statusMaintenance *prometheus.Desc
	statusNeedsDbUpgrade *prometheus.Desc
	statusExtendedSupport *prometheus.Desc

	// System metrics
	systemInfo       *prometheus.Desc
	freeSpace        *prometheus.Desc
	cpuLoad          *prometheus.Desc
	cpuCount         *prometheus.Desc
	memTotal         *prometheus.Desc
	memFree          *prometheus.Desc
	swapTotal        *prometheus.Desc
	swapFree         *prometheus.Desc

	// Apps metrics
	appsInstalled       *prometheus.Desc
	appsUpdatesAvailable *prometheus.Desc

	// Update metrics
	updateAvailable *prometheus.Desc

	// Storage metrics
	usersTotal          *prometheus.Desc
	filesTotal          *prometheus.Desc
	storagesTotal       *prometheus.Desc
	storagesLocalTotal  *prometheus.Desc
	storagesHomeTotal   *prometheus.Desc
	storagesOtherTotal  *prometheus.Desc

	// Shares metrics
	sharesTotal                 *prometheus.Desc
	sharesUserTotal             *prometheus.Desc
	sharesGroupsTotal           *prometheus.Desc
	sharesLinkTotal             *prometheus.Desc
	sharesMailTotal             *prometheus.Desc
	sharesRoomTotal             *prometheus.Desc
	sharesLinkNoPasswordTotal   *prometheus.Desc
	sharesFederatedSentTotal    *prometheus.Desc
	sharesFederatedReceivedTotal *prometheus.Desc

	// Server metrics
	phpMemoryLimit        *prometheus.Desc
	phpUploadMaxFilesize  *prometheus.Desc
	phpOpcacheMemoryUsed  *prometheus.Desc
	phpOpcacheMemoryFree  *prometheus.Desc
	phpOpcacheHitRate     *prometheus.Desc
	databaseSize          *prometheus.Desc

	// Active users metrics
	activeUsers *prometheus.Desc

	// Scrape metrics
	scrapeSuccess *prometheus.Desc
}

func NewNextcloudCollector(baseURL, token string) *NextcloudCollector {
	return &NextcloudCollector{
		baseURL: baseURL,
		token:   token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},

		// Status metrics (from /status.php)
		statusInfo: prometheus.NewDesc(
			"nextcloud_status_info",
			"Nextcloud status information",
			[]string{"version", "versionstring", "productname", "edition"}, nil,
		),
		statusInstalled: prometheus.NewDesc(
			"nextcloud_status_installed",
			"Nextcloud installation status (1 = installed, 0 = not installed)",
			nil, nil,
		),
		statusMaintenance: prometheus.NewDesc(
			"nextcloud_status_maintenance",
			"Nextcloud maintenance mode (1 = enabled, 0 = disabled)",
			nil, nil,
		),
		statusNeedsDbUpgrade: prometheus.NewDesc(
			"nextcloud_status_needs_db_upgrade",
			"Nextcloud needs database upgrade (1 = yes, 0 = no)",
			nil, nil,
		),
		statusExtendedSupport: prometheus.NewDesc(
			"nextcloud_status_extended_support",
			"Nextcloud extended support status (1 = enabled, 0 = disabled)",
			nil, nil,
		),

		// System metrics
		systemInfo: prometheus.NewDesc(
			"nextcloud_system_info",
			"Nextcloud system information",
			[]string{"version"}, nil,
		),
		freeSpace: prometheus.NewDesc(
			"nextcloud_system_freespace_bytes",
			"Free disk space in bytes",
			nil, nil,
		),
		cpuLoad: prometheus.NewDesc(
			"nextcloud_system_cpuload",
			"CPU load average",
			[]string{"interval"}, nil,
		),
		cpuCount: prometheus.NewDesc(
			"nextcloud_system_cpu_count",
			"Number of CPUs",
			nil, nil,
		),
		memTotal: prometheus.NewDesc(
			"nextcloud_system_mem_total_bytes",
			"Total memory in bytes",
			nil, nil,
		),
		memFree: prometheus.NewDesc(
			"nextcloud_system_mem_free_bytes",
			"Free memory in bytes",
			nil, nil,
		),
		swapTotal: prometheus.NewDesc(
			"nextcloud_system_swap_total_bytes",
			"Total swap in bytes",
			nil, nil,
		),
		swapFree: prometheus.NewDesc(
			"nextcloud_system_swap_free_bytes",
			"Free swap in bytes",
			nil, nil,
		),

		// Apps metrics
		appsInstalled: prometheus.NewDesc(
			"nextcloud_apps_installed_total",
			"Number of installed apps",
			nil, nil,
		),
		appsUpdatesAvailable: prometheus.NewDesc(
			"nextcloud_apps_updates_available_total",
			"Number of app updates available",
			nil, nil,
		),

		// Update metrics
		updateAvailable: prometheus.NewDesc(
			"nextcloud_update_available",
			"Nextcloud update available (1 = yes, 0 = no)",
			[]string{"available_version"}, nil,
		),

		// Storage metrics
		usersTotal: prometheus.NewDesc(
			"nextcloud_users_total",
			"Total number of users",
			nil, nil,
		),
		filesTotal: prometheus.NewDesc(
			"nextcloud_files_total",
			"Total number of files",
			nil, nil,
		),
		storagesTotal: prometheus.NewDesc(
			"nextcloud_storages_total",
			"Total number of storages",
			nil, nil,
		),
		storagesLocalTotal: prometheus.NewDesc(
			"nextcloud_storages_local_total",
			"Number of local storages",
			nil, nil,
		),
		storagesHomeTotal: prometheus.NewDesc(
			"nextcloud_storages_home_total",
			"Number of home storages",
			nil, nil,
		),
		storagesOtherTotal: prometheus.NewDesc(
			"nextcloud_storages_other_total",
			"Number of other storages",
			nil, nil,
		),

		// Shares metrics
		sharesTotal: prometheus.NewDesc(
			"nextcloud_shares_total",
			"Total number of shares",
			nil, nil,
		),
		sharesUserTotal: prometheus.NewDesc(
			"nextcloud_shares_user_total",
			"Number of user shares",
			nil, nil,
		),
		sharesGroupsTotal: prometheus.NewDesc(
			"nextcloud_shares_groups_total",
			"Number of group shares",
			nil, nil,
		),
		sharesLinkTotal: prometheus.NewDesc(
			"nextcloud_shares_link_total",
			"Number of link shares",
			nil, nil,
		),
		sharesMailTotal: prometheus.NewDesc(
			"nextcloud_shares_mail_total",
			"Number of mail shares",
			nil, nil,
		),
		sharesRoomTotal: prometheus.NewDesc(
			"nextcloud_shares_room_total",
			"Number of room shares",
			nil, nil,
		),
		sharesLinkNoPasswordTotal: prometheus.NewDesc(
			"nextcloud_shares_link_no_password_total",
			"Number of link shares without password",
			nil, nil,
		),
		sharesFederatedSentTotal: prometheus.NewDesc(
			"nextcloud_shares_federated_sent_total",
			"Number of federated shares sent",
			nil, nil,
		),
		sharesFederatedReceivedTotal: prometheus.NewDesc(
			"nextcloud_shares_federated_received_total",
			"Number of federated shares received",
			nil, nil,
		),

		// Server metrics
		phpMemoryLimit: prometheus.NewDesc(
			"nextcloud_php_memory_limit_bytes",
			"PHP memory limit in bytes",
			nil, nil,
		),
		phpUploadMaxFilesize: prometheus.NewDesc(
			"nextcloud_php_upload_max_filesize_bytes",
			"PHP upload max filesize in bytes",
			nil, nil,
		),
		phpOpcacheMemoryUsed: prometheus.NewDesc(
			"nextcloud_php_opcache_memory_used_bytes",
			"PHP OPcache used memory in bytes",
			nil, nil,
		),
		phpOpcacheMemoryFree: prometheus.NewDesc(
			"nextcloud_php_opcache_memory_free_bytes",
			"PHP OPcache free memory in bytes",
			nil, nil,
		),
		phpOpcacheHitRate: prometheus.NewDesc(
			"nextcloud_php_opcache_hit_rate",
			"PHP OPcache hit rate percentage",
			nil, nil,
		),
		databaseSize: prometheus.NewDesc(
			"nextcloud_database_size_bytes",
			"Database size in bytes",
			nil, nil,
		),

		// Active users metrics
		activeUsers: prometheus.NewDesc(
			"nextcloud_active_users",
			"Number of active users",
			[]string{"period"}, nil,
		),

		// Scrape metrics
		scrapeSuccess: prometheus.NewDesc(
			"nextcloud_scrape_success",
			"Whether the scrape was successful (1 = success, 0 = failure)",
			nil, nil,
		),
	}
}

func (c *NextcloudCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.statusInfo
	ch <- c.statusInstalled
	ch <- c.statusMaintenance
	ch <- c.statusNeedsDbUpgrade
	ch <- c.statusExtendedSupport
	ch <- c.systemInfo
	ch <- c.freeSpace
	ch <- c.cpuLoad
	ch <- c.cpuCount
	ch <- c.memTotal
	ch <- c.memFree
	ch <- c.swapTotal
	ch <- c.swapFree
	ch <- c.appsInstalled
	ch <- c.appsUpdatesAvailable
	ch <- c.updateAvailable
	ch <- c.usersTotal
	ch <- c.filesTotal
	ch <- c.storagesTotal
	ch <- c.storagesLocalTotal
	ch <- c.storagesHomeTotal
	ch <- c.storagesOtherTotal
	ch <- c.sharesTotal
	ch <- c.sharesUserTotal
	ch <- c.sharesGroupsTotal
	ch <- c.sharesLinkTotal
	ch <- c.sharesMailTotal
	ch <- c.sharesRoomTotal
	ch <- c.sharesLinkNoPasswordTotal
	ch <- c.sharesFederatedSentTotal
	ch <- c.sharesFederatedReceivedTotal
	ch <- c.phpMemoryLimit
	ch <- c.phpUploadMaxFilesize
	ch <- c.phpOpcacheMemoryUsed
	ch <- c.phpOpcacheMemoryFree
	ch <- c.phpOpcacheHitRate
	ch <- c.databaseSize
	ch <- c.activeUsers
	ch <- c.scrapeSuccess
}

func (c *NextcloudCollector) Collect(ch chan<- prometheus.Metric) {
	// Fetch status data
	status, err := c.fetchStatus()
	if err != nil {
		log.Printf("Error fetching status: %v", err)
	} else {
		// Status metrics
		ch <- prometheus.MustNewConstMetric(c.statusInfo, prometheus.GaugeValue, 1,
			status.Version, status.VersionString, status.ProductName, status.Edition)
		ch <- prometheus.MustNewConstMetric(c.statusInstalled, prometheus.GaugeValue, boolToFloat(status.Installed))
		ch <- prometheus.MustNewConstMetric(c.statusMaintenance, prometheus.GaugeValue, boolToFloat(status.Maintenance))
		ch <- prometheus.MustNewConstMetric(c.statusNeedsDbUpgrade, prometheus.GaugeValue, boolToFloat(status.NeedsDbUpgrade))
		ch <- prometheus.MustNewConstMetric(c.statusExtendedSupport, prometheus.GaugeValue, boolToFloat(status.ExtendedSupport))
	}

	// Fetch serverinfo data
	data, err := c.fetchData()
	if err != nil {
		log.Printf("Error fetching data: %v", err)
		ch <- prometheus.MustNewConstMetric(c.scrapeSuccess, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.scrapeSuccess, prometheus.GaugeValue, 1)

	nc := data.OCS.Data.Nextcloud
	srv := data.OCS.Data.Server
	users := data.OCS.Data.ActiveUsers

	// System metrics
	ch <- prometheus.MustNewConstMetric(c.systemInfo, prometheus.GaugeValue, 1, nc.System.Version)
	ch <- prometheus.MustNewConstMetric(c.freeSpace, prometheus.GaugeValue, float64(nc.System.FreeSpace))

	if len(nc.System.CPULoad) >= 3 {
		ch <- prometheus.MustNewConstMetric(c.cpuLoad, prometheus.GaugeValue, nc.System.CPULoad[0], "1m")
		ch <- prometheus.MustNewConstMetric(c.cpuLoad, prometheus.GaugeValue, nc.System.CPULoad[1], "5m")
		ch <- prometheus.MustNewConstMetric(c.cpuLoad, prometheus.GaugeValue, nc.System.CPULoad[2], "15m")
	}

	ch <- prometheus.MustNewConstMetric(c.cpuCount, prometheus.GaugeValue, float64(nc.System.CPUNum))
	// Memory values from API are in KB, convert to bytes
	ch <- prometheus.MustNewConstMetric(c.memTotal, prometheus.GaugeValue, float64(nc.System.MemTotal)*1024)
	ch <- prometheus.MustNewConstMetric(c.memFree, prometheus.GaugeValue, float64(nc.System.MemFree)*1024)
	ch <- prometheus.MustNewConstMetric(c.swapTotal, prometheus.GaugeValue, float64(nc.System.SwapTotal)*1024)
	ch <- prometheus.MustNewConstMetric(c.swapFree, prometheus.GaugeValue, float64(nc.System.SwapFree)*1024)

	// Apps metrics
	ch <- prometheus.MustNewConstMetric(c.appsInstalled, prometheus.GaugeValue, float64(nc.System.Apps.NumInstalled))
	ch <- prometheus.MustNewConstMetric(c.appsUpdatesAvailable, prometheus.GaugeValue, float64(nc.System.Apps.NumUpdatesAvailable))

	// Update metrics
	updateVal := 0.0
	if nc.System.Update.Available {
		updateVal = 1.0
	}
	ch <- prometheus.MustNewConstMetric(c.updateAvailable, prometheus.GaugeValue, updateVal, nc.System.Update.AvailableVersion)

	// Storage metrics
	ch <- prometheus.MustNewConstMetric(c.usersTotal, prometheus.GaugeValue, float64(nc.Storage.NumUsers))
	ch <- prometheus.MustNewConstMetric(c.filesTotal, prometheus.GaugeValue, float64(nc.Storage.NumFiles))
	ch <- prometheus.MustNewConstMetric(c.storagesTotal, prometheus.GaugeValue, float64(nc.Storage.NumStorages))
	ch <- prometheus.MustNewConstMetric(c.storagesLocalTotal, prometheus.GaugeValue, float64(nc.Storage.NumStoragesLocal))
	ch <- prometheus.MustNewConstMetric(c.storagesHomeTotal, prometheus.GaugeValue, float64(nc.Storage.NumStoragesHome))
	ch <- prometheus.MustNewConstMetric(c.storagesOtherTotal, prometheus.GaugeValue, float64(nc.Storage.NumStoragesOther))

	// Shares metrics
	ch <- prometheus.MustNewConstMetric(c.sharesTotal, prometheus.GaugeValue, float64(nc.Shares.NumShares))
	ch <- prometheus.MustNewConstMetric(c.sharesUserTotal, prometheus.GaugeValue, float64(nc.Shares.NumSharesUser))
	ch <- prometheus.MustNewConstMetric(c.sharesGroupsTotal, prometheus.GaugeValue, float64(nc.Shares.NumSharesGroups))
	ch <- prometheus.MustNewConstMetric(c.sharesLinkTotal, prometheus.GaugeValue, float64(nc.Shares.NumSharesLink))
	ch <- prometheus.MustNewConstMetric(c.sharesMailTotal, prometheus.GaugeValue, float64(nc.Shares.NumSharesMail))
	ch <- prometheus.MustNewConstMetric(c.sharesRoomTotal, prometheus.GaugeValue, float64(nc.Shares.NumSharesRoom))
	ch <- prometheus.MustNewConstMetric(c.sharesLinkNoPasswordTotal, prometheus.GaugeValue, float64(nc.Shares.NumSharesLinkNoPassword))
	ch <- prometheus.MustNewConstMetric(c.sharesFederatedSentTotal, prometheus.GaugeValue, float64(nc.Shares.NumFedSharesSent))
	ch <- prometheus.MustNewConstMetric(c.sharesFederatedReceivedTotal, prometheus.GaugeValue, float64(nc.Shares.NumFedSharesReceived))

	// Server metrics
	ch <- prometheus.MustNewConstMetric(c.phpMemoryLimit, prometheus.GaugeValue, float64(srv.PHP.MemoryLimit))
	ch <- prometheus.MustNewConstMetric(c.phpUploadMaxFilesize, prometheus.GaugeValue, float64(srv.PHP.UploadMaxFilesize))
	ch <- prometheus.MustNewConstMetric(c.phpOpcacheMemoryUsed, prometheus.GaugeValue, float64(srv.PHP.OPcache.MemoryUsage.UsedMemory))
	ch <- prometheus.MustNewConstMetric(c.phpOpcacheMemoryFree, prometheus.GaugeValue, float64(srv.PHP.OPcache.MemoryUsage.FreeMemory))
	ch <- prometheus.MustNewConstMetric(c.phpOpcacheHitRate, prometheus.GaugeValue, srv.PHP.OPcache.OPcacheStatistics.OPcacheHitRate)

	// Database size (parse string to int)
	if dbSize, err := strconv.ParseInt(srv.Database.Size, 10, 64); err == nil {
		ch <- prometheus.MustNewConstMetric(c.databaseSize, prometheus.GaugeValue, float64(dbSize))
	}

	// Active users metrics
	ch <- prometheus.MustNewConstMetric(c.activeUsers, prometheus.GaugeValue, float64(users.Last5Minutes), "5min")
	ch <- prometheus.MustNewConstMetric(c.activeUsers, prometheus.GaugeValue, float64(users.Last1Hour), "1hour")
	ch <- prometheus.MustNewConstMetric(c.activeUsers, prometheus.GaugeValue, float64(users.Last24Hours), "24hours")
	ch <- prometheus.MustNewConstMetric(c.activeUsers, prometheus.GaugeValue, float64(users.Last7Days), "7days")
	ch <- prometheus.MustNewConstMetric(c.activeUsers, prometheus.GaugeValue, float64(users.Last1Month), "1month")
	ch <- prometheus.MustNewConstMetric(c.activeUsers, prometheus.GaugeValue, float64(users.Last3Months), "3months")
	ch <- prometheus.MustNewConstMetric(c.activeUsers, prometheus.GaugeValue, float64(users.Last6Months), "6months")
	ch <- prometheus.MustNewConstMetric(c.activeUsers, prometheus.GaugeValue, float64(users.LastYear), "1year")
}

func (c *NextcloudCollector) fetchStatus() (*StatusResponse, error) {
	url := c.baseURL + "/status.php"
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
	url := c.baseURL + "/ocs/v2.php/apps/serverinfo/api/v1/info?format=json&skipApps=false&skipUpdate=false"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("NC-Token", c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Command line flags
	baseURL := flag.String("url", "", "Nextcloud base URL (e.g., https://cloud.example.com)")
	token := flag.String("token", "", "NC-Token for authentication")
	listenAddr := flag.String("listen", ":9205", "Address to listen on")
	flag.Parse()

	// Use environment variables as fallback
	if *baseURL == "" {
		*baseURL = getEnv("NEXTCLOUD_URL", "")
	}
	if *token == "" {
		*token = getEnv("NC_TOKEN", "")
	}
	if *listenAddr == ":9205" {
		*listenAddr = getEnv("LISTEN_ADDR", ":9205")
	}

	// Validate required parameters
	if *baseURL == "" {
		log.Fatal("Nextcloud URL is required. Set via -url flag or NEXTCLOUD_URL environment variable")
	}
	if *token == "" {
		log.Fatal("NC-Token is required. Set via -token flag or NC_TOKEN environment variable")
	}

	// Create and register collector
	collector := NewNextcloudCollector(*baseURL, *token)
	prometheus.MustRegister(collector)

	// Setup HTTP server
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
<head><title>Nextcloud Exporter</title></head>
<body>
<h1>Nextcloud Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
</body>
</html>`))
	})

	log.Printf("Starting Nextcloud exporter on %s", *listenAddr)
	log.Printf("Fetching metrics from: %s", *baseURL)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}