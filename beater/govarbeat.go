package beater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/urso/govarbeat/config"
)

type Govarbeat struct {
	Configuration *config.Config

	wg   sync.WaitGroup
	done chan struct{}

	client publisher.Client

	workers []*worker
}

type worker struct {
	name   string
	host   string
	period time.Duration
	client *http.Client
	done   <-chan struct{}
}

// Creates beater
func New() *Govarbeat {
	return &Govarbeat{
		done: make(chan struct{}),
	}
}

/// *** Beater interface methods ***///

func (bt *Govarbeat) Config(b *beat.Beat) error {

	// Load beater configuration
	err := cfgfile.Read(&bt.Configuration, "")
	if err != nil {
		return fmt.Errorf("Error reading config file: %v", err)
	}

	return nil
}

func (bt *Govarbeat) Setup(b *beat.Beat) error {
	for name, cfg := range bt.Configuration.Govarbeat.Remotes {
		defaultPeriod := 1 * time.Second
		defaultTimeout := 30 * time.Second

		period, err := configDuration(cfg.Period, defaultPeriod)
		if err != nil {
			return err
		}

		timeout, err := configDuration(cfg.Timeout, defaultTimeout)
		if err != nil {
			return err
		}

		for _, host := range cfg.Hosts {
			bt.workers = append(bt.workers, &worker{
				done:   bt.done,
				host:   host,
				name:   name,
				period: period,
				client: &http.Client{
					Timeout: timeout,
				},
			})
		}
	}

	if len(bt.workers) == 0 {
		return fmt.Errorf("No workers configured")
	}

	return nil
}

func (bt *Govarbeat) Run(b *beat.Beat) error {
	logp.Info("govarbeat is running! Hit CTRL-C to stop it.")

	for _, w := range bt.workers {
		bt.wg.Add(1)
		go func(worker *worker) {
			defer bt.wg.Done()
			worker.run(b.Events)
		}(w)
	}

	bt.wg.Wait()
	return nil
}

func (bt *Govarbeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (bt *Govarbeat) Stop() {
	close(bt.done)
}

func (w *worker) run(client publisher.Client) {
	ticker := time.NewTicker(w.period)
	for {
		stats := map[string]float64{}
		now := time.Now()
		last := now

		for {
			select {
			case <-w.done:
				return
			case <-ticker.C:
			}

			data, err := w.readStats()
			if err != nil {
				logp.Warn("Failed to read vars from %v: %v", w.host, err)
				break
			}

			last = now
			now = time.Now()
			dt := now.Sub(last).Seconds()

			event := common.MapStr{
				"@timestamp": common.Time(now),
				"type":       w.name,
				"remote":     w.host,
			}
			for name, v := range data {
				if old, ok := stats[name]; ok {
					event[fmt.Sprintf("d_%s", name)] = (v - old) / dt
				}
				event[fmt.Sprintf("total_%s", name)] = v
				stats[name] = v
			}
			client.PublishEvent(event)
		}
	}
}

func (w *worker) readStats() (map[string]float64, error) {
	resp, err := w.client.Get(fmt.Sprintf("http://%v/debug/vars", w.host))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var rawData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&rawData)
	if err != nil {
		return nil, err
	}

	data := map[string]float64{}
	for k, v := range rawData {
		switch val := v.(type) {
		case int:
			data[k] = float64(val)
		case float64:
			data[k] = val
		}
	}

	return data, nil
}

func configDuration(cfg string, d time.Duration) (time.Duration, error) {
	if cfg != "" {
		return time.ParseDuration(cfg)
	}
	return d, nil
}
