package datadoghq_test

import (
	"testing"
	"time"

	"github.com/savaki/datadoghq"
)

func TestMarshal(t *testing.T) {
	p := datadoghq.Point{
		Timestamp: time.Unix(1234, 0),
		Value:     56.78,
	}
	data, err:=p.MarshalJSON()
	if err != nil {
		t.Fail()
	}
	if string(data) != "[1234,56.78]" {
		t.Fail()
	}
}
