package app

import (
	"context"
	"database/sql"
	"encoding/csv"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"

	"github.com/maksadbek/statsfornba/internal/controller"
	"github.com/maksadbek/statsfornba/internal/model"
	repository "github.com/maksadbek/statsfornba/internal/repository/postgres"
	"github.com/maksadbek/statsfornba/pkg/platform/kafka"
)

type Config struct {
	HTTPServer struct {
		Addr string `default:":8080"`
	}

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
		c.String(http.StatusOK, "ok")
	})

	r.GET("/:team/:season", a.TeamStatsHandler())
	r.GET("/:team/:season/:player", a.PlayerStatsHandler())
	r.POST("/upload", a.UploadStatsHandler())

	srv := &http.Server{
		Addr:    c.HTTPServer.Addr,
		Handler: r.Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server Shutdown:", err)
	}

	<-ctx.Done()
	log.Println("timeout of 5 seconds, server exiting")

	return nil
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
			log.Println("failed to open file", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		defer f.Close()

		csvReader := csv.NewReader(f)

		head, err := csvReader.Read()
		if err != nil {
			log.Println("failed to read the first row from csv file", err)
			c.AbortWithError(http.StatusBadRequest, err)

			return
		}

		idx := map[string]int{}
		for i, h := range head {
			idx[h] = i
		}

		for {
			rec, err := csvReader.Read()
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err)

				break
			}

			team := rec[idx["team"]]
			player := rec[idx["player"]]
			season := rec[idx["season"]]

			if team == "" || player == "" || season == "" {
				log.Println("team, player or season cannot be empty", rec)
				continue
			}

			points, err := strconv.Atoi(rec[idx["points"]])
			if err != nil {
				log.Println("invalid points value, skipping", err, rec)
				continue
			}

			rebounds, err := strconv.Atoi(rec[idx["rebounds"]])
			if err != nil {
				log.Println("invalid rebounds value, skipping", err, rec)
				continue
			}

			assists, err := strconv.Atoi(rec[idx["assists"]])
			if err != nil {
				log.Println("invalid assists value, skipping", err, rec)
				continue
			}

			steals, err := strconv.Atoi(rec[idx["steals"]])
			if err != nil {
				log.Println("invalid steals value, skipping", err, rec)
				continue
			}

			blocks, err := strconv.Atoi(rec[idx["blocks"]])
			if err != nil {
				log.Println("invalid blocks value, skipping", err, rec)
				continue
			}

			fouls, err := strconv.Atoi(rec[idx["fouls"]])
			if err != nil {
				log.Println("invalid fouls value, skipping", err, rec)
				continue
			}

			if fouls < 0 || fouls > 6 {
				log.Println("invalid fouls value, must be between 0 and 6, skipping", err, rec)
				continue
			}

			turnovers, err := strconv.Atoi(rec[idx["turnovers"]])
			if err != nil {
				log.Println("invalid turnovers value, skipping", err, rec)
				continue
			}

			minutesPlayed, err := strconv.ParseFloat(rec[idx["minutes played"]], 32)
			if err != nil {
				log.Println("invalid turnovers value, skipping", err, rec)
				continue
			}

			minutes, seconds := math.Modf(minutesPlayed)

			if points < 0 ||
				rebounds < 0 ||
				blocks < 0 ||
				turnovers < 0 ||
				minutesPlayed < 0 || minutesPlayed > 48.0 ||
				steals < 0 ||
				assists < 0 {
				log.Println("invalid values, stats must be positive and minutes played must be less than 48.0", rec)
				continue
			}

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
				SecondsPlayed: uint32(minutes)*60 + uint32(seconds*10),
			}

			log.Println("publishing the event to kafka", stat)

			err = a.stats.Add(&stat)
			if err != nil {
				log.Println("failed to publish", err)
				continue
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
			c.AbortWithError(http.StatusInternalServerError, err)
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
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}
