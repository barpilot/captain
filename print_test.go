package captain // import "github.com/harbur/captain"

import (
	"testing"
)

func TestPrintInfo(t *testing.T) {
	pInfo("test info %s", "message")
}

func TestPrintErr(t *testing.T) {
	pError("test err %s", "message")
}

func TestPrintDebug(t *testing.T) {
	Debug = true
	defer func() { Debug = false }()
	pDebug("test debug %s", "message")
}
