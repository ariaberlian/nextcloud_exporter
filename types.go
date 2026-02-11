package main

// OCSResponse is the main response structure from Nextcloud serverinfo API
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

// NextcloudData contains system, storage, and shares information
type NextcloudData struct {
	System  SystemData  `json:"system"`
	Storage StorageData `json:"storage"`
	Shares  SharesData  `json:"shares"`
}

// SystemData contains system-level information
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
		NumInstalled        int `json:"num_installed"`
		NumUpdatesAvailable int `json:"num_updates_available"`
	} `json:"apps"`
	Update struct {
		Available        bool   `json:"available"`
		AvailableVersion string `json:"available_version"`
	} `json:"update"`
}

// StorageData contains storage statistics
type StorageData struct {
	NumUsers         int `json:"num_users"`
	NumFiles         int `json:"num_files"`
	NumStorages      int `json:"num_storages"`
	NumStoragesLocal int `json:"num_storages_local"`
	NumStoragesHome  int `json:"num_storages_home"`
	NumStoragesOther int `json:"num_storages_other"`
}

// SharesData contains sharing statistics
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

// ServerData contains server configuration information
type ServerData struct {
	Webserver string `json:"webserver"`
	PHP       struct {
		Version           string `json:"version"`
		MemoryLimit       int64  `json:"memory_limit"`
		MaxExecutionTime  int    `json:"max_execution_time"`
		UploadMaxFilesize int64  `json:"upload_max_filesize"`
		OPcache           struct {
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

// ActiveUsersData contains active user statistics
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

// StatusResponse is the response from /status.php
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
