package tui

import (
	"fmt"
	"strings"

	"github.com/cybertortuga/aitriage/internal/scanner/deps"
	lipgloss "github.com/charmbracelet/lipgloss"
)

// RenderDependencyTree generates a premium ASCII tree representation of the dependency graph.
// Features: depth-based coloring, children count badges, cycle detection, summary stats.
func RenderDependencyTree(graph deps.DependencyGraph) string {
	if len(graph.Nodes) == 0 {
		return ""
	}

	// Find roots (nodes that are not children of any other node)
	isChild := make(map[string]bool)
	totalEdges := 0
	for _, children := range graph.Edges {
		totalEdges += len(children)
		for _, child := range children {
			isChild[child] = true
		}
	}

	var roots []string
	nodeMap := make(map[string]deps.Dependency)
	for _, node := range graph.Nodes {
		id := node.ID()
		nodeMap[id] = node
		if !isChild[id] {
			roots = append(roots, id)
		}
	}

	// If no roots found but nodes exist (circular or complex), pick the first node
	if len(roots) == 0 && len(graph.Nodes) > 0 {
		roots = append(roots, graph.Nodes[0].ID())
	}

	var lines []string
	visited := make(map[string]int)
	maxDepth := 0

	connectorStyle := lipgloss.NewStyle().Foreground(colorOutline)
	versionStyle := lipgloss.NewStyle().Foreground(colorGray)
	rootBadgeStyle := lipgloss.NewStyle().
		Foreground(colorBG).
		Background(colorSecondary).
		Padding(0, 1).
		Bold(true)
	typeBadgeStyle := lipgloss.NewStyle().
		Foreground(colorBG).
		Background(colorOutline).
		Padding(0, 1)
	countStyle := lipgloss.NewStyle().Foreground(colorGray).Italic(true)
	cycleStyle := lipgloss.NewStyle().Foreground(colorSecondary).Italic(true)

	var renderNode func(string, string, bool, int)
	renderNode = func(name string, indent string, isLast bool, depth int) {
		if depth > 20 { // Safety depth limit
			return
		}
		if depth > maxDepth {
			maxDepth = depth
		}

		// Cycle detection
		if visited[name] > 0 {
			branch := "├── "
			if isLast {
				branch = "└── "
			}
			lines = append(lines, connectorStyle.Render(indent+branch)+cycleStyle.Render(name+" ↻ cycle"))
			return
		}
		visited[name]++
		defer func() { visited[name]-- }()

		branch := ""
		if depth > 0 {
			branch = "├── "
			if isLast {
				branch = "└── "
			}
		}

		// Node metadata
		node, ok := nodeMap[name]
		nodeText := name
		if ok {
			nodeText = node.Name
			if node.Version != "" {
				nodeText += versionStyle.Render(" " + node.Version)
			}
			if node.Type == "root" {
				nodeText += " " + rootBadgeStyle.Render("ROOT")
			} else if node.Type != "" && node.Type != "prod" {
				nodeText += " " + typeBadgeStyle.Render(strings.ToUpper(node.Type))
			}
		}

		// Children count badge
		childCount := len(graph.Edges[name])
		if childCount > 0 {
			nodeText += " " + countStyle.Render(fmt.Sprintf("(%d deps)", childCount))
		}

		// Depth-based color: root=bold+underline, shallow=primary, deep=gray
		var nodeStyle lipgloss.Style
		switch {
		case depth == 0:
			nodeStyle = lipgloss.NewStyle().Bold(true).Foreground(colorPrimary).Underline(true)
		case depth <= 2:
			nodeStyle = lipgloss.NewStyle().Foreground(colorPrimary)
		case depth <= 4:
			nodeStyle = lipgloss.NewStyle().Foreground(colorTextVariant)
		default:
			nodeStyle = lipgloss.NewStyle().Foreground(colorGray)
		}

		lines = append(lines, connectorStyle.Render(indent+branch)+nodeStyle.Render(nodeText))

		children := graph.Edges[name]
		newIndent := indent
		if isLast {
			newIndent += "    "
		} else {
			newIndent += "│   "
		}

		for i, child := range children {
			renderNode(child, newIndent, i == len(children)-1, depth+1)
		}
	}

	for i, root := range roots {
		renderNode(root, "", i == len(roots)-1, 0)
	}

	// Summary stats line
	summaryStyle := lipgloss.NewStyle().Foreground(colorOutline)
	statStyle := lipgloss.NewStyle().Foreground(colorGray)
	statValStyle := lipgloss.NewStyle().Foreground(colorPrimaryDim).Bold(true)

	lines = append(lines, "")
	summary := summaryStyle.Render("─── ") +
		statStyle.Render("NODES: ") + statValStyle.Render(fmt.Sprintf("%d", len(graph.Nodes))) +
		statStyle.Render(" │ EDGES: ") + statValStyle.Render(fmt.Sprintf("%d", totalEdges)) +
		statStyle.Render(" │ ROOTS: ") + statValStyle.Render(fmt.Sprintf("%d", len(roots))) +
		statStyle.Render(" │ MAX_DEPTH: ") + statValStyle.Render(fmt.Sprintf("%d", maxDepth)) +
		summaryStyle.Render(" ───")
	lines = append(lines, summary)

	return strings.Join(lines, "\n")
}

// BuildTopologyNodes flattens the dependency graph into a navigable list
// respecting the current expand/collapse state.
func BuildTopologyNodes(graph deps.DependencyGraph, expanded map[string]bool) []TopologyNode {
	if len(graph.Nodes) == 0 {
		return nil
	}

	// Find roots
	isChild := make(map[string]bool)
	for _, children := range graph.Edges {
		for _, child := range children {
			isChild[child] = true
		}
	}

	nodeMap := make(map[string]deps.Dependency)
	var roots []string
	for _, node := range graph.Nodes {
		id := node.ID()
		nodeMap[id] = node
		if !isChild[id] {
			roots = append(roots, id)
		}
	}
	if len(roots) == 0 && len(graph.Nodes) > 0 {
		roots = append(roots, graph.Nodes[0].ID())
	}

	var result []TopologyNode
	visited := make(map[string]bool)

	var walk func(id string, depth int)
	walk = func(id string, depth int) {
		if depth > 20 || visited[id] {
			return
		}
		visited[id] = true
		defer func() { visited[id] = false }()

		node, ok := nodeMap[id]
		children := graph.Edges[id]
		hasKids := len(children) > 0

		tn := TopologyNode{
			ID:       id,
			Depth:    depth,
			HasKids:  hasKids,
			IsRoot:   !isChild[id],
			Children: children,
		}
		if ok {
			tn.Name = node.Name
			tn.Version = node.Version
		} else {
			tn.Name = id
		}
		result = append(result, tn)

		if hasKids && expanded[id] {
			for _, child := range children {
				walk(child, depth+1)
			}
		}
	}

	for _, root := range roots {
		walk(root, 0)
	}

	return result
}

// rebuildTopologyNodes refreshes the flattened topology node list from the current graph
func (m *DashboardModel) rebuildTopologyNodes() {
	m.TopologyNodes = BuildTopologyNodesFiltered(m.DepGraph, m.TopologyExpanded, m.TopologyFilter)
	if m.TopologyCursor >= len(m.TopologyNodes) {
		m.TopologyCursor = len(m.TopologyNodes) - 1
	}
	if m.TopologyCursor < 0 {
		m.TopologyCursor = 0
	}
}

// BuildTopologyNodesFiltered builds the nodes and then filters them by query
func BuildTopologyNodesFiltered(graph deps.DependencyGraph, expanded map[string]bool, filter string) []TopologyNode {
	allNodes := BuildTopologyNodes(graph, expanded)
	if filter == "" {
		return allNodes
	}

	filterLower := strings.ToLower(filter)
	var filtered []TopologyNode
	for _, n := range allNodes {
		if strings.Contains(strings.ToLower(n.Name), filterLower) || strings.Contains(strings.ToLower(n.ID), filterLower) || strings.Contains(strings.ToLower(n.Version), filterLower) {
			filtered = append(filtered, n)
		}
	}
	return filtered
}
