package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/github/github-mcp-server/internal/ghmcp"
	"github.com/github/github-mcp-server/pkg/github"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// These variables are set by the build process using ldflags.
var version = "version"
var commit = "commit"
var date = "date"

var (
	rootCmd = &cobra.Command{
		Use:     "server",
		Short:   "GitHub MCP Server",
		Long:    `A GitHub MCP server that handles various tools and resources.`,
		Version: fmt.Sprintf("Version: %s\nCommit: %s\nBuild Date: %s", version, commit, date),
	}

	stdioCmd = &cobra.Command{
		Use:   "stdio",
		Short: "Start stdio server",
		Long:  `Start a server that communicates via standard input/output streams using JSON-RPC messages.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			token := viper.GetString("personal_access_token")
			if token == "" {
				return errors.New("GITHUB_PERSONAL_ACCESS_TOKEN not set")
			}

			var enabledToolsets []string
			if err := viper.UnmarshalKey("toolsets", &enabledToolsets); err != nil {
				return fmt.Errorf("failed to unmarshal toolsets: %w", err)
			}

			stdioServerConfig := ghmcp.StdioServerConfig{
				Version:              version,
				Host:                 viper.GetString("host"),
				Token:                token,
				EnabledToolsets:      enabledToolsets,
				DynamicToolsets:      viper.GetBool("dynamic_toolsets"),
				ReadOnly:             viper.GetBool("read-only"),
				ExportTranslations:   viper.GetBool("export-translations"),
				EnableCommandLogging: viper.GetBool("enable-command-logging"),
				LogFilePath:          viper.GetString("log-file"),
			}
			return ghmcp.RunStdioServer(stdioServerConfig)
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.SetGlobalNormalizationFunc(wordSepNormalizeFunc)

	rootCmd.SetVersionTemplate("{{.Short}}\n{{.Version}}\n")

	rootCmd.PersistentFlags().StringSlice("toolsets", github.DefaultTools, "An optional comma separated list of groups of tools to allow, defaults to enabling all")
	rootCmd.PersistentFlags().Bool("dynamic-toolsets", false, "Enable dynamic toolsets")
	rootCmd.PersistentFlags().Bool("read-only", false, "Restrict the server to read-only operations")
	rootCmd.PersistentFlags().String("log-file", "", "Path to log file")
	rootCmd.PersistentFlags().Bool("enable-command-logging", false, "When enabled, the server will log all command requests and responses to the log file")
	rootCmd.PersistentFlags().Bool("export-translations", false, "Save translations to a JSON file")
	rootCmd.PersistentFlags().String("gh-host", "", "Specify the GitHub hostname (for GitHub Enterprise etc.)")

	_ = viper.BindPFlag("toolsets", rootCmd.PersistentFlags().Lookup("toolsets"))
	_ = viper.BindPFlag("dynamic_toolsets", rootCmd.PersistentFlags().Lookup("dynamic-toolsets"))
	_ = viper.BindPFlag("read-only", rootCmd.PersistentFlags().Lookup("read-only"))
	_ = viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file"))
	_ = viper.BindPFlag("enable-command-logging", rootCmd.PersistentFlags().Lookup("enable-command-logging"))
	_ = viper.BindPFlag("export-translations", rootCmd.PersistentFlags().Lookup("export-translations"))
	_ = viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("gh-host"))

	rootCmd.AddCommand(stdioCmd)
}

func initConfig() {
	viper.SetEnvPrefix("github")
	viper.AutomaticEnv()
}

func wordSepNormalizeFunc(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	from := []string{"_"}
	to := "-"
	for _, sep := range from {
		name = strings.ReplaceAll(name, sep, to)
	}
	return pflag.NormalizedName(name)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	
    // 🔍 환경 변수 디버깅 로그 추가
    fmt.Println("DEBUG: GITHUB_PERSONAL_ACCESS_TOKEN =", os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN"))



	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "GitHub MCP Server is running")
	})

	http.HandleFunc("/run-stdio", func(w http.ResponseWriter, r *http.Request) {
		err := stdioCmd.RunE(stdioCmd, []string{})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to run stdio server: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "Stdio server started")
	})

	http.HandleFunc("/tools", toolsHandler)

	fmt.Printf("Listening on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start HTTP server: %v\n", err)
		os.Exit(1)
	}
	
	

}


type Tool struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    InputSpec   interface{} `json:"input_spec"`
    OutputSpec  interface{} `json:"output_spec"`
}

func toolsHandler(w http.ResponseWriter, r *http.Request) {
    tools := []Tool{
        {
            Name:        "Airbnb Search",
            Description: "Search Airbnb listings",
            InputSpec: map[string]string{
                "location":  "string",
                "check_in":  "date",
                "check_out": "date",
            },
            OutputSpec: map[string]string{
                "listings": "array",
            },
        },
        {
            Name:        "Airbnb Listing Details",
            Description: "Get details for a specific listing",
            InputSpec: map[string]string{
                "listing_id": "string",
            },
            OutputSpec: map[string]string{
                "details": "object",
            },
        },
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{"tools": tools})
}




// package main

// import (
// 	"errors"
// 	"fmt"
// 	"os"
// 	"strings"

// 	"github.com/github/github-mcp-server/internal/ghmcp"
// 	"github.com/github/github-mcp-server/pkg/github"
// 	"github.com/spf13/cobra"
// 	"github.com/spf13/pflag"
// 	"github.com/spf13/viper"
// )

// // These variables are set by the build process using ldflags.
// var version = "version"
// var commit = "commit"
// var date = "date"

// var (
// 	rootCmd = &cobra.Command{
// 		Use:     "server",
// 		Short:   "GitHub MCP Server",
// 		Long:    `A GitHub MCP server that handles various tools and resources.`,
// 		Version: fmt.Sprintf("Version: %s\nCommit: %s\nBuild Date: %s", version, commit, date),
// 	}

// 	stdioCmd = &cobra.Command{
// 		Use:   "stdio",
// 		Short: "Start stdio server",
// 		Long:  `Start a server that communicates via standard input/output streams using JSON-RPC messages.`,
// 		RunE: func(_ *cobra.Command, _ []string) error {
// 			token := viper.GetString("personal_access_token")
// 			if token == "" {
// 				return errors.New("GITHUB_PERSONAL_ACCESS_TOKEN not set")
// 			}

// 			// If you're wondering why we're not using viper.GetStringSlice("toolsets"),
// 			// it's because viper doesn't handle comma-separated values correctly for env
// 			// vars when using GetStringSlice.
// 			// https://github.com/spf13/viper/issues/380
// 			var enabledToolsets []string
// 			if err := viper.UnmarshalKey("toolsets", &enabledToolsets); err != nil {
// 				return fmt.Errorf("failed to unmarshal toolsets: %w", err)
// 			}

// 			stdioServerConfig := ghmcp.StdioServerConfig{
// 				Version:              version,
// 				Host:                 viper.GetString("host"),
// 				Token:                token,
// 				EnabledToolsets:      enabledToolsets,
// 				DynamicToolsets:      viper.GetBool("dynamic_toolsets"),
// 				ReadOnly:             viper.GetBool("read-only"),
// 				ExportTranslations:   viper.GetBool("export-translations"),
// 				EnableCommandLogging: viper.GetBool("enable-command-logging"),
// 				LogFilePath:          viper.GetString("log-file"),
// 			}
// 			return ghmcp.RunStdioServer(stdioServerConfig)
// 		},
// 	}
// )

// func init() {
// 	cobra.OnInitialize(initConfig)
// 	rootCmd.SetGlobalNormalizationFunc(wordSepNormalizeFunc)

// 	rootCmd.SetVersionTemplate("{{.Short}}\n{{.Version}}\n")

// 	// Add global flags that will be shared by all commands
// 	rootCmd.PersistentFlags().StringSlice("toolsets", github.DefaultTools, "An optional comma separated list of groups of tools to allow, defaults to enabling all")
// 	rootCmd.PersistentFlags().Bool("dynamic-toolsets", false, "Enable dynamic toolsets")
// 	rootCmd.PersistentFlags().Bool("read-only", false, "Restrict the server to read-only operations")
// 	rootCmd.PersistentFlags().String("log-file", "", "Path to log file")
// 	rootCmd.PersistentFlags().Bool("enable-command-logging", false, "When enabled, the server will log all command requests and responses to the log file")
// 	rootCmd.PersistentFlags().Bool("export-translations", false, "Save translations to a JSON file")
// 	rootCmd.PersistentFlags().String("gh-host", "", "Specify the GitHub hostname (for GitHub Enterprise etc.)")

// 	// Bind flag to viper
// 	_ = viper.BindPFlag("toolsets", rootCmd.PersistentFlags().Lookup("toolsets"))
// 	_ = viper.BindPFlag("dynamic_toolsets", rootCmd.PersistentFlags().Lookup("dynamic-toolsets"))
// 	_ = viper.BindPFlag("read-only", rootCmd.PersistentFlags().Lookup("read-only"))
// 	_ = viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file"))
// 	_ = viper.BindPFlag("enable-command-logging", rootCmd.PersistentFlags().Lookup("enable-command-logging"))
// 	_ = viper.BindPFlag("export-translations", rootCmd.PersistentFlags().Lookup("export-translations"))
// 	_ = viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("gh-host"))

// 	// Add subcommands
// 	rootCmd.AddCommand(stdioCmd)
// }

// func initConfig() {
// 	// Initialize Viper configuration
// 	viper.SetEnvPrefix("github")
// 	viper.AutomaticEnv()

// }

// func main() {
// 	if err := rootCmd.Execute(); err != nil {
// 		fmt.Fprintf(os.Stderr, "%v\n", err)
// 		os.Exit(1)
// 	}
// }

// func wordSepNormalizeFunc(_ *pflag.FlagSet, name string) pflag.NormalizedName {
// 	from := []string{"_"}
// 	to := "-"
// 	for _, sep := range from {
// 		name = strings.ReplaceAll(name, sep, to)
// 	}
// 	return pflag.NormalizedName(name)
// }
