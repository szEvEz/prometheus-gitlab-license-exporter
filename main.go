package main

import (
	"crypto/tls"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

const namespace = "gitlab_license"

var (
	tr = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
)

type gitlabLicenseCollector struct {
	id               *prometheus.Desc
	startsAt         *prometheus.Desc
	expiresAt        *prometheus.Desc
	historicalMax    *prometheus.Desc
	maximumUserCount *prometheus.Desc
	expired          *prometheus.Desc
	overage          *prometheus.Desc
	userLimit        *prometheus.Desc
	activeUsers      *prometheus.Desc
}

func newGitlabLicenseCollector() *gitlabLicenseCollector {
	return &gitlabLicenseCollector{
		id: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "id"),
			"ID of the license",
			nil,
			nil,
		),
		startsAt: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "starts_at"),
			"Date the license starts at",
			nil,
			nil,
		),
		expiresAt: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "expires_at"),
			"Date the license expires at",
			nil,
			nil,
		),
		historicalMax: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "historical_max"),
			"This is the highest peak of users on your installation since the license started",
			nil,
			nil,
		),
		maximumUserCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "maximum_user_count"),
			"This is the highest peak of users on your installation since the license started",
			nil,
			nil,
		),
		expired: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "expired"),
			"Expiry status of the license",
			nil,
			nil,
		),
		overage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "overage"),
			"The difference between the number of billable users and the licensed number of users",
			nil,
			nil,
		),
		userLimit: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "user_limit"),
			"The number of users the license is licensed for",
			nil,
			nil,
		),
		activeUsers: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "active_users"),
			"Current active users that consume a license",
			nil,
			nil,
		),
	}
}

func (collector *gitlabLicenseCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.id
	ch <- collector.startsAt
	ch <- collector.expiresAt
	ch <- collector.historicalMax
	ch <- collector.maximumUserCount
	ch <- collector.expired
	ch <- collector.overage
	ch <- collector.userLimit
	ch <- collector.activeUsers
}

type gitlabLicenseMetrics struct {
	ID               float64 `json:"id"`
	StartsAt         string  `json:"starts_at"`
	ExpiresAt        string  `json:"expires_at"`
	HistoricalMax    float64 `json:"historical_max"`
	MaximumUserCount float64 `json:"maximum_user_count"`
	Expired          bool    `json:"expired"`
	Overage          float64 `json:"overage"`
	UserLimit        float64 `json:"user_limit"`
	ActiveUsers      float64 `json:"active_users"`
}

func (collector *gitlabLicenseCollector) Collect(ch chan<- prometheus.Metric) {

	gitlabURL, gitlabToken := validateEnvVars()

	req, err := http.NewRequest("GET", gitlabURL+"/api/v4/license", nil)
	if err != nil {
		log.Info(err)
	}

	req.Header.Add("Authorization", "Bearer "+gitlabToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Warn(err)
		return
	} else if resp.StatusCode == 401 {
		log.Info("Unauthorized - Status Code", resp.StatusCode)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Info(err)
	}

	var g gitlabLicenseMetrics
	err = json.Unmarshal(body, &g)
	if err != nil {
		log.Info("unmarshal error", err)
	}

	log.Info("Metrics retrieved")

	ch <- prometheus.MustNewConstMetric(collector.id, prometheus.GaugeValue, g.ID)
	ch <- prometheus.MustNewConstMetric(collector.startsAt, prometheus.GaugeValue, parseStringDateTofloat(g.StartsAt))
	ch <- prometheus.MustNewConstMetric(collector.expiresAt, prometheus.GaugeValue, parseStringDateTofloat(g.ExpiresAt))
	ch <- prometheus.MustNewConstMetric(collector.historicalMax, prometheus.GaugeValue, g.HistoricalMax)
	ch <- prometheus.MustNewConstMetric(collector.maximumUserCount, prometheus.GaugeValue, g.MaximumUserCount)
	ch <- prometheus.MustNewConstMetric(collector.expired, prometheus.GaugeValue, isExpired(g.Expired))
	ch <- prometheus.MustNewConstMetric(collector.overage, prometheus.GaugeValue, g.Overage)
	ch <- prometheus.MustNewConstMetric(collector.userLimit, prometheus.GaugeValue, g.UserLimit)
	ch <- prometheus.MustNewConstMetric(collector.activeUsers, prometheus.GaugeValue, g.ActiveUsers)
}

func validateEnvVars() (string, string) {
	gitlabURL, set := os.LookupEnv("GITLAB_URL")
	if !set {
		log.Fatal("GITLAB_URL environment variable is not set")
	}

	gitlabToken, set := os.LookupEnv("GITLAB_TOKEN")
	if !set {
		log.Fatal("GITLAB_TOKEN environment variable is not set")
	}

	return gitlabURL, gitlabToken
}

func isExpired(expired bool) float64 {
	switch expired {
	case false:
		return 0
	case true:
		return 1
	default:
		return 0
	}
}

func parseStringDateTofloat(date string) float64 {
	ca, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Info(err)
	}

	result := float64(ca.Unix())

	return result
}

func main() {

	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	validateEnvVars()

	log.Info("Service started on localhost:9191/metrics")
	gitlabLicenseCollector := newGitlabLicenseCollector()

	reg := prometheus.NewRegistry()
	reg.MustRegister(gitlabLicenseCollector)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":9191", nil))
}
