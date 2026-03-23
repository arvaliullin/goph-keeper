package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/arvaliullin/goph-keeper/internal/client/config"
	"github.com/stretchr/testify/assert"
)

func captureOutput(f func()) string {
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = origStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestExecute_Register(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "register", "-l", "test", "-p", "pass", "-s", ts.URL}

	// Create temp dir for config
	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Успешная регистрация!")

	cfg, _ := config.LoadConfig()
	assert.Equal(t, "test-token", cfg.Token)
}

func TestExecute_Login(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": "login-token"})
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "login", "-l", "test", "-p", "pass", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Успешный вход!")
}

func TestExecute_Add(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "new-sec-id"})
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "add", "-t", "text", "-d", "mydata", "-m", "mymeta", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Секрет добавлен")
}

func TestExecute_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": "1", "type": "text", "metadata": "dGVzdC1tZXRh"},
		})
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "list", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Найдено секретов")
	assert.Contains(t, output, "Meta:")
}

func TestExecute_Logout(t *testing.T) {
	home := t.TempDir()
	os.Setenv("HOME", home)

	cfgDir := home + "/.gophkeeper"
	os.MkdirAll(cfgDir, 0700)
	os.WriteFile(cfgDir+"/config.json", []byte(`{"token":"old-token","server":"http://localhost:8080"}`), 0600)

	os.Args = []string{"keeper", "logout"}

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Выход выполнен")

	cfg, err := config.LoadConfig()
	assert.NoError(t, err)
	assert.Empty(t, cfg.Token)
}

func TestExecute_Update(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "sec-1"})
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "update", "sec-1", "-d", "newdata", "-m", "newmeta", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Секрет обновлен")
}

func TestExecute_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": "1", "type": "text", "data": "eW8="})
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "get", "1", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "ID: 1")
}

func TestExecute_Delete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "delete", "sec-1", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Секрет удален")
}

func TestExecute_Sync(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": "1", "type": "text", "data": "eW8="},
		})
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "sync", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Получено изменений: 1")
}

func TestExecute_Version(t *testing.T) {
	SetBuildInfo("v1.0.0", "2026-01-01")

	os.Args = []string{"keeper", "version"}

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Build version: v1.0.0")
	assert.Contains(t, output, "Build date: 2026-01-01")
}

func TestExecute_AddBinaryType(t *testing.T) {
	os.Args = []string{"keeper", "add", "-t", "binary", "-d", "data", "-s", "http://localhost:9999"}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "add-binary")
}

func TestExecute_DownloadBinary(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("binary-content"))
	}))
	defer ts.Close()

	home := t.TempDir()
	os.Setenv("HOME", home)
	outPath := home + "/out.bin"

	os.Args = []string{"keeper", "download-binary", "sec-1", "-o", outPath, "-s", ts.URL}

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Файл сохранен")
	data, err := os.ReadFile(outPath)
	assert.NoError(t, err)
	assert.Equal(t, "binary-content", string(data))
}

func TestExecute_GetBinaryWithOutput(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path == "/api/v1/secrets/sec-1" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"id": "sec-1", "type": "binary"})
			return
		}
		if r.URL.Path == "/api/v1/binary/sec-1" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("bin-data"))
			return
		}
	}))
	defer ts.Close()

	home := t.TempDir()
	os.Setenv("HOME", home)
	outPath := home + "/get-out.bin"

	os.Args = []string{"keeper", "get", "sec-1", "-o", outPath, "-s", ts.URL}

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Бинарные данные сохранены")
}

func TestExecute_ListNoMeta(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": "1", "type": "text"},
		})
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "list", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Найдено секретов")
	assert.NotContains(t, output, "Meta:")
}

func TestExecute_SyncWithPreviousSync(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.URL.Query().Get("updated_after"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	}))
	defer ts.Close()

	home := t.TempDir()
	os.Setenv("HOME", home)
	cfgDir := home + "/.gophkeeper"
	os.MkdirAll(cfgDir, 0700)
	os.WriteFile(cfgDir+"/config.json", []byte(`{"token":"t","server":"`+ts.URL+`","last_sync":"2026-01-01T00:00:00Z"}`), 0600)

	os.Args = []string{"keeper", "sync"}

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Получено изменений: 0")
}

func TestExecute_GetDeletedSecret(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "1", "type": "text", "deleted_at": "2026-01-01T00:00:00Z",
		})
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "get", "1", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Deleted at:")
}

func TestExecute_DownloadBinaryError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	home := t.TempDir()
	os.Setenv("HOME", home)

	os.Args = []string{"keeper", "download-binary", "sec-1", "-o", home + "/out.bin", "-s", ts.URL}

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Ошибка скачивания")
}
