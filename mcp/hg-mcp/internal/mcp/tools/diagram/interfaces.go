package diagram

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Regex patterns for scanning Go source files.
var (
	// Matches: type FooInterface interface {
	reInterface = regexp.MustCompile(`^\s*type\s+(\w+)\s+interface\s*\{`)

	// Matches: type FooStruct struct {
	reStruct = regexp.MustCompile(`^\s*type\s+(\w+)\s+struct\s*\{`)

	// Matches: func (r *Receiver) MethodName(args...) returnType
	// Captures: receiver-name, method-name, params, return
	reMethod = regexp.MustCompile(`^\s*func\s+\(\s*\w+\s+\*?(\w+)\s*\)\s+(\w+)\s*\(([^)]*)\)\s*(.*)`)

	// Matches a method signature inside an interface block: MethodName(args) return
	reInterfaceMethod = regexp.MustCompile(`^\s+(\w+)\s*\(([^)]*)\)\s*(.*)`)
)

// handleDiagramInterfaces implements the aftrs_diagram_interfaces tool.
func handleDiagramInterfaces(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath, errResult := tools.RequireStringParam(req, "repo_path")
	if errResult != nil {
		return errResult, nil
	}

	pkgFilter := tools.OptionalStringParam(req, "package_filter", "./...")
	format := tools.OptionalStringParam(req, "format", "mermaid")
	ifaceFilter := tools.GetStringParam(req, "interface_filter")

	// Step 1: Get package list via go list.
	packages, _, err := listPackages(ctx, repoPath, pkgFilter)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("go list failed: %w", err)), nil
	}

	if len(packages) == 0 {
		return tools.TextResult("No packages found matching the filter."), nil
	}

	// Step 2: Scan Go files for types and methods.
	var allInterfaces []InterfaceInfo
	var allStructs []StructInfo

	// Collect methods per struct across all packages.
	structMethods := make(map[string][]string) // "pkg.StructName" -> methods

	for _, pkg := range packages {
		for _, goFile := range pkg.GoFiles {
			filePath := filepath.Join(pkg.Dir, goFile)
			ifaces, structs, methods := scanGoFile(filePath, pkg.Name)
			allInterfaces = append(allInterfaces, ifaces...)
			allStructs = append(allStructs, structs...)

			for structName, methodList := range methods {
				key := pkg.Name + "." + structName
				structMethods[key] = append(structMethods[key], methodList...)
			}
		}
	}

	// Attach collected methods to struct info.
	for i, s := range allStructs {
		key := s.Package + "." + s.Name
		if methods, ok := structMethods[key]; ok {
			allStructs[i].Methods = methods
		}
	}

	// Apply interface_filter if provided.
	if ifaceFilter != "" {
		var filtered []InterfaceInfo
		for _, iface := range allInterfaces {
			if iface.Name == ifaceFilter {
				filtered = append(filtered, iface)
			}
		}
		allInterfaces = filtered
	}

	// Step 3: Determine implementations (heuristic: method name+signature matching).
	impls := findImplementations(allInterfaces, allStructs)

	// If interface_filter is set, only include structs that implement a matched interface.
	if ifaceFilter != "" {
		implStructs := make(map[string]bool)
		for _, impl := range impls {
			implStructs[impl.Package+"."+impl.Struct] = true
		}
		var filteredStructs []StructInfo
		for _, s := range allStructs {
			if implStructs[s.Package+"."+s.Name] {
				filteredStructs = append(filteredStructs, s)
			}
		}
		allStructs = filteredStructs
	}

	// Step 4: Emit diagram.
	var source string
	switch format {
	case "d2":
		source = EmitD2Interfaces(allInterfaces, allStructs, impls)
	default:
		source = EmitMermaidInterfaces(allInterfaces, allStructs, impls)
	}

	output := DiagramOutput{
		Source:    source,
		Format:   format,
		TypeCount: len(allInterfaces) + len(allStructs),
		EdgeCount: len(impls),
	}

	return tools.JSONResult(output), nil
}

// scanGoFile parses a single Go file for interface definitions, struct definitions,
// and method declarations. It returns them grouped.
func scanGoFile(filePath, pkgName string) ([]InterfaceInfo, []StructInfo, map[string][]string) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, nil, nil
	}
	defer f.Close()

	var interfaces []InterfaceInfo
	var structs []StructInfo
	methods := make(map[string][]string) // structName -> method signatures

	scanner := bufio.NewScanner(f)

	// State machine for parsing interface method blocks.
	var inInterface string
	var ifaceMethods []string
	braceDepth := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Track interface body.
		if inInterface != "" {
			// Count braces to detect end of interface block.
			braceDepth += strings.Count(line, "{") - strings.Count(line, "}")
			if braceDepth <= 0 {
				// End of interface block.
				interfaces = append(interfaces, InterfaceInfo{
					Package: pkgName,
					Name:    inInterface,
					Methods: ifaceMethods,
				})
				inInterface = ""
				ifaceMethods = nil
				continue
			}

			// Try to match a method signature inside the interface.
			if m := reInterfaceMethod.FindStringSubmatch(line); m != nil {
				methodName := m[1]
				params := normalizeParams(m[2])
				ret := strings.TrimSpace(m[3])

				// Skip embedded interfaces (single word, starts with uppercase, no parens).
				sig := methodName + "(" + params + ")"
				if ret != "" {
					sig += " " + ret
				}
				ifaceMethods = append(ifaceMethods, sig)
			}
			continue
		}

		// Check for interface definition.
		if m := reInterface.FindStringSubmatch(line); m != nil {
			inInterface = m[1]
			ifaceMethods = nil
			braceDepth = strings.Count(line, "{") - strings.Count(line, "}")
			if braceDepth <= 0 {
				// Single-line empty interface.
				interfaces = append(interfaces, InterfaceInfo{
					Package: pkgName,
					Name:    inInterface,
					Methods: nil,
				})
				inInterface = ""
			}
			continue
		}

		// Check for struct definition.
		if m := reStruct.FindStringSubmatch(line); m != nil {
			structs = append(structs, StructInfo{
				Package: pkgName,
				Name:    m[1],
			})
			continue
		}

		// Check for method definition (receiver methods).
		if m := reMethod.FindStringSubmatch(line); m != nil {
			receiver := m[1]
			methodName := m[2]
			params := normalizeParams(m[3])
			ret := strings.TrimSpace(m[4])

			// Strip opening brace from return value.
			ret = strings.TrimSuffix(strings.TrimSpace(ret), "{")
			ret = strings.TrimSpace(ret)

			sig := methodName + "(" + params + ")"
			if ret != "" {
				sig += " " + ret
			}
			methods[receiver] = append(methods[receiver], sig)
		}
	}

	// Handle case where file ends while still inside an interface block.
	if inInterface != "" {
		interfaces = append(interfaces, InterfaceInfo{
			Package: pkgName,
			Name:    inInterface,
			Methods: ifaceMethods,
		})
	}

	return interfaces, structs, methods
}

// normalizeParams strips parameter names from a parameter list, leaving only types.
// "ctx context.Context, name string" -> "context.Context, string"
func normalizeParams(params string) string {
	params = strings.TrimSpace(params)
	if params == "" {
		return ""
	}

	parts := strings.Split(params, ",")
	var types []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		fields := strings.Fields(part)
		if len(fields) == 0 {
			continue
		}
		// The last field is the type.
		types = append(types, fields[len(fields)-1])
	}
	return strings.Join(types, ", ")
}

// findImplementations checks which structs satisfy which interfaces
// by comparing method signatures (name + normalized signature).
func findImplementations(interfaces []InterfaceInfo, structs []StructInfo) []ImplEdge {
	var impls []ImplEdge

	for _, iface := range interfaces {
		if len(iface.Methods) == 0 {
			continue
		}

		// Build a set of required method names (just the name portion for matching).
		ifaceMethodNames := make(map[string]string) // methodName -> full signature
		for _, m := range iface.Methods {
			name := methodNameFromSig(m)
			ifaceMethodNames[name] = m
		}

		for _, s := range structs {
			if s.Name == iface.Name {
				continue // Skip if same name (unlikely but possible).
			}

			structMethodNames := make(map[string]bool)
			for _, m := range s.Methods {
				name := methodNameFromSig(m)
				structMethodNames[name] = true
			}

			// Check if struct has all interface methods (by name).
			allMatch := true
			for name := range ifaceMethodNames {
				if !structMethodNames[name] {
					allMatch = false
					break
				}
			}

			if allMatch {
				impls = append(impls, ImplEdge{
					Struct:    s.Name,
					Interface: iface.Name,
					Package:   s.Package,
				})
			}
		}
	}

	return impls
}

// methodNameFromSig extracts the method name from a signature like "Name() string".
func methodNameFromSig(sig string) string {
	idx := strings.Index(sig, "(")
	if idx < 0 {
		return sig
	}
	return sig[:idx]
}
