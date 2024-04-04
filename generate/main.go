package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	installer, _ := pkg.NewDownloadInstaller("{{  .DownloadUrlTemplate }}", ctx)
	env := pkg.NewEnv("{{ .HomeDir }}", "{{ .Name }}", "{{ .BinaryName }}", installer)

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
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Listing all installed versions")
			// Implement the logic to list all installed versions here
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

	// Create a new command
	cmd := exec.Command(binaryName, "path")

	// Run the command and capture the output
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing command:", err)
		return
	}

	// Store the output in the dst variable
	dst := strings.TrimSpace(string(out))

	// Create a new command with dst and the command-line arguments
	cmd = exec.Command(dst, args...)

	// Set the command's Stdin and Stdout to the main process's Stdin and Stdout
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	// Run the command
	_ = cmd.Run()
}
`

func main() {
	var downloadUrlTemplate, homeDir, name, binaryName string

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
			if _, err = os.Stat(binaryEnvDir); os.IsNotExist(err) {
				err = os.Mkdir(binaryEnvDir, 0755)
				if err != nil {
					return err
				}
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
				HomeDir             string
				Name                string
				BinaryName          string
			}{
				DownloadUrlTemplate: downloadUrlTemplate,
				HomeDir:             homeDir,
				Name:                name,
				BinaryName:          binaryName,
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
			if _, err = os.Stat(dirPath); os.IsNotExist(err) {
				err = os.Mkdir(dirPath, 0755)
				if err != nil {
					return err
				}
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

			//err = replaceFirstLine(fmt.Sprintf("module %s", name))
			//if err != nil {
			//	return err
			//}
			// Run 'go install' in the root folder
			goModInitCmd := exec.Command("go", "mod", "init", binaryName)
			goModInitCmd.Dir = binaryEnvDir
			err = goModInitCmd.Run()
			if err != nil {
				return fmt.Errorf("failed to run 'go mod init' in the env folder: %w", err)
			}
			goModTidyCmd := exec.Command("go", "mod", "tidy")
			goModTidyCmd.Dir = binaryEnvDir
			err = goModTidyCmd.Run()
			if err != nil {
				return fmt.Errorf("failed to run 'go mod tidy' in the env folder: %w", err)
			}
			installCmd := exec.Command("go", "install")
			installCmd.Dir = binaryEnvDir
			err = installCmd.Run()
			if err != nil {
				return fmt.Errorf("failed to run 'go install' in the env folder: %w", err)
			}

			// Run 'go install' in the ../{name} folder
			installCmd = exec.Command("go", "install")
			installCmd.Dir = dirPath
			err = installCmd.Run()
			if err != nil {
				return fmt.Errorf("failed to run 'go install' in the ../%s folder: %w", name, err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&downloadUrlTemplate, "url", "u", "", "Download URL template")
	cmd.Flags().StringVarP(&homeDir, "dir", "d", "", "Home directory")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Environment name")
	cmd.Flags().StringVarP(&binaryName, "binary", "b", "", "Binary name")

	if err := cmd.Execute(); err != nil {
		fmt.Println("Error executing command:", err)
	}
}

func replaceFirstLine(newLine string) error {
	// Open the file in read mode
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	gomodPath := filepath.Join(dir, "..", "go.mod")
	file, err := os.Open(gomodPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	// Read all lines into a slice
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Replace the first line
	lines[0] = newLine

	// Open the file in write mode
	file, err = os.Create(gomodPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	// Write the updated lines back to the file
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, _ = fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}
