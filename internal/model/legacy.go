package model

import "time"

type LegacyLoginHandler func(account *AccountDTO, expTime time.Duration) (string, error)
