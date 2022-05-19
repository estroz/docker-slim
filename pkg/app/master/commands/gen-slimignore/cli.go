package genslimignore

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli/v2"

	"github.com/docker-slim/docker-slim/pkg/app"
	"github.com/docker-slim/docker-slim/pkg/slimignore"
	"github.com/docker-slim/docker-slim/pkg/slimignore/generate"
)

const (
	Name  = "gen-slimignore"
	Usage = "Generate a " + slimignore.SlimIgnoreFile + " file"
	Alias = "gs"

	flagFile   = "file"
	flagStdout = "stdout"

	flagLanguages     = "languages"
	flagListLanguages = "list-languages"
	flagKeepCerts     = "keep-certs"
)

var CLI = &cli.Command{
	Name:    Name,
	Aliases: []string{Alias},
	Usage:   Usage,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    flagFile,
			Aliases: []string{"f"},
			Usage:   "File to write slimignore to",
			Value:   slimignore.SlimIgnoreFile,
		},
		&cli.BoolFlag{
			Name:  flagStdout,
			Usage: "Write slimignore to stdout",
			Value: false,
		},
		&cli.StringSliceFlag{
			Name:  flagLanguages,
			Usage: "Languages in the image, ex. python3,nodejs",
			Value: cli.NewStringSlice(),
		},
		&cli.BoolFlag{
			Name:  flagKeepCerts,
			Usage: "Keep cert dirs in the image",
			Value: true,
		},

		&cli.BoolFlag{
			Name:    flagListLanguages,
			Aliases: []string{"l"},
			Usage:   "List all available languages",
			Value:   false,
		},
	},
	Action: func(ctx *cli.Context) error {
		xc := app.NewExecutionContext(Name)

		if ctx.Bool(flagListLanguages) {
			doList(ctx, xc)
			return nil
		}

		doGen(ctx, xc)

		return nil
	},
}

func doGen(ctx *cli.Context, xc *app.ExecutionContext) {

	var languages []string
	for _, l := range ctx.StringSlice(flagLanguages) {
		languages = append(languages, strings.Split(l, ",")...)
	}

	appStacks, err := generate.MakeAppStacks(languages)
	if err != nil {
		exitErr(xc, "param.error."+flagLanguages, err)
	}

	toStdout := ctx.Bool(flagStdout)
	outFile := ctx.String(flagFile)

	if toStdout && ctx.IsSet(flagFile) {
		exit(xc, "param.error.conflict", fmt.Sprintf("cannot set %s and %s together", flagStdout, flagFile))
	}

	g := generate.Generator{
		Stacks:    appStacks,
		KeepCerts: ctx.Bool(flagKeepCerts),
	}

	if toStdout {
		err = g.Run(ctx.Context, os.Stdout)
	} else {
		err = g.RunFile(ctx.Context, outFile)
		fmt.Printf("Wrote file output to: %s\n", outFile)
		app.ShowCommunityInfo()
	}
	if err != nil {
		exitErr(xc, "generator.error.run", err)
	}
}

func doList(ctx *cli.Context, xc *app.ExecutionContext) {

	allAppStacks := generate.ListAllAppStacks()

	keys := make([]string, len(allAppStacks))
	i := 0
	for name := range allAppStacks {
		keys[i] = name
		i++
	}
	sort.Strings(keys)

	tw := tabwriter.NewWriter(os.Stdout, 4, 4, 2, '\t', 0)

	for _, name := range keys {
		_ = allAppStacks[name]
		line := fmt.Sprintf("%s\tTypical %s language .slimignore paths, including certs.\n", name, strings.Title(name))
		if _, err := tw.Write([]byte(line)); err != nil {
			exitErr(xc, "list.error.write", err)
		}
	}

	if err := tw.Flush(); err != nil {
		exitErr(xc, "list.error.flush", err)
	}
}

type ovars = app.OutVars

func exitErr(xc *app.ExecutionContext, typ string, err error) {
	exit(xc, typ, err.Error())
}

func exit(xc *app.ExecutionContext, typ, errStr string) {
	xc.Out.Error(typ, errStr)
	xc.Out.State("exited", ovars{
		"exit.code": -1,
	})
	xc.Exit(-1)
}
