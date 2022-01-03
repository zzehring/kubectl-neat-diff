package main

import (
	"fmt"
	"github.com/go-clix/cli"
	neat "github.com/zzehring/kubectl-neat/v2/cmd"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	log.SetFlags(0)

	cmd := cli.Command{
		Use:   "kubectl-neat-diff [file1] [file2]",
		Short: "Remove fields from kubectl diff that carry low / no information",
		Args:  cli.ArgsExact(2),
	}

	ignoreLinesRegexes := cmd.Flags().StringSliceP("ignore-matching-lines", "I", []string{},
		"Ignore changes whose lines all match RegExp.")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		if err := neatifyDir(args[0]); err != nil {
			return err
		}
		if err := neatifyDir(args[1]); err != nil {
			return err
		}

		diffArgs := formDiffCmdArguments(*ignoreLinesRegexes, args)

		c := exec.Command("diff", diffArgs...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	if err := cmd.Execute(); err != nil {
		log.Fatalln("Error:", err)
	}
}

func neatifyDir(dir string) error {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		filename := filepath.Join(dir, fi.Name())
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		n, err := neat.NeatYAMLOrJSON(data, "same")
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(filename, []byte(n), fi.Mode()); err != nil {
			return err
		}
	}

	return nil
}

func formDiffCmdArguments(ignoreLinesRegexes, files []string) []string {
	args := []string{"-uN"}

	for _, ignoreLinesRegex := range ignoreLinesRegexes {
		args = append(args, fmt.Sprintf("-I %s", ignoreLinesRegex))
	}

	args = append(args, files...)
	return args
}
