package logging

import (
	"fmt"
	"io/ioutil"
	"os/exec"

	"testing"
)

const (
	LOG_FILE = "./filelog.log"
	TENBYTES = "1234567890"
)

func TestFilelogRotate(t *testing.T) {
	fl, err := NewFileLogWriter(LOG_FILE, true, 100)
	if err != nil {
		t.Error(err)
	}

	log := MustGetLogger("filelog")
	fl = fl.SetRotateHourly(true)
	backend := NewLogBackend(fl, "", 0)
	SetBackend(backend)

	for i := 0; i < 11; i++ {
		log.Info(TENBYTES)
	}

	fl.Close()

	data, err := ioutil.ReadFile(LOG_FILE)
	if err != nil {
		t.Error(err)
	}

	if string(data) != TENBYTES+"\n" {
		t.Error("file rotate error")
		fmt.Printf("expect:%s, got:%s", TENBYTES+"\n", string(data))
	}

	cmd := exec.Command("bash", "-c", "rm *.log*")
	cmd.Run()

}
