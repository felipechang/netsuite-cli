# NetSuite CLI

A powerful Command Line Interface (CLI) tool for managing NetSuite projects. This tool streamlines the process of initializing NetSuite projects and generating SuiteScript templates, helping developers follow best practices and save time.

## Features

- **Project Initialization**: Quickly set up a new NetSuite project structure compatible with SuiteCloud CLI.
- **Script Generation**: Automatically generate TypeScript and XML templates for various SuiteScript types.
- **Smart Defaults**: Remembers your user and company details for faster project setup.
- **Interactive Prompts**: Guided prompts for easy configuration and file generation.

## Prerequisites

Before using `netsuite-cli`, ensure you have the following installed:

- [Go](https://go.dev/) (version 1.21 or higher)
- [Node.js](https://nodejs.org/) and npm
- Oracle NetSuite SuiteCloud CLI:
  ```bash
  npm install -g @oracle/suitecloud-cli
  ```

## Installation

To install the NetSuite CLI, run:

```bash
go install github.com/felipechang/netsuite-cli@latest
```
Alternatively, you can build it from source:

```bash
git clone https://github.com/felipechang/netsuite-cli.git
cd netsuite-cli
go install
```

## Usage

### Creating a New Project

Initialize a new NetSuite project with the `create` command:

```bash
netsuite-cli create --name my-project
```

You can also run it interactively:

```bash
netsuite-cli create
```

This command will:
1. Create a standard SDF project structure.
2. Generate necessary configuration files (`package.json`, `suitecloud.config.js`, `tsconfig.json`).
3. Optionally set up the SuiteCloud account.

**Flags:**
- `--name` / `-n`: Specify the project name.
- `--skip-setup` / `-s`: Skip the account setup step.
- `--output` / `-o`: Output directory (default: current directory).

### Adding Scripts

Once inside a project created by `netsuite-cli`, you can easily add new SuiteScripts using the `add` command.

```bash
netsuite-cli add [script-type] [script-name]
```

**Example:**

```bash
netsuite-cli add suitelet my_custom_suitelet
```

This will generate both the TypeScript source file and the corresponding XML definition file.

### Supported Script Types

The CLI supports generating templates for the following script types:

- **bundle**: Group related scripts together.
- **client**: Client-side scripts for UI customization.
- **formclient**: Scripts attached to forms for custom logic.
- **mapreduce**: Handle large amounts of data processing.
- **massupdate**: Programmatic custom updates to fields.
- **portlet**: Dashboard portlet scripts.
- **restlet**: RESTful API endpoints for external integration.
- **scheduled**: Scheduled background processing.
- **suitelet**: Custom pages and backend logic.
- **userevent**: Event-driven scripts for record actions.
- **workflowaction**: Custom logic for workflows.
- **common**: TypeScript definitions and shared code.

## Configuration

The CLI stores user preferences (Company Name, User Name, Email) in a `.netsuite-cli` file in your home directory. Project-specific configuration is stored in a `.netsuite-cli` file within the project root.

## Development

1. Clone the repository.
2. Install dependencies: `go mod download`.
3. Run the tool: `go run main.go`.

## License

[MIT License](LICENSE)
