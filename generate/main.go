package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

const EnvMainTemplate = `package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

    "github.com/lonegunmanb/genv/pkg"
	"github.com/spf13/cobra"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	downloadInstaller, _ := pkg.NewDownloadInstaller("{{  .DownloadUrlTemplate }}", ctx)
	goBuildInstaller := pkg.NewGoBuildInstaller("{{ .GoBuildRepoUrl }}", "{{ .BinaryName }}", "{{ .GoBuildSubFolder }}", ctx)
	fallbackInstaller := pkg.NewFallbackInstaller(downloadInstaller, goBuildInstaller)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err.Error())
	}
	env := pkg.NewEnv(homeDir, "{{ .Name }}", "{{ .BinaryName }}", fallbackInstaller)

	// Listen for interrupt signal (Ctrl + C) and cancel the context when received
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			cancel()
		}
	}()

	var rootCmd = &cobra.Command{Use: "{{ .Name }}"}

	var cmdInstall = &cobra.Command{
		Use:   "install [version]",
		Short: "Install a specific version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			fmt.Printf("Installing version: %s\n", version)
			return env.Install(version)
		},
	}

	var cmdUse = &cobra.Command{
		Use:   "use [version]",
		Short: "Use a specific version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			fmt.Printf("Using version: %s\n", version)
			return env.Use(version)
		},
	}

    var cmdBinaryPath = &cobra.Command{
		Use:   "path",
		Short: "Get the full path to current binary",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := env.CurrentBinaryPath()
			if err != nil {
				return err
			}
			if path == nil {
				return fmt.Errorf("no version selected, please run use first")
			}
			fmt.Print(*path)
			return nil
		},
	}

	var cmdUninstall = &cobra.Command{
		Use:   "uninstall [version]",
		Short: "Uninstall a specific version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			fmt.Printf("Uninstalling version: %s\n", version)
			return env.Uninstall(version)
		},
	}

	var cmdList = &cobra.Command{
		Use:   "list",
		Short: "List all installed versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			installed, err := env.ListInstalled()
			if err != nil {
				return err
			}
			for _, i := range installed {
				fmt.Println(i)
			}
			return nil
		},
	}

	rootCmd.AddCommand(cmdInstall, cmdUse, cmdUninstall, cmdList, cmdBinaryPath)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error executing command:", err)
	}
}
`

const DummyMainTemplate = `
package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var binaryName = "{{ .Name }}"

func main() {
	if runtime.GOOS == "windows" {
		binaryName = fmt.Sprintf("%s.exe", binaryName)
	}
	// Get the command-line arguments
	args := os.Args[1:]

	// Store the output in the dst variable
	dst, err := currentBinaryPath()
	if err != nil {
		os.Exit(1)
	}

	if dst == "" || strings.Contains(dst, "no version") {
		cmd := exec.Command(binaryName, "use", defaultVersion())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
		dst, err = currentBinaryPath()
		if err != nil {
			os.Exit(1)
		}
	}

	// Create a new command with dst and the command-line arguments
	cmd := exec.Command(dst, args...)

	// Set the command's Stdin and Stdout to the main process's Stdin and Stdout
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	// Run the command
	_ = cmd.Run()
}

func currentBinaryPath() (string, error) {
	// Create a new command
	cmd := exec.Command(binaryName, "path")

	// Run the command and capture the output
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing command:", err)
		return "", err
	}
	return string(out), nil
}

func defaultVersion() string {
	v := os.Getenv("{{ .UpperName }}_DEFAULT_VERSION")
	if v == "" {
		v = "latest"
	}
	return v
}
`

func main() {
	var downloadUrlTemplate, name, binaryName, gitRepo, gitSubFolder string

	var cmd = &cobra.Command{
		Use:   "genv",
		Short: "genv is a CLI tool for managing environments",
		RunE: func(cmd *cobra.Command, args []string) error {
			tplt, err := template.New("main").Parse(EnvMainTemplate)
			if err != nil {
				return err
			}
			pwd, err := os.Getwd()
			if err != nil {
				return err
			}
			binaryEnvDir := filepath.Join(pwd, "..", name)
			if _, err = os.Stat(binaryEnvDir); err == nil {
				_ = os.RemoveAll(binaryEnvDir)
			}
			err = os.Mkdir(binaryEnvDir, 0755)
			if err != nil {
				return err
			}
			dst := filepath.Clean(filepath.Join(binaryEnvDir, "main.go"))
			file, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defer func() {
				_ = file.Close()
			}()
			envData := struct {
				DownloadUrlTemplate string
				Name                string
				UpperName           string
				BinaryName          string
				GoBuildSubFolder    string
				GoBuildRepoUrl      string
			}{
				DownloadUrlTemplate: downloadUrlTemplate,
				Name:                name,
				UpperName:           strings.ToUpper(name),
				BinaryName:          binaryName,
				GoBuildRepoUrl:      gitRepo,
				GoBuildSubFolder:    gitSubFolder,
			}
			err = tplt.Execute(file, envData)
			if err != nil {
				return err
			}

			// Parse the DummyMainTemplate
			tplt, err = template.New("dummyMain").Parse(DummyMainTemplate)
			if err != nil {
				return err
			}

			// Check if the directory ../{binaryName} exists, if not create it
			dirPath := filepath.Join(pwd, "..", binaryName)
			if _, err = os.Stat(dirPath); err == nil {
				_ = os.RemoveAll(dirPath)
			}
			err = os.Mkdir(dirPath, 0755)
			if err != nil {
				return err
			}

			// Open the file ../{name}/main.go in write mode
			dst = filepath.Join(dirPath, "main.go")
			file, err = os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defer func() {
				_ = file.Close()
			}()

			// Execute the template with envData and write the output to the file
			err = tplt.Execute(file, envData)
			if err != nil {
				return err
			}

			err = executeCommand(binaryEnvDir, "go", "mod", "init", name)
			if err != nil {
				return fmt.Errorf("failed to run 'go mod init' in the env folder: %w", err)
			}
			err = executeCommand(binaryEnvDir, "go", "mod", "tidy")
			if err != nil {
				return fmt.Errorf("failed to run 'go mod tidy' in the env folder: %w", err)
			}
			err = executeCommand(binaryEnvDir, "go", "install")
			if err != nil {
				return fmt.Errorf("failed to run 'go install' in the env folder: %w", err)
			}

			// Run 'go install' in the ../{name} folder
			err = executeCommand(dirPath, "go", "install")
			if err != nil {
				return fmt.Errorf("failed to run 'go install' in the ../%s folder: %w", binaryName, err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&downloadUrlTemplate, "url", "u", "", "Download URL template")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Environment name")
	cmd.Flags().StringVarP(&binaryName, "binary", "b", "", "Binary name")
	cmd.Flags().StringVarP(&gitRepo, "git-repo", "", "", "Git Repository URL for Go build installer")
	cmd.Flags().StringVarP(&gitSubFolder, "git-sub-folder", "", "", "SubFolder For Go build installer")

	if err := cmd.Execute(); err != nil {
		fmt.Println("Error executing command:", err)
	}
}

func executeCommand(wd string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = wd
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
