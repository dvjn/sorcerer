package main

import (
	"fmt"
	"net/http"

	"github.com/dvjn/sorcerer/internal/api"
	"github.com/dvjn/sorcerer/internal/auth"
	"github.com/dvjn/sorcerer/internal/config"
	"github.com/dvjn/sorcerer/internal/distribution"
	"github.com/dvjn/sorcerer/internal/logger"
	"github.com/dvjn/sorcerer/internal/store"
	"github.com/rs/zerolog/log"
)

func main() {
	logger.Initialize()
	log.Info().Msg("starting sorcerer")

	config, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	logger.Configure(&config.Log)
	log.Debug().Msg("initialized config")
	log.Trace().Interface("config", config).Send()

	errors := config.Validate()
	if len(errors) > 0 {
		log.Fatal().Errs("errors", errors).Msg("config validations failed")
	}

	auth, err := auth.New(&config.Auth, &log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize auth")
	}
	log.Debug().Msg("initialized auth")

	store, err := store.New(&config.Store)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize store")
	}
	log.Debug().Msg("initialized store")

	distribution := distribution.New(store, auth.DistributionMiddleware())
	log.Debug().Msg("initialized distribution")

	api := api.New(distribution.Router(), auth.Router())
	log.Debug().Msg("initialized api")

	log.Info().Msgf("listening on port %d", config.Server.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.Server.Port), api.Router())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}
