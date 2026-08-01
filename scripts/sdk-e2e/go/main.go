package main

import (
	"context"
	"fmt"
	"os"
	"time"

	rexec "github.com/PipeOpsHQ/rexec-go"
)

func main() {
	url := os.Getenv("URL")
	token := os.Getenv("TOKEN")
	client := rexec.NewClient(url, token)
	ctx := context.Background()

	fmt.Println("[go] list...")
	list, err := client.Containers.List(ctx)
	must(err)
	fmt.Println("[go] list count", len(list))

	fmt.Println("[go] create...")
	c, err := client.Containers.Create(ctx, &rexec.CreateContainerRequest{
		Image: "ubuntu",
		Name:  fmt.Sprintf("go-e2e-%d", time.Now().Unix()),
	})
	must(err)
	fmt.Println("[go] created", c.ID, c.Status, c.Image)

	list, err = client.Containers.List(ctx)
	must(err)
	fmt.Println("[go] list count", len(list))

	got, err := client.Containers.Get(ctx, c.ID)
	must(err)
	fmt.Println("[go] get", got.ID, got.Status)

	must(client.Containers.Delete(ctx, c.ID))
	fmt.Println("[go] OK")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
