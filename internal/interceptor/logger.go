package interceptor

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"time"

	"connectrpc.com/connect"
)

// ReqRespLogger は connect.Interceptor を実装するロギングインターセプタです。
type ReqRespLogger struct {
	logger *slog.Logger
}

// NewReqRespLogger は ReqRespLogger の新しいインスタンスを生成します。
func NewReqRespLogger(logger *slog.Logger) *ReqRespLogger {
	return &ReqRespLogger{
		logger: logger,
	}
}

// --- Unary RPC ---

// WrapUnary は Unary RPC の処理をラップします。
func (i *ReqRespLogger) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		start := time.Now()
		i.logUnaryStart(ctx, req) // リクエスト開始ログ

		var code connect.Code
		res, err := next(ctx, req) // 本体処理の実行

		if err != nil {
			code = connect.CodeOf(err)
		} else {
			code = 0 // OK
		}

		// リクエスト終了ログ
		i.logUnaryEnd(ctx, req, res, err, code, time.Since(start))
		return res, err
	}
}

func (i *ReqRespLogger) logUnaryStart(ctx context.Context, req connect.AnyRequest) {
	// 🔵 接続確立時のログ
	i.logger.InfoContext(ctx, "🔵 Unary Request Start",
		slog.String("procedure", req.Spec().Procedure),
		slog.Any("request_body", req.Any()), // DEBUGレベル推奨
	)
}

func (i *ReqRespLogger) logUnaryEnd(
	ctx context.Context,
	req connect.AnyRequest,
	res connect.AnyResponse,
	err error,
	code connect.Code,
	duration time.Duration,
) {
	// 🔴 接続終了時のログ
	if err != nil {
		i.logger.ErrorContext(ctx, "🔴 Unary Request End",
			slog.String("procedure", req.Spec().Procedure),
			slog.Duration("duration", duration),
			slog.String("code", code.String()),
			slog.String("error", err.Error()),
		)
	} else {
		i.logger.InfoContext(ctx, "🔴 Unary Request End",
			slog.String("procedure", req.Spec().Procedure),
			slog.Duration("duration", duration),
			slog.String("code", code.String()),
			slog.Any("response_body", res.Any()), // DEBUGレベル推奨
		)
	}
}

// --- Streaming RPC (Handler) ---

// WrapStreamingHandler はサーバーサイドのストリーミングをラップします。
func (i *ReqRespLogger) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		start := time.Now()

		// 🔵 接続確立時のログ
		i.logger.InfoContext(ctx, "🔵 Handler Stream Start",
			slog.String("procedure", conn.Spec().Procedure),
		)

		// 🔴 接続終了時のログ
		defer func() {
			duration := time.Since(start)
			i.logger.InfoContext(ctx, "🔴 Handler Stream Finished",
				slog.String("procedure", conn.Spec().Procedure),
				slog.Duration("duration", duration),
			)
		}()

		// loggingHandlerConn で conn をラップ
		wrappedConn := &loggingHandlerConn{
			StreamingHandlerConn: conn,
			ctx:                  ctx,
			logger:               i.logger,
		}

		// ラップした接続を使って本体処理(next)を実行
		return next(ctx, wrappedConn)
	}
}

// loggingHandlerConn はサーバーサイドの送受信をフックします。
type loggingHandlerConn struct {
	connect.StreamingHandlerConn
	ctx    context.Context
	logger *slog.Logger
}

// Receive メソッドをオーバーライド
func (c *loggingHandlerConn) Receive(msg any) error {
	err := c.StreamingHandlerConn.Receive(msg)
	if err != nil && !errors.Is(err, io.EOF) {
		c.logger.ErrorContext(c.ctx, "Handler Stream Receive Error",
			slog.String("procedure", c.Spec().Procedure),
			slog.String("error", err.Error()),
		)
	} else if err == nil {
		// 🟢 受信成功ログ (DEBUGレベル推奨)
		c.logger.DebugContext(c.ctx, "🟢 Handler Stream Receive",
			slog.String("procedure", c.Spec().Procedure),
			slog.Any("message", msg),
		)
	}
	return err
}

// Send メソッドをオーバーライド
func (c *loggingHandlerConn) Send(msg any) error {
	err := c.StreamingHandlerConn.Send(msg)
	if err != nil {
		c.logger.ErrorContext(c.ctx, "Handler Stream Send Error",
			slog.String("procedure", c.Spec().Procedure),
			slog.String("error", err.Error()),
		)
	} else {
		// 🟢 送信成功ログ (DEBUGレベル推奨)
		c.logger.DebugContext(c.ctx, "🟢 Handler Stream Send",
			slog.String("procedure", c.Spec().Procedure),
			slog.Any("message", msg),
		)
	}
	return err
}

// --- Streaming RPC (Client) ---

// WrapStreamingClient はクライアントサイドのストリーミングをラップします。
func (i *ReqRespLogger) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		// 1. next() を呼び、実際の接続(conn)を取得
		conn := next(ctx, spec)

		// 🔵 タイミング1: 接続確立時（1回のみ）
		i.logger.InfoContext(ctx, "🔵 Client Stream Start",
			slog.String("procedure", spec.Procedure),
		)

		// 2. 取得した conn をラップして返す
		return &loggingClientConn{
			StreamingClientConn: conn,
			logger:              i.logger,
			spec:                spec,
			ctx:                 ctx,
		}
	}
}

// loggingClientConn はクライアントサイドの送受信とクローズをフックします。
type loggingClientConn struct {
	connect.StreamingClientConn
	logger *slog.Logger
	spec   connect.Spec
	ctx    context.Context
}

// Send メソッドをオーバーライド
func (c *loggingClientConn) Send(msg any) error {
	err := c.StreamingClientConn.Send(msg)
	if err != nil {
		c.logger.ErrorContext(c.ctx, "Client Stream Send Error",
			slog.String("procedure", c.spec.Procedure),
			slog.String("error", err.Error()),
		)
	} else {
		// 🟢 タイミング2: クライアントがサーバーへメッセージを送信するたびに実行
		c.logger.DebugContext(c.ctx, "🟢 Client Stream Send",
			slog.String("procedure", c.spec.Procedure),
			slog.Any("message", msg),
		)
	}
	return err
}

// Receive メソッドをオーバーライド
func (c *loggingClientConn) Receive(msg any) error {
	err := c.StreamingClientConn.Receive(msg)
	if err != nil && !errors.Is(err, io.EOF) {
		c.logger.ErrorContext(c.ctx, "Client Stream Receive Error",
			slog.String("procedure", c.spec.Procedure),
			slog.String("error", err.Error()),
		)
	} else if err == nil {
		// 🟢 タイミング2: クライアントがサーバーからメッセージを受信するたびに実行
		c.logger.DebugContext(c.ctx, "🟢 Client Stream Receive",
			slog.String("procedure", c.spec.Procedure),
			slog.Any("message", msg),
		)
	}
	return err
}

// CloseRequest メソッドをオーバーライド
func (c *loggingClientConn) CloseRequest() error {
	err := c.StreamingClientConn.CloseRequest()
	// 🔴 タイミング3a: クライアントが送信を終了する時に実行(1回のみ)
	if err != nil {
		c.logger.ErrorContext(c.ctx, "Client Stream CloseRequest failed",
			slog.String("procedure", c.spec.Procedure),
			slog.String("error", err.Error()),
		)
	} else {
		c.logger.InfoContext(c.ctx, "🔴 Client Stream CloseRequest",
			slog.String("procedure", c.spec.Procedure),
		)
	}
	return err
}

// CloseResponse メソッドをオーバーライド
func (c *loggingClientConn) CloseResponse() error {
	err := c.StreamingClientConn.CloseResponse()
	// 🔴 タイミング3b: クライアントが受信を終了する時に実行(1回のみ)
	if err != nil {
		c.logger.ErrorContext(c.ctx, "Client Stream CloseResponse failed",
			slog.String("procedure", c.spec.Procedure),
			slog.String("error", err.Error()),
		)
	} else {
		c.logger.InfoContext(c.ctx, "🔴 Client Stream CloseResponse",
			slog.String("procedure", c.spec.Procedure),
		)
	}
	return err
}

var _ connect.Interceptor = (*ReqRespLogger)(nil)
