# 記事: connect-goインターセプタ実装パターンガイド - Unary/Streaming対応 のサンプルコード

このリポジトリには、以下のすべてが含まれています：

* **完全なインターセプタ実装** (`internal/interceptor/logger.go`)
* **サーバー実装** (Unary, Server/Client/Bidirectional Streaming)
* **クライアント実装** (各RPC種別の呼び出し例)

サーバーとクライアントを実際に動かすことで、この記事で説明したログがどのような順番で出力されるかを確認できます。

```bash
# サーバーの起動
make run-server

# 別のターミナルで各RPCを実行
make get-user        # Unary RPC
make list-users      # Server Streaming
make update-users    # Client Streaming
make chat            # Bidirectional Streaming
```

各コマンドを実行すると、サーバー側とクライアント側の両方で、🔵（開始）→ 🟢（送受信）→ 🔴（終了）の順番でログが出力される様子を確認できます。

特に、Bidirectional Streaming（chat）では、クライアントが複数のメッセージを送信（🟢 Send）した後に接続をクローズ（🔴 CloseRequest）し、その後サーバーからのレスポンスを受信（🟢 Receive）する様子が観察できます。これにより、ストリーミングの非同期性とインターセプタのロギングタイミングが理解しやすくなります。

実際のBidirectional Streaming（chat）の実行結果は以下のようになります：

Client側

![client](./images/connect-goのUnaryとStreamのインターセプタ解説/client-bidirectional-streaming.png)

Server側

![server](./images/connect-goのUnaryとStreamのインターセプタ解説/server-bidirectional-streaming.png)
