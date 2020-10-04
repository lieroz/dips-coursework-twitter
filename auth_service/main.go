package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
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
	rdb                         *redis.Client
	rdbTimeout                  = 100 * time.Millisecond
	serviceJwtExpirationTimeout = 30 * time.Minute
	clientJwtExpirationTimeout  = 5 * time.Minute
	jwtKey                      = []byte("super_secret_key")
	internalService             = "InternalService"
)

type Claims struct {
	Data string `json:"data"`
	jwt.StandardClaims
}

func generateToken(timeout time.Duration, data string) (time.Time, string, error) {
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
		return expirationTime, "", err
	}

	return expirationTime, tokenString, nil
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
	_, tokenString, err = generateToken(serviceJwtExpirationTimeout, "InternalService")
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

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func signUp(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	creds := &UserCredentials{}
	if err = json.Unmarshal(body, &creds); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), rdbTimeout)
	defer cancel()

	if exists, err := rdb.Exists(ctx, creds.Username).Result(); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		if exists == 1 {
			log.Error().Msgf("user '%s' is already signed up", creds.Username)
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	expirationTime, token, err := generateToken(clientJwtExpirationTimeout, creds.Username)
	if err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := rdb.Set(ctx, creds.Username, creds.Password, 0).Err(); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: expirationTime,
	})
}

func signIn(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	creds := &UserCredentials{}
	if err = json.Unmarshal(body, &creds); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), rdbTimeout)
	defer cancel()

	expectedPassword, err := rdb.Get(ctx, creds.Username).Result()
	if err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if expectedPassword != creds.Password {
		log.Error().Msgf("%s passwords don't match", creds.Username)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	expirationTime, token, err := generateToken(clientJwtExpirationTimeout, creds.Username)
	if err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: expirationTime,
	})
}

func refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tknStr := c.Value
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), rdbTimeout)
	defer cancel()

	if exists, err := rdb.Exists(ctx, claims.Data).Result(); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		if exists == 0 {
			log.Error().Msgf("user '%s' is not in storage", claims.Data)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 10*time.Second {
		w.WriteHeader(http.StatusOK)
		return
	}

	expirationTime := time.Now().Add(clientJwtExpirationTimeout)
	claims.ExpiresAt = expirationTime.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   tokenString,
		Expires: expirationTime,
	})
	w.WriteHeader(http.StatusCreated)
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	log.Logger = log.With().Caller().Logger()

	var configPath string
	flag.StringVar(&configPath, "config", "conf.json", "config file path")

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
	r.HandleFunc("/signup", signUp).Methods("POST")
	r.HandleFunc("/signin", signIn).Methods("POST")
	r.HandleFunc("/refresh", refresh).Methods("GET")

	if _, tokenString, err := generateToken(0, internalService); err != nil {
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
