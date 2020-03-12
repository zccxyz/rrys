package model

import (
	"time"
)

type Movies struct {
	Vid            uint64
	Cnname         string
	Enname         string
	Channel        string
	Area           string
	Category       string
	Tvstation      string
	Lang           string
	PlayStatus     string
	Rank           uint64
	Views          uint64
	Score          float64
	PublishYear    uint64
	Itemupdate     uint64
	Poster         string
	FavoriteStatus uint64
	Season         uint64
	Episode        uint64
	Premiere       string
	Zimuzu         string
	Aliasname      string
	ScoreCounts    uint64
	Content        string
	CloseResource  uint64
	Website        string
	Level          string
	Director       string
	Writer         string
	Actor          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
