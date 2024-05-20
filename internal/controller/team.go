package controller

import (
	"context"
	"fmt"

	"github.com/maksadbek/statsfornba/internal/model"
	repository "github.com/maksadbek/statsfornba/internal/repository/postgres"
)

type Team struct {
	statsRepo *repository.Stats
}

func NewTeam(r *repository.Stats) *Team {
	return &Team{
		statsRepo: r,
	}
}

func (t *Team) GetAverageStats(ctx context.Context, team, season string) (*model.AverageStat, error) {
	fmt.Println(">>>>>>>>>>>", t)

	stat, err := t.statsRepo.GetTeamAverage(ctx, team, season)
	if err != nil {
		return nil, err
	}

	return stat, nil
}
