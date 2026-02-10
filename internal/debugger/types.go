package debugger

// EvalScope describes the goroutine and frame for evaluation.
type EvalScope struct {
	GoroutineID  int64 `json:"GoroutineID"`
	Frame        int   `json:"Frame"`
	DeferredCall int   `json:"DeferredCall"`
}

// LoadConfig describes how to load values from the target's memory.
type LoadConfig struct {
	FollowPointers     bool `json:"followPointers"`
	MaxVariableRecurse int  `json:"maxVariableRecurse"`
	MaxStringLen       int  `json:"maxStringLen"`
	MaxArrayValues     int  `json:"maxArrayValues"`
	MaxStructFields    int  `json:"maxStructFields"`
}

// DefaultLoadConfig returns a sensible default LoadConfig.
func DefaultLoadConfig() LoadConfig {
	return LoadConfig{
		FollowPointers:     true,
		MaxVariableRecurse: 1,
		MaxStringLen:       64,
		MaxArrayValues:     64,
		MaxStructFields:    -1,
	}
}

// Breakpoint represents a Delve breakpoint.
type Breakpoint struct {
	ID           int               `json:"id"`
	Name         string            `json:"name"`
	Addr         uint64            `json:"addr"`
	Addrs        []uint64          `json:"addrs"`
	File         string            `json:"file"`
	Line         int               `json:"line"`
	FunctionName string            `json:"functionName,omitempty"`
	Cond         string            `json:"Cond"`
	Tracepoint   bool              `json:"continue"`
	Goroutine    bool              `json:"goroutine"`
	Stacktrace   int               `json:"stacktrace"`
	Variables    []string          `json:"variables,omitempty"`
	LoadArgs     *LoadConfig       `json:"LoadArgs,omitempty"`
	LoadLocals   *LoadConfig       `json:"LoadLocals,omitempty"`
	HitCount     map[string]uint64 `json:"hitCount"`
	TotalHitCount uint64           `json:"totalHitCount"`
	Disabled     bool              `json:"disabled"`
}

// Function represents function information.
type Function struct {
	Name      string `json:"name"`
	Value     uint64 `json:"value"`
	Type      byte   `json:"type"`
	GoType    uint64 `json:"goType"`
	Optimized bool   `json:"optimized"`
}

// Location holds program location information.
type Location struct {
	PC       uint64    `json:"pc"`
	File     string    `json:"file"`
	Line     int       `json:"line"`
	Function *Function `json:"function,omitempty"`
}

// Variable describes a variable.
type Variable struct {
	Name     string     `json:"name"`
	Addr     uint64     `json:"addr"`
	OnlyAddr bool       `json:"onlyAddr"`
	Type     string     `json:"type"`
	RealType string     `json:"realType"`
	Kind     int        `json:"kind"`
	Value    string     `json:"value"`
	Len      int64      `json:"len"`
	Cap      int64      `json:"cap"`
	Children []Variable `json:"children"`
	Base     uint64     `json:"base"`
	Unreadable string   `json:"unreadable"`
}

// Thread represents a thread in the debugged process.
type Thread struct {
	ID             int             `json:"id"`
	PC             uint64          `json:"pc"`
	File           string          `json:"file"`
	Line           int             `json:"line"`
	Function       *Function       `json:"function,omitempty"`
	GoroutineID    int64           `json:"goroutineID"`
	Breakpoint     *Breakpoint     `json:"breakPoint,omitempty"`
	BreakpointInfo *BreakpointInfo `json:"breakPointInfo,omitempty"`
	ReturnValues   []Variable      `json:"returnValues,omitempty"`
}

// BreakpointInfo contains information about the current breakpoint.
type BreakpointInfo struct {
	Stacktrace []Stackframe `json:"stacktrace,omitempty"`
	Goroutine  *Goroutine   `json:"goroutine,omitempty"`
	Variables  []Variable   `json:"variables,omitempty"`
	Arguments  []Variable   `json:"arguments,omitempty"`
	Locals     []Variable   `json:"locals,omitempty"`
}

// Goroutine represents a goroutine.
type Goroutine struct {
	ID             int64             `json:"id"`
	CurrentLoc     Location          `json:"currentLoc"`
	UserCurrentLoc Location          `json:"userCurrentLoc"`
	GoStatementLoc Location          `json:"goStatementLoc"`
	StartLoc       Location          `json:"startLoc"`
	ThreadID       int               `json:"threadID"`
	Status         uint64            `json:"status"`
	WaitSince      int64             `json:"waitSince"`
	WaitReason     int64             `json:"waitReason"`
	Unreadable     string            `json:"unreadable"`
	Labels         map[string]string `json:"labels,omitempty"`
}

// Stackframe describes one frame in a stack trace.
type Stackframe struct {
	Location
	Locals    []Variable `json:"Locals,omitempty"`
	Arguments []Variable `json:"Arguments,omitempty"`
	Bottom    bool       `json:"Bottom,omitempty"`
	Err       string     `json:"Err,omitempty"`
}

// DebuggerState represents the current state of the debugger.
type DebuggerState struct {
	Pid               int          `json:"pid"`
	TargetCommandLine string       `json:"targetCommandLine"`
	Running           bool         `json:"Running"`
	CurrentThread     *Thread      `json:"currentThread,omitempty"`
	SelectedGoroutine *Goroutine   `json:"currentGoroutine,omitempty"`
	Threads           []*Thread    `json:"Threads,omitempty"`
	NextInProgress    bool         `json:"NextInProgress"`
	Exited            bool         `json:"exited"`
	ExitStatus        int          `json:"exitStatus"`
	When              string       `json:"When"`
}

// DebuggerCommand is a command to change the debugger's execution state.
type DebuggerCommand struct {
	Name                 string      `json:"name"`
	ThreadID             int         `json:"threadID,omitempty"`
	GoroutineID          int64       `json:"goroutineID,omitempty"`
	ReturnInfoLoadConfig *LoadConfig `json:"ReturnInfoLoadConfig,omitempty"`
	Expr                 string      `json:"expr,omitempty"`
}

// Command name constants matching Delve's API.
const (
	CmdContinue        = "continue"
	CmdStep            = "step"
	CmdNext            = "next"
	CmdStepOut         = "stepOut"
	CmdStepInstruction = "stepInstruction"
	CmdHalt            = "halt"
)

// --- RPC request/response types ---

type CreateBreakpointIn struct {
	Breakpoint Breakpoint `json:"Breakpoint"`
	LocExpr    string     `json:"LocExpr,omitempty"`
}

type CreateBreakpointOut struct {
	Breakpoint Breakpoint `json:"Breakpoint"`
}

type ClearBreakpointIn struct {
	Id   int    `json:"Id"`
	Name string `json:"Name"`
}

type ClearBreakpointOut struct {
	Breakpoint *Breakpoint `json:"Breakpoint"`
}

type ListBreakpointsIn struct {
	All bool `json:"All"`
}

type ListBreakpointsOut struct {
	Breakpoints []*Breakpoint `json:"Breakpoints"`
}

type StateIn struct {
	NonBlocking bool `json:"NonBlocking"`
}

type StateOut struct {
	State *DebuggerState `json:"State"`
}

type CommandOut struct {
	State DebuggerState `json:"State"`
}

type StacktraceIn struct {
	Id    int64       `json:"Id"`
	Depth int         `json:"Depth"`
	Full  bool        `json:"Full"`
	Cfg   *LoadConfig `json:"Cfg,omitempty"`
}

type StacktraceOut struct {
	Locations []Stackframe `json:"Locations"`
}

type ListLocalVarsIn struct {
	Scope EvalScope  `json:"Scope"`
	Cfg   LoadConfig `json:"Cfg"`
}

type ListLocalVarsOut struct {
	Variables []Variable `json:"Variables"`
}

type ListFunctionArgsIn struct {
	Scope EvalScope  `json:"Scope"`
	Cfg   LoadConfig `json:"Cfg"`
}

type ListFunctionArgsOut struct {
	Args []Variable `json:"Args"`
}

type EvalIn struct {
	Scope EvalScope   `json:"Scope"`
	Expr  string      `json:"Expr"`
	Cfg   *LoadConfig `json:"Cfg,omitempty"`
}

type EvalOut struct {
	Variable *Variable `json:"Variable"`
}

type ListGoroutinesIn struct {
	Start int `json:"Start"`
	Count int `json:"Count"`
}

type ListGoroutinesOut struct {
	Goroutines []*Goroutine `json:"Goroutines"`
	Nextg      int          `json:"Nextg"`
}
