package main

import (
	"log"
	"time"
)

func (s *server) createTimer(job string) {
	if _, ok := s.timeKeeper[job]; ok {
		log.Print("Reseting timer for job ", job)
		s.timeKeeper[job].Stop()
		delete(s.timeKeeper, job)
	}

	log.Printf("Creating timer for job '%s' with quiet period of %d seconds", job, s.param.proxy.QuietPeriod)

	timer := time.AfterFunc(time.Second*time.Duration(s.param.proxy.QuietPeriod), func() {
		log.Print("Quiet period exceeded for job ", job)
		s.triggerJob(job)
		if _, ok := s.timeKeeper[job]; ok {
			log.Print("Deleting timer for job ", job)
			delete(s.timeKeeper, job)
		}
	})

	s.timeKeeper[job] = timer
	if _, ok := s.timeKeeper[job]; ok {
		log.Print("Timer saved in time keeper")
	}

	return
}

func (s *server) createRefreshJob() {
	ticker := time.NewTicker(s.mappingRefreshInterval)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				err := s.refreshMapping()
				if err != nil {
					log.Print(err)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
