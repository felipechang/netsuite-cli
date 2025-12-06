package cmd

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var (
	projectNameFlag string
	skipSetupFlag   bool
	outputDirFlag   string
)

var initTemplateFS embed.FS

// initCmd represents the create command
var initCmd = &cobra.Command{
	Use:   "create",
	Short: "Initialize a new NetSuite project",
	Long: `Initialize a new NetSuite project by creating the project structure,
generating configuration files, and setting up the account.`,
	Run: func(cmd *cobra.Command, args []string) {
		runInit()
	},
}

func init() {
	initCmd.Flags().StringVarP(&projectNameFlag, "name", "n", "", "Project name (required)")
	initCmd.Flags().BoolVarP(&skipSetupFlag, "skip-setup", "s", false, "Skip account setup step")
	initCmd.Flags().StringVarP(&outputDirFlag, "output", "o", ".", "Output directory for the project (default: current directory)")

	rootCmd.AddCommand(initCmd)
}

// getSuiteCloudCommand checks for the availability of the suitecloud CLI command.
func getSuiteCloudCommand() string {
	if _, err := exec.LookPath("suitecloud"); err == nil {
		return "suitecloud"
	}
	if _, err := exec.LookPath("suitecloud.cmd"); err == nil {
		return "suitecloud.cmd"
	}
	return ""
}

// runInit executes the project initialization process.
func runInit() {
	suiteCloudCmd := getSuiteCloudCommand()
	if suiteCloudCmd == "" {
		fmt.Println("Error: suitecloud CLI is not available in the command line.")
		fmt.Println("Please install it using: npm install -g @oracle/suitecloud-cli")
		os.Exit(1)
	}

	userConfig, err := LoadUserConfig()
	if err != nil {
		fmt.Printf("Warning: Failed to load user configuration: %v\n", err)
	}

	projectName := strings.TrimSpace(projectNameFlag)
	if projectName == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter project name: ")
		var err error
		projectName, err = reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading project name: %v\n", err)
			os.Exit(1)
		}
		projectName = strings.TrimSpace(projectName)
	}

	if projectName == "" {
		fmt.Println("Error: Project name cannot be empty.")
		fmt.Println("Use --name or -n flag to specify project name, or provide it interactively.")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	defaultCompanyName := ""
	if userConfig != nil && userConfig.CompanyName != "" {
		defaultCompanyName = userConfig.CompanyName
	}
	fmt.Print("Enter company name")
	if defaultCompanyName != "" {
		fmt.Printf(" (default: %s)", defaultCompanyName)
	}
	fmt.Print(": ")
	companyName, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading company name: %v\n", err)
		os.Exit(1)
	}
	companyName = strings.TrimSpace(companyName)
	if companyName == "" {
		if defaultCompanyName != "" {
			companyName = defaultCompanyName
		} else {
			fmt.Println("Error: Company name cannot be empty.")
			os.Exit(1)
		}
	}

	defaultUserName := ""
	if userConfig != nil && userConfig.UserName != "" {
		defaultUserName = userConfig.UserName
	} else {
		currentUser, err := user.Current()
		if err == nil && currentUser != nil {
			parts := strings.Split(currentUser.Username, "\\")
			if len(parts) > 1 {
				defaultUserName = parts[1]
			} else {
				defaultUserName = currentUser.Username
			}
		}
	}

	fmt.Print("Enter user name")
	if defaultUserName != "" {
		fmt.Printf(" (default: %s)", defaultUserName)
	}
	fmt.Print(": ")
	userName, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading user name: %v\n", err)
		os.Exit(1)
	}
	userName = strings.TrimSpace(userName)
	if userName == "" {
		if defaultUserName != "" {
			userName = defaultUserName
		} else {
			fmt.Println("Error: User name cannot be empty.")
			os.Exit(1)
		}
	}

	defaultUserEmail := ""
	if userConfig != nil && userConfig.UserEmail != "" {
		defaultUserEmail = userConfig.UserEmail
	}
	fmt.Print("Enter user email")
	if defaultUserEmail != "" {
		fmt.Printf(" (default: %s)", defaultUserEmail)
	}
	fmt.Print(": ")
	userEmail, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading user email: %v\n", err)
		os.Exit(1)
	}
	userEmail = strings.TrimSpace(userEmail)
	if userEmail == "" {
		if defaultUserEmail != "" {
			userEmail = defaultUserEmail
		} else {
			fmt.Println("Error: User email cannot be empty.")
			os.Exit(1)
		}
	}

	if strings.ContainsAny(projectName, `<>:"/\|?*`) {
		fmt.Println("Error: Project name contains invalid characters.")
		os.Exit(1)
	}

	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	outputDir := outputDirFlag
	if outputDir == "." {
		outputDir = wd
	} else if !filepath.IsAbs(outputDir) {
		outputDir = filepath.Join(wd, outputDir)
	}

	projectDir := filepath.Join(outputDir, projectName)

	if _, err := os.Stat(projectDir); err == nil {
		fmt.Printf("Error: Project directory '%s' already exists.\n", projectDir)
		os.Exit(1)
	}

	const projectType = "ACCOUNTCUSTOMIZATION"
	fmt.Printf("Creating project '%s' (type: %s)...\n", projectName, projectType)

	originalDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	if err := os.Chdir(outputDir); err != nil {
		fmt.Printf("Error changing to output directory: %v\n", err)
		os.Exit(1)
	}
	defer os.Chdir(originalDir)

	createCmd := exec.Command(suiteCloudCmd, "project:create", "--type", projectType, "--projectname", projectName)
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr
	createCmd.Stdin = os.Stdin

	if err := createCmd.Run(); err != nil {
		fmt.Printf("Error creating project: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		fmt.Printf("Error: Project directory '%s' was not created.\n", projectDir)
		os.Exit(1)
	}

	suiteScriptsDir := filepath.Join(projectDir, "src", "FileCabinet", "SuiteScripts")
	projectFolderPath := filepath.Join(suiteScriptsDir, projectName)
	if err := os.MkdirAll(projectFolderPath, 0755); err != nil {
		fmt.Printf("Warning: Failed to create project folder in SuiteScripts: %v\n", err)
	} else {
		fmt.Printf("Created project folder: %s\n", projectFolderPath)
	}

	objectsDir := filepath.Join(projectDir, "src", "Objects")
	objectsProjectFolderPath := filepath.Join(objectsDir, projectName)
	if err := os.MkdirAll(objectsProjectFolderPath, 0755); err != nil {
		fmt.Printf("Warning: Failed to create project folder in Objects: %v\n", err)
	} else {
		fmt.Printf("Created project folder: %s\n", objectsProjectFolderPath)
	}

	fmt.Println("Generating configuration files...")

	templateData := map[string]string{
		"ProjectName": projectName,
	}

	createFileFromTemplate(filepath.Join(projectDir, "package.json"), "templates/package.json.tmpl", templateData)
	createFileFromTemplate(filepath.Join(projectDir, "suitecloud.config.js"), "templates/suitecloud.config.js.tmpl", templateData)
	createFileFromTemplate(filepath.Join(projectDir, "tsconfig.json"), "templates/tsconfig.json.tmpl", templateData)
	createFileFromTemplate(filepath.Join(projectDir, ".gitignore"), "templates/.gitignore.tmpl", templateData)

	if !skipSetupFlag {
		fmt.Println("Setting up account...")
		setupCmd := exec.Command(suiteCloudCmd, "account:setup")
		setupCmd.Dir = projectDir
		setupCmd.Stdout = os.Stdout
		setupCmd.Stderr = os.Stderr
		setupCmd.Stdin = os.Stdin

		if err := setupCmd.Run(); err != nil {
			fmt.Printf("Warning: Account setup encountered an error: %v\n", err)
			fmt.Printf("You can run 'suitecloud account:setup' manually in the project directory.\n")
		} else {
			fmt.Println("Account setup completed successfully.")
		}
	} else {
		fmt.Println("Skipping account setup (--skip-setup flag used).")
	}

	config := &ProjectConfig{
		ProjectName: projectName,
		CompanyName: companyName,
		UserName:    userName,
		UserEmail:   userEmail,
	}
	if err := SaveConfig(projectDir, config); err != nil {
		fmt.Printf("Warning: Failed to save configuration: %v\n", err)
	} else {
		fmt.Println("Configuration saved to .netsuite-cli file")
	}

	userConfigToSave := &UserConfig{
		CompanyName: companyName,
		UserName:    userName,
		UserEmail:   userEmail,
	}
	if err := SaveUserConfig(userConfigToSave); err != nil {
		fmt.Printf("Warning: Failed to save user configuration: %v\n", err)
	} else {
		fmt.Println("User configuration saved to .netsuite-cli file")
	}

	fmt.Printf("\n✓ Initialization complete!\n")
	fmt.Printf("Project created at: %s\n", projectDir)
	fmt.Printf("To get started, run: cd %s\n", projectDir)
}

// createFile creates a file with the specified content.
func createFile(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Printf("Error creating %s: %v\n", path, err)
		os.Exit(1)
	}
}

// createFileFromTemplate creates a file by executing a template with the provided data.
func createFileFromTemplate(path, templatePath string, data map[string]string) {
	tmplContent, err := initTemplateFS.ReadFile(templatePath)
	if err != nil {
		fmt.Printf("Error reading template %s: %v\n", templatePath, err)
		os.Exit(1)
	}

	tmpl, err := template.New("config").Parse(string(tmplContent))
	if err != nil {
		fmt.Printf("Error parsing template %s: %v\n", templatePath, err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		fmt.Printf("Error executing template %s: %v\n", templatePath, err)
		os.Exit(1)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		fmt.Printf("Error creating %s: %v\n", path, err)
		os.Exit(1)
	}
}
