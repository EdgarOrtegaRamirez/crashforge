package main

import (
	"fmt"
	"os"

	"github.com/EdgarOrtegaRamirez/crashforge/analyzer"
	"github.com/EdgarOrtegaRamirez/crashforge/parser"
	"github.com/EdgarOrtegaRamirez/crashforge/reporter"
	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
	verbose bool
	format  string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "crashforge",
		Short: "Crash & Error Log Analyzer",
		Long:  "Analyze crash logs and stack traces from multiple programming languages.",
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "text", "output format (text, json, markdown)")

	rootCmd.AddCommand(newParseCmd())
	rootCmd.AddCommand(newAnalyzeCmd())
	rootCmd.AddCommand(newWatchCmd())
	rootCmd.AddCommand(newVersionCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newParseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "parse [file]",
		Short: "Parse a stack trace from a file or stdin",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text, err := readInput(args)
			if err != nil {
				return err
			}

			p := parser.New()
			info, err := p.Parse(text)
			if err != nil {
				return err
			}

			r := reporter.New(reporter.Format(format), verbose)
			fmt.Print(r.ReportSingle(info))
			return nil
		},
	}
}

func newAnalyzeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze [file]",
		Short: "Analyze multiple stack traces from a log file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text, err := readInput(args)
			if err != nil {
				return err
			}

			p := parser.New()
			errors := p.ParseMultiple(text)

			if len(errors) == 0 {
				fmt.Println("No stack traces found in input.")
				return nil
			}

			a := analyzer.New()
			result := a.Analyze(errors)

			r := reporter.New(reporter.Format(format), verbose)
			fmt.Print(r.Report(result))
			return nil
		},
	}
	return cmd
}

func newWatchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "watch [file]",
		Short: "Watch a log file for new errors",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Watching %s for errors... (Ctrl+C to stop)\n", args[0])
			fmt.Println("File watching not yet implemented. Use 'analyze' command instead.")
			return nil
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("crashforge v%s\n", version)
		},
	}
}

func readInput(args []string) (string, error) {
	if len(args) > 0 {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return string(data), nil
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", fmt.Errorf("no input provided. Pass a file or pipe input via stdin")
	}

	var data []byte
	buf := make([]byte, 4096)
	for {
		n, err := os.Stdin.Read(buf)
		data = append(data, buf[:n]...)
		if err != nil {
			break
		}
	}
	return string(data), nil
}
