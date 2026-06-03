package engine

// State 表示策略引擎的运行状态
type State string

const (
	StateCreated State = "created" // 初始状态
	StateLoaded  State = "loaded"  // 已加载 StrategyBundle
	StateRunning State = "running" // 正在运行
	StatePaused  State = "paused"  // 已暂停
	StateStopped State = "stopped" // 已停止
)

func (s State) IsActive() bool {
	return s == StateRunning || s == StatePaused
}

func (s State) CanStart() bool {
	return s == StateLoaded || s == StatePaused
}

func (s State) CanPause() bool {
	return s == StateRunning
}

func (s State) CanStop() bool {
	return s == StateRunning || s == StatePaused
}
