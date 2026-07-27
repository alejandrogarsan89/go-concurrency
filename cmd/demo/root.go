package main

import "github.com/spf13/cobra"

// rootCmd builds the top-level `demo` command with all pattern subcommands.
func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "demo",
		Short: "Runnable demos for the go-concurrency patterns library",
		Long: "go-concurrency demo runner.\n\n" +
			"Explore Go concurrency and parallelism patterns from the command line.\n" +
			"Examples:\n" +
			"  demo waitgroup --tasks 8\n" +
			"  demo generator --n 10 --take 3\n" +
			"  demo fanin --sources 3 --per 4",
		SilenceUsage: true,
	}
	root.AddCommand(waitgroupCmd())
	root.AddCommand(generatorCmd())
	root.AddCommand(faninCmd())
	root.AddCommand(poolCmd())
	root.AddCommand(pipelineCmd())
	return root
}
