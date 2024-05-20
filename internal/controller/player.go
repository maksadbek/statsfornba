package controller

import (
	"context"

	"github.com/maksadbek/statsfornba/internal/model"
	repository "github.com/maksadbek/statsfornba/internal/repository/postgres"
)

type Player struct {
	statsRepo *repository.Stats
}

func NewPlayer(r *repository.Stats) *Player {
	return &Player{
		statsRepo: r,
	}
}

func (p *Player) GetAverageStats(ctx context.Context, player, team, season string) (*model.AverageStat, error) {
	stat, err := p.statsRepo.GetPlayerAverage(ctx, player, team, season)
	if err != nil {
		return nil, err
	}

	return stat, nil
}
