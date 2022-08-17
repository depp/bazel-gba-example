package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gbajam/romimage"
)

type strflag struct {
	isset bool
	value string
}

func (f *strflag) Set(x string) error {
	*f = strflag{true, x}
	return nil
}

func (f *strflag) String() string {
	return f.value
}

var flagTitle strflag

func trimSuffix(s string) string {
	if i := strings.LastIndexByte(s, '.'); i != -1 {
		return s[:i]
	}
	return s
}

func mainE() error {
	// Args.
	flag.Var(&flagTitle, "title", "Game title. Defaults to input filename, without extension.")
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		return errors.New("missing argument: <program>")
	}
	if len(args) > 2 {
		return fmt.Errorf("too many arguments: got %d, expected 1 or 2", len(args))
	}
	progname := args[0]
	var outname string
	if len(args) >= 2 {
		outname = args[1]
	}
	title := flagTitle.value
	if !flagTitle.isset {
		title = strings.ToUpper(trimSuffix(filepath.Base(progname)))
	}
	if len(title) > romimage.TitleLength {
		old := title
		title = title[:romimage.TitleLength]
		fmt.Fprintf(os.Stderr, "Warning: title is too long, truncating %q to %q (%d bytes)\n", old, title, romimage.TitleLength)
	}

	// Execute.
	data, err := romimage.Make(progname, &romimage.Info{
		Title: title,
	})
	if err != nil {
		return err
	}

	if outname != "" {
		if err := ioutil.WriteFile(outname, data, 0666); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if err := mainE(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
