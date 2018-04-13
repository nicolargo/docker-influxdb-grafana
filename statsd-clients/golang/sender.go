package main

import (
	"github.com/cihub/seelog"
	"gopkg.in/alexcesaro/statsd.v2"
	"io"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"
)

type StatsdSender interface {
	io.Closer
	Send() error
}

type statsdSender struct {
	statsdAddr string
	closeChan  chan bool
	stats      *statsd.Client
	wg         *sync.WaitGroup
}

func NewStatsdSender(statsdAddr string) StatsdSender {
	return &statsdSender{
		statsdAddr: statsdAddr,
		closeChan:  make(chan bool),
		wg:         &sync.WaitGroup{},
	}
}

func (s *statsdSender) Send() error {

	var err error
	var host string
	if host, err = os.Hostname(); err != nil {
		seelog.Warn("hostname error: ", err)
		host = "default"
	}

	seelog.Infof(`connecting to statsd at '%s'`, s.statsdAddr)
	s.stats, err = statsd.New(
		statsd.Address(s.statsdAddr),
		statsd.TagsFormat(statsd.InfluxDB),
		statsd.Tags("cluster", "tigcluster", "host", host, "service", "golang-statsd"),
	)
	if err != nil {
		return seelog.Errorf("statsdFeed statsd.New error: %v", err)
	}

	go func() {
		s.wg.Add(1)
		defer s.wg.Done()

		t := time.NewTicker(5 * time.Second)
		defer t.Stop()

		g := newGrowShrinker(0, 100)

		for {
			select {
			case <-t.C:
				seelog.Tracef(`firing stats`)

				cnt := int(math.Floor((rand.NormFloat64() / 0.1) + 50))
				s.stats.Count("mycount", cnt)

				dur := math.Floor(rand.ExpFloat64() / 0.001)
				s.stats.Timing("mytiming", dur)

				qlen := g.sample()
				s.stats.Gauge("myguage", qlen)

				seelog.Debugf(`fired stats - cnt: %d  dur: %f  qlen: %d`, cnt, dur, qlen)

				break

			case <-s.closeChan:
				seelog.Infof(`closing stats timer`)
				return

			}
		}
	}()

	return nil
}

func (s *statsdSender) Close() error {

	if s.stats != nil {
		s.stats.Close()
	}
	s.closeChan <- true
	s.wg.Wait()

	return nil
}

type growShrinker struct {
	dir, val, min, max int
}

func newGrowShrinker(min, max int) *growShrinker {
	return &growShrinker{0, min + (max-min)/2, min, max}
}

func (g *growShrinker) sample() int {

	dirDelta := int(math.Floor((rand.NormFloat64() / 0.5) + .5))

	if (dirDelta < 0) && (g.val <= g.min) {
		g.val = g.min
		g.dir = 0
		return g.val
	}
	if (dirDelta > 0) && (g.val >= g.max) {
		g.val = g.max
		g.dir = 0
		return g.val
	}

	g.dir += dirDelta
	g.val += g.dir
	g.val = int(math.Min(math.Max(float64(g.val), float64(g.min)), float64(g.max)))

	return g.val
}
