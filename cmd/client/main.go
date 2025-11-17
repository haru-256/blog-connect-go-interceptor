package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/sync/errgroup"

	myservice "github.com/haru-256/blog-connect-go-interceptor/gen/go/myservice/v1"
	myserviceconnect "github.com/haru-256/blog-connect-go-interceptor/gen/go/myservice/v1/myservicev1connect"
	"github.com/haru-256/blog-connect-go-interceptor/internal/interceptor"
)

func callGetUser(ctx context.Context, client myserviceconnect.MyServiceClient, logger *slog.Logger) error {
	req := connect.NewRequest(&myservice.GetUserRequest{
		UserId: "user123",
	})
	res, err := client.GetUser(ctx, req)
	if err != nil {
		logger.ErrorContext(ctx, "Error calling GetUser", "error", err)
		return err
	}
	logger.InfoContext(ctx, "GetUser completed", "response", res.Msg)
	return nil
}

func callListUsers(ctx context.Context, client myserviceconnect.MyServiceClient, logger *slog.Logger) error {
	req := connect.NewRequest(&myservice.ListUsersRequest{
		PageSize: 2,
	})
	stream, err := client.ListUsers(ctx, req)
	if err != nil {
		logger.ErrorContext(ctx, "Error calling ListUsers", "error", err)
		return err
	}
	for stream.Receive() {
		res := stream.Msg()
		logger.InfoContext(ctx, "ListUsers Response received", "response", res)
	}
	if err := stream.Err(); err != nil {
		logger.ErrorContext(ctx, "Stream error", "error", err)
		return err
	}
	return nil
}

func callUpdateUsers(ctx context.Context, client myserviceconnect.MyServiceClient, logger *slog.Logger) error {
	stream := client.UpdateUsers(ctx)
	// 1回目の送信
	req := &myservice.UpdateUsersRequest{
		Users: []*myservice.User{
			{UserId: "user1", Name: "Updated_user1"},
			{UserId: "user2", Name: "Updated_user2"},
		},
	}
	if err := stream.Send(req); err != nil {
		logger.ErrorContext(ctx, "Error sending UpdateUsers request", "error", err)
		return err
	}
	// 2回目の送信
	req = &myservice.UpdateUsersRequest{
		Users: []*myservice.User{
			{UserId: "user3", Name: "Updated_user3"},
		},
	}
	if err := stream.Send(req); err != nil {
		logger.ErrorContext(ctx, "Error sending UpdateUsers request", "error", err)
		return err
	}
	res, err := stream.CloseAndReceive()
	if err != nil {
		logger.ErrorContext(ctx, "Error receiving UpdateUsers response", "error", err)
		return err
	}
	logger.InfoContext(ctx, "UpdateUsers completed", "response", res.Msg)
	return nil
}

func callChat(ctx context.Context, client myserviceconnect.MyServiceClient, logger *slog.Logger) error {
	stream := client.Chat(ctx)
	eg, egCtx := errgroup.WithContext(ctx)

	// 送信用 goroutine
	eg.Go(func() error {
		messages := []string{"Hello", "How are you?", "Goodbye"}
		for _, msg := range messages {
			select {
			case <-egCtx.Done():
				return egCtx.Err()
			default:
			}
			if err := stream.Send(&myservice.ChatRequest{Text: msg}); err != nil {
				logger.ErrorContext(ctx, "Error sending Chat message", "error", err)
				return err
			}
		}
		if err := stream.CloseRequest(); err != nil {
			logger.ErrorContext(ctx, "Error closing Chat request", "error", err)
			return err
		}
		return nil
	})

	// 受信用 goroutine
	eg.Go(func() error {
		for {
			res, err := stream.Receive()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				logger.ErrorContext(ctx, "Error receiving Chat response", "error", err)
				return err
			}
			logger.InfoContext(ctx, "Chat Response received", "response", res)
		}
	})

	return eg.Wait()
}

func main() {
	op := flag.String("op", "get-user", "GetUserサービスを呼び出す")
	flag.Parse()

	// ロガーの準備 (DEBUGレベルでペイロードも出力)
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}),
	)
	// インターセプタのインスタンス化
	loggingInterceptor := interceptor.NewReqRespLogger(logger)
	// クライアントの初期化時にオプションとして渡す
	client := &http.Client{
		Transport: &http2.Transport{
			// http2.Transport doesn't complain the URL scheme isn't 'https'
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				// h2c (HTTP/2 Cleartext) 用に TLS なしで接続
				return net.Dial(network, addr)
			},
		},
	}
	myServiceClient := myserviceconnect.NewMyServiceClient(
		client,
		"http://localhost:8081",
		connect.WithGRPC(),
		connect.WithInterceptors(loggingInterceptor),
	)

	ctx := context.Background()
	var err error
	switch *op {
	case "get-user":
		err = callGetUser(ctx, myServiceClient, logger)
	case "list-users":
		err = callListUsers(ctx, myServiceClient, logger)
	case "update-users":
		err = callUpdateUsers(ctx, myServiceClient, logger)
	case "chat":
		err = callChat(ctx, myServiceClient, logger)
	default:
		logger.ErrorContext(ctx, "Unknown operation", "operation", *op)
		return
	}

	if err != nil {
		return
	}
}
