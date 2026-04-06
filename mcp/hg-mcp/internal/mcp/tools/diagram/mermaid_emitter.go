package diagram

import (
	"fmt"
	"sort"
	"strings"
)

// EmitMermaidPackages generates a Mermaid graph TD diagram showing package dependencies.
func EmitMermaidPackages(packages []PackageInfo, edges []ImportEdge, modulePath string) string {
	var sb strings.Builder

	sb.WriteString("graph TD\n")

	// Group packages by their top-level directory.
	groups := make(map[string][]PackageInfo)
	var rootPkgs []PackageInfo

	for _, pkg := range packages {
		if pkg.ShortPath == "." || !strings.Contains(pkg.ShortPath, "/") {
			rootPkgs = append(rootPkgs, pkg)
		} else {
			parts := strings.SplitN(pkg.ShortPath, "/", 2)
			groups[parts[0]] = append(groups[parts[0]], pkg)
		}
	}

	// Determine the module short name.
	moduleShort := modulePath
	if idx := strings.LastIndex(modulePath, "/"); idx >= 0 {
		moduleShort = modulePath[idx+1:]
	}

	// Emit root-level packages (no subgraph needed if they stand alone).
	if len(rootPkgs) > 0 && len(groups) > 0 {
		sb.WriteString(fmt.Sprintf("    subgraph %s\n", mermaidLabel(moduleShort)))
		for _, pkg := range rootPkgs {
			id := mermaidID(pkg.ShortPath)
			sb.WriteString(fmt.Sprintf("        %s[\"%s\"]\n", id, pkg.Name))
		}
		sb.WriteString("    end\n")
	} else {
		for _, pkg := range rootPkgs {
			id := mermaidID(pkg.ShortPath)
			sb.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", id, pkg.Name))
		}
	}

	// Emit grouped packages as subgraphs (sorted for deterministic output).
	groupNames := make([]string, 0, len(groups))
	for g := range groups {
		groupNames = append(groupNames, g)
	}
	sort.Strings(groupNames)

	for _, groupName := range groupNames {
		pkgs := groups[groupName]
		sb.WriteString(fmt.Sprintf("    subgraph %s\n", mermaidLabel(groupName)))
		for _, pkg := range pkgs {
			id := mermaidID(pkg.ShortPath)
			sb.WriteString(fmt.Sprintf("        %s[\"%s\"]\n", id, pkg.Name))
		}
		sb.WriteString("    end\n")
	}

	// Emit edges.
	for _, edge := range edges {
		fromID := mermaidID(edge.From)
		toID := mermaidID(edge.To)
		sb.WriteString(fmt.Sprintf("    %s --> %s\n", fromID, toID))
	}

	return sb.String()
}

// EmitMermaidInterfaces generates a Mermaid class diagram showing interface implementations.
func EmitMermaidInterfaces(interfaces []InterfaceInfo, structs []StructInfo, impls []ImplEdge) string {
	var sb strings.Builder

	sb.WriteString("classDiagram\n")

	// Emit interfaces.
	for _, iface := range interfaces {
		sb.WriteString(fmt.Sprintf("    class %s {\n", iface.Name))
		sb.WriteString("        <<interface>>\n")
		for _, m := range iface.Methods {
			sb.WriteString(fmt.Sprintf("        +%s\n", mermaidMethodSig(m)))
		}
		sb.WriteString("    }\n")
	}

	// Emit structs.
	for _, s := range structs {
		sb.WriteString(fmt.Sprintf("    class %s {\n", s.Name))
		for _, m := range s.Methods {
			sb.WriteString(fmt.Sprintf("        +%s\n", mermaidMethodSig(m)))
		}
		sb.WriteString("    }\n")
	}

	// Emit implementation edges.
	for _, impl := range impls {
		sb.WriteString(fmt.Sprintf("    %s ..|> %s\n", impl.Struct, impl.Interface))
	}

	return sb.String()
}

// mermaidID converts a package path to a valid Mermaid node ID.
func mermaidID(path string) string {
	if path == "." {
		return "root"
	}
	r := strings.NewReplacer("/", "_", "-", "_", ".", "_")
	return r.Replace(path)
}

// mermaidLabel returns a label safe for Mermaid subgraph names.
func mermaidLabel(name string) string {
	// Mermaid subgraph names can contain spaces and most characters.
	return name
}

// mermaidMethodSig formats a method signature for Mermaid class diagrams.
// Mermaid uses: +methodName(params) returnType
func mermaidMethodSig(sig string) string {
	// The signature is already in "Name(params) return" format.
	// Mermaid uses ~ for generics, but we keep it simple.
	// Escape angle brackets that Mermaid might interpret as HTML.
	sig = strings.ReplaceAll(sig, "<", "~")
	sig = strings.ReplaceAll(sig, ">", "~")
	return sig
}
