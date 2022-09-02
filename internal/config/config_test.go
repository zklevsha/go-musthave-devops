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
	"time"
)

var testServerJSON = ServerConfigJSON{
	ServerAddress:  "http://server.test",
	Key:            "test_hash",
	DSN:            "postgres://username:password@localhost:5432/database_name",
	PrivateKeyPath: "/tmp/test/private.pem",
	StoreInterval:  "3m",
	StoreFile:      "/tmp/test.json",
}

var testAgentConfig = AgentConfigJSON{
	ServerAddress:  "http://server.test",
	PollInterval:   "1s",
	ReportInterval: "3s",
	PublicKeyPath:  "/tmp/test/public.pem",
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
	tconfPollInterval, _ := time.ParseDuration(tconf.PollInterval)
	tconfReportInterval, _ := time.ParseDuration(tconf.ReportInterval)
	fname := "/tmp/TestLoadAgentConfig.json"
	createJson(fname, tconf)
	tt := []struct {
		name string
		args []string
		want AgentConfig
	}{
		{name: "no flags", args: []string{},
			want: AgentConfig{ServerAddress: serverAddressDefault,
				PollInterval: pollIntervalDefault, ReportInterval: reportIntervalDefault}},
		{name: "all flags", args: []string{"-a", "test_socket", "-c", "test_file.json",
			"-crypto-key", "test.pem", "-k", "test_hash", "-p", "5s", "-r", "20s"},
			want: AgentConfig{ServerAddress: "test_socket", Key: "test_hash",
				PollInterval: time.Second * 5, ReportInterval: time.Second * 20,
				PublicKeyPath: "test.pem"}},
		{name: "read from file", args: []string{"-c", fname},
			want: AgentConfig{ServerAddress: tconf.ServerAddress,
				Key: tconf.Key, PollInterval: tconfPollInterval,
				ReportInterval: tconfReportInterval, PublicKeyPath: tconf.PublicKeyPath}},
		{name: "bad duration", args: []string{"-p", "bad", "-r", "bad"},
			want: AgentConfig{ServerAddress: serverAddressDefault,
				PollInterval: pollIntervalDefault, ReportInterval: reportIntervalDefault}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res := GetAgentConfig(tc.args)
			if res != tc.want {
				t.Errorf("AgentConfig mismatch: have: %v,  want: %v", res, tc.want)
			}
		})
	}
	deleteJson(fname)
}

func TestAgentConfigEnv(t *testing.T) {
	want := AgentConfig{PollInterval: time.Second * 25,
		ReportInterval: time.Second * 14, ServerAddress: "test_serv",
		Key: "test_hash", PublicKeyPath: "public.pem"}
	t.Run("Get agent config with env variables", func(t *testing.T) {
		t.Setenv("POLL_INTERVAL", want.PollInterval.String())
		t.Setenv("REPORT_INTERVAL", want.ReportInterval.String())
		t.Setenv("ADDRESS", want.ServerAddress)
		t.Setenv("KEY", want.Key)
		t.Setenv("CRYPTO_KEY", want.PublicKeyPath)
		t.Setenv("CONFIG", "test.json")
		res := GetAgentConfig([]string{})
		if res != want {
			t.Errorf("AgentConfig mismatch: have: %v,  want: %v", res, want)
		}
	})

}

func TestGetServerConfigFlags(t *testing.T) {
	tconf := testServerJSON
	tconfStoreInterval, _ := time.ParseDuration(tconf.StoreInterval)
	fname := "/tmp/TestLoadServerConfig.json"
	createJson(fname, tconf)
	tt := []struct {
		name string
		args []string
		want ServerConfig
	}{
		{name: "no flags", args: []string{},
			want: ServerConfig{ServerAddress: serverAddressDefault,
				StoreFile: storeFileDefault, StoreInterval: storeIntervalDefault,
				Restore: false}},
		{name: "all flags", args: []string{
			"-a", "server", "-c", "config.json", "-f", "/tmp/test.json",
			"-k", "hash",
			"-d", "postgress//test:5432/tesd_db",
			"-i", "1s", "-r", "-crypto-key", "private.pem"},
			want: ServerConfig{
				ServerAddress: "server", Key: "hash", DSN: "postgress//test:5432/tesd_db",
				StoreFile: "/tmp/test.json", StoreInterval: time.Second,
				Restore: true, UseDB: true, PrivateKeyPath: "private.pem"}},
		{name: "read from file", args: []string{"-c", fname},
			want: ServerConfig{ServerAddress: tconf.ServerAddress,
				Key: tconf.Key, DSN: tconf.DSN,
				StoreFile: tconf.StoreFile, StoreInterval: tconfStoreInterval,
				Restore: false, UseDB: true, PrivateKeyPath: tconf.PrivateKeyPath}},
		{name: "bad store interval value", args: []string{"-i", "bad"},
			want: ServerConfig{ServerAddress: serverAddressDefault,
				StoreFile: storeFileDefault, StoreInterval: storeIntervalDefault,
				Restore: false, UseDB: false}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res := GetServerConfig(tc.args)
			if res != tc.want {
				t.Errorf("serverConfig mismatch: have: %v, want:%v", res, tc.want)
			}
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
		t.Setenv("CRYPTO_KEY", "private.pem")
		t.Setenv("CONFIG", "test.json")
		res := GetServerConfig([]string{})
		want := ServerConfig{ServerAddress: "testServ", StoreInterval: time.Second,
			StoreFile: "storeFile", Restore: true, UseDB: true, Key: "test_hash", DSN: "test_dsn",
			PrivateKeyPath: "private.pem"}
		if res != want {
			t.Errorf("ServerConfig mismatch: have: %v, want: %v", res, want)
		}
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
