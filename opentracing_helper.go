package gormopentracing

import (
	"context"

	"go.opencensus.io/trace"
	"gorm.io/gorm"
)

const (
	_prefix      = "gorm.opentracing"
	_errorTagKey = "error"
)

var (
	// span.Tag keys
	_tableTagKey = keyWithPrefix("table")
	// span.Log keys
	//_errorLogKey        = keyWithPrefix("error")
	//_resultLogKey       = keyWithPrefix("result")
	_sqlLogKey          = keyWithPrefix("sql")
	_rowsAffectedLogKey = keyWithPrefix("rowsAffected")
)

func keyWithPrefix(key string) string {
	return _prefix + "." + key
}

var (
	opentracingSpanKey = "opentracing:span"
	// json               = jsoniter.ConfigCompatibleWithStandardLibrary
)

func (p opentracingPlugin) injectBefore(db *gorm.DB, op operationName) {
	// make sure context could be used
	if db == nil {
		return
	}

	if db.Statement == nil || db.Statement.Context == nil {
		db.Logger.Error(
			context.TODO(),
			"could not inject sp from nil Statement.Context or nil Statement",
		)
		return
	}

	if trace.FromContext(db.Statement.Context) == nil {
		if !p.opt.logWithoutRoot {
			if p.opt.debug {
				panic("trace gorm without parent span")
			}
			return
		}
	}

	_, span := trace.StartSpan(db.Statement.Context, op.String())
	if span == nil {
		return
	}
	db.InstanceSet(opentracingSpanKey, span)
}

func (p opentracingPlugin) extractAfter(db *gorm.DB) {
	// make sure context could be used
	if db == nil {
		return
	}
	if db.Statement == nil || db.Statement.Context == nil {
		db.Logger.Error(
			context.TODO(),
			"could not extract sp from nil Statement.Context or nil Statement",
		)
		return
	}

	// extract sp from db context
	//sp := opentracing.SpanFromContext(db.Statement.Context)
	v, ok := db.InstanceGet(opentracingSpanKey)
	if !ok || v == nil {
		return
	}

	sp, ok := v.(*trace.Span)
	if !ok || sp == nil {
		return
	}
	defer sp.End()

	// tag and log fields we want.
	tag(sp, db, p.opt.logSqlParameters)
}

// tag called after operation
func tag(span *trace.Span, db *gorm.DB, logSqlVariables bool) {
	if err := db.Error; err != nil {
		span.SetStatus(trace.Status{
			Code:    trace.StatusCodeUnknown,
			Message: db.Error.Error(),
		})
	}

	span.AddAttributes(
		trace.StringAttribute(_tableTagKey, db.Statement.Table),
		trace.Int64Attribute(_rowsAffectedLogKey, db.Statement.RowsAffected),
		trace.StringAttribute(_sqlLogKey, appendSql(db, logSqlVariables)),
	)
}

func appendSql(db *gorm.DB, logSqlVariables bool) string {
	if logSqlVariables {
		return db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	} else {
		return db.Statement.SQL.String()
	}
}
