package parallel

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestMap(t *testing.T) {
	in := []int{1, 2, 3, 4, 5, 6, 7, 8}
	expected := []int{1, 4, 9, 16, 25, 36, 49, 64}

	out, err := Map(context.Background(), in, func(_ context.Context, num int) (int, error) {
		t.Logf("Squaring %d", num)
		time.Sleep(time.Duration(num) * 50 * time.Millisecond)
		return num * num, nil
	}, 2)

	if err != nil {
		t.Error("Unexpected error ", err)
		return
	}

	if !reflect.DeepEqual(out, expected) {
		t.Errorf("Output not expected\nOut: %v\nExp: %v", out, expected)
	}
}

func TestMapCancel(t *testing.T) {
	in := []int{1, 2, 3, 4, 5, 6, 7, 8}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	out, err := Map(ctx, in, func(_ context.Context, num int) (int, error) {
		t.Logf("Squaring %d", num)
		time.Sleep(time.Duration(num) * 50 * time.Millisecond)
		return num * num, nil
	}, 3)

	if out != nil {
		t.Errorf("Expected no output, but got %v", out)
	}

	if err == nil {
		t.Errorf("Expected error, but got none")
	} else {
		t.Logf("Got expected error: %s", err.Error())
	}
}
