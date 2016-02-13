package beater

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
		defaultDuration := 1 * time.Second
		d := defaultDuration
		if cfg.Period != "" {
			var err error
			d, err = time.ParseDuration(cfg.Period)
			if err != nil {
				return err
			}
		}

		bt.workers = append(bt.workers, &worker{
			done:   bt.done,
			host:   cfg.Host,
			name:   name,
			period: d,
		})
	}

	if len(bt.workers) == 0 {
		return fmt.Errorf("No workers configured")
	}

	return nil
}

func (bt *Govarbeat) Run(b *beat.Beat) error {
	logp.Info("demobeat is running! Hit CTRL-C to stop it.")

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

			data, err := readStats(w.host)
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

func readStats(host string) (map[string]float64, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v/debug/vars", host))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var rawData map[string]interface{}
	err = json.Unmarshal(body, &rawData)
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
