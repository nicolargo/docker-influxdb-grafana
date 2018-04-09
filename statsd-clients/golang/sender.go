package main

import (
	"github.com/cihub/seelog"
	"gopkg.in/alexcesaro/statsd.v2"
	"io"
	"math"
	"math/rand"
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

	seelog.Infof(`connecting to statsd at '%s'`, s.statsdAddr)
	var err error
	s.stats, err = statsd.New(
		statsd.Address(s.statsdAddr),
		statsd.TagsFormat(statsd.InfluxDB),
		statsd.Tags("host", "pgoMac", "service", "golang-statsd"),
	)
	if err != nil {
		return seelog.Errorf("statsdFeed statsd.New error: %v", err)
	}

	go func() {
		s.wg.Add(1)
		defer s.wg.Done()

		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				seelog.Tracef(`firing stats`)

				cnt := int(math.Floor(rand.ExpFloat64() / 100.0))
				s.stats.Count("mycount", cnt)

				dur := int(math.Floor(rand.ExpFloat64() / 50.0))
				s.stats.Timing("mytiming", dur)

				seelog.Debugf(`fired stats - cnt: %d  dur: %d`, cnt, dur)

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
