package queuer

import "testing"

func TestSendTask(t *testing.T) {
	queuer := NewQueuer()
	// queuer.Run()
	queuer.SendTask()
}
