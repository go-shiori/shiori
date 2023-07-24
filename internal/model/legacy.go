package model

import "time"

type LegacyLoginHandler func(account Account, expTime time.Duration) (string, error)
