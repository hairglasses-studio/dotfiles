package diagram

import (
	"fmt"
	"sort"
	"strings"
)

// EmitD2Packages generates a D2 diagram showing package dependencies.
// Packages are grouped into containers by their first path segment after the module root.
func EmitD2Packages(packages []PackageInfo, edges []ImportEdge, modulePath string) string {
	var sb strings.Builder

	sb.WriteString("direction: down\n\n")

	// Determine the module short name for the root container.
	moduleShort := modulePath
	if idx := strings.LastIndex(modulePath, "/"); idx >= 0 {
		moduleShort = modulePath[idx+1:]
	}

	// Group packages by their top-level directory (first path segment of ShortPath).
	groups := make(map[string][]PackageInfo) // group name -> packages
	var rootPkgs []PackageInfo

	for _, pkg := range packages {
		if pkg.ShortPath == "." || !strings.Contains(pkg.ShortPath, "/") {
			rootPkgs = append(rootPkgs, pkg)
		} else {
			parts := strings.SplitN(pkg.ShortPath, "/", 2)
			groups[parts[0]] = append(groups[parts[0]], pkg)
		}
	}

	// Emit root container.
	sb.WriteString(fmt.Sprintf("%s: \"%s\" {\n", d2ID(moduleShort), moduleShort))

	// Emit root-level packages.
	for _, pkg := range rootPkgs {
		id := d2ID(pkg.ShortPath)
		label := pkg.Name
		if pkg.ShortPath == "." {
			label = moduleShort
			id = d2ID(moduleShort + "_root")
		}
		sb.WriteString(fmt.Sprintf("    %s: \"%s\" {\n        shape: package\n    }\n", id, label))
	}

	// Emit grouped packages (sorted for deterministic output).
	groupNames := make([]string, 0, len(groups))
	for g := range groups {
		groupNames = append(groupNames, g)
	}
	sort.Strings(groupNames)

	for _, groupName := range groupNames {
		pkgs := groups[groupName]
		sb.WriteString(fmt.Sprintf("    %s: \"%s\" {\n", d2ID(groupName), groupName))
		for _, pkg := range pkgs {
			// Use the leaf name as the ID within the group.
			leafPath := pkg.ShortPath
			if idx := strings.LastIndex(leafPath, "/"); idx >= 0 {
				leafPath = leafPath[idx+1:]
			}
			sb.WriteString(fmt.Sprintf("        %s: \"%s\" {\n            shape: package\n        }\n", d2ID(leafPath), pkg.Name))
		}
		sb.WriteString("    }\n")
	}

	sb.WriteString("}\n\n")

	// Emit edges.
	for _, edge := range edges {
		fromD2 := d2PackagePath(edge.From, moduleShort)
		toD2 := d2PackagePath(edge.To, moduleShort)
		sb.WriteString(fmt.Sprintf("%s -> %s: imports\n", fromD2, toD2))
	}

	return sb.String()
}

// EmitD2Interfaces generates a D2 diagram showing interface implementations.
func EmitD2Interfaces(interfaces []InterfaceInfo, structs []StructInfo, impls []ImplEdge) string {
	var sb strings.Builder

	sb.WriteString("direction: down\n\n")

	// Emit interfaces as class shapes.
	for _, iface := range interfaces {
		id := d2ID(iface.Name)
		sb.WriteString(fmt.Sprintf("%s: \"%s\" {\n", id, iface.Name))
		sb.WriteString("    shape: class\n")
		sb.WriteString("    style.font-color: \"#57c7ff\"\n")
		for _, m := range iface.Methods {
			sb.WriteString(fmt.Sprintf("    %s\n", d2EscapeMethod(m)))
		}
		sb.WriteString("}\n\n")
	}

	// Emit structs as class shapes.
	for _, s := range structs {
		id := d2ID(s.Name)
		sb.WriteString(fmt.Sprintf("%s: \"%s\" {\n", id, s.Name))
		sb.WriteString("    shape: class\n")
		for _, m := range s.Methods {
			sb.WriteString(fmt.Sprintf("    %s\n", d2EscapeMethod(m)))
		}
		sb.WriteString("}\n\n")
	}

	// Emit implementation edges (dashed).
	for _, impl := range impls {
		sb.WriteString(fmt.Sprintf("%s -> %s: implements {\n    style.stroke-dash: 3\n}\n",
			d2ID(impl.Struct), d2ID(impl.Interface)))
	}

	return sb.String()
}

// d2ID converts a name to a valid D2 identifier by replacing non-alphanumeric chars.
func d2ID(name string) string {
	if name == "." {
		return "root"
	}
	r := strings.NewReplacer("/", "_", "-", "_", ".", "_")
	return r.Replace(name)
}

// d2PackagePath converts a package short path to a D2 container path.
// e.g., "internal/mcp/tools" -> "moduleShort.internal.mcp.tools"
func d2PackagePath(shortPath, moduleShort string) string {
	if shortPath == "." {
		return d2ID(moduleShort) + "." + d2ID(moduleShort+"_root")
	}

	parts := strings.Split(shortPath, "/")
	if len(parts) == 1 {
		return d2ID(moduleShort) + "." + d2ID(parts[0])
	}

	// group.leaf
	group := parts[0]
	leaf := parts[len(parts)-1]
	return d2ID(moduleShort) + "." + d2ID(group) + "." + d2ID(leaf)
}

// d2EscapeMethod wraps a method signature for use inside a D2 class.
// D2 class fields use the format: fieldName: type
// For method signatures we emit them as-is; D2 treats them as labels.
func d2EscapeMethod(sig string) string {
	// Escape characters that D2 might interpret specially.
	sig = strings.ReplaceAll(sig, "\"", "'")
	// Wrap []Type in quotes to avoid D2 parsing issues.
	if strings.Contains(sig, "[") || strings.Contains(sig, "]") {
		return fmt.Sprintf("\"%s\"", sig)
	}
	return sig
}
