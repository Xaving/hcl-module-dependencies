package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// FileMapping is a schema for Terraform file.
type Modules struct {
	ModuleMappings []ModuleMapping `hcl:"module,block"`
	Remain         hcl.Body        `hcl:",remain"`
}

// ModuleMapping is a schema for "module" block in Terraform file.
type ModuleMapping struct {
	Name   string   `hcl:"name,label"`
	Source string   `hcl:"source"`
	Remain hcl.Body `hcl:",remain"`
}

//Structure for variable buffer file
type Variables struct {
	VariableMappings []VariableMapping `hcl:"variable,block"`
	Remain           hcl.Body          `hcl:",remain"`
}

// VariableMapping is a schema for "variable" block
type VariableMapping struct {
	Name   string   `hcl:"name,label"`
	Remain hcl.Body `hcl:",remain"`
}

// TODO: deal with source of the form : ./module/newmod

// LoadTerraformFiles loads terraform files from a given dir.
func main() {

	///////////////////////////////
	// first part: parse tf file //
	//////////////////////////////
	p := hclparse.NewParser()
	// for test purpose file and default value are fixed
	default_owner := "aamkye"
	default_repo := "bday"
	default_branch := "master"
	default_path := "terraform"
	f, diags := p.ParseHCLFile("test6.tf")
	if diags.HasErrors() {
		log.Fatalf("diags: %v", diags)
	}
	ms := &Modules{}
	diags = gohcl.DecodeBody(f.Body, nil, ms)
	if diags.HasErrors() {
		fmt.Println("Error")
		return
	}

	/////////////////////////////////////
	//Second Part: fetch file in github//
	/////////////////////////////////////

	for _, mod := range ms.ModuleMappings {

		// Parse module source Address
		fmt.Printf("#############: %s\n", mod.Source)
		var url string
		// Source contains prefix ./ or ../
		fmt.Println("######", mod.Source, "\n")
		if string(mod.Source[0]) == "." {
			// todo make the path from .. and .
			// for test purpose take the suffix
			extra_path := strings.TrimPrefix(mod.Source, "./")
			// Use default default values
			url = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s/%s/",
				default_owner, default_repo, default_branch, default_path, extra_path)
		} else {
			// Parse adress
			url = parseModuleAddress(mod.Source)
		}
		fmt.Println("######", url, "\n")
		// Fetch the variable.tf
		vars, err := fetchFile(url, "variables.tf")
		if err != nil {
			log.Fatal("Cannot request file: %s", err)
		}

		// Fetch the main.tf
		main, err := fetchFile(url, "main.tf")
		if err != nil {
			log.Fatal("Cannot request file: %s", err)
		}

		/////////////////////////////////
		//Third Part: parse the buffer //
		////////////////////////////////
		variables, diagsv := parseVariable(vars)
		if diagsv.HasErrors() {
			log.Fatalf("Error while parsing the buffer for variables")
		}

		modules, diagsm := parseModule(main)
		no_main_file := false
		if diagsm.HasErrors() {
			no_main_file = true
			//log.Fatalf("Error while parsing the buffer for modules")
		}

		/////////////////////////////////
		// Fourth Part: Format a result//
		/////////////////////////////////

		m := ms.ModuleMappings[0]
		fmt.Printf("\n\n\n##########################\n")
		fmt.Printf("Module\n")
		fmt.Printf("\tName: %s\n", m.Name)
		fmt.Printf("\tSource: %s\n", m.Source)
		fmt.Printf("\tVariable: \n")
		for _, v := range variables {
			fmt.Printf("\t\tVariable: %s\n", v)
		}
		if !no_main_file {
			fmt.Println("SUB-module", modules, "\n")
		} else {
			fmt.Println("No main.tf found\n")
		}
	}
}

func parseModuleAddress(source string) string {
	//##############
	// manage source of different form: ssh:
	//git@github.com/, ../module, ./module, github.com/,git::ssh://git@github.com/
	//##############
	// trim prefix
	t := source
	for _, p := range []string{"git@github.com/", "github.com/", "git::ssh://git@github.com/"} {
		t = strings.TrimPrefix(t, p)
	}
	// split string with ?ref=
	s := strings.Split(t, "?ref=")
	// extract path if any
	var path string
	n := s[0]
	if strings.Contains(n, "//") {
		c := strings.Split(n, "//")
		path = c[1]
		n = c[0]
	}
	// extract owner, repo
	var owner string
	var repo string
	p := strings.Split(n, "/")
	owner = p[0]
	repo = strings.TrimSuffix(p[1], ".git")
	// check if a specific branch
	var branch string
	// If there is a branch
	if len(s) != 1 {
		branch = s[1]
	} else {
		// check if repo is main or master
		branch = "main"
		address := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s/variable.tf",
			owner, repo, branch, path)
		// Check the address
		_, err := http.Get(address)
		if err != nil {
			// change main to master
			branch = "master"
		}
	}
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s/", owner, repo, branch, path)
}

func parseVariable(input []byte) ([]string, hcl.Diagnostics) {
	// Create the parser from the buffer
	p := hclparse.NewParser()
	pi, diags := p.ParseHCL(input, "from_variable_file")
	if diags.HasErrors() {
		return nil, diags
	}
	// Get the variables from the parser
	vs := &Variables{}
	diags = gohcl.DecodeBody(pi.Body, nil, vs)
	if diags.HasErrors() {
		fmt.Println("Error")
		return nil, diags
	}
	// Return the list of variable name
	var variables []string
	for _, v := range vs.VariableMappings {
		variables = append(variables, v.Name)
	}
	return variables, diags
}

func parseModule(input []byte) ([]string, hcl.Diagnostics) {
	// Create the parser from the buffer
	fmt.Printf("#############: %s\n", "pass1")
	p := hclparse.NewParser()
	pi, diags := p.ParseHCL(input, "from_variable_file")
	if diags.HasErrors() {
		fmt.Println(diags)
		return nil, diags
	}
	// Get the Modules from the parser
	fmt.Printf("#############: %s\n", "pass2")
	ms := &Modules{}
	diags = gohcl.DecodeBody(pi.Body, nil, ms)
	if diags.HasErrors() {
		fmt.Println("Error")
		return nil, diags
	}
	// Return the list of module name
	fmt.Printf("#############: %s\n", "pass3")
	var mods []string
	for _, m := range ms.ModuleMappings {
		mods = append(mods, m.Name)
	}
	return mods, diags
}

func fetchFile(address, filename string) ([]byte, error) {
	// Compose the url
	url := address + filename
	// Get the file
	r, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	// return the result in a buffer
	var data []byte
	s := bufio.NewScanner(r.Body)
	for s.Scan() {
		data = append(data, s.Bytes()...)
		data = append(data, '\n')
	}
	return data, nil
}
