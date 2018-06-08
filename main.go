package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
	"github.com/tobischo/gokeepasslib"
	"github.com/xperimental/locko2keepass/lckexp"
)

var masterPassword = "default"

func main() {
	log.SetFlags(0)

	pflag.Parse()
	files := pflag.Args()

	for i, file := range files {
		log.Printf("Processing %d: %s", i, file)

		data, err := lckexp.ReadExport(file)
		if err != nil {
			log.Printf("Error reading export: %s", err)
			continue
		}

		database, err := convertToKeepass(data.RootEntries)
		if err != nil {
			log.Printf("Error converting to Keepass database: %s", err)
			continue
		}

		if err := writeKeepass(file+".kdbx", database); err != nil {
			log.Printf("Error writing Keepass file: %s", err)
			continue
		}
	}
}

func convertToKeepass(rootEntries []*lckexp.LockoEntry) (*gokeepasslib.Group, error) {
	root := gokeepasslib.NewGroup()
	root.Name = "Locko Import"

	for _, e := range rootEntries {
		if len(e.Children) > 0 {
			subGroup, err := convertGroup(e)
			if err != nil {
				return nil, err
			}
			root.Groups = append(root.Groups, *subGroup)
			continue
		}

		entry := convertEntry(e)
		root.Entries = append(root.Entries, entry)
	}

	return &root, nil
}

func convertGroup(in *lckexp.LockoEntry) (*gokeepasslib.Group, error) {
	group := gokeepasslib.NewGroup()
	group.Name = in.Title

	for _, e := range in.Children {
		if len(e.Children) > 0 {
			subGroup, err := convertGroup(e)
			if err != nil {
				return nil, err
			}
			group.Groups = append(group.Groups, *subGroup)
			continue
		}

		entry := convertEntry(e)
		group.Entries = append(group.Entries, entry)
	}

	return &group, nil
}

func mkValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value}}
}

func mkProtectedValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value, Protected: true}}
}

func convertEntry(in *lckexp.LockoEntry) gokeepasslib.Entry {
	entry := gokeepasslib.NewEntry()
	if in.Title != "" {
		entry.Values = append(entry.Values, mkValue("Title", in.Title))
	}

	if in.Username != "" {
		entry.Values = append(entry.Values, mkValue("UserName", in.Username))
	}

	if in.Password != "" {
		entry.Values = append(entry.Values, mkProtectedValue("Password", in.Password))
	}
	return entry
}

func writeKeepass(fileName string, rootGroup *gokeepasslib.Group) error {
	db := &gokeepasslib.Database{
		Signature:   &gokeepasslib.DefaultSig,
		Headers:     gokeepasslib.NewFileHeaders(),
		Credentials: gokeepasslib.NewPasswordCredentials(masterPassword),
		Content: &gokeepasslib.DBContent{
			Meta: gokeepasslib.NewMetaData(),
			Root: &gokeepasslib.RootData{
				Groups: []gokeepasslib.Group{*rootGroup},
			},
		},
	}

	if err := db.LockProtectedEntries(); err != nil {
		return fmt.Errorf("error locking database: %s", err)
	}

	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file: %s", err)
	}
	defer file.Close()

	encoder := gokeepasslib.NewEncoder(file)
	if err := encoder.Encode(db); err != nil {
		return fmt.Errorf("error encoding database: %s", err)
	}

	return nil
}
