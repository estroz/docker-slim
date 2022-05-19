package genslimkeep

import (
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli/v2"

	"github.com/docker-slim/docker-slim/pkg/app"
	"github.com/docker-slim/docker-slim/pkg/slimkeep"
	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate"
	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate/appstack"
)

const (
	Name  = "gen-slimkeep"
	Usage = "Generate a " + slimkeep.SlimKeepFile + " file"
	Alias = "gs"

	flagFile      = "file"
	flagStdout    = "stdout"
	flagLanguages = "languages"
	flagKeepCerts = "keep-certs"

	flagListLanguages = "list-languages"

	flagBuilder     = "builder"
	flagFromImage   = "from-image"
	flagBuildPrefix = "build-prefix"
)

var CLI = &cli.Command{
	Name:    Name,
	Aliases: []string{Alias},
	Usage:   Usage,
	Flags: []cli.Flag{
		// Gen flags
		&cli.StringFlag{
			Name:    flagFile,
			Aliases: []string{"f"},
			Usage:   "File to write slimkeep to",
			Value:   slimkeep.SlimKeepFile,
		},
		&cli.BoolFlag{
			Name:  flagStdout,
			Usage: "Write slimkeep to stdout",
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

		// List flags
		&cli.BoolFlag{
			Name:    flagListLanguages,
			Aliases: []string{"l"},
			Usage:   "List all available languages",
			Value:   false,
		},

		// Build flags
		&cli.StringFlag{
			Name:  flagBuilder,
			Usage: "Language to use to build a slimkeep for the current dir",
		},
		&cli.StringFlag{
			Name:  flagBuildPrefix,
			Usage: "In-container directory prefix to append to the language's slimkeep entry",
		},
		&cli.StringFlag{
			Name:  flagFromImage,
			Usage: "Build a slimkeep from an image's file tree (not implemented yet)",
		},
	},
	Action: func(ctx *cli.Context) error {
		xc := app.NewExecutionContext(Name)

		listLangs := ctx.Bool(flagListLanguages)
		builder := ctx.String(flagBuilder)

		switch {
		case listLangs:
			doList(ctx, xc)
		case builder != "":
			doBuild(ctx, xc, builder)
		default:
			doGen(ctx, xc)
		}

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

	runStacks(ctx, xc, appStacks...)
}

func runStacks(ctx *cli.Context, xc *app.ExecutionContext, appStacks ...appstack.AppStack) {
	toStdout := ctx.Bool(flagStdout)
	outFile := ctx.String(flagFile)

	if toStdout && ctx.IsSet(flagFile) {
		exit(xc, "param.error.conflict", fmt.Sprintf("cannot set %s and %s together", flagStdout, flagFile))
	}

	g := generate.Generator{
		Stacks:    appStacks,
		KeepCerts: ctx.Bool(flagKeepCerts),
	}

	var err error
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

	tw := tabwriter.NewWriter(os.Stdout, 20, 10, 2, '\t', 0)

	header := "NAME\tBUILDER\tDESC\n"
	if _, err := tw.Write([]byte(header)); err != nil {
		exitErr(xc, "list.error.write", err)
	}

	for _, name := range keys {
		stack := allAppStacks[name]

		builderID := ""
		if _, isBuilder := stack.(appstack.Builder); isBuilder {
			builderID = "yes"
		}

		desc := fmt.Sprintf("Typical %s language .slimkeep paths, including certs.", strings.Title(name))
		line := fmt.Sprintf("%s\t%s\t%s\n", name, builderID, desc)
		if _, err := tw.Write([]byte(line)); err != nil {
			exitErr(xc, "list.error.write", err)
		}
	}

	if err := tw.Flush(); err != nil {
		exitErr(xc, "list.error.flush", err)
	}
}

func doBuild(ctx *cli.Context, xc *app.ExecutionContext, builderName string) {

	stackName := strings.TrimSuffix(builderName, ".builder")

	stack, err := generate.GetAppStack(stackName)
	if err != nil {
		exitErr(xc, "params.error.get", err)
	}

	builder, isBuilder := stack.(appstack.Builder)
	if !isBuilder {
		exit(xc, "builder.error.get", "language does not have a builder")
	}

	prefix := ctx.String(flagBuildPrefix)

	fromImage := ctx.String(flagFromImage)
	var buildFS fs.FS
	if fromImage == "" {
		if prefix != "" && prefix[len(prefix)-1] == '/' {
			prefix = prefix[:len(prefix)-1]
		}
		buildFS = os.DirFS(".")
	} else {
		if prefix != "" {
			exit(xc, "params.error", "build-prefix cannot be set with from-image")
		}
		if buildFS, err = newImageFS(fromImage); err != nil {
			exitErr(xc, "builder.error.image.walk", err)
		}
	}

	if err := builder.Build(buildFS, prefix); err != nil {
		exitErr(xc, "builder.error.build", err)
	}

	runStacks(ctx, xc, stack)
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
