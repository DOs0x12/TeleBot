package broker

import "testing"

func TestCastCommandToTopicName(t *testing.T) {
	testComName := "/test"
	want := "test"
	res := castCommandToTopicName(testComName)
	if res != want {
		t.Errorf("The result was incorrect, get: %v, want: %v", res, want)
	}
}
