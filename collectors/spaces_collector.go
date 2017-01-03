package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type SpacesCollector struct {
	namespace                             string
	cfClient                              *cfclient.Client
	spaceInfoMetric                       *prometheus.GaugeVec
	spacesTotalMetric                     prometheus.Gauge
	lastSpacesScrapeErrorMetric           prometheus.Gauge
	lastSpacesScrapeTimestampMetric       prometheus.Gauge
	lastSpacesScrapeDurationSecondsMetric prometheus.Gauge
}

func NewSpacesCollector(namespace string, cfClient *cfclient.Client) *SpacesCollector {
	spaceInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "space",
			Name:      "info",
			Help:      "Labeled Cloud Foundry Space information with a constant '1' value.",
		},
		[]string{"space_id", "space_name"},
	)

	spacesTotalMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "spaces",
			Name:      "total",
			Help:      "Total number of Cloud Foundry Spaces.",
		},
	)

	lastSpacesScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_spaces_scrape_error",
			Help:      "Whether the last scrape of Spaces metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
	)

	lastSpacesScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_spaces_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Spaces metrics from Cloud Foundry.",
		},
	)

	lastSpacesScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_spaces_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Spaces metrics from Cloud Foundry.",
		},
	)

	return &SpacesCollector{
		namespace:                             namespace,
		cfClient:                              cfClient,
		spaceInfoMetric:                       spaceInfoMetric,
		spacesTotalMetric:                     spacesTotalMetric,
		lastSpacesScrapeErrorMetric:           lastSpacesScrapeErrorMetric,
		lastSpacesScrapeTimestampMetric:       lastSpacesScrapeTimestampMetric,
		lastSpacesScrapeDurationSecondsMetric: lastSpacesScrapeDurationSecondsMetric,
	}
}

func (c SpacesCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	c.spaceInfoMetric.Reset()

	spaces, err := c.cfClient.ListSpaces()
	if err != nil {
		log.Errorf("Error while listing spaces: %v", err)
		c.reportErrorMetric(true, ch)
		return
	}

	for _, space := range spaces {
		c.spaceInfoMetric.WithLabelValues(
			space.Guid,
			space.Name,
		).Set(float64(1))
	}

	c.spaceInfoMetric.Collect(ch)

	c.spacesTotalMetric.Set(float64(len(spaces)))
	c.spacesTotalMetric.Collect(ch)

	c.lastSpacesScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastSpacesScrapeTimestampMetric.Collect(ch)

	c.lastSpacesScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastSpacesScrapeDurationSecondsMetric.Collect(ch)

	c.reportErrorMetric(false, ch)
}

func (c SpacesCollector) Describe(ch chan<- *prometheus.Desc) {
	c.spaceInfoMetric.Describe(ch)
	c.spacesTotalMetric.Describe(ch)
	c.lastSpacesScrapeErrorMetric.Describe(ch)
	c.lastSpacesScrapeTimestampMetric.Describe(ch)
	c.lastSpacesScrapeDurationSecondsMetric.Describe(ch)
}

func (c SpacesCollector) reportErrorMetric(errorHappend bool, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if errorHappend {
		errorMetric = float64(1)
	}

	c.lastSpacesScrapeErrorMetric.Set(errorMetric)
	c.lastSpacesScrapeErrorMetric.Collect(ch)
}
