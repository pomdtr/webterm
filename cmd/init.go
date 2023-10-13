package cmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
)

const manifestName = "com.pomdtr.popcorn.json"

var (
	//go:embed manifest.json
	manifest []byte
	//go:embed entrypoint.sh.gotmpl
	entrypoint []byte
)

var (
	manifestTmpl   = template.Must(template.New("manifest").Parse(string(manifest)))
	entrypointTmpl = template.Must(template.New("entrypoint").Parse(string(entrypoint)))
	manifestPaths  = map[string]string{
		"chrome":      filepath.Join(xdg.DataHome, "Google", "Chrome", "NativeMessagingHosts", manifestName),
		"chrome-beta": filepath.Join(xdg.DataHome, "Google", "Chrome Beta", "NativeMessagingHosts", manifestName),
		"edge":        filepath.Join(xdg.DataHome, "microsoft", "edge", "NativeMessagingHosts", manifestName),
		"brave":       filepath.Join(xdg.DataHome, "BraveSoftware", "Brave-Browser", "NativeMessagingHosts", manifestName),
		"vivaldi":     filepath.Join(xdg.DataHome, "vivaldi", "NativeMessagingHosts", manifestName),
		"arc":         filepath.Join(xdg.DataHome, "Google", "Chrome", "NativeMessagingHosts", manifestName),
	}
)

func NewCmdInit() *cobra.Command {
	flags := struct {
		Browser          string
		ExtensionID      string
		ProfileDirectory string
	}{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init configuration for a browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("unable to get user home directory: %w", err)
			}

			manifestPath, ok := manifestPaths[flags.Browser]
			if !ok {
				return fmt.Errorf("invalid browser: %s", flags.Browser)
			}

			if flags.ProfileDirectory != "" {
				manifestPath = filepath.Join(manifestPath, flags.ProfileDirectory)
			}

			cmd.Printf("Writing manifest file to %s\n", manifestPath)
			if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
				return fmt.Errorf("unable to create manifest directory: %w", err)
			}

			manifestBuffer := bytes.Buffer{}
			if err := manifestTmpl.Execute(&manifestBuffer, map[string]string{
				"homeDir":     homeDir,
				"extensionID": flags.ExtensionID,
			}); err != nil {
				return fmt.Errorf("unable to execute manifest template: %w", err)
			}

			if err := os.WriteFile(manifestPath, manifestBuffer.Bytes(), 0644); err != nil {
				return fmt.Errorf("unable to write manifest file: %w", err)
			}
			cmd.Printf("Manifest file written successfully\n")

			if err := os.MkdirAll(filepath.Join(homeDir, ".local", "bin"), 0755); err != nil {
				return fmt.Errorf("unable to create entrypoint directory: %w", err)
			}

			entrypointBuffer := bytes.Buffer{}
			execPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("unable to get executable path: %w", err)
			}
			if err := entrypointTmpl.Execute(&entrypointBuffer, map[string]string{
				"popcornBin": execPath,
				"browser":    flags.Browser,
			}); err != nil {
				return fmt.Errorf("unable to execute entrypoint template: %w", err)
			}

			entrypointPath := filepath.Join(homeDir, ".local", "bin", "popcorn.sh")
			cmd.Printf("Writing entrypoint file to %s\n", entrypointPath)
			if err := os.WriteFile(entrypointPath, entrypointBuffer.Bytes(), 0755); err != nil {
				return fmt.Errorf("unable to write entrypoint file: %w", err)
			}
			cmd.Printf("Entrypoint file written successfully\n")

			return nil
		},
	}

	cmd.Flags().StringVar(&flags.Browser, "browser", "", "Browser to install the extension for")
	cmd.MarkFlagRequired("browser")
	cmd.RegisterFlagCompletionFunc("browser", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		for browser := range manifestPaths {
			completions = append(completions, browser)
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.Flags().StringVar(&flags.ExtensionID, "extension-id", "", "Extension ID to install")
	cmd.MarkFlagRequired("extension-id")

	cmd.Flags().StringVar(&flags.ProfileDirectory, "profile-directory", "", "Profile Directory")

	return cmd
}
