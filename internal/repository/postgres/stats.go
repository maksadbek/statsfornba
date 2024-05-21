package repository

import (
	"context"
	"database/sql"

	"github.com/maksadbek/statsfornba/internal/model"
)

type Stats struct {
	db *sql.DB
}

func NewStats(db *sql.DB) (*Stats, error) {
	return &Stats{
		db: db,
	}, nil
}

func (s *Stats) Add(ctx context.Context, stat *model.Stat) error {
	query := `insert into player_avg (name, team, season, points, rebounds, assists, steals, blocks, fouls, turnovers, seconds_played)
            values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
            on conflict (name, team, season)
            do update set
                points = player_avg.points + EXCLUDED.points,
                rebounds = player_avg.rebounds + EXCLUDED.rebounds,
                assists = player_avg.assists + EXCLUDED.assists,
                steals = player_avg.steals + EXCLUDED.steals,
                blocks = player_avg.blocks + EXCLUDED.blocks,
                fouls = player_avg.fouls + EXCLUDED.fouls,
                turnovers = player_avg.turnovers + EXCLUDED.turnovers,
                seconds_played = player_avg.seconds_played + EXCLUDED.seconds_played,
                game_count = player_avg.game_count + 1`

	_, err := s.db.ExecContext(ctx, query,
		stat.Player,
		stat.Team,
		stat.Season,
		stat.Points,
		stat.Rebounds,
		stat.Assists,
		stat.Steals,
		stat.Blocks,
		stat.Fouls,
		stat.Turnovers,
		stat.SecondsPlayed,
	)

	if err != nil {
		return err
	}

	queryTeam := `insert into team_avg (name, season, points, rebounds, assists, steals, blocks, fouls, turnovers, seconds_played)
    values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    on conflict (name, season)
    do update set
        points = team_avg.points + EXCLUDED.points,
        rebounds = team_avg.rebounds + EXCLUDED.rebounds,
        assists = team_avg.assists + EXCLUDED.assists,
        steals = team_avg.steals + EXCLUDED.steals,
        blocks = team_avg.blocks + EXCLUDED.blocks,
        fouls = team_avg.fouls + EXCLUDED.fouls,
        turnovers = team_avg.turnovers + EXCLUDED.turnovers,
        seconds_played = team_avg.seconds_played + EXCLUDED.seconds_played,
        game_count = team_avg.game_count + 1`

	_, err = s.db.ExecContext(ctx, queryTeam,
		stat.Team,
		stat.Season,
		stat.Points,
		stat.Rebounds,
		stat.Assists,
		stat.Steals,
		stat.Blocks,
		stat.Fouls,
		stat.Turnovers,
		stat.SecondsPlayed,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *Stats) GetTeamAverage(ctx context.Context, team, season string) (*model.AverageStat, error) {
	query := `select points, rebounds, assists, steals, blocks, fouls, turnovers, seconds_played, game_count from team_avg where name = $1 and season = $2`

	var points, rebounds, assists, steals, fouls, turnovers, secondsPlayed, blocks, gameCount int
	err := s.db.QueryRowContext(ctx, query, team, season).Scan(
		&points,
		&rebounds,
		&assists,
		&steals,
		&blocks,
		&fouls,
		&turnovers,
		&secondsPlayed,
		&gameCount,
	)
	if err != nil {
		return nil, err
	}

	averageStat := model.AverageStat{
		Points:        float32(points) / float32(gameCount),
		Assists:       float32(assists) / float32(gameCount),
		Rebounds:      float32(rebounds) / float32(gameCount),
		Steals:        float32(steals) / float32(gameCount),
		Blocks:        float32(blocks) / float32(gameCount),
		Fouls:         float32(fouls) / float32(gameCount),
		Turnovers:     float32(turnovers) / float32(gameCount),
		MinutesPlayed: float32(secondsPlayed) / float32(60) / float32(gameCount),
	}

	return &averageStat, nil
}

func (s *Stats) GetPlayerAverage(ctx context.Context, player, team, season string) (*model.AverageStat, error) {
	query := `select points, rebounds, assists, steals, blocks, fouls, turnovers, seconds_played, game_count from player_avg where name = $1 and team = $2 and season = $3`

	var points, rebounds, assists, steals, fouls, turnovers, secondsPlayed, blocks, gameCount int
	err := s.db.QueryRowContext(ctx, query, player, team, season).Scan(
		&points,
		&rebounds,
		&assists,
		&steals,
		&blocks,
		&fouls,
		&turnovers,
		&secondsPlayed,
		&gameCount,
	)
	if err != nil {
		return nil, err
	}

	averageStat := model.AverageStat{
		Points:        float32(points) / float32(gameCount),
		Assists:       float32(assists) / float32(gameCount),
		Rebounds:      float32(rebounds) / float32(gameCount),
		Steals:        float32(steals) / float32(gameCount),
		Blocks:        float32(blocks) / float32(gameCount),
		Fouls:         float32(fouls) / float32(gameCount),
		Turnovers:     float32(turnovers) / float32(gameCount),
		MinutesPlayed: float32(secondsPlayed) / float32(60) / float32(gameCount),
	}

	return &averageStat, nil
}
