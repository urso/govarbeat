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
	wg   sync.WaitGroup
	done chan struct{}

	client publisher.Client

	workers []*worker
}

type worker struct {
	host   string
	name   string
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

	cfg := &config.GovarbeatConfig{}
	err := cfgfile.Read(&cfg, "")
	if err != nil {
		return fmt.Errorf("Error reading config file: %v", err)
	}

	for name, cfg := range cfg.Remotes {
		defaultDuration := 1 * time.Second
		d := defaultDuration
		if cfg.Period != "" {
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

func (bt *Govarbeat) Setup(b *beat.Beat) error {
	return nil
}

func (bt *Govarbeat) Run(b *beat.Beat) error {
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
		stats := make(map[string]float64)
		now := time.Now()
		last := now
		for {
			select {
			case <-w.done:
				return
			case <-ticker.C:
			}

			resp, err := http.Get(fmt.Sprintf("http://%s/debug/vars", w.host))
			if err != nil {
				logp.Info("Failed to retrieve variables: %v", err)
				break
			}

			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				logp.Info("Error reading response: %v", err)
				break
			}

			last = now
			now = time.Now()
			dt := now.Sub(last).Seconds()

			var data map[string]interface{}
			err = json.Unmarshal(body, &data)
			if err != nil {
				logp.Warn("Failed to decode json: %v", err)
				break
			}

			event := common.MapStr{
				"@timestamp": common.Time(now),
				"type":       w.name,
			}
			for name, raw := range data {
				has := true
				var total float64
				switch v := raw.(type) {
				case int:
					total = float64(v)
				case float64:
					total = v
				default:
					has = false
				}
				if !has {
					continue
				}

				if old, ok := stats[name]; ok {
					event[fmt.Sprintf("total_%s", name)] = total
					event[fmt.Sprintf("d_%s", name)] = (total - old) / dt
				}
				stats[name] = total
			}
			client.PublishEvent(event)
		}
	}
}
