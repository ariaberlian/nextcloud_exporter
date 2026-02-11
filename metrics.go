package main

import "github.com/prometheus/client_golang/prometheus"

// MetricDescriptors holds all prometheus metric descriptors
type MetricDescriptors struct {
	// Status metrics (from /status.php)
	StatusInfo            *prometheus.Desc
	StatusInstalled       *prometheus.Desc
	StatusMaintenance     *prometheus.Desc
	StatusNeedsDbUpgrade  *prometheus.Desc
	StatusExtendedSupport *prometheus.Desc

	// System metrics
	SystemInfo *prometheus.Desc
	FreeSpace  *prometheus.Desc
	CPULoad    *prometheus.Desc
	CPUCount   *prometheus.Desc
	MemTotal   *prometheus.Desc
	MemFree    *prometheus.Desc
	SwapTotal  *prometheus.Desc
	SwapFree   *prometheus.Desc

	// Apps metrics
	AppsInstalled        *prometheus.Desc
	AppsUpdatesAvailable *prometheus.Desc

	// Update metrics
	UpdateAvailable *prometheus.Desc

	// Storage metrics
	UsersTotal         *prometheus.Desc
	FilesTotal         *prometheus.Desc
	StoragesTotal      *prometheus.Desc
	StoragesLocalTotal *prometheus.Desc
	StoragesHomeTotal  *prometheus.Desc
	StoragesOtherTotal *prometheus.Desc

	// Shares metrics
	SharesTotal                  *prometheus.Desc
	SharesUserTotal              *prometheus.Desc
	SharesGroupsTotal            *prometheus.Desc
	SharesLinkTotal              *prometheus.Desc
	SharesMailTotal              *prometheus.Desc
	SharesRoomTotal              *prometheus.Desc
	SharesLinkNoPasswordTotal    *prometheus.Desc
	SharesFederatedSentTotal     *prometheus.Desc
	SharesFederatedReceivedTotal *prometheus.Desc

	// Server metrics
	PHPMemoryLimit       *prometheus.Desc
	PHPUploadMaxFilesize *prometheus.Desc
	PHPOpcacheMemoryUsed *prometheus.Desc
	PHPOpcacheMemoryFree *prometheus.Desc
	PHPOpcacheHitRate    *prometheus.Desc
	DatabaseSize         *prometheus.Desc

	// Active users metrics
	ActiveUsers *prometheus.Desc

	// Scrape metrics
	ScrapeSuccess *prometheus.Desc
}

// NewMetricDescriptors creates all metric descriptors
func NewMetricDescriptors() *MetricDescriptors {
	return &MetricDescriptors{
		// Status metrics (from /status.php)
		StatusInfo: prometheus.NewDesc(
			"nextcloud_status_info",
			"Nextcloud status information",
			[]string{"version", "versionstring", "productname", "edition"}, nil,
		),
		StatusInstalled: prometheus.NewDesc(
			"nextcloud_status_installed",
			"Nextcloud installation status (1 = installed, 0 = not installed)",
			nil, nil,
		),
		StatusMaintenance: prometheus.NewDesc(
			"nextcloud_status_maintenance",
			"Nextcloud maintenance mode (1 = enabled, 0 = disabled)",
			nil, nil,
		),
		StatusNeedsDbUpgrade: prometheus.NewDesc(
			"nextcloud_status_needs_db_upgrade",
			"Nextcloud needs database upgrade (1 = yes, 0 = no)",
			nil, nil,
		),
		StatusExtendedSupport: prometheus.NewDesc(
			"nextcloud_status_extended_support",
			"Nextcloud extended support status (1 = enabled, 0 = disabled)",
			nil, nil,
		),

		// System metrics
		SystemInfo: prometheus.NewDesc(
			"nextcloud_system_info",
			"Nextcloud system information",
			[]string{"version"}, nil,
		),
		FreeSpace: prometheus.NewDesc(
			"nextcloud_system_freespace_bytes",
			"Free disk space in bytes",
			nil, nil,
		),
		CPULoad: prometheus.NewDesc(
			"nextcloud_system_cpuload",
			"CPU load average",
			[]string{"interval"}, nil,
		),
		CPUCount: prometheus.NewDesc(
			"nextcloud_system_cpu_count",
			"Number of CPUs",
			nil, nil,
		),
		MemTotal: prometheus.NewDesc(
			"nextcloud_system_mem_total_bytes",
			"Total memory in bytes",
			nil, nil,
		),
		MemFree: prometheus.NewDesc(
			"nextcloud_system_mem_free_bytes",
			"Free memory in bytes",
			nil, nil,
		),
		SwapTotal: prometheus.NewDesc(
			"nextcloud_system_swap_total_bytes",
			"Total swap in bytes",
			nil, nil,
		),
		SwapFree: prometheus.NewDesc(
			"nextcloud_system_swap_free_bytes",
			"Free swap in bytes",
			nil, nil,
		),

		// Apps metrics
		AppsInstalled: prometheus.NewDesc(
			"nextcloud_apps_installed_total",
			"Number of installed apps",
			nil, nil,
		),
		AppsUpdatesAvailable: prometheus.NewDesc(
			"nextcloud_apps_updates_available_total",
			"Number of app updates available",
			nil, nil,
		),

		// Update metrics
		UpdateAvailable: prometheus.NewDesc(
			"nextcloud_update_available",
			"Nextcloud update available (1 = yes, 0 = no)",
			[]string{"available_version"}, nil,
		),

		// Storage metrics
		UsersTotal: prometheus.NewDesc(
			"nextcloud_users_total",
			"Total number of users",
			nil, nil,
		),
		FilesTotal: prometheus.NewDesc(
			"nextcloud_files_total",
			"Total number of files",
			nil, nil,
		),
		StoragesTotal: prometheus.NewDesc(
			"nextcloud_storages_total",
			"Total number of storages",
			nil, nil,
		),
		StoragesLocalTotal: prometheus.NewDesc(
			"nextcloud_storages_local_total",
			"Number of local storages",
			nil, nil,
		),
		StoragesHomeTotal: prometheus.NewDesc(
			"nextcloud_storages_home_total",
			"Number of home storages",
			nil, nil,
		),
		StoragesOtherTotal: prometheus.NewDesc(
			"nextcloud_storages_other_total",
			"Number of other storages",
			nil, nil,
		),

		// Shares metrics
		SharesTotal: prometheus.NewDesc(
			"nextcloud_shares_total",
			"Total number of shares",
			nil, nil,
		),
		SharesUserTotal: prometheus.NewDesc(
			"nextcloud_shares_user_total",
			"Number of user shares",
			nil, nil,
		),
		SharesGroupsTotal: prometheus.NewDesc(
			"nextcloud_shares_groups_total",
			"Number of group shares",
			nil, nil,
		),
		SharesLinkTotal: prometheus.NewDesc(
			"nextcloud_shares_link_total",
			"Number of link shares",
			nil, nil,
		),
		SharesMailTotal: prometheus.NewDesc(
			"nextcloud_shares_mail_total",
			"Number of mail shares",
			nil, nil,
		),
		SharesRoomTotal: prometheus.NewDesc(
			"nextcloud_shares_room_total",
			"Number of room shares",
			nil, nil,
		),
		SharesLinkNoPasswordTotal: prometheus.NewDesc(
			"nextcloud_shares_link_no_password_total",
			"Number of link shares without password",
			nil, nil,
		),
		SharesFederatedSentTotal: prometheus.NewDesc(
			"nextcloud_shares_federated_sent_total",
			"Number of federated shares sent",
			nil, nil,
		),
		SharesFederatedReceivedTotal: prometheus.NewDesc(
			"nextcloud_shares_federated_received_total",
			"Number of federated shares received",
			nil, nil,
		),

		// Server metrics
		PHPMemoryLimit: prometheus.NewDesc(
			"nextcloud_php_memory_limit_bytes",
			"PHP memory limit in bytes",
			nil, nil,
		),
		PHPUploadMaxFilesize: prometheus.NewDesc(
			"nextcloud_php_upload_max_filesize_bytes",
			"PHP upload max filesize in bytes",
			nil, nil,
		),
		PHPOpcacheMemoryUsed: prometheus.NewDesc(
			"nextcloud_php_opcache_memory_used_bytes",
			"PHP OPcache used memory in bytes",
			nil, nil,
		),
		PHPOpcacheMemoryFree: prometheus.NewDesc(
			"nextcloud_php_opcache_memory_free_bytes",
			"PHP OPcache free memory in bytes",
			nil, nil,
		),
		PHPOpcacheHitRate: prometheus.NewDesc(
			"nextcloud_php_opcache_hit_rate",
			"PHP OPcache hit rate percentage",
			nil, nil,
		),
		DatabaseSize: prometheus.NewDesc(
			"nextcloud_database_size_bytes",
			"Database size in bytes",
			nil, nil,
		),

		// Active users metrics
		ActiveUsers: prometheus.NewDesc(
			"nextcloud_active_users",
			"Number of active users",
			[]string{"period"}, nil,
		),

		// Scrape metrics
		ScrapeSuccess: prometheus.NewDesc(
			"nextcloud_scrape_success",
			"Whether the scrape was successful (1 = success, 0 = failure)",
			nil, nil,
		),
	}
}

// DescribeAll sends all metric descriptors to the channel
func (m *MetricDescriptors) DescribeAll(ch chan<- *prometheus.Desc) {
	ch <- m.StatusInfo
	ch <- m.StatusInstalled
	ch <- m.StatusMaintenance
	ch <- m.StatusNeedsDbUpgrade
	ch <- m.StatusExtendedSupport
	ch <- m.SystemInfo
	ch <- m.FreeSpace
	ch <- m.CPULoad
	ch <- m.CPUCount
	ch <- m.MemTotal
	ch <- m.MemFree
	ch <- m.SwapTotal
	ch <- m.SwapFree
	ch <- m.AppsInstalled
	ch <- m.AppsUpdatesAvailable
	ch <- m.UpdateAvailable
	ch <- m.UsersTotal
	ch <- m.FilesTotal
	ch <- m.StoragesTotal
	ch <- m.StoragesLocalTotal
	ch <- m.StoragesHomeTotal
	ch <- m.StoragesOtherTotal
	ch <- m.SharesTotal
	ch <- m.SharesUserTotal
	ch <- m.SharesGroupsTotal
	ch <- m.SharesLinkTotal
	ch <- m.SharesMailTotal
	ch <- m.SharesRoomTotal
	ch <- m.SharesLinkNoPasswordTotal
	ch <- m.SharesFederatedSentTotal
	ch <- m.SharesFederatedReceivedTotal
	ch <- m.PHPMemoryLimit
	ch <- m.PHPUploadMaxFilesize
	ch <- m.PHPOpcacheMemoryUsed
	ch <- m.PHPOpcacheMemoryFree
	ch <- m.PHPOpcacheHitRate
	ch <- m.DatabaseSize
	ch <- m.ActiveUsers
	ch <- m.ScrapeSuccess
}
