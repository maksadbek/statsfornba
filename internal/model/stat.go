package model

const (
	StatPoints        = "points"
	StatRebounds      = "rebounds"
	StatAssists       = "assists"
	StatSteals        = "steals"
	StatBlocks        = "blocks"
	StatFouls         = "fouls"
	StatTurnovers     = "turnovers"
	StatMinutesPlayed = "minutes_played"
)

var AllStats = [...]string{
	StatPoints,
	StatRebounds,
	StatAssists,
	StatSteals,
	StatBlocks,
	StatFouls,
	StatTurnovers,
	StatMinutesPlayed,
}

type Stat struct {
	Player string
	Team   string
	Season string

	Points        uint32
	Rebounds      uint32
	Assists       uint32
	Steals        uint32
	Blocks        uint32
	Fouls         uint32
	Turnovers     uint32
	MinutesPlayed uint32
}

type AverageStat struct {
	Points        float32
	Rebounds      float32
	Assists       float32
	Steals        float32
	Blocks        float32
	Fouls         float32
	Turnovers     float32
	MinutesPlayed float32
}
