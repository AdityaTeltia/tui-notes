package models

import (
	"regexp"
	"strings"
)

var todoRegex = regexp.MustCompile(`(?m)^\s*[-*]\s+\[([ xX])\]\s+(.+)$`)

type TodoItem struct {
	Text      string
	Completed bool
	LineNum   int
}

// ExtractTodos extracts all todo items from note content
func (n *Note) ExtractTodos() []TodoItem {
	todos := []TodoItem{}
	lines := strings.Split(n.Content, "\n")
	
	for i, line := range lines {
		matches := todoRegex.FindStringSubmatch(line)
		if len(matches) >= 3 {
			completed := strings.ToLower(matches[1]) == "x"
			todos = append(todos, TodoItem{
				Text:      strings.TrimSpace(matches[2]),
				Completed: completed,
				LineNum:   i,
			})
		}
	}
	
	return todos
}

// ToggleTodo toggles a todo item at the given line number
func (n *Note) ToggleTodo(lineNum int) bool {
	lines := strings.Split(n.Content, "\n")
	if lineNum < 0 || lineNum >= len(lines) {
		return false
	}
	
	line := lines[lineNum]
	matches := todoRegex.FindStringSubmatch(line)
	if len(matches) >= 2 {
		// Toggle the checkbox
		if matches[1] == " " {
			// Mark as completed
			lines[lineNum] = strings.Replace(line, "[ ]", "[x]", 1)
		} else {
			// Mark as incomplete
			lines[lineNum] = strings.Replace(line, "[x]", "[ ]", 1)
			lines[lineNum] = strings.Replace(lines[lineNum], "[X]", "[ ]", 1)
		}
		n.Content = strings.Join(lines, "\n")
		return true
	}
	
	return false
}

// CountTodos returns the count of completed and total todos
func (n *Note) CountTodos() (completed, total int) {
	todos := n.ExtractTodos()
	total = len(todos)
	for _, todo := range todos {
		if todo.Completed {
			completed++
		}
	}
	return completed, total
}

