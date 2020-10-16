package exporter

import (
	"time"

	"github.com/nm-morais/deMMon-exporter/types/metrics"
	lv "github.com/nm-morais/deMMon-exporter/types/metrics/utils"
	"github.com/nm-morais/go-babel/pkg/peer"
	"github.com/nm-morais/go-babel/pkg/protocol"
)

type ExporterConf struct {
	ImporterAddr    peer.Peer
	MaxRedials      int
	RedialTimeout   time.Duration
	ExportFrequency time.Duration
}

type Exporter struct {
	proto      protocol.Protocol
	counters   *lv.Space
	gauges     *lv.Space
	histograms *lv.Space
	tags       map[string]string
	confs      ExporterConf
}

func New(confs ExporterConf, tags map[string]string) *Exporter {
	e := &Exporter{
		confs:      confs,
		counters:   lv.NewSpace(),
		gauges:     lv.NewSpace(),
		histograms: lv.NewSpace(),
		tags:       tags,
	}
	e.proto = NewExporterProto(confs, e)
	return e
}

// Proto returns the babel proto of the exporter.
func (e *Exporter) Proto() protocol.Protocol {
	return e.proto
}

// NewCounter returns an Influx counter.
func (e *Exporter) NewCounter(name string) *InfluxCounter {
	return &InfluxCounter{
		name: name,
		obs:  e.counters.Observe,
	}
}

// NewGauge returns an Influx gauge.
func (e *Exporter) NewGauge(name string) *InfluxGauge {
	return &InfluxGauge{
		name: name,
		obs:  e.gauges.Observe,
		add:  e.gauges.Add,
	}
}

// NewHistogram returns an Influx histogram.
func (e *Exporter) NewHistogram(name string) *InfluxHist {
	return &InfluxHist{
		name: name,
		obs:  e.histograms.Observe,
	}
}

// Export flushes the buffered content of the metrics to the writer, in an
// Influx BatchPoints format. WriteTo abides best-effort semantics, so
// observations are lost if there is a problem with the write. Clients should not
// need to explicitely export metrics
func (e *Exporter) Export() (err error) {
	// bp, err := influxdb.NewBatchPoints(e.confs.BpConf)
	// if err != nil {
	// 	return err
	// }

	// now := time.Now()

	// e.counters.Reset().Walk(func(name string, lvs lv.LabelValues, values []float64) bool {
	// 	tags := mergeTags(e.tags, lvs)
	// 	var p *influxdb.Point
	// 	fields := map[string]interface{}{"count": sum(values)}
	// 	p, err = influxdb.NewPoint(name, tags, fields, now)
	// 	if err != nil {
	// 		return false
	// 	}
	// 	bp.AddPoint(p)
	// 	return true
	// })
	// if err != nil {
	// 	return err
	// }

	// e.gauges.Reset().Walk(func(name string, lvs lv.LabelValues, values []float64) bool {
	// 	tags := mergeTags(e.tags, lvs)
	// 	var p *influxdb.Point
	// 	fields := map[string]interface{}{"value": last(values)}
	// 	p, err = influxdb.NewPoint(name, tags, fields, now)
	// 	if err != nil {
	// 		return false
	// 	}
	// 	bp.AddPoint(p)
	// 	return true
	// })
	// if err != nil {
	// 	return err
	// }

	// e.histograms.Reset().Walk(func(name string, lvs lv.LabelValues, values []float64) bool {
	// 	histogram := generic.NewHistogram(name, 50)
	// 	tags := mergeTags(e.tags, lvs)
	// 	var p *influxdb.Point
	// 	for _, v := range values {
	// 		histogram.Observe(v)
	// 	}
	// 	fields := map[string]interface{}{
	// 		"p50": histogram.Quantile(0.50),
	// 		"p90": histogram.Quantile(0.90),
	// 		"p95": histogram.Quantile(0.95),
	// 		"p99": histogram.Quantile(0.99),
	// 	}

	// 	p, err = influxdb.NewPoint(name, tags, fields, now)

	// 	if err != nil {
	// 		return false
	// 	}
	// 	bp.AddPoint(p)
	// 	return true
	// })
	// if err != nil {
	// 	return err
	// }

	// notifErr := pkg.SendNotification(protocoltypes.NewMetricNotification(bp))
	// if notifErr != nil {
	// 	return fmt.Errorf(notifErr.Reason())
	// }
	return
}

func mergeTags(tags map[string]string, labelValues []string) map[string]string {
	if len(labelValues)%2 != 0 {
		panic("mergeTags received a labelValues with an odd number of strings")
	}
	ret := make(map[string]string, len(tags)+len(labelValues)/2)
	for k, v := range tags {
		ret[k] = v
	}
	for i := 0; i < len(labelValues); i += 2 {
		ret[labelValues[i]] = labelValues[i+1]
	}
	return ret
}

func sum(a []float64) float64 {
	var v float64
	for _, f := range a {
		v += f
	}
	return v
}

func last(a []float64) float64 {
	return a[len(a)-1]
}

type observeFunc func(name string, lvs lv.LabelValues, value float64)

// Counter is an Influx counter. Observations are forwarded to an Influx
// object, and aggregated (summed) per timeseries.
type InfluxCounter struct {
	name string
	lvs  lv.LabelValues
	obs  observeFunc
}

// With implements metrics.Counter.
func (c *InfluxCounter) With(labelValues ...string) metrics.Counter {
	return &InfluxCounter{
		name: c.name,
		lvs:  c.lvs.With(labelValues...),
		obs:  c.obs,
	}
}

// Add implements metrics.Counter.
func (c *InfluxCounter) Add(delta float64) {
	c.obs(c.name, c.lvs, delta)
}

// Gauge is an Influx gauge. Observations are forwarded to a Dogstatsd
// object, and aggregated (the last observation selected) per timeseries.
type InfluxGauge struct {
	name string
	lvs  lv.LabelValues
	obs  observeFunc
	add  observeFunc
}

// With implements metrics.Gauge.
func (g *InfluxGauge) With(labelValues ...string) metrics.Gauge {
	return &InfluxGauge{
		name: g.name,
		lvs:  g.lvs.With(labelValues...),
		obs:  g.obs,
		add:  g.add,
	}
}

// Set implements metrics.Gauge.
func (g *InfluxGauge) Set(value float64) {
	g.obs(g.name, g.lvs, value)
}

// Add implements metrics.Gauge.
func (g *InfluxGauge) Add(delta float64) {
	g.add(g.name, g.lvs, delta)
}

// InfluxHist is an Influx histrogram. Observations are aggregated into a
// generic.InfluxHist and emitted as per-quantile gauges to the Influx server.
type InfluxHist struct {
	name string
	lvs  lv.LabelValues
	obs  observeFunc
}

// With implements metrics.Histogram.
func (h *InfluxHist) With(labelValues ...string) metrics.Histogram {
	return &InfluxHist{
		name: h.name,
		lvs:  h.lvs.With(labelValues...),
		obs:  h.obs,
	}
}

// Observe implements metrics.Histogram.
func (h *InfluxHist) Observe(value float64) {
	h.obs(h.name, h.lvs, value)
}
