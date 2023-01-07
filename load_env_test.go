package load_env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

type InnerStruct struct {
	Host string `env:"name=HOST,required"`
	Port int    `env:"name=PORT,required"`
}
type TestConfigWithStruct struct {
	Level  string `env:"name=LEVEL,required"`
	Config InnerStruct
}

func TestLoadEnvWithStruct(t *testing.T) {
	err := os.Setenv("PORT", "8080")
	assert.Nil(t, err, "should set env")

	err = os.Setenv("LEVEL", "debug")
	assert.Nil(t, err, "should set env")

	err = os.Setenv("HOST", "localhost")
	assert.Nil(t, err, "should set env")

	config := TestConfigWithStruct{}
	err = LoadEnv(&config)
	assert.Nil(t, err, "should not return error")

	assert.Equal(t, 8080, config.Config.Port, "should load env")
	assert.Equal(t, "debug", config.Level, "should load env")
	assert.Equal(t, "localhost", config.Config.Host, "should load env")
}
