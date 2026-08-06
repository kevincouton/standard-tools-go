package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kevincouton/standard-tools-go/internal/agent"
	"github.com/kevincouton/standard-tools-go/internal/marketdata"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: standard-tools <server|audit>")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "server":
		fmt.Println("use cmd/server to run the server")
	case "audit":
		if len(os.Args) < 3 || os.Args[2] != "verify" {
			fmt.Println("usage: standard-tools audit verify")
			os.Exit(1)
		}
		cache := marketdata.NewInMemoryCache()
		svc := marketdata.NewService("synthetic", cache)
		svc.Register(&marketdata.SyntheticProvider{})
		dispatcher := agent.NewDispatcher(svc)
		res, err := dispatcher.Dispatch(context.Background(), agent.ToolCall{Name: "health", Arguments: []byte(`{}`)})
		if err != nil {
			fmt.Println("audit verify failed:", err)
			os.Exit(1)
		}
		fmt.Println("audit verify:", string(res.Output))
	default:
		fmt.Printf("unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
