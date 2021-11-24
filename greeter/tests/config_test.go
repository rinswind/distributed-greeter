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
  Dsn: "mysql://{{ .User }}:{{ .Password }}@tcp({{ .Endpoint }})/{{ .Name }}"
#  Dsn: "mysql://static-user:static-password@tcp(static-endpoint)/static-db"
  Name: yaml-db-name
  User: yaml-db-user
  Endpoint: yaml-db-endpoint:3306
DbConfigDir: db-creds

Redis:
  Dsn: "redis://{{ .Endpoint }}/{{ .Db }}"
  Db: 100
  Endpoint: yaml-redis-endpoint
RedisConfigDir: redis-creds

AccessTokenConfigDir: at-creds
`)

	setupEnv(t, "DB_ENDPOINT", "env-db-endpoint")

	setupDir(t, "db-creds", map[string]string{"db_user": "file-db-user", "DB_PASSWORD": "file-db-password"})
	setupDir(t, "redis-creds", map[string]string{"redis_password": "file-redis-password", "redis-endpoint": "file-redis-endpoint"})
	setupDir(t, "at-creds", map[string]string{"at_access_token_secret": "file-at-secret", "at_refresh_token_secret": "file-rt-secret"})

	cfg := config.ReadConfig()
	t.Logf("%+v\n", cfg)

	assertTrue(t, "DB_ENDPOINT override", cfg.Db.Endpoint == "env-db-endpoint")
	assertTrue(t, "Db.User file override", cfg.Db.User == "file-db-user")
	assertTrue(t, "Db.Password file override", cfg.Db.Password == "file-db-password")
	assertTrue(t, "Db.Dsn is correct", cfg.Db.Dsn == "mysql://file-db-user:file-db-password@tcp(env-db-endpoint)/yaml-db-name")
	assertTrue(t, "Redis.Dsn is correct", cfg.Redis.Dsn == "redis://file-redis-endpoint/100")
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
