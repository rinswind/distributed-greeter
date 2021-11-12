package tests

import (
	"bufio"
	"os"
	"path"
	"testing"

	"github.com/rinswind/distributed-greeter/greeter/internal/config"
)

func TestConfigLoad(t *testing.T) {
	setupFile(t, "config.yaml", `
Http:
  Port: 9090
  CertFile: ca.crt

Db:
  Endpoint: db-test
DbConfigDir: db-creds

Redis:
  Endpoint: redis-test
RedisConfigDir: redis-creds

AccessTokenConfigDir: at-creds
`)

	setupEnv(t, "DB_ENDPOINT", "env-db-test")

	setupDir(t, "db-creds", map[string]string{"db_user": "file-user", "DB_PASSWORD": "file-password"})
	setupDir(t, "redis-creds", map[string]string{"redis_access_key": "redis-key"})
	setupDir(t, "at-creds", map[string]string{"at_access_token_secret": "at-secret", "at_refresh_token_secret": "rt-secret"})

	cfg := config.ReadConfig()
	t.Logf("%+v\n", cfg)

	assertTrue(t, "DB_ENDPOINT override", cfg.Db.Endpoint == "env-db-test")
	assertTrue(t, "Db.User file override", cfg.Db.User == "file-user")
	assertTrue(t, "Db.Password file override", cfg.Db.Password == "file-password")
}

func setupEnv(t *testing.T, key string, val string) string {
	oldVal := os.Getenv(key)
	os.Setenv(key, val)
	t.Cleanup(func() { os.Setenv(key, oldVal) })
	return oldVal
}

func setupFile(t *testing.T, file string, content string) *os.File {
	f, err := os.Create(file)
	checkError(t, err)
	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = w.WriteString(content)
	checkError(t, err)
	w.Flush()

	t.Cleanup(func() { os.Remove(file) })
	return f
}

func setupDir(t *testing.T, dir string, files map[string]string) {
	err := os.Mkdir(dir, 0700)
	checkError(t, err)
	t.Cleanup(func() { os.Remove(dir) })

	for name, val := range files {
		setupFile(t, path.Join(dir, name), val)
	}
}

func checkError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err.Error())
	}
}

func assertTrue(t *testing.T, msg string, assertion bool) {
	if !assertion {
		t.Fatalf("Assertion failed: %v", msg)
	}
}
