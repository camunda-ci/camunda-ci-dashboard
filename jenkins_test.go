package dashboard

import (
	"fmt"
	"testing"
)

func TestJenkinsGetQueue(t *testing.T) {
	t.SkipNow()
	jenkins := NewJenkinsClient("https://app.camunda.com/jenkins", "foo", "bar")

	queue, _ := jenkins.GetQueue()
	for item := range queue.Items {
		fmt.Println(item)
	}
}
