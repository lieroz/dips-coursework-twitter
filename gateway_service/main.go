package main

import (
	"context"
	"io"
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
)

var (
	usersClient  pb.UsersClient
	tweetsClient pb.TweetsClient
)

const (
	timeout = 200 * time.Millisecond
)

/*
 TODO:
 1. add oauth
*/

func SignUp(w http.ResponseWriter, r *http.Request) {
	msgProto := &pb.CreateRequest{}
	if err := jsonpb.Unmarshal(r.Body, msgProto); err != nil {
		log.Error().Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

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
			}
			io.WriteString(w, s.Message())
		}
	}
}

func LogIn(w http.ResponseWriter, r *http.Request) {
}

func LogOut(w http.ResponseWriter, r *http.Request) {
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
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
			}
			io.WriteString(w, s.Message())
		}
	}
}

func GetUserInfoSummary(w http.ResponseWriter, r *http.Request) {
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
			}
			io.WriteString(w, s.Message())
		} else {
			m := jsonpb.Marshaler{}
			m.Marshal(w, in)
		}
	}
}

func GetFollowers(w http.ResponseWriter, r *http.Request) {
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
	}
}

func Follow(w http.ResponseWriter, r *http.Request) {
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
			}
			io.WriteString(w, s.Message())
		}
	}
}

func Unfollow(w http.ResponseWriter, r *http.Request) {
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
			}
			io.WriteString(w, s.Message())
		}
	}
}

func Tweet(w http.ResponseWriter, r *http.Request) {
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
			}
			io.WriteString(w, s.Message())
		}
	}
}

func GetUserTweets(w http.ResponseWriter, r *http.Request) {
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

	perRPC := oauth.NewOauthAccess(fetchToken())
	creds, err := credentials.NewClientTLSFromFile(data.Path("x509/ca_cert.pem"), "x.test.example.com")
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	opts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(perRPC),
		grpc.WithTransportCredentials(creds),
	}

	opts = append(opts, grpc.WithBlock())

	if conn, err := grpc.Dial("users:8001", opts...); err != nil {
		log.Fatal().Err(err).Send()
	} else {
		defer conn.Close()
		usersClient = pb.NewUsersClient(conn)
	}

	if conn, err := grpc.Dial("tweets:8002", opts...); err != nil {
		log.Fatal().Err(err).Send()
	} else {
		defer conn.Close()
		tweetsClient = pb.NewTweetsClient(conn)
	}

	r := mux.NewRouter()
	r.HandleFunc("/signup", SignUp).Methods("POST")
	r.HandleFunc("/login", LogIn).Methods("POST")   // TODO
	r.HandleFunc("/logout", LogOut).Methods("POST") // TODO
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

	log.Info().Msg("Start listen")
	if err := http.ListenAndServe("localhost:8080", r); err != nil {
		log.Fatal().Err(err).Send()
	}
}

func fetchToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken: "some-secret-token",
	}
}
