package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/examples/data"
	status "google.golang.org/grpc/status"

	pb "github.com/lieroz/dips-coursework-twitter/protos"
	"github.com/lieroz/dips-coursework-twitter/tools"
)

var (
	usersClient  pb.UsersClient
	usersConn    *grpc.ClientConn
	tweetsClient pb.TweetsClient
	tweetsConn   *grpc.ClientConn
)

const (
	timeout = 200 * time.Millisecond
)

/*
 TODO:
 1. add oauth
*/

type UserSignUpRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Firstname   string `json:"firstname"`
	Lastname    string `json:"lastname"`
	Description string `json:"description"`
}

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	userRequest := &UserSignUpRequest{}
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, userRequest); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		authRequest := &UserCredentials{
			Username: userRequest.Username,
			Password: userRequest.Password,
		}

		body, err = json.Marshal(authRequest)
		if err != nil {
			log.Error().Err(err).Send()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp, err := http.Post(fmt.Sprintf("http://%s:%d/signup",
			tools.Conf.AuthServiceHost, tools.Conf.AuthPort), "application/json",
			bytes.NewReader(body))
		if err != nil {
			log.Error().Err(err).Send()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var tokenCookie *http.Cookie = nil
		cookies := resp.Cookies()

		for _, cookie := range cookies {
			if cookie.Name == "token" {
				tokenCookie = cookie
				break
			}
		}

		if tokenCookie == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		msgProto := &pb.CreateRequest{
			Username:    userRequest.Username,
			Firstname:   userRequest.Firstname,
			Lastname:    userRequest.Lastname,
			Description: userRequest.Description,
		}

		if _, err := usersClient.CreateUser(ctx, msgProto); err != nil {
			s, _ := status.FromError(err)

			switch s.Code() {
			case codes.InvalidArgument:
				w.WriteHeader(http.StatusBadRequest)
			case codes.AlreadyExists:
				w.WriteHeader(http.StatusConflict)
			case codes.Internal:
				w.WriteHeader(http.StatusInternalServerError)
			case codes.DeadlineExceeded:
				w.WriteHeader(http.StatusRequestTimeout)
			case codes.Unauthenticated:
				w.WriteHeader(http.StatusUnauthorized)
				reconnect()
			}
			io.WriteString(w, s.Message())
		} else {
			http.SetCookie(w, tokenCookie)
		}
	}
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post(fmt.Sprintf("http://%s:%d/signin",
		tools.Conf.AuthServiceHost, tools.Conf.AuthPort), "application/json",
		r.Body)
	if err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var tokenCookie *http.Cookie = nil
	cookies := resp.Cookies()

	for _, cookie := range cookies {
		if cookie.Name == "token" {
			tokenCookie = cookie
			break
		}
	}

	if tokenCookie == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, tokenCookie)
}

func checkToken(r *http.Request) bool {
	c, err := r.Cookie("token")
	if err != nil {
		log.Error().Err(err).Send()
		return false
	}

	token := c.Value
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/refresh",
		tools.Conf.AuthServiceHost, tools.Conf.AuthPort), nil)
	if err != nil {
		log.Error().Err(err).Send()
		return false
	}

	req.AddCookie(&http.Cookie{Name: "token", Value: token})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Send()
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotModified {
		return false
	}

	return true
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	if !checkToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	msgProto := &pb.DeleteRequest{}
	if err := jsonpb.Unmarshal(r.Body, msgProto); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if _, err := usersClient.DeleteUser(ctx, msgProto); err != nil {
			s, _ := status.FromError(err)
			switch s.Code() {
			case codes.InvalidArgument:
				w.WriteHeader(http.StatusBadRequest)
			case codes.Internal:
				w.WriteHeader(http.StatusInternalServerError)
			case codes.DeadlineExceeded:
				w.WriteHeader(http.StatusRequestTimeout)
			case codes.Unauthenticated:
				w.WriteHeader(http.StatusUnauthorized)
				reconnect()
			}
			io.WriteString(w, s.Message())
		}
	}
}

func GetUserInfoSummary(w http.ResponseWriter, r *http.Request) {
	if !checkToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	msgProto := &pb.GetSummaryRequest{}
	if err := jsonpb.Unmarshal(r.Body, msgProto); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if in, err := usersClient.GetUserInfoSummary(ctx, msgProto); err != nil {
			s, _ := status.FromError(err)
			switch s.Code() {
			case codes.InvalidArgument:
				w.WriteHeader(http.StatusBadRequest)
			case codes.NotFound:
				w.WriteHeader(http.StatusNotFound)
			case codes.Internal:
				w.WriteHeader(http.StatusInternalServerError)
			case codes.DeadlineExceeded:
				w.WriteHeader(http.StatusRequestTimeout)
			case codes.Unauthenticated:
				w.WriteHeader(http.StatusUnauthorized)
				reconnect()
			}
			io.WriteString(w, s.Message())
		} else {
			m := jsonpb.Marshaler{}
			m.Marshal(w, in)
		}
	}
}

func GetFollowers(w http.ResponseWriter, r *http.Request) {
	if !checkToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	msgProto := &pb.GetUsersRequest{}
	if err := jsonpb.Unmarshal(r.Body, msgProto); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		msgProto.Role = pb.Role_Follower
		if stream, err := usersClient.GetUsers(ctx, msgProto); err != nil {
			s, _ := status.FromError(err)
			switch s.Code() {
			case codes.InvalidArgument:
				w.WriteHeader(http.StatusBadRequest)
			case codes.Internal:
				w.WriteHeader(http.StatusInternalServerError)
			case codes.DeadlineExceeded:
				w.WriteHeader(http.StatusRequestTimeout)
			case codes.Unauthenticated:
				w.WriteHeader(http.StatusUnauthorized)
				reconnect()
			}
			io.WriteString(w, s.Message())
		} else {
			m := jsonpb.Marshaler{}
			first := true

			for {
				in, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						break
					} else {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}

				if !in.GetError() {
					if first {
						io.WriteString(w, "{")
						first = false
					} else {
						io.WriteString(w, ",")
					}

					m.Marshal(w, in.GetUser())
				}
			}

			io.WriteString(w, "}")
		}
	}
}

func GetFollowing(w http.ResponseWriter, r *http.Request) {
	if !checkToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	msgProto := &pb.GetUsersRequest{}
	if err := jsonpb.Unmarshal(r.Body, msgProto); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		msgProto.Role = pb.Role_Followed
		if stream, err := usersClient.GetUsers(ctx, msgProto); err != nil {
			s, _ := status.FromError(err)
			switch s.Code() {
			case codes.InvalidArgument:
				w.WriteHeader(http.StatusBadRequest)
			case codes.Internal:
				w.WriteHeader(http.StatusInternalServerError)
			case codes.DeadlineExceeded:
				w.WriteHeader(http.StatusRequestTimeout)
			case codes.Unauthenticated:
				w.WriteHeader(http.StatusUnauthorized)
				reconnect()
			}
			io.WriteString(w, s.Message())
		} else {
			m := jsonpb.Marshaler{}
			first := true

			for {
				in, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						break
					} else {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}

				if !in.GetError() {
					if first {
						io.WriteString(w, "{")
						first = false
					} else {
						io.WriteString(w, ",")
					}

					m.Marshal(w, in.GetUser())
				}
			}

			io.WriteString(w, "}")
		}
	}
}

func GetUserTimeline(w http.ResponseWriter, r *http.Request) {
	if !checkToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userProto := &pb.GetTimelineRequest{}
	if err := jsonpb.Unmarshal(r.Body, userProto); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if in, err := usersClient.GetTimeline(ctx, userProto); err != nil {
			s, _ := status.FromError(err)
			switch s.Code() {
			case codes.InvalidArgument:
				w.WriteHeader(http.StatusBadRequest)
			case codes.Internal:
				w.WriteHeader(http.StatusInternalServerError)
			case codes.DeadlineExceeded:
				w.WriteHeader(http.StatusRequestTimeout)
			case codes.Unauthenticated:
				w.WriteHeader(http.StatusUnauthorized)
				reconnect()
			}
			io.WriteString(w, s.Message())
		} else {
			if in.GetTimeline() != nil {
				tweetsProto := &pb.GetTweetsRequest{Tweets: in.Timeline}
				if stream, err := tweetsClient.GetTweets(ctx, tweetsProto); err != nil {
					s, _ := status.FromError(err)
					switch s.Code() {
					case codes.InvalidArgument:
						w.WriteHeader(http.StatusBadRequest)
					case codes.Internal:
						w.WriteHeader(http.StatusInternalServerError)
					case codes.DeadlineExceeded:
						w.WriteHeader(http.StatusRequestTimeout)
					case codes.Unauthenticated:
						w.WriteHeader(http.StatusUnauthorized)
						reconnect()
					}
					io.WriteString(w, s.Message())
				} else {
					m := jsonpb.Marshaler{}
					first := true

					for {
						in, err := stream.Recv()
						if err != nil {
							if err == io.EOF {
								break
							} else {
								w.WriteHeader(http.StatusInternalServerError)
								return
							}
						}

						if !in.GetError() {
							if first {
								io.WriteString(w, "[")
								first = false
							} else {
								io.WriteString(w, ",")
							}

							m.Marshal(w, in.GetTweet())
						}
					}

					io.WriteString(w, "]")
				}
			}
		}
	}
}

func Follow(w http.ResponseWriter, r *http.Request) {
	if !checkToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	msgProto := &pb.FollowRequest{}
	if err := jsonpb.Unmarshal(r.Body, msgProto); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if _, err := usersClient.Follow(ctx, msgProto); err != nil {
			s, _ := status.FromError(err)
			switch s.Code() {
			case codes.InvalidArgument:
				w.WriteHeader(http.StatusBadRequest)
			case codes.Canceled:
				w.WriteHeader(http.StatusConflict)
			case codes.NotFound:
				w.WriteHeader(http.StatusConflict)
			case codes.Internal:
				w.WriteHeader(http.StatusInternalServerError)
			case codes.DeadlineExceeded:
				w.WriteHeader(http.StatusRequestTimeout)
			case codes.Unauthenticated:
				w.WriteHeader(http.StatusUnauthorized)
				reconnect()
			}
			io.WriteString(w, s.Message())
		}
	}
}

func Unfollow(w http.ResponseWriter, r *http.Request) {
	if !checkToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	msgProto := &pb.FollowRequest{}
	if err := jsonpb.Unmarshal(r.Body, msgProto); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if _, err := usersClient.Unfollow(ctx, msgProto); err != nil {
			s, _ := status.FromError(err)
			switch s.Code() {
			case codes.InvalidArgument:
				w.WriteHeader(http.StatusBadRequest)
			case codes.Canceled:
				w.WriteHeader(http.StatusConflict)
			case codes.NotFound:
				w.WriteHeader(http.StatusConflict)
			case codes.Internal:
				w.WriteHeader(http.StatusInternalServerError)
			case codes.DeadlineExceeded:
				w.WriteHeader(http.StatusRequestTimeout)
			case codes.Unauthenticated:
				w.WriteHeader(http.StatusUnauthorized)
				reconnect()
			}
			io.WriteString(w, s.Message())
		}
	}
}

func Tweet(w http.ResponseWriter, r *http.Request) {
	if !checkToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	msgProto := &pb.CreateTweetRequest{}
	if err := jsonpb.Unmarshal(r.Body, msgProto); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if _, err := tweetsClient.CreateTweet(ctx, msgProto); err != nil {
			s, _ := status.FromError(err)
			switch s.Code() {
			case codes.InvalidArgument:
				w.WriteHeader(http.StatusBadRequest)
			case codes.Internal:
				w.WriteHeader(http.StatusInternalServerError)
			case codes.DeadlineExceeded:
				w.WriteHeader(http.StatusRequestTimeout)
			case codes.Unauthenticated:
				w.WriteHeader(http.StatusUnauthorized)
				reconnect()
			}
			io.WriteString(w, s.Message())
		}
	}
}

func GetUserTweets(w http.ResponseWriter, r *http.Request) {
	if !checkToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	msgProto := &pb.GetTweetsRequest{}
	if err := jsonpb.Unmarshal(r.Body, msgProto); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if stream, err := tweetsClient.GetTweets(ctx, msgProto); err != nil {
			s, _ := status.FromError(err)
			switch s.Code() {
			case codes.InvalidArgument:
				w.WriteHeader(http.StatusBadRequest)
			case codes.Internal:
				w.WriteHeader(http.StatusInternalServerError)
			case codes.DeadlineExceeded:
				w.WriteHeader(http.StatusRequestTimeout)
			case codes.Unauthenticated:
				w.WriteHeader(http.StatusUnauthorized)
				reconnect()
			}
			io.WriteString(w, s.Message())
		} else {
			m := jsonpb.Marshaler{}
			first := true

			for {
				in, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						break
					} else {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}

				if !in.GetError() {
					if first {
						io.WriteString(w, "{")
						first = false
					} else {
						io.WriteString(w, ",")
					}

					m.Marshal(w, in.GetTweet())
				}
			}

			io.WriteString(w, "}")
		}
	}
}

func DeleteTweets(w http.ResponseWriter, r *http.Request) {
	if !checkToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	msgProto := &pb.DeleteTweetsRequest{}
	if err := jsonpb.Unmarshal(r.Body, msgProto); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if _, err := tweetsClient.DeleteTweets(ctx, msgProto); err != nil {
			s, _ := status.FromError(err)
			switch s.Code() {
			case codes.InvalidArgument:
				w.WriteHeader(http.StatusBadRequest)
			case codes.Internal:
				w.WriteHeader(http.StatusInternalServerError)
			case codes.DeadlineExceeded:
				w.WriteHeader(http.StatusRequestTimeout)
			case codes.Unauthenticated:
				w.WriteHeader(http.StatusUnauthorized)
				reconnect()
			}
			io.WriteString(w, s.Message())
		}
	}
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

	ticker := time.NewTicker(5 * time.Minute)
	done := make(chan bool)

	go func() {
		connect()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				reconnect()
			}
		}
	}()

	r := mux.NewRouter()
	r.HandleFunc("/signup", SignUp).Methods("POST")
	r.HandleFunc("/signin", SignIn).Methods("POST")
	r.HandleFunc("/user/delete", DeleteUser).Methods("DELETE")
	r.HandleFunc("/user/summary", GetUserInfoSummary).Methods("GET")
	r.HandleFunc("/user/followers", GetFollowers).Methods("GET")
	r.HandleFunc("/user/following", GetFollowing).Methods("GET")
	r.HandleFunc("/user/timeline", GetUserTimeline).Methods("GET")
	r.HandleFunc("/user/follow", Follow).Methods("POST")
	r.HandleFunc("/user/unfollow", Unfollow).Methods("POST")
	r.HandleFunc("/tweets/tweet", Tweet).Methods("POST")
	r.HandleFunc("/tweets", GetUserTweets).Methods("GET")
	r.HandleFunc("/tweets/delete", DeleteTweets).Methods("DELETE")

	listenAddress := fmt.Sprintf("0.0.0.0:%d", tools.Conf.GatewayPort)
	log.Info().Msgf("Start listen on port %s", listenAddress)
	if err := http.ListenAndServe(listenAddress, r); err != nil {
		log.Fatal().Err(err).Send()
	}
	done <- true
}

func connect() {
	perRPC := oauth.NewOauthAccess(fetchToken())
	creds, err := credentials.NewClientTLSFromFile(data.Path("x509/ca_cert.pem"), "x.test.example.com")
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	opts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(perRPC),
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
	}

	if usersConn, err = grpc.Dial(fmt.Sprintf("%s:%d",
		tools.Conf.UsersServiceHost, tools.Conf.UsersPort), opts...); err != nil {
		log.Error().Err(err).Send()
	} else {
		usersClient = pb.NewUsersClient(usersConn)
	}

	if tweetsConn, err = grpc.Dial(fmt.Sprintf("%s:%d",
		tools.Conf.TweetsServiceHost, tools.Conf.TweetsPort), opts...); err != nil {
		log.Error().Err(err).Send()
	} else {
		tweetsClient = pb.NewTweetsClient(tweetsConn)
	}
}

func reconnect() {
	usersConn.Close()
	tweetsConn.Close()

	connect()
}

var authToken string

func fetchToken() *oauth2.Token {
	r, err := http.Get(fmt.Sprintf("http://%s:%d/service/token", tools.Conf.AuthServiceHost, tools.Conf.AuthPort))
	if err != nil {
		log.Error().Err(err).Send()
		goto exit
	}

	if r.StatusCode == http.StatusCreated || r.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Send()
			goto exit
		}
		authToken = string(body)
	}

exit:
	return &oauth2.Token{
		AccessToken: authToken,
	}
}
