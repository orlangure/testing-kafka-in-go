package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/kafka"
	"github.com/orlangure/testing-kafka-in-go/consumer"
	"github.com/orlangure/testing-kafka-in-go/handler"
	"github.com/orlangure/testing-kafka-in-go/producer"
	"github.com/orlangure/testing-kafka-in-go/reporter"
	"github.com/stretchr/testify/require"
)

func TestApp(t *testing.T) {
	container, err := gnomock.Start(
		kafka.Preset(kafka.WithTopics("events")),
		gnomock.WithDebugMode(), gnomock.WithLogWriter(os.Stdout),
		gnomock.WithContainerName("kafka"),
	)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, gnomock.Stop(container))
	}()

	ctx, cancel := context.WithCancel(context.Background())

	p := producer.New(container.Address(kafka.BrokerPort), "events")
	c := consumer.New(container.Address(kafka.BrokerPort), "events")
	r := reporter.New(ctx, c)
	mux := handler.Mux(p, r)

	s := httptest.NewServer(mux)
	rep := getReport(t, s.URL)
	require.Empty(t, rep)

	events := []string{"login", "order"}
	accounts := []string{"a", "b", "c"}

	for i := 0; i < 10; i++ {
		ev := events[i%2]
		acc := accounts[i%3]
		sendEvent(t, s.URL, ev, acc)
	}

	cancel()

	require.NoError(t, p.Close())
	require.NoError(t, c.Close())

	time.Sleep(time.Second * 1)

	rep = getReport(t, s.URL)
	require.NotEmpty(t, rep)

	aStats, bStats, cStats := rep["a"], rep["b"], rep["c"]
	require.Len(t, aStats, 2)
	require.Len(t, bStats, 2)
	require.Len(t, cStats, 2)

	require.Equal(t, 2, aStats["login"])
	require.Equal(t, 1, aStats["order"])

	require.Equal(t, 1, bStats["login"])
	require.Equal(t, 2, bStats["order"])

	require.Equal(t, 1, cStats["login"])
	require.Equal(t, 1, cStats["order"])
}

func getReport(t *testing.T, url string) reporter.Stats {
	res, err := http.Post(url+"/stats", "application/json", nil)
	require.NoError(t, err)

	bs, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	require.Equal(t, http.StatusOK, res.StatusCode, string(bs))

	var s reporter.Stats
	require.NoError(t, json.Unmarshal(bs, &s))

	return s
}

func sendEvent(t *testing.T, url string, tt, id string) {
	buf := bytes.NewBufferString(fmt.Sprintf(`{"type":"%s","account_id":"%s"}`, tt, id))
	res, err := http.Post(url+"/event", "application/json", buf)
	require.NoError(t, err)

	bs, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	require.Equal(t, http.StatusAccepted, res.StatusCode, string(bs))
}
