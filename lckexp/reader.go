package lckexp

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type LockoExport struct {
	RootEntries []*LockoEntry
}

type LockoEntry struct {
	UUID     string
	Title    string
	Username string
	Password string
	Children []*LockoEntry
}

type RawEntry struct {
	UUID         string    `json:"uuid"`
	Title        string    `json:"title"`
	Type         int       `json:"type"`
	SortIndex    int       `json:"sortIndex"`
	DateModified float64   `json:"dateModified"`
	DateCreated  float64   `json:"dateCreated"`
	Data         EntryData `json:"data"`
	/*{
	"dateLastUsed" : 0,
	"dateFavorited" : 0,
	"dateTrashed" : 0
	}*/
}

func (e RawEntry) Username() string {
	username, ok := e.Data.Fields["username"]
	if !ok {
		return ""
	}

	return fmt.Sprintf("%s", username)
}

func (e RawEntry) Password() string {
	password, ok := e.Data.Fields["password"]
	if !ok {
		return ""
	}

	return fmt.Sprintf("%s", password)
}

type EntryData struct {
	Fields          map[string]interface{} `json:"fields"`
	PasswordHistory EntryPasswordHistory   `json:"passwordHistory"`
}

type EntryPasswordHistory struct {
	Password []HistoricPassword `json:"password"`
}

type HistoricPassword struct {
	Value     string  `json:"value"`
	Timestamp float64 `json:"timestamp"`
}

func ReadExport(fileName string) (*LockoExport, error) {
	archive, err := zip.OpenReader(fileName)
	if err != nil {
		return nil, fmt.Errorf("error opening archive: %s", err)
	}

	entries := []*LockoEntry{}
	for _, info := range archive.File {
		if info.FileInfo().IsDir() {
			continue
		}

		rawEntry, err := readRawEntry(info)
		if err != nil {
			return nil, fmt.Errorf("error reading item %s: %s", info.Name, err)
		}

		converted := &LockoEntry{
			UUID:     rawEntry.UUID,
			Title:    rawEntry.Title,
			Username: rawEntry.Username(),
			Password: rawEntry.Password(),
		}

		id := strings.TrimSuffix(info.Name, ".item")
		dir, base := splitName(id)

		if dir == "." {
			entries = append(entries, converted)
			continue
		}

		parent, ok := findParent(entries, dir)
		if !ok {
			return nil, fmt.Errorf("can not find parent %s for %s", dir, base)
		}

		parent.Children = append(parent.Children, converted)
	}

	return &LockoExport{
		RootEntries: entries,
	}, nil
}

func readRawEntry(info *zip.File) (*RawEntry, error) {
	file, err := info.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening: %s", err)
	}
	defer file.Close()

	var entry RawEntry
	if err := json.NewDecoder(file).Decode(&entry); err != nil {
		return nil, fmt.Errorf("parse error: %s", err)
	}

	return &entry, nil
}

func splitName(name string) (string, string) {
	dir := filepath.Dir(name)
	base := filepath.Base(name)
	return dir, base
}

func findParent(entries []*LockoEntry, name string) (*LockoEntry, bool) {
	dir, base := splitName(name)

	if dir != "." {
		// TODO implement sub-folders
		return nil, false
	}

	for _, e := range entries {
		if e.UUID == base {
			return e, true
		}
	}

	return nil, false
}
