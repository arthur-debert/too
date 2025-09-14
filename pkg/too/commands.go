// Package too provides a api facade for the actual api.
// This is intended to facilitate integration with the cli and other tools.
package too

import (
	"fmt"
	
	"github.com/arthur-debert/too/pkg/too/models"
	// Keep these imports for now for backward compatibility
	cmdDataPath "github.com/arthur-debert/too/pkg/too/commands/datapath"
	cmdFormats "github.com/arthur-debert/too/pkg/too/commands/formats"
	cmdInit "github.com/arthur-debert/too/pkg/too/commands/init"
)

// Command option types
type (
	// Keep these for special commands that aren't unified yet
	InitOptions         = cmdInit.Options
	ShowDataPathOptions = cmdDataPath.Options
	ListFormatsOptions  = cmdFormats.Options
	
	// Simplified option types for unified commands
	AddOptions struct {
		CollectionPath string
		ParentPath     string
	}
	
	ModifyOptions struct {
		CollectionPath string
	}
	
	CompleteOptions struct {
		CollectionPath string
	}
	
	ReopenOptions struct {
		CollectionPath string
	}
	
	CleanOptions struct {
		CollectionPath string
	}
	
	SearchOptions struct {
		CollectionPath string
		CaseSensitive  bool
	}
	
	ListOptions struct {
		CollectionPath string
		ShowDone       bool
		ShowAll        bool
	}
	
	MoveOptions struct {
		CollectionPath string
	}
	
	SwapOptions = MoveOptions
)

// Command result types - using unified ChangeResult for most commands
type (
	// Keep these for special commands
	InitResult         = cmdInit.Result
	ShowDataPathResult = cmdDataPath.Result
	ListFormatsResult  = cmdFormats.Result
	FormatInfo         = cmdFormats.Format
	
	// Legacy result types for backward compatibility
	AddResult struct {
		Todo         *models.Todo
		PositionPath string
		AllTodos     []*models.Todo
		TotalCount   int
		DoneCount    int
	}
	
	ModifyResult struct {
		Todo       *models.Todo
		OldText    string
		NewText    string
		AllTodos   []*models.Todo
		TotalCount int
		DoneCount  int
	}
	
	CompleteResult struct {
		Todo       *models.Todo
		OldStatus  string
		NewStatus  string
		AllTodos   []*models.Todo
		TotalCount int
		DoneCount  int
	}
	
	ReopenResult = CompleteResult
	
	CleanResult struct {
		RemovedTodos []*models.Todo
		RemovedCount int
		ActiveCount  int                // Number of active todos remaining
		ActiveTodos  []*models.Todo  // Active todos for display
		AllTodos     []*models.Todo  // Alias for ActiveTodos
		TotalCount   int
		DoneCount    int
	}
	
	SearchResult struct {
		Query        string
		MatchedTodos []*models.Todo
		TotalCount   int
	}
	
	ListResult struct {
		Todos      []*models.Todo
		TotalCount int
		DoneCount  int
	}
	
	MoveResult struct {
		Todo         *models.Todo
		OldParentUID string
		NewParentUID string
		OldPath      string  // For compatibility
		NewPath      string  // For compatibility
		AllTodos     []*models.Todo
		TotalCount   int
		DoneCount    int
	}
	
	SwapResult = MoveResult
)

// Init initializes a new todo collection
func Init(opts InitOptions) (*InitResult, error) {
	return cmdInit.Execute(opts)
}

// Add adds a new todo to the collection
func Add(text string, opts AddOptions) (*AddResult, error) {
	changeOpts := map[string]interface{}{
		"collectionPath": opts.CollectionPath,
		"parent":         opts.ParentPath,
	}
	
	result, err := ExecuteUnifiedCommand("add", []string{text}, changeOpts)
	if err != nil {
		return nil, err
	}
	
	// Convert to legacy AddResult
	if len(result.AffectedTodos) == 0 {
		return nil, fmt.Errorf("no todo created")
	}
	
	todo := result.AffectedTodos[0]
	return &AddResult{
		Todo:         todo,
		PositionPath: todo.PositionPath,
		AllTodos:     result.AllTodos,
		TotalCount:   result.TotalCount,
		DoneCount:    result.DoneCount,
	}, nil
}

// AddAsChange executes Add and returns a unified ChangeResult
func AddAsChange(text string, opts AddOptions) (*ChangeResult, error) {
	changeOpts := map[string]interface{}{
		"collectionPath": opts.CollectionPath,
		"parent":         opts.ParentPath,
	}
	
	return ExecuteUnifiedCommand("add", []string{text}, changeOpts)
}

// Modify modifies the text of an existing todo by position
func Modify(position string, newText string, opts ModifyOptions) (*ModifyResult, error) {
	changeOpts := map[string]interface{}{
		"collectionPath": opts.CollectionPath,
	}
	
	result, err := ExecuteUnifiedCommand("edit", []string{position, newText}, changeOpts)
	if err != nil {
		return nil, err
	}
	
	// Convert to legacy ModifyResult
	if len(result.AffectedTodos) == 0 {
		return nil, fmt.Errorf("no todo modified")
	}
	
	todo := result.AffectedTodos[0]
	return &ModifyResult{
		Todo:       todo,
		OldText:    "", // Not tracked in unified version
		NewText:    todo.Text,
		AllTodos:   result.AllTodos,
		TotalCount: result.TotalCount,
		DoneCount:  result.DoneCount,
	}, nil
}

// Complete marks a todo as complete by position path
func Complete(positionPath string, opts CompleteOptions) (*CompleteResult, error) {
	changeOpts := map[string]interface{}{
		"collectionPath": opts.CollectionPath,
	}
	
	result, err := ExecuteUnifiedCommand("complete", []string{positionPath}, changeOpts)
	if err != nil {
		return nil, err
	}
	
	// Convert to legacy CompleteResult
	if len(result.AffectedTodos) == 0 {
		return nil, fmt.Errorf("no todo completed")
	}
	
	todo := result.AffectedTodos[0]
	return &CompleteResult{
		Todo:       todo,
		OldStatus:  string(models.StatusPending),
		NewStatus:  string(models.StatusDone),
		AllTodos:   result.AllTodos,
		TotalCount: result.TotalCount,
		DoneCount:  result.DoneCount,
	}, nil
}

// Reopen marks a todo as pending by position path
func Reopen(positionPath string, opts ReopenOptions) (*ReopenResult, error) {
	changeOpts := map[string]interface{}{
		"collectionPath": opts.CollectionPath,
	}
	
	result, err := ExecuteUnifiedCommand("reopen", []string{positionPath}, changeOpts)
	if err != nil {
		return nil, err
	}
	
	// Convert to legacy ReopenResult
	if len(result.AffectedTodos) == 0 {
		return nil, fmt.Errorf("no todo reopened")
	}
	
	todo := result.AffectedTodos[0]
	return &ReopenResult{
		Todo:       todo,
		OldStatus:  string(models.StatusDone),
		NewStatus:  string(models.StatusPending),
		AllTodos:   result.AllTodos,
		TotalCount: result.TotalCount,
		DoneCount:  result.DoneCount,
	}, nil
}

// Clean removes finished todos from the collection
func Clean(opts CleanOptions) (*CleanResult, error) {
	changeOpts := map[string]interface{}{
		"collectionPath": opts.CollectionPath,
	}
	
	result, err := ExecuteUnifiedCommand("clean", []string{}, changeOpts)
	if err != nil {
		return nil, err
	}
	
	return &CleanResult{
		RemovedTodos: result.AffectedTodos,
		RemovedCount: len(result.AffectedTodos),
		ActiveCount:  len(result.AllTodos),  // Active todos remaining
		ActiveTodos:  result.AllTodos,       // Active todos for display
		AllTodos:     result.AllTodos,       // Alias
		TotalCount:   result.TotalCount,
		DoneCount:    result.DoneCount,
	}, nil
}

// Search searches for todos containing the query string
func Search(query string, opts SearchOptions) (*SearchResult, error) {
	changeOpts := map[string]interface{}{
		"collectionPath": opts.CollectionPath,
		"query":          query,
	}
	
	result, err := ExecuteUnifiedCommand("search", []string{query}, changeOpts)
	if err != nil {
		return nil, err
	}
	
	return &SearchResult{
		Query:        query,
		MatchedTodos: result.AllTodos, // Search filters AllTodos
		TotalCount:   result.TotalCount,
	}, nil
}

// List returns todos from the collection with optional filtering
func List(opts ListOptions) (*ListResult, error) {
	changeOpts := map[string]interface{}{
		"collectionPath": opts.CollectionPath,
		"done":           opts.ShowDone,
		"all":            opts.ShowAll,
	}
	
	result, err := ExecuteUnifiedCommand("list", []string{}, changeOpts)
	if err != nil {
		return nil, err
	}
	
	return &ListResult{
		Todos:      result.AllTodos,
		TotalCount: result.TotalCount,
		DoneCount:  result.DoneCount,
	}, nil
}

// Move moves a todo from one parent to another
func Move(sourcePath string, destParentPath string, opts MoveOptions) (*MoveResult, error) {
	changeOpts := map[string]interface{}{
		"collectionPath": opts.CollectionPath,
	}
	
	result, err := ExecuteUnifiedCommand("move", []string{sourcePath, destParentPath}, changeOpts)
	if err != nil {
		return nil, err
	}
	
	// Convert to legacy MoveResult
	if len(result.AffectedTodos) == 0 {
		return nil, fmt.Errorf("no todo moved")
	}
	
	todo := result.AffectedTodos[0]
	return &MoveResult{
		Todo:         todo,
		OldParentUID: "", // Not tracked in unified version
		NewParentUID: todo.ParentID,
		OldPath:      sourcePath,
		NewPath:      todo.PositionPath,
		AllTodos:     result.AllTodos,
		TotalCount:   result.TotalCount,
		DoneCount:    result.DoneCount,
	}, nil
}

// Swap moves a todo from one parent to another (alias for Move)
func Swap(sourcePath string, destParentPath string, opts SwapOptions) (*SwapResult, error) {
	return Move(sourcePath, destParentPath, opts)
}

// ShowDataPath shows the path to the data file
func ShowDataPath(opts ShowDataPathOptions) (*ShowDataPathResult, error) {
	return cmdDataPath.Execute(opts)
}

// ListFormats returns the list of available output formats
func ListFormats(opts ListFormatsOptions) (*ListFormatsResult, error) {
	return cmdFormats.Execute(opts)
}
