package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	myservice "github.com/haru-256/blog-connect-go-interceptor/gen/go/myservice/v1"
	myserviceconnect "github.com/haru-256/blog-connect-go-interceptor/gen/go/myservice/v1/myservicev1connect"
	"github.com/haru-256/blog-connect-go-interceptor/internal/interceptor"
)

// myServiceImpl は MyService の実装です。
type myServiceImpl struct {
	logger *slog.Logger
	myserviceconnect.UnimplementedMyServiceHandler
}

func NewMyServiceImpl(logger *slog.Logger) *myServiceImpl {
	return &myServiceImpl{
		logger: logger,
	}
}

// GetUser (Unary RPC) の実装
func (s *myServiceImpl) GetUser(
	ctx context.Context,
	req *connect.Request[myservice.GetUserRequest],
) (*connect.Response[myservice.GetUserResponse], error) {
	s.logger.InfoContext(ctx, "--- [Server Logic] GetUser called ---")
	// ダミーのロジック
	if req.Msg.UserId == "error" {
		s.logger.ErrorContext(ctx, "User not found", "userId", req.Msg.UserId)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	}

	res := &myservice.GetUserResponse{
		User: &myservice.User{
			UserId: req.Msg.UserId,
			Name:   "haru256",
		},
	}
	return connect.NewResponse(res), nil
}

// ListUsers (Server Streaming RPC) の実装
func (s *myServiceImpl) ListUsers(
	ctx context.Context,
	req *connect.Request[myservice.ListUsersRequest],
	stream *connect.ServerStream[myservice.ListUsersResponse],
) error {
	s.logger.InfoContext(ctx, "--- [Server Logic] ListUsers called ---")
	// ダミーデータ
	allUsers := []*myservice.User{
		{UserId: "user1", Name: "Alice"},
		{UserId: "user2", Name: "Bob"},
		{UserId: "user3", Name: "Charlie"},
		{UserId: "user4", Name: "Diana"},
		{UserId: "user5", Name: "Eve"},
	}

	// ページネーションのページサイズと開始位置の取得
	pageSize := int(req.Msg.PageSize)
	if pageSize <= 0 {
		pageSize = 2 // デフォルトページサイズ
	}
	startIndex := 0
	if req.Msg.PageToken != "" {
		fmt.Sscanf(req.Msg.PageToken, "page-%d", &startIndex)
	}
	// 指定されたページ以降について、ページサイズごとにユーザーデータを送信
	for startIndex < len(allUsers) {
		endIndex := min(startIndex+pageSize, len(allUsers))

		s.logger.InfoContext(ctx, "Sending page", "startIndex", startIndex, "endIndex", endIndex)

		users := allUsers[startIndex:endIndex]
		var nextPageToken string
		if endIndex < len(allUsers) {
			nextPageToken = fmt.Sprintf("page-%d", endIndex)
		}

		if err := stream.Send(&myservice.ListUsersResponse{
			Users:         users,
			NextPageToken: nextPageToken,
		}); err != nil {
			return connect.NewError(connect.CodeUnknown, err)
		}

		startIndex = endIndex
	}
	return nil
}

func (s *myServiceImpl) UpdateUsers(ctx context.Context, stream *connect.ClientStream[myservice.UpdateUsersRequest]) (*connect.Response[myservice.UpdateUsersResponse], error) {
	s.logger.InfoContext(ctx, "--- [Server Logic] UpdateUsers client stream opened ---")
	updatedCount := 0
	// クライアントから来るメッセージを順に受け取る
	for stream.Receive() {
		msg := stream.Msg()
		batchSize := len(msg.Users)
		s.logger.InfoContext(ctx, "Received user batch", "count", batchSize)
		updatedCount += batchSize
	}

	// 受信中にエラーがあれば処理
	if err := stream.Err(); err != nil {
		return nil, connect.NewError(connect.CodeOf(err), err)
	}

	res := &myservice.UpdateUsersResponse{
		UpdatedCount: int32(updatedCount),
	}
	return connect.NewResponse(res), nil
}

// Chat (Bidirectional Streaming RPC) の実装
func (s *myServiceImpl) Chat(
	ctx context.Context,
	stream *connect.BidiStream[myservice.ChatRequest, myservice.ChatResponse],
) error {
	s.logger.InfoContext(ctx, "--- [Server Logic] Chat bidirectional stream opened ---")
	for {
		// クライアントからのメッセージを受信
		req, err := stream.Receive()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// クライアントが送信を終了し、ストリームを正常にクローズ
				s.logger.InfoContext(ctx, "--- [Server Logic] Chat stream closed by client (EOF) ---")
				return nil
			}
			// その他のエラー（ネットワークエラーなど）
			return connect.NewError(connect.CodeOf(err), err)
		}

		s.logger.InfoContext(ctx, "--- [Server Logic] Received message ---", "text", req.Text)

		// クライアントに返信
		if err := stream.Send(&myservice.ChatResponse{
			Text: fmt.Sprintf("Server echoes: %s", req.Text),
		}); err != nil {
			return connect.NewError(connect.CodeUnknown, err)
		}
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ロガーの準備 (DEBUGレベルでペイロードも出力)
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}),
	)
	// インターセプタのインスタンス化
	loggingInterceptor := interceptor.NewReqRespLogger(logger)
	// ハンドラの初期化時にオプションとして渡す
	mux := http.NewServeMux()
	path, handler := myserviceconnect.NewMyServiceHandler(
		NewMyServiceImpl(logger),
		connect.WithInterceptors(loggingInterceptor), // <- ここで適用
	)
	mux.Handle(path, handler)

	// reflection
	reflector := grpcreflect.NewStaticReflector(
		myserviceconnect.MyServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// サーバーの起動
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8081"
	}
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		logger.ErrorContext(ctx, "Failed to start server", "error", err)
		return
	}
	srv := &http.Server{
		Addr:    ln.Addr().String(),
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
	go func() {
		logger.InfoContext(ctx, "Server started", "address", ln.Addr().String())
		if srvErr := srv.Serve(ln); srvErr != nil && !errors.Is(srvErr, http.ErrServerClosed) {
			logger.ErrorContext(ctx, "Server failed", "error", srvErr)
			return
		}
	}()
	<-ctx.Done()

	// サーバーをグレースフルシャットダウン
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.ErrorContext(ctx, "Server shutdown failed", "error", err)
		return
	} else {
		logger.InfoContext(ctx, "Server stopped gracefully")
	}
}
