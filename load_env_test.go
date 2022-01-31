package load_nev

import (
	"os"
	"testing"
)
import "github.com/stretchr/testify/assert"

type TestStruct struct {
	Port int `env:"name=PORT,default=8080"`
}

func TestOnlyAcceptPointerToStruct(t *testing.T) {
	err := os.Setenv("PORT", "8080")
	if err != nil {
		panic(err)
	}

	err = LoadEnv("test")
	assert.NotNil(t, err, "should return error")

	i := 1
	err = LoadEnv(&i)

	err = LoadEnv(TestStruct{})
	assert.NotNil(t, err, "should return error")

	err = LoadEnv(&TestStruct{})
	assert.Nil(t, err, "should not return error")
}

func TestLoadEnv(t *testing.T) {
	config := TestStruct{}
	err := LoadEnv(&config)
	assert.Nil(t, err, "should not return error")

	assert.Equal(t, 8080, config.Port, "should load env")
}
