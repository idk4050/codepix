package stream_test

import (
	"fmt"
	"testing"
)

func TestStream(t *testing.T) {
	client, commandHandler, pixKeyRepo := StreamWithMocks()

	type test struct {
		description string
		fn          func(*testing.T)
	}
	tests := []test{
		{"start", Start(client, commandHandler, pixKeyRepo)},
		{"confirm", Confirm(client, commandHandler)},
		{"complete", Complete(client, commandHandler)},
		{"fail", Fail(client, commandHandler)},
	}
	for i, test := range tests {
		t.Run(fmt.Sprint(i, "_", test.description), test.fn)
	}
}

func TestStreamIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	client, repo, pixKeyRepo, creator, tearDown := Stream()
	defer tearDown()

	type test struct {
		description string
		fn          func(*testing.T)
	}
	tests := []test{
		{"start", StartIntegration(client, repo, pixKeyRepo, creator)},
		{"confirm", ConfirmIntegration(client, repo, creator)},
		{"complete", CompleteIntegration(client, repo, creator)},
		{"fail", FailIntegration(client, repo, creator)},
	}
	for i, test := range tests {
		t.Run(fmt.Sprint(i, "_", test.description), test.fn)
	}
}
