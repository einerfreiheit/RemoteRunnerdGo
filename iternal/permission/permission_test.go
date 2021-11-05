package permission

import (
	"reflect"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	perm := NewPermissioner()
	data := []byte("ping pong pang")
	perm.Read(data)
	expected := map[string]bool{"ping": true, "pong": true, "pang": true}
	if reflect.DeepEqual(expected, perm.permitted) != true {
		t.Error("Read test failed")
	}
}

func split(str string) (sb []byte, ss []string) {
	return []byte(str), strings.Split(str, " ")

}
func TestCheck(t *testing.T) {
	perm := NewPermissioner()
	data := "ping pong pang"
	sb, ss := split(data)
	perm.Read(sb)

	rangeTest := func(ss []string) {
		for _, cmd := range ss {
			to_test := []string{cmd}
			if !perm.Check(to_test) {
				t.Error("Read test failed")
			}
		}
	}
	rangeTest(ss)

	data = "ping pong"
	sb, ss = split(data)
	perm.Read(sb)
	rangeTest(ss)

	to_test := []string{"pang"}
	if perm.Check(to_test) {
		t.Error("Read reload test failed")
	}

	to_test = []string{"pong & pang"}
	if perm.Check(to_test) {
		t.Error("Read & test failed")
	}
}
