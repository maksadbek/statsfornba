package app

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"

	"github.com/maksadbek/statsfornba/internal/controller"
	"github.com/maksadbek/statsfornba/internal/model"
	repository "github.com/maksadbek/statsfornba/internal/repository/postgres"
	"github.com/maksadbek/statsfornba/pkg/platform/kafka"
)

type Config struct {
	Kafka struct {
		Addr  []string `required:"true"`
		Topic string   `required:"true"`
	}

	Postgres struct {
		DSN string `required:"true"`
	}
}

type App struct {
	playerController *controller.Player
	teamController   *controller.Team
	stats            *controller.Stats
}

func Run() error {
	var c Config
	err := envconfig.Process("api", &c)
	if err != nil {
		return err
	}

	conn, err := sql.Open("postgres", c.Postgres.DSN)
	if err != nil {
		return err
	}

	statsRepo, err := repository.NewStats(conn)
	if err != nil {
		return err
	}

	publisher, err := kafka.NewPublisher(c.Kafka.Addr, c.Kafka.Topic, "api")
	if err != nil {
		return err
	}

	a := App{
		playerController: controller.NewPlayer(statsRepo),
		teamController:   controller.NewTeam(statsRepo),
		stats:            controller.NewStats(publisher),
	}

	r := gin.Default()

	r.GET("/status", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.GET("/:team/:season", a.TeamStatsHandler())
	r.GET("/:team/:season/:player", a.PlayerStatsHandler())
	r.POST("/upload", a.UploadStatsHandler())

	return r.Run(":8080")
}

func (a *App) UploadStatsHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, "get form error: %s", err)
			return
		}

		f, err := file.Open()
		if err != nil {
			fmt.Println(err)
			return
		}

		defer f.Close()

		csvReader := csv.NewReader(f)

		head, err := csvReader.Read()
		if err != nil {
			fmt.Println(err)

			return
		}

		idx := map[string]int{}

		for i, h := range head {
			idx[h] = i
		}

		for {
			rec, err := csvReader.Read()
			if err != nil {
				break
			}

			team := rec[idx["team"]]
			player := rec[idx["player"]]
			season := rec[idx["season"]]
			points, _ := strconv.Atoi(rec[idx["points"]])
			rebounds, _ := strconv.Atoi(rec[idx["rebounds"]])
			assists, _ := strconv.Atoi(rec[idx["assists"]])
			steals, _ := strconv.Atoi(rec[idx["steals"]])
			blocks, _ := strconv.Atoi(rec[idx["blocks"]])
			fouls, _ := strconv.Atoi(rec[idx["fouls"]])
			turnovers, _ := strconv.Atoi(rec[idx["turnovers"]])
			minutesPlayed, _ := strconv.Atoi(rec[idx["minutes played"]])

			stat := model.Stat{
				Team:          team,
				Season:        season,
				Player:        player,
				Points:        uint32(points),
				Rebounds:      uint32(rebounds),
				Assists:       uint32(assists),
				Steals:        uint32(steals),
				Blocks:        uint32(blocks),
				Fouls:         uint32(fouls),
				Turnovers:     uint32(turnovers),
				MinutesPlayed: uint32(minutesPlayed),
			}

			fmt.Println("publish", stat)
			err = a.stats.Add(&stat)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (a *App) PlayerStatsHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		team := c.Param("team")
		season := c.Param("season")
		player := c.Param("player")

		stats, err := a.playerController.GetAverageStats(c.Request.Context(), player, team, season)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

func (a *App) TeamStatsHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		team := c.Param("team")
		season := c.Param("season")

		stats, err := a.teamController.GetAverageStats(c.Request.Context(), team, season)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}
