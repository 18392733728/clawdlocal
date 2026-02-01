package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileReadTool implements a tool for reading files
type FileReadTool struct{}

func (t *FileReadTool) Name() string {
	return "file_read"
}

func (t *FileReadTool) Description() string {
	return "Read the contents of a file from the filesystem"
}

func (t *FileReadTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"filepath": "string - Path to the file to read",
	}
}

func (t *FileReadTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	filepathParam, ok := params["filepath"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'filepath' parameter")
	}

	// Security: Ensure we're only reading from allowed directories
	allowedBase := "./workspace"
	absPath, err := filepath.Abs(filepathParam)
	if err != nil {
		return nil, err
	}
	
	allowedAbs, err := filepath.Abs(allowedBase)
	if err != nil {
		return nil, err
	}
	
	if !isSubPath(allowedAbs, absPath) {
		return nil, fmt.Errorf("access denied: file path must be within %s", allowedBase)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	return string(data), nil
}

// FileWriteTool implements a tool for writing files
type FileWriteTool struct{}

func (t *FileWriteTool) Name() string {
	return "file_write"
}

func (t *FileWriteTool) Description() string {
	return "Write content to a file in the filesystem"
}

func (t *FileWriteTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"filepath": "string - Path to the file to write",
		"content":  "string - Content to write to the file",
	}
}

func (t *FileWriteTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	filepathParam, ok := params["filepath"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'filepath' parameter")
	}
	
	content, ok := params["content"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'content' parameter")
	}

	// Security: Ensure we're only writing to allowed directories
	allowedBase := "./workspace"
	absPath, err := filepath.Abs(filepathParam)
	if err != nil {
		return nil, err
	}
	
	allowedAbs, err := filepath.Abs(allowedBase)
	if err != nil {
		return nil, err
	}
	
	if !isSubPath(allowedAbs, absPath) {
		return nil, fmt.Errorf("access denied: file path must be within %s", allowedBase)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	err = os.WriteFile(absPath, []byte(content), 0644)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("File written successfully to %s", filepathParam),
	}, nil
}

// FileListTool implements a tool for listing directory contents
type FileListTool struct{}

func (t *FileListTool) Name() string {
	return "file_list"
}

func (t *FileListTool) Description() string {
	return "List contents of a directory"
}

func (t *FileListTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"dirpath": "string - Path to the directory to list (optional, defaults to './workspace')",
	}
}

func (t *FileListTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dirpath := "./workspace"
	if d, ok := params["dirpath"].(string); ok && d != "" {
		dirpath = d
	}

	// Security: Ensure we're only listing from allowed directories
	allowedBase := "./workspace"
	absPath, err := filepath.Abs(dirpath)
	if err != nil {
		return nil, err
	}
	
	allowedAbs, err := filepath.Abs(allowedBase)
	if err != nil {
		return nil, err
	}
	
	if !isSubPath(allowedAbs, absPath) {
		return nil, fmt.Errorf("access denied: directory path must be within %s", allowedBase)
	}

	files, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}
		
		result = append(result, map[string]interface{}{
			"name":    file.Name(),
			"is_dir":  file.IsDir(),
			"size":    info.Size(),
			"modtime": info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	return result, nil
}

// Helper function to check if a path is within a base path
func isSubPath(base, path string) bool {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return false
	}
	
	// Check if the relative path starts with ".." which would mean it's outside the base
	return !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}