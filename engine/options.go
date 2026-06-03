package engine

// Options 配置 Engine 的运行参数
type Options struct {
	// PriceCheckInterval 股价检查 cron 表达式，为空时使用默认值 "0 */5 * * * *"
	PriceCheckInterval string

	// StrictMode 严格模式：编译/加载时遇到非致命错误也返回错误
	StrictMode bool

	// AutoStart 加载 Bundle 后是否自动启动
	AutoStart bool

	// EnableRiskGuard 是否启用风控拦截（预留开关）
	EnableRiskGuard bool

	// EnableAdapterManager 是否启用适配器管理（预留开关）
	EnableAdapterManager bool
}

// DefaultOptions 返回默认配置
func DefaultOptions() Options {
	return Options{
		PriceCheckInterval:   "0 */5 * * * *",
		StrictMode:           false,
		AutoStart:            false,
		EnableRiskGuard:      true,
		EnableAdapterManager: true,
	}
}

// Option 函数式配置选项
type Option func(*Options)

func WithPriceCheckInterval(spec string) Option {
	return func(o *Options) {
		o.PriceCheckInterval = spec
	}
}

func WithStrictMode(v bool) Option {
	return func(o *Options) {
		o.StrictMode = v
	}
}

func WithAutoStart(v bool) Option {
	return func(o *Options) {
		o.AutoStart = v
	}
}
