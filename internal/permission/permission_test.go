package permission

import (
	"strings"
	"testing"
)

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
			request := []string{cmd}
			if !perm.Check(request) {
				t.Error("Read test failed")
			}
		}
	}
	rangeTest(ss)

	data = "ping pong"
	sb, ss = split(data)
	perm.Read(sb)
	rangeTest(ss)

	request := []string{"pang"}
	if perm.Check(request) {
		t.Error("Read reload test failed")
	}

	request = []string{"pong & pang"}
	if perm.Check(request) {
		t.Error("Read & test failed")
	}
}
