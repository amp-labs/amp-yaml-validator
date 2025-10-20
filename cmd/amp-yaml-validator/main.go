package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	validator "github.com/amp-labs/amp-yaml-validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	separatorLength = 60 // Length of separator lines in output
)

var (
	cfgFile      string
	strictMode   bool
	skipProvider bool
	skipAsync    bool
)

var rootCmd = &cobra.Command{
	Use:   "amp-yaml-validator [path-to-amp.yaml]",
	Short: "Validate Ampersand amp.yaml manifest files",
	Long: `A command-line tool to validate Ampersand amp.yaml manifest files.

The validator checks for schema compliance, business logic rules, and best practices.
Exit codes:
  0 - Validation passed (no errors, warnings are OK)
  1 - Validation failed (one or more errors found)`,
	Args: cobra.ExactArgs(1),
	Run:  runValidation,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.amp-yaml-validator.yaml)")
	rootCmd.Flags().BoolVar(&strictMode, "strict", false, "treat warnings as errors")
	rootCmd.Flags().BoolVar(&skipProvider, "skip-provider", false, "skip provider-specific validation")
	rootCmd.Flags().BoolVar(&skipAsync, "skip-async", false, "skip async error prevention validation")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".amp-yaml-validator")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		_, _ = fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
}

func runValidation(cmd *cobra.Command, args []string) {
	filePath := args[0]

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		_, _ = fmt.Fprintf(os.Stderr, "Error: File not found: %s\n", filePath)
		os.Exit(1)
	}

	// Build validator options
	var opts []validator.Option
	if strictMode {
		opts = append(opts, validator.WithStrictMode(true))
	}

	if skipProvider {
		opts = append(opts, validator.WithSkipProviderValidation())
	}

	if skipAsync {
		opts = append(opts, validator.WithSkipAsyncValidation())
	}

	// Validate the file
	result, err := validator.ValidateFile(cmd.Context(), filePath, opts...)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error validating file: %v\n", err)
		os.Exit(1)
	}

	// Print results
	printResults(filePath, result)

	// Exit with appropriate code
	if !result.Valid {
		os.Exit(1)
	}

	os.Exit(0)
}

func printResults(filePath string, result *validator.ValidationResult) {
	fileName := filepath.Base(filePath)

	fmt.Printf("Validating: %s\n", fileName)          //nolint:forbidigo // CLI output to user
	fmt.Println(strings.Repeat("=", separatorLength)) //nolint:forbidigo // CLI output to user

	if result.Valid && len(result.Warnings) == 0 {
		fmt.Println("✓ Validation passed with no issues!") //nolint:forbidigo // CLI output to user

		return
	}

	// Print errors
	if len(result.Errors) > 0 {
		fmt.Printf("\n%s Errors (%d):\n", getSymbol("error"), len(result.Errors)) //nolint:forbidigo // CLI output

		for i, issue := range result.Errors {
			printIssue(i+1, issue)
		}
	}

	// Print warnings
	if len(result.Warnings) > 0 {
		fmt.Printf("\n%s Warnings (%d):\n", getSymbol("warning"), len(result.Warnings)) //nolint:forbidigo // CLI output

		for i, issue := range result.Warnings {
			printIssue(i+1, issue)
		}
	}

	// Print summary
	fmt.Println()                                     //nolint:forbidigo // CLI output to user
	fmt.Println(strings.Repeat("=", separatorLength)) //nolint:forbidigo // CLI output to user

	if result.Valid {
		fmt.Printf("✓ Validation passed with %d warning(s)\n", len(result.Warnings)) //nolint:forbidigo // CLI output
	} else {
		//nolint:forbidigo // CLI output to user
		fmt.Printf("✗ Validation failed with %d error(s) and %d warning(s)\n",
			len(result.Errors), len(result.Warnings))
	}
}

func printIssue(num int, issue validator.ValidationIssue) {
	fmt.Printf("\n  %d. [%s] %s\n", num, issue.Rule, issue.Message) //nolint:forbidigo // CLI output to user

	if issue.Path != "" {
		fmt.Printf("     Path: %s\n", issue.Path) //nolint:forbidigo // CLI output to user
	}

	if issue.Line > 0 {
		if issue.Column > 0 {
			fmt.Printf("     Location: line %d, column %d\n", issue.Line, issue.Column) //nolint:forbidigo // CLI output
		} else {
			fmt.Printf("     Location: line %d\n", issue.Line) //nolint:forbidigo // CLI output to user
		}
	}

	if issue.Suggestion != "" {
		fmt.Printf("     Suggestion: %s\n", issue.Suggestion) //nolint:forbidigo // CLI output to user
	}
}

func getSymbol(severity string) string {
	switch severity {
	case "error":
		return "✗"
	case "warning":
		return "⚠"
	default:
		return "•"
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
