package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONPath(t *testing.T) {
	assert.Equal(t, "home/test.json", JSONPath("home", "test"))
	assert.Equal(t, "home/test.txt", JSONPath("home", "test.txt"))
}

func TestRemoveExt(t *testing.T) {
	assert.Equal(t, "test", RemoveExt("test.txt"))
}

func TestIsJSONExt(t *testing.T) {
	assert.Equal(t, true, IsJSONExt("test.json"))
	assert.Equal(t, false, IsJSONExt("test.txt"))
}

func TestHasSuffix(t *testing.T) {
	assert.Equal(t, true, HasSuffix("test-sub.json", "-", "sub"))
	assert.Equal(t, true, HasSuffix("test-test-sub.json", "-", "sub"))
	assert.Equal(t, false, HasSuffix("test-t.txt", "-", "sub"))
	assert.Equal(t, false, HasSuffix("test.txt", "-", "sub"))
}

func TestRemoveSuffix(t *testing.T) {
	assert.Equal(t, "test", RemoveSuffix("test-sub.json", "-", "sub"))
	assert.Equal(t, "---test--aaa--test$$11333##-", RemoveSuffix("---test--aaa--test$$11333##--sub.json", "-", "sub"))
	assert.Equal(t, "-----a----", RemoveSuffix("-----a-----sub.json", "-", "sub"))

}
