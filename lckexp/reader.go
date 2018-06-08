package lckexp

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"strings"
)

type LockoExport struct {
	RawEntries  map[string]RawEntry
	RootEntries []LockoEntry
}

type LockoEntry struct {
	UUID     string
	Title    string
	Username string
	Password string
	Children []LockoEntry
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

	rawEntries := make(map[string]RawEntry)
	for _, info := range archive.File {
		if info.FileInfo().IsDir() {
			continue
		}

		file, err := info.Open()
		if err != nil {
			return nil, fmt.Errorf("error opening item %s: %s", info.Name, err)
		}
		defer file.Close()

		var entry RawEntry
		if err := json.NewDecoder(file).Decode(&entry); err != nil {
			return nil, fmt.Errorf("error reading item %s: %s", info.Name, err)
		}

		rawEntries[info.Name] = entry
	}

	entries, err := convertEntries(rawEntries)
	if err != nil {
		return nil, fmt.Errorf("error converting entries: %s", err)
	}

	return &LockoExport{
		RawEntries:  rawEntries,
		RootEntries: entries,
	}, nil
}

func convertEntries(raw map[string]RawEntry) ([]LockoEntry, error) {
	rootEntryNames := []string{}
	for k := range raw {
		if strings.ContainsRune(k, '/') {
			continue
		}

		rootEntryNames = append(rootEntryNames, k)
	}

	entries := []LockoEntry{}
	for _, name := range rootEntryNames {
		entry := raw[name]
		entries = append(entries, LockoEntry{
			UUID:     entry.UUID,
			Title:    entry.Title,
			Username: entry.Username(),
			Password: entry.Password(),
		})
		delete(raw, name)
	}

	return entries, nil
}
