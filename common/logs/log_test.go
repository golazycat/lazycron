package logs

import "testing"

func TestInitLoggers(t *testing.T) {

	err := InitLoggers("./error.log")
	if err != nil {
		t.Errorf("Error Init: %v", err)
		return
	}

	Info.Printf("Hello log")
	Warn.Printf("Hello WARN")
	Error.Printf("Error!")
}
