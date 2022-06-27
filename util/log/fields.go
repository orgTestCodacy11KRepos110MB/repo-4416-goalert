package log

import (
	"context"

	"github.com/google/uuid"
)

// Fields are used to add values in structured logging.
type Fields map[string]interface{}
type logContextField string

// SetRequestID will assign a unique ID to the context for tracing.
func SetRequestID(ctx context.Context) context.Context {
	return context.WithValue(ctx, logContextKeyRequestID, uuid.New().String())
}

// ContextFields will return the current set of fields associated with a context.
func ContextFields(ctx context.Context) Fields {
	f, _ := ctx.Value(logContextKeyFieldList).([]string)
	m := make(Fields, len(f))
	for _, f := range f {
		m[f] = ctx.Value(logContextField(f))
	}
	return m
}

// WithField will return a context with the specified field set to value.
func WithField(ctx context.Context, field string, value interface{}) context.Context {
	f, _ := ctx.Value(logContextKeyFieldList).([]string)

	var hasField bool
	// Search for the field in the existing slice.
	for _, fn := range f {
		if field == fn {
			hasField = true
			break
		}
	}

	if !hasField {
		// If the field is missing (i.e. it's new) we need to add it to the
		// list of fields.
		//
		// So we create a copy of the slice -- as we don't want to
		// modify the existing one, since it's used by the parent
		// context.
		fList := make([]string, len(f), len(f)+1)
		copy(fList, f)
		fList = append(fList, field)
		f = fList
	}

	ctx = context.WithValue(ctx, logContextKeyFieldList, f)
	ctx = context.WithValue(ctx, logContextField(field), value)

	return ctx
}

// WithFields will return a context with the provided fields set.
func WithFields(ctx context.Context, fields Fields) context.Context {
	if fields == nil {
		return ctx
	}
	for field, value := range fields {
		ctx = WithField(ctx, field, value)
	}
	return ctx
}

// RequestID will return the associated RequestID or empty string if missing.
func RequestID(ctx context.Context) string {
	v, _ := ctx.Value(logContextKeyRequestID).(string)
	return v
}
