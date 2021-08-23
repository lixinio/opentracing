package gormopentracing

type options struct {
	logWithoutRoot   bool // logWithoutRoot 如果没有父Span，也记录
	logSqlParameters bool // 记录sql参数
	debug            bool
}

func defaultOption() *options {
	return &options{
		logWithoutRoot:   false,
		logSqlParameters: true,
		debug:            false,
	}
}

type applyOption func(o *options)

func WithLogWithoutRoot(logWithoutRoot bool) applyOption {
	return func(o *options) {
		o.logWithoutRoot = logWithoutRoot
	}
}

func WithSqlParameters(logSqlParameters bool) applyOption {
	return func(o *options) {
		o.logSqlParameters = logSqlParameters
	}
}

func WithDebug(debug bool) applyOption {
	return func(o *options) {
		o.debug = debug
	}
}
