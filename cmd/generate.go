// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"github.com/a-cordier/brp/langs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a binary resources source file",
	Long: `
	TODO: Write command description
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		ns, err := dirToNS(dir)

		if err != nil {
			return err
		}

		lang, _ := cmd.Flags().GetString("lang")

		if cmd.Flags().Changed("ns") {
			ns, _ = cmd.Flags().GetString("ns")
		}

		if err = generate(dir, lang, output, ns); err != nil {
			return err
		}
		return nil
	},
}

var (
	languages string
	folder    string
	output    string
	ns        string
)

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&languages, "lang", "l", "cpp", "A destination language")
	generateCmd.Flags().StringVarP(&folder, "dir", "d", "", "Your resource directory")
	generateCmd.Flags().StringVarP(&output, "output", "o", folder, "Output file without extension")
	generateCmd.Flags().StringVarP(&ns, "ns", "n", folder, "Namespace to access your resources")

	if err := generateCmd.MarkFlagRequired("dir"); err != nil {
		panic(err)
	}
}

func generate(dir, lang, output, ns string) error {
	source, err := langs.NewSource(lang, output, ns)
	if err != nil {
		return err
	}
	if err := addFiles(dir, source); err != nil {
		return err
	}
	tmpl, err := template.New(source.Name).Funcs(template.FuncMap{"Join": strings.Join}).Parse(source.Template)
	if err != nil {
		return err
	}

	file := os.Stdout

	if "" != source.Name {
		file, err = os.Create(source.GetFileName())
		if err != nil {
			return err
		}
	}

	if err := tmpl.Execute(file, source); err != nil {
		return err
	}

	return nil

}

func dirToNS(dir string) (ns string, err error) {
	abs, err := filepath.Abs(dir)
	ns = filepath.Base(abs)
	return
}

func newFile(path, dir string) (*langs.File, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &langs.File{
		ID:   fileID(path, dir),
		Data: format(data),
	}, nil
}

func fileID(path, src string) string {
	return strings.TrimPrefix(strings.TrimPrefix(path, src), "/")
}

func addFiles(dir string, source *langs.Source) error {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		file, err := newFile(path, dir)
		if err != nil {
			return err
		}
		source.Files = append(source.Files, file)
		return nil
	})
	return err
}

func format(data []byte) [][]string {
	hex := make([][]string, 0)

	for i := 0; i < len(data); i += 16 {
		chunks := data[i:min(i+16, len(data))]
		row := make([]string, len(chunks))

		for i, c := range chunks {
			row[i] = fmt.Sprintf("0x%02x", c)
		}

		hex = append(hex, row)
	}

	return hex
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
