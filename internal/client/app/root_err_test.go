package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute_RegisterError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "register", "-l", "test", "-p", "pass", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Ошибка регистрации")
}

func TestExecute_LoginError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "login", "-l", "test", "-p", "wrong", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Ошибка входа")
}

func TestExecute_AddError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "add", "-t", "login", "-d", "mydata", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Ошибка добавления")
}

func TestExecute_UpdateError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "update", "sec-1", "-d", "newdata", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Ошибка обновления")
}

func TestExecute_ListError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "list", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Ошибка получения списка")
}

func TestExecute_GetError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "get", "999", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Ошибка получения секрета")
}

func TestExecute_DeleteError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "delete", "999", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Ошибка удаления")
}

func TestExecute_SyncError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	os.Args = []string{"keeper", "sync", "-s", ts.URL}

	home := t.TempDir()
	os.Setenv("HOME", home)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "Ошибка синхронизации")
}

func TestExecute_InvalidConfig(t *testing.T) {
	os.Args = []string{"keeper", "list"}

	home := t.TempDir()
	os.Setenv("HOME", home)
	// write invalid json to config
	cfgPath := home + "/.gophkeeper/config.json"
	os.MkdirAll(home+"/.gophkeeper", 0700)
	os.WriteFile(cfgPath, []byte("{invalid"), 0600)

	output := captureOutput(func() {
		Execute()
	})

	assert.Contains(t, output, "invalid character")
}
