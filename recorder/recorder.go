package recorder

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/webmakersteve/elastic_prom_recorder/config"
	"strings"
	"time"
)

type recordingRule struct {
	gauge prometheus.Gauge
	query *strings.Reader
	name  string
}

type Group struct {
	name          string
	elasticClient *elasticsearch.Client
	fields        *log.Fields
	rules         []*recordingRule
	duration      time.Duration
	index         string
}

func NewGroup(group *config.Group) (Group, error) {
	var g = Group{
		name: group.Name,
	}

	cfg := elasticsearch.Config{
		Addresses: group.Elasticsearch.Addresses,
	}

	if group.Elasticsearch.Username != nil {
		cfg.Username = *group.Elasticsearch.Username
	}

	if group.Elasticsearch.Password != nil {
		cfg.Password = *group.Elasticsearch.Password
	}

	es, err := elasticsearch.NewClient(cfg)

	if err != nil {
		return g, err
	}

	g.fields = &log.Fields{
		"name": group.Name,
	}
	g.elasticClient = es

	rulesArray := make([]*recordingRule, len(group.Rules))

	duration, err := time.ParseDuration(group.Interval)

	if err != nil {
		return g, err
	}

	g.duration = duration
	g.index = group.Elasticsearch.Index

	for i, rule := range group.Rules {
		labels := prometheus.Labels{
			"recording_group": group.Name,
		}

		for k, v := range rule.Labels {
			labels[k] = v
		}

		name := rule.Record + ":" + group.Interval
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        name,
			ConstLabels: labels,
			Help:        "The total number of processed events",
		})

		if !json.Valid([]byte(rule.Query)) {
			return g, errors.New("JSON for query is invalid")
		}

		rulesArray[i] = &recordingRule{
			gauge: gauge,
			query: strings.NewReader(rule.Query),
			name:  name,
		}
	}

	g.rules = rulesArray

	return g, nil
}

func Execute(group *Group, ctx context.Context) {
	logger := log.WithFields(*group.fields)

	for true {
		logger.Info("Running for group")

		for _, r := range group.rules {
			// Execute the recording rule
			ruleLogger := logger.WithFields(log.Fields{
				"rule": r.name,
			})

			ruleLogger.Info("Executing rule")

			err := executeRule(group, r)

			if err != nil {
				ruleLogger.Warnf("Failed to execute rule: %s", err)
			}
		}

		time.Sleep(group.duration)
	}

	<-ctx.Done()
}

func executeRule(group *Group, rule *recordingRule) error {
	es := group.elasticClient

	result, err := es.Search(
		es.Search.WithBody(rule.query),
		es.Search.WithIndex(group.index),
	)

	if err != nil {
		return err
	}

	defer result.Body.Close()

	response := result.String()

	log.Info(response)

	// Parse out the body and set the gauge
	rule.gauge.Set(1)

	return nil
}
