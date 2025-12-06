package cmd

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/spf13/cobra"
)

var scriptTypeConfigs = []struct {
	name  string
	usage string
}{
	{"bundle", "Bundle scripts can be of type customization or configuration, allowing you to group related scripts together"},
	{"client", "Client scripts are executed by predefined event triggers in the client browser, enabling you to customize the user interface"},
	{"formclient", "Form Client scripts are attached to forms, allowing you to add custom logic and functionality to form submissions"},
	{"mapreduce", "Map/Reduce scripts are designed to handle large amounts of data, making them ideal for data processing and analysis tasks"},
	{"massupdate", "Mass update scripts allow you to programmatically perform custom updates to fields that are not available through general mass updates"},
	{"portlet", "Portlet scripts are run on the server and are rendered in the NetSuite dashboard, providing a way to customize the dashboard with custom functionality"},
	{"restlet", "RESTlet is a SuiteScript that you make available for other applications to call, enabling integration with external services and systems"},
	{"scheduled", "Scheduled scripts are executed (processed) with SuiteCloud Processors, allowing you to automate tasks and processes at specific times or intervals"},
	{"suitelet", "Suitelets are extensions of the SuiteScript API that allow you to build custom NetSuite pages and backend logic"},
	{"userevent", "User event scripts are executed when users perform actions on records, such as create, load, update, copy, delete, or submit, enabling you to automate tasks"},
	{"workflowaction", "Workflow action scripts are good for custom logic or managing sublist fields which are not currently available"},
	{"common", "Holds TypeScript definitions for your scripts, providing a way to define the structure and types of your code"},
}

// ScriptTemplates holds the content for TypeScript and XML templates.
type ScriptTemplates struct {
	TypeScript string
	XML        string
}

// getRecordType maps a script type to its corresponding NetSuite record type.
func getRecordType(scriptType string) string {
	recordTypeMap := map[string]string{
		"client":         "clientscript",
		"mapreduce":      "mapreducescript",
		"massupdate":     "massupdatescript",
		"portlet":        "portlet",
		"restlet":        "restlet",
		"scheduled":      "scheduledscript",
		"suitelet":       "suitelet",
		"userevent":      "usereventscript",
		"workflowaction": "workflowactionscript",
	}
	if recordType, ok := recordTypeMap[scriptType]; ok {
		return recordType
	}
	return ""
}

// toSnakeCase converts a string to snake_case.
func toSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	runes := []rune(s)
	prevLower := false

	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 && prevLower {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
			prevLower = false
		} else if unicode.IsLower(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
			prevLower = true
		} else {
			if i > 0 && result.Len() > 0 && result.String()[result.Len()-1] != '_' {
				result.WriteRune('_')
			}
			prevLower = false
		}
	}

	snake := result.String()
	re := regexp.MustCompile(`_+`)
	snake = re.ReplaceAllString(snake, "_")
	snake = strings.Trim(snake, "_")

	return strings.ToLower(snake)
}

var templateFS embed.FS

// GetTemplates retrieves the TypeScript and XML templates for a given script type.
func GetTemplates(scriptType string) ScriptTemplates {
	tsPath := fmt.Sprintf("templates/%s.ts.tmpl", scriptType)
	xmlPath := fmt.Sprintf("templates/%s.xml.tmpl", scriptType)

	tsContent, err := templateFS.ReadFile(tsPath)
	if err != nil {
		fmt.Printf("Warning: Could not read TypeScript template for %s: %v\n", scriptType, err)
		tsContent = []byte("")
	}

	xmlContent, err := templateFS.ReadFile(xmlPath)
	if err != nil {
		fmt.Printf("Warning: Could not read XML template for %s: %v\n", scriptType, err)
		xmlContent = []byte("")
	}

	return ScriptTemplates{
		TypeScript: string(tsContent),
		XML:        string(xmlContent),
	}
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new NetSuite script",
	Long:  `Generate a new NetSuite script from a template.`,
}

func init() {
	rootCmd.AddCommand(addCmd)

	for _, config := range scriptTypeConfigs {
		c := config
		subCmd := &cobra.Command{
			Use:   c.name + " [name]",
			Short: c.usage,
			Args:  cobra.MaximumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				runAdd(c.name, args)
			},
		}
		addCmd.AddCommand(subCmd)
	}
}

// TemplateData holds the data used to render script templates.
type TemplateData struct {
	Project      string
	ProjectName  string
	Description  string
	Date         string
	CompanyName  string
	UserName     string
	UserEmail    string
	ScriptName   string
	ScriptId     string
	ScriptPath   string
	DeploymentId string
	RecordType   string
}

// runAdd executes the logic for adding a new script.
func runAdd(scriptType string, args []string) {
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Not a project folder. Please run 'netsuite-cli create'")
		os.Exit(1)
	}

	scriptName := ""
	if len(args) > 0 {
		scriptName = args[0]
	}

	projectName := config.ProjectName
	defaultScriptName := toSnakeCase(projectName)

	if scriptName == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter script name")
		if defaultScriptName != "" {
			fmt.Printf(" (default: %s)", defaultScriptName)
		}
		fmt.Print(": ")
		var err error
		scriptName, err = reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading script name: %v\n", err)
			os.Exit(1)
		}
		scriptName = strings.TrimSpace(scriptName)
		if scriptName == "" {
			scriptName = defaultScriptName
		}
	}

	if scriptName == "" {
		fmt.Println("Error: Script name is required")
		os.Exit(1)
	}
	companyName := config.CompanyName
	userName := config.UserName
	userEmail := config.UserEmail

	reader := bufio.NewReader(os.Stdin)
	defaultDescription := scriptName + " description"
	fmt.Print("Enter script description")
	if defaultDescription != "" {
		fmt.Printf(" (default: %s)", defaultDescription)
	}
	fmt.Print(": ")
	description, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading description: %v\n", err)
		os.Exit(1)
	}
	description = strings.TrimSpace(description)
	if description == "" {
		description = defaultDescription
	}

	recordType := ""
	if scriptType == "userevent" || scriptType == "workflowaction" {
		fmt.Print("Enter record type (e.g., CUSTOMER, SALESORDER, INVOICE): ")
		recordTypeInput, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading record type: %v\n", err)
			os.Exit(1)
		}
		recordType = strings.TrimSpace(recordTypeInput)
		if recordType == "" {
			fmt.Println("Error: Record type is required for " + scriptType + " scripts")
			os.Exit(1)
		}
	}

	scriptId := strings.ReplaceAll(strings.ToLower(scriptName), " ", "_")
	deploymentId := "customdeploy_" + scriptId

	companyPrefix := GetCompanyPrefix(companyName)

	prefixedFileName := companyPrefix + "_" + scriptName
	tsFileNameWithType := prefixedFileName + "_" + scriptType

	data := TemplateData{
		Project:      projectName,
		ProjectName:  projectName,
		Description:  description,
		Date:         time.Now().Format("2006-01-02"),
		CompanyName:  companyName,
		UserName:     userName,
		UserEmail:    userEmail,
		ScriptName:   scriptName,
		ScriptId:     "customscript_" + scriptId,
		ScriptPath:   "SuiteScripts/" + projectName + "/" + tsFileNameWithType + ".ts",
		DeploymentId: deploymentId,
		RecordType:   recordType,
	}

	templates := GetTemplates(scriptType)

	suiteScriptsDir, err := findSuiteScriptsDir()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	selectedFolder, scriptPathPrefix := selectScriptFolder(suiteScriptsDir)

	osPath := strings.ReplaceAll(selectedFolder, "/", string(filepath.Separator))
	targetDir := filepath.Join(suiteScriptsDir, osPath)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", targetDir, err)
		os.Exit(1)
	}

	if selectedFolder != "" {
		data.ScriptPath = scriptPathPrefix + selectedFolder + "/" + tsFileNameWithType + ".ts"
	} else {
		data.ScriptPath = scriptPathPrefix + tsFileNameWithType + ".ts"
	}

	tsFileName := tsFileNameWithType + ".ts"
	tsPath := filepath.Join(targetDir, tsFileName)

	renderAndWrite(tsPath, templates.TypeScript, data)
	fmt.Printf("Created %s\n", tsPath)

	if templates.XML != "" && scriptType != "common" {
		objectsDir, err := findObjectsDir()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		recordType := getRecordType(scriptType)
		if recordType == "" {
			fmt.Printf("Warning: No record type found for script type '%s'. XML file not created.\n", scriptType)
		} else {
			xmlTargetDir := filepath.Join(objectsDir, projectName, recordType)
			if err := os.MkdirAll(xmlTargetDir, 0755); err != nil {
				fmt.Printf("Error creating XML directory %s: %v\n", xmlTargetDir, err)
				os.Exit(1)
			}

			xmlFileName := prefixedFileName + ".xml"
			xmlPath := filepath.Join(xmlTargetDir, xmlFileName)
			renderAndWrite(xmlPath, templates.XML, data)
			fmt.Printf("Created %s\n", xmlPath)
		}
	}
}

// renderAndWrite renders a template with data and writes it to the specified path.
func renderAndWrite(path string, tmplStr string, data TemplateData) {
	tmpl, err := template.New("script").Parse(tmplStr)
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		fmt.Printf("Error writing file %s: %v\n", path, err)
		os.Exit(1)
	}
}

// findSuiteScriptsDir locates the SuiteScripts directory in the project.
func findSuiteScriptsDir() (string, error) {
	possiblePaths := []string{
		"src/FileCabinet/SuiteScripts",
		"src/SuiteScripts",
		"SuiteScripts",
	}

	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path, nil
		}
	}

	var basePath string
	if _, err := os.Stat("src/FileCabinet"); err == nil {
		basePath = "src/FileCabinet/SuiteScripts"
	} else if _, err := os.Stat("src"); err == nil {
		basePath = "src/SuiteScripts"
	} else {
		basePath = "SuiteScripts"
	}

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create SuiteScripts directory: %v", err)
	}

	return basePath, nil
}

// findObjectsDir locates the Objects directory in the project.
func findObjectsDir() (string, error) {
	possiblePaths := []string{
		"src/Objects",
		"Objects",
	}

	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path, nil
		}
	}

	var basePath string
	if _, err := os.Stat("src"); err == nil {
		basePath = "src/Objects"
	} else {
		basePath = "Objects"
	}

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create Objects directory: %v", err)
	}

	return basePath, nil
}

// FolderOption represents a folder selection option in the interactive menu.
type FolderOption struct {
	Path     string
	Display  string
	FullPath string
}

// selectScriptFolder allows the user to interactively select a folder for the script.
func selectScriptFolder(suiteScriptsDir string) (string, string) {
	folders := findAllFolders(suiteScriptsDir, "")

	scriptPathPrefix := "SuiteScripts/"

	if len(folders) == 0 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\nNo folders found under SuiteScripts. Place script in SuiteScripts root? (y/n): ")
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			os.Exit(1)
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled. Script not created.")
			os.Exit(0)
		}
		return "", scriptPathPrefix
	}

	return displayScrollableMenu(folders, scriptPathPrefix)
}

// findAllFolders recursively finds all directories starting from baseDir.
func findAllFolders(baseDir string, relativePath string) []FolderOption {
	var folders []FolderOption

	fullPath := baseDir
	if relativePath != "" {
		fullPath = filepath.Join(baseDir, relativePath)
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return folders
	}

	for _, entry := range entries {
		if entry.IsDir() {
			entryRelativePath := entry.Name()
			if relativePath != "" {
				entryRelativePath = filepath.Join(relativePath, entry.Name())
			}

			netSuitePath := strings.ReplaceAll(entryRelativePath, string(filepath.Separator), "/")

			depth := strings.Count(entryRelativePath, string(filepath.Separator))
			indent := strings.Repeat("  ", depth)
			displayName := indent + entry.Name()
			if depth > 0 {
				displayName += " (" + netSuitePath + ")"
			}

			folders = append(folders, FolderOption{
				Path:     netSuitePath,
				Display:  displayName,
				FullPath: filepath.Join(fullPath, entry.Name()),
			})

			subfolders := findAllFolders(baseDir, entryRelativePath)
			folders = append(folders, subfolders...)
		}
	}

	return folders
}

// displayScrollableMenu shows a scrollable menu of folder options to the user.
func displayScrollableMenu(folders []FolderOption, scriptPathPrefix string) (string, string) {
	const pageSize = 20
	reader := bufio.NewReader(os.Stdin)
	currentPage := 0
	totalPages := (len(folders) + pageSize - 1) / pageSize

	for {
		fmt.Print("\n")
		fmt.Println("Available folders under SuiteScripts:")
		fmt.Println("  0. SuiteScripts (root)")
		fmt.Println(strings.Repeat("-", 60))

		start := currentPage * pageSize
		end := start + pageSize
		if end > len(folders) {
			end = len(folders)
		}

		for i := start; i < end; i++ {
			fmt.Printf("  %d. %s\n", i+1, folders[i].Display)
		}

		if totalPages > 1 {
			fmt.Printf("\nPage %d of %d", currentPage+1, totalPages)
			if currentPage > 0 {
				fmt.Print(" (p: previous page")
			}
			if currentPage < totalPages-1 {
				if currentPage > 0 {
					fmt.Print(", n: next page")
				} else {
					fmt.Print(" (n: next page")
				}
			}
			if currentPage > 0 || currentPage < totalPages-1 {
				fmt.Print(")")
			}
		}

		fmt.Print("\nSelect folder (0 for root, number to select")
		if totalPages > 1 {
			fmt.Print(", 'n' for next page, 'p' for previous page")
		}
		fmt.Print("): ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading selection: %v\n", err)
			os.Exit(1)
		}

		input = strings.TrimSpace(strings.ToLower(input))

		if totalPages > 1 {
			if input == "n" && currentPage < totalPages-1 {
				currentPage++
				continue
			}
			if input == "p" && currentPage > 0 {
				currentPage--
				continue
			}
		}

		selection, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Invalid selection. Please enter a number")
			if totalPages > 1 {
				fmt.Print(" or 'n'/'p' for navigation")
			}
			fmt.Println()
			time.Sleep(1 * time.Second)
			continue
		}

		if selection == 0 {
			return "", scriptPathPrefix
		}

		if selection < 1 || selection > len(folders) {
			fmt.Printf("Invalid selection. Please choose between 0 and %d\n", len(folders))
			time.Sleep(1 * time.Second)
			continue
		}

		return folders[selection-1].Path, scriptPathPrefix
	}
}
