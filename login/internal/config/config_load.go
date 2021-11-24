package config

import (
	"bytes"
	"context"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/sethvargo/go-envconfig"
	"gopkg.in/yaml.v2"
)

type dirLookup struct {
	path string
}

func (dl *dirLookup) Lookup(key string) (string, bool) {
	kebabKey := strings.ReplaceAll(key, "_", "-")
	names := []string{
		key,
		strings.ToLower(key),
		kebabKey,
		strings.ToLower(kebabKey),
	}

	//	var errors []error

	for _, name := range names {
		val, err := os.ReadFile(path.Join(dl.path, name))
		if err == nil {
			return string(val), true
		}
		//		errors = append(errors, err)
	}

	// for _, e := range errors {
	// 	log.Printf("Failed to load var %v file: %v", key, e)
	// }

	return "", false
}

func loadDir(data interface{}, path string) {
	err := envconfig.ProcessWith(context.Background(), data, &dirLookup{path})
	if err != nil {
		log.Printf("Failed to load dir: %v", err)
	}
}

func loadYaml(data interface{}, file string) {
	f, err := os.Open(file)
	if err != nil {
		log.Printf("Failed to load file %v: %v", file, err)
		return
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(data)
	if err != nil {
		log.Printf("Failed to decode file %v: %v", file, err)
	}
}

func loadEnv(data interface{}) {
	err := envconfig.Process(context.Background(), data)
	if err != nil {
		log.Printf("Failed to load env: %v", err)
	}
}

func expandTemplate(expand *string, args interface{}) {
	buff := bytes.NewBufferString("")
	template.Must(template.New(*expand).Parse(*expand)).Execute(buff, args)
	*expand = buff.String()
}
