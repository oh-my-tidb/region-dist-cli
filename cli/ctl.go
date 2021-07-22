// Copyright 2021 TiKV Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/disksing/region-dist-cli/cli/command"

	"github.com/chzyer/readline"
	"github.com/mattn/go-shellwords"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	cobra.EnablePrefixMatching = true
}

// GetRootCmd is exposed for integration tests. But it can be embedded into another suite, too.
func GetRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "pd-analyze",
		Short: "Placement Driver Analyze",
	}

	rootCmd.PersistentFlags().StringP("pd", "u", "http://172.16.4.3:35434/", "address of pd")
	rootCmd.AddCommand(
		command.NewHotRegionCommand(),
		command.NewRegionCommand(),
	)
	rootCmd.Flags().ParseErrorsWhitelist.UnknownFlags = true
	rootCmd.SilenceErrors = true
	return rootCmd
}

// MainStart start main command
func MainStart(args []string) {
	rootCmd := GetRootCmd()
	rootCmd.Flags().BoolP("version", "V", false, "Print version information and exit.")

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if v, err := cmd.Flags().GetBool("interact"); err == nil && v {
			readlineCompleter := readline.NewPrefixCompleter(genCompleter(cmd)...)
			loop(cmd.PersistentFlags(), readlineCompleter)
		}
	}

	rootCmd.SetArgs(args)
	rootCmd.ParseFlags(args)
	rootCmd.SetOutput(os.Stdout)

	if err := rootCmd.Execute(); err != nil {
		rootCmd.Println(err)
		os.Exit(1)
	}
}

func loop(persistentFlags *pflag.FlagSet, readlineCompleter readline.AutoCompleter) {
	l, err := readline.NewEx(&readline.Config{
		Prompt:            "\033[31m»\033[0m ",
		HistoryFile:       "/tmp/readline.tmp",
		AutoComplete:      readlineCompleter,
		InterruptPrompt:   "^C",
		EOFPrompt:         "^D",
		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	getREPLCmd := func() *cobra.Command {
		rootCmd := GetRootCmd()
		persistentFlags.VisitAll(func(flag *pflag.Flag) {
			if flag.Changed {
				rootCmd.PersistentFlags().Set(flag.Name, flag.Value.String())
			}
		})
		rootCmd.LocalFlags().MarkHidden("pd")
		rootCmd.LocalFlags().MarkHidden("cacert")
		rootCmd.LocalFlags().MarkHidden("cert")
		rootCmd.LocalFlags().MarkHidden("key")
		rootCmd.SetOutput(os.Stdout)
		return rootCmd
	}

	for {
		line, err := l.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				break
			} else if err == io.EOF {
				break
			}
			continue
		}
		if line == "exit" {
			os.Exit(0)
		}
		args, err := shellwords.Parse(line)
		if err != nil {
			fmt.Printf("parse command err: %v\n", err)
			continue
		}

		rootCmd := getREPLCmd()
		rootCmd.SetArgs(args)
		rootCmd.ParseFlags(args)
		if err := rootCmd.Execute(); err != nil {
			rootCmd.Println(err)
		}
	}
}

func genCompleter(cmd *cobra.Command) []readline.PrefixCompleterInterface {
	pc := []readline.PrefixCompleterInterface{}

	for _, v := range cmd.Commands() {
		if v.HasFlags() {
			flagsPc := []readline.PrefixCompleterInterface{}
			flagUsages := strings.Split(strings.Trim(v.Flags().FlagUsages(), " "), "\n")
			for i := 0; i < len(flagUsages)-1; i++ {
				flagsPc = append(flagsPc, readline.PcItem(strings.Split(strings.Trim(flagUsages[i], " "), " ")[0]))
			}
			flagsPc = append(flagsPc, genCompleter(v)...)
			pc = append(pc, readline.PcItem(strings.Split(v.Use, " ")[0], flagsPc...))
		} else {
			pc = append(pc, readline.PcItem(strings.Split(v.Use, " ")[0], genCompleter(v)...))
		}
	}
	return pc
}
