package meta

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Station-Manager/errors"
	"github.com/goccy/go-json"
)

type SqliteMeta struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

func LoadSqliteDatabaseList(databaseDir string) ([]SqliteMeta, error) {
	const op errors.Op = "meta.LoadSqliteDatabaseList"
	databaseDir = strings.TrimSpace(databaseDir)
	if databaseDir == "" {
		return nil, errors.New(op).Msg("Database directory cannot be empty")
	}

	data, err := readFile(filepath.Join(databaseDir, ".meta.json"))
	if err != nil {
		return nil, errors.New(op).Err(err)
	}

	var meta []SqliteMeta
	if err = json.Unmarshal(data, &meta); err != nil {
		return nil, errors.New(op).Err(err)
	}

	return meta, nil
}

func readFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var data []byte
	if data, err = io.ReadAll(file); err != nil {
		return nil, err
	}

	return data, nil
}
