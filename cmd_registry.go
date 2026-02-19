package main

import (
	"fmt"
	"os"
)

func cmdRegister(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: ko register #<tag>")
		return 1
	}

	tag := CleanTag(args[0])
	if tag == "" {
		fmt.Fprintln(os.Stderr, "usage: ko register #<tag>")
		return 1
	}

	regPath := RegistryPath()
	if regPath == "" {
		fmt.Fprintln(os.Stderr, "ko register: cannot determine config directory")
		return 1
	}

	reg, err := LoadRegistry(regPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko register: %v\n", err)
		return 1
	}

	// Determine project root from .tickets directory or cwd
	projectRoot, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko register: %v\n", err)
		return 1
	}

	reg.Projects[tag] = projectRoot

	// Detect and store ticket prefix
	ticketsDir := resolveTicketsDir(projectRoot)
	if prefix := detectPrefixFromDir(ticketsDir); prefix != "" {
		reg.Prefixes[tag] = prefix
	}

	if err := SaveRegistry(regPath, reg); err != nil {
		fmt.Fprintf(os.Stderr, "ko register: %v\n", err)
		return 1
	}

	fmt.Printf("registered %s -> %s\n", tag, projectRoot)
	return 0
}

func cmdDefault(args []string) int {
	regPath := RegistryPath()
	if regPath == "" {
		fmt.Fprintln(os.Stderr, "ko default: cannot determine config directory")
		return 1
	}

	reg, err := LoadRegistry(regPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko default: %v\n", err)
		return 1
	}

	// No args: show current default
	if len(args) == 0 {
		if reg.Default == "" {
			fmt.Println("no default project set")
		} else {
			fmt.Println(reg.Default)
		}
		return 0
	}

	tag := CleanTag(args[0])

	// Verify tag is registered
	if _, ok := reg.Projects[tag]; !ok {
		fmt.Fprintf(os.Stderr, "ko default: '%s' is not registered\n", tag)
		return 1
	}

	reg.Default = tag

	if err := SaveRegistry(regPath, reg); err != nil {
		fmt.Fprintf(os.Stderr, "ko default: %v\n", err)
		return 1
	}

	fmt.Printf("default: %s\n", tag)
	return 0
}

func cmdProjects(args []string) int {
	regPath := RegistryPath()
	if regPath == "" {
		fmt.Fprintln(os.Stderr, "ko projects: cannot determine config directory")
		return 1
	}

	if _, err := os.Stat(regPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "ko projects: no project registry found (%s)\n", regPath)
		return 1
	}

	reg, err := LoadRegistry(regPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko projects: %v\n", err)
		return 1
	}

	if len(reg.Projects) == 0 {
		fmt.Println("no projects registered")
		return 0
	}

	keys := make([]string, 0, len(reg.Projects))
	for k := range reg.Projects {
		keys = append(keys, k)
	}
	sortStrings(keys)

	for _, k := range keys {
		marker := "  "
		if k == reg.Default {
			marker = "* "
		}
		fmt.Printf("%s%s\t%s\n", marker, k, reg.Projects[k])
	}
	return 0
}


// findProjectRoot returns the absolute path to the project root.
// If a tickets directory is found, derives root via ProjectRoot.
// Otherwise, returns the current working directory.
func findProjectRoot() (string, error) {
	ticketsDir, err := FindTicketsDir()
	if err == nil {
		return ProjectRoot(ticketsDir), nil
	}
	return os.Getwd()
}
