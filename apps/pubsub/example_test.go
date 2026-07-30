package pubsub_test

import (
	"fmt"
	"sort"

	"github.com/alejandrogarsan89/go-concurrency/apps/pubsub"
)

func ExampleBroker() {
	b := pubsub.New[string]()
	defer b.Close()

	alice, _ := b.Subscribe("news", 4)
	bob, _ := b.Subscribe("news", 4)

	b.Publish("news", "hello")
	b.Publish("news", "world")

	// Each subscriber gets its own copy of every message.
	var got []string
	for i := 0; i < 2; i++ {
		got = append(got, "alice:"+<-alice)
		got = append(got, "bob:"+<-bob)
	}
	sort.Strings(got)
	for _, line := range got {
		fmt.Println(line)
	}
	// Output:
	// alice:hello
	// alice:world
	// bob:hello
	// bob:world
}
