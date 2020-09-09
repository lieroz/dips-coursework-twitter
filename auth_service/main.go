package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/lieroz/dips-coursework-twitter/tools"
)

var (
	rdb                  *redis.Client
	rdbTimeout           = 100 * time.Millisecond
	jwtExpirationTimeout = 30 * time.Minute
	jwtKey               = []byte("super_secret_key")
	internalService      = "InternalService"
)

type Claims struct {
	Data string `json:"data"`
	jwt.StandardClaims
}

func generateToken(timeout time.Duration, data string) (string, error) {
	expirationTime := time.Now().Add(timeout)
	claims := &Claims{
		Data: data,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func serviceToken(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), rdbTimeout)
	defer cancel()

	var tokenString string

	if t, err := rdb.Get(ctx, internalService).Result(); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		tokenString = t
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if token.Valid {
		io.WriteString(w, tokenString)
		return
	}

	if err != nil {
		if e, ok := err.(*jwt.ValidationError); ok {
			if e.Errors == jwt.ValidationErrorExpired {
				goto expired
			}
		}

		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusBadRequest)
		return
	}

expired:
	tokenString, err = generateToken(jwtExpirationTimeout, "InternalService")
	if err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := rdb.Set(ctx, internalService, tokenString, 0).Err(); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, tokenString)
	}
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	log.Logger = log.With().Caller().Logger()

	var configPath string
	flag.StringVar(&configPath, "config", "compose-conf.json", "config file path")

	flag.Parse()

	err := tools.ParseConfig(configPath)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     tools.Conf.RedisUrl,
		PoolSize: tools.Conf.RedisPoolSize,
	})

	r := mux.NewRouter()
	r.HandleFunc("/service/token", serviceToken).Methods("GET")

	if tokenString, err := generateToken(0, internalService); err != nil {
		log.Fatal().Err(err).Send()
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), rdbTimeout)
		defer cancel()

		if err := rdb.Set(ctx, internalService, tokenString, 0).Err(); err != nil {
			log.Fatal().Err(err).Send()
		}
	}

	listenAddress := fmt.Sprintf("0.0.0.0:%d", tools.Conf.AuthPort)
	log.Info().Msgf("Start listen on port %s", listenAddress)
	if err := http.ListenAndServe(listenAddress, r); err != nil {
		log.Fatal().Err(err).Send()
	}
}
