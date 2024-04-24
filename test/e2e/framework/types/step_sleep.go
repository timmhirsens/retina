//go:build e2eframework

package types

import (
	"log"
	"time"
)

type Sleep struct {
	Duration time.Duration
}

func (c *Sleep) Run() error {
	log.Printf("sleeping for %s...\n", c.Duration.String())
	time.Sleep(c.Duration)
	return nil
}

func (c *Sleep) Stop() error {
	return nil
}

func (c *Sleep) Prevalidate() error {
	return nil
}
