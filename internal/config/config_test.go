package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
)

var testServerJSON = ServerConfigJSON{
	ServerAddress:  "http://server.test",
	Key:            "test_hash",
	DSN:            "postgres://username:password@localhost:5432/database_name",
	PrivateKeyPath: "/tmp/test/private.pem",
}

var testAgentConfig = AgentConfigJSON{
	ServerAddress:  "http://server.test",
	PollInterval:   "1s",
	ReportInterval: "3s",
	PublicKeyPath:  "/tmp/test/private.pem",
	Key:            "test_hash",
}

// creating json file
func createJson(fname string, s interface{}) {

	file, err := json.Marshal(s)
	if err != nil {
		log.Fatalf("failed to marshal struct: %s", err.Error())
	}

	err = ioutil.WriteFile(fname, file, 0644)
	if err != nil {
		log.Fatalf("cant create %s: %s", fname, err.Error())
	}
}

// removing json file
func deleteJson(fname string) {
	err := os.Remove(fname)
	if err != nil {
		log.Fatalf("cant delete %s: %s", fname, err.Error())
	}
}

func TestGetAgentConfig(t *testing.T) {
	tconf := testAgentConfig
	fname := "/tmp/TestLoadAgentConfig.json"
	createJson(fname, tconf)
	tt := []struct {
		name string
		args []string
	}{
		{name: "no flags", args: []string{}},
		{name: "all flags", args: []string{"-a", "test_socket", "-c", "test_file.json",
			"-crypto-key", "test.pem", "-k", "test_hash", "-p", "1s", "-r", "1"}},
		{name: "read from file", args: []string{"-c", fname}},
		{name: "bad duration", args: []string{"-p", "bad", "-r", "bad"}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			GetAgentConfig(tc.args)
		})
	}
	deleteJson(fname)
}

func TestAgentConfigEnv(t *testing.T) {
	t.Run("Get agent config with env variables", func(t *testing.T) {
		t.Setenv("POLL_INTERVAL", "1s")
		t.Setenv("REPORT_INTERVAL", "1s")
		t.Setenv("ADDRESS", "test_serv")
		t.Setenv("KEY", "test_hash")
		t.Setenv("CRYPTO_KEY", "test.pem")
		t.Setenv("CONFIG", "test.json")
		GetAgentConfig([]string{})
	})
}

func TestGetServerConfigFlags(t *testing.T) {
	tconf := testServerJSON
	fname := "/tmp/TestLoadServerConfig.json"
	createJson(fname, tconf)
	tt := []struct {
		name string
		args []string
	}{
		{name: "no flags", args: []string{}},
		{name: "all flags", args: []string{"-a", "server", "-c",
			"config.json", "-d", "postgress//test:5432/tesd_db", "-f",
			"/tmp/test.json", "-i", "1s", "-k", "hash"}},
		{name: "read from file", args: []string{"-c", fname}},
		{name: "bad store interval value", args: []string{"-i", "bad"}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			GetServerConfig(tc.args)
		})
	}
	deleteJson(fname)
}

func TestGetServerConfigEnv(t *testing.T) {
	t.Run("Get server config with env variables", func(t *testing.T) {
		t.Setenv("ADDRESS", "testServ")
		t.Setenv("STORE_INTERVAL", "1s")
		t.Setenv("STORE_FILE", "storeFile")
		t.Setenv("RESTORE", "true")
		t.Setenv("KEY", "test_hash")
		t.Setenv("DATABASE_DSN", "test_dsn")
		t.Setenv("CRYPTO_KEY", "test")
		t.Setenv("CONFIG", "test.json")
		GetServerConfig([]string{})
	})
}

func TestLoadServerConfig(t *testing.T) {
	tc := testServerJSON
	fname := "/tmp/TestLoadServerConfig.json"
	createJson(fname, tc)
	t.Run("TestLoadServerConfig", func(t *testing.T) {
		c, err := loadServerConfig(fname)
		if err != nil {
			t.Errorf("cant loadServerConfig: %s", err.Error())
		}

		if !reflect.DeepEqual(tc, c) {
			t.Errorf("structs does not match: have: %v, want: %v", c, tc)
		}
	})
	deleteJson(fname)
}

func TestLoadAgentConfig(t *testing.T) {
	tc := testAgentConfig
	fname := "/tmp/TestLoadAgentConfig.json"
	createJson(fname, tc)
	t.Run("TestLoadAgentConfig", func(t *testing.T) {
		c, err := loadAgentConfig(fname)
		if err != nil {
			t.Errorf("cant loadAgentConfig: %s", err.Error())
		}

		if !reflect.DeepEqual(tc, c) {
			t.Errorf("structs does not match: have: %v, want: %v", c, tc)
		}
	})
	deleteJson(fname)
}

func TestIsFlagPassed(t *testing.T) {
	t.Run("TestIsFlagPassed", func(t *testing.T) {
		f := flag.NewFlagSet("test", flag.ExitOnError)
		var b bool
		f.BoolVar(&b, "t", true, "test flag")
		args := []string{"-t"}
		err := f.Parse(args)
		if err != nil {
			t.Errorf("failed to parse args: %s", err.Error())
		}
		fmt.Println(f.Args())
		if !isFlagPassed("t", f) {
			t.Error("failed to detect passed bool flag 't'")
		}
	})
}
