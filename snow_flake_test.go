package snowflake

import "testing"

func TestNewSnowFlake(t *testing.T) {
	_, err := NewSnowFlake(func() int64 { return 1 })
	if err != nil {
		t.Fatalf("error creating SnowFlake")
	}
}
