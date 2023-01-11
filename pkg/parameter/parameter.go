package parameter

import "time"

// Base interface for types with parameters.
type WithParameters interface {
	GetParameter(key string) (any, bool)
	SetParameter(key string, value interface{})
}

// HasParameter check if parameter with given key exists.
func HasParameter(w WithParameters, key string) bool {
	_, ok := w.GetParameter(key)
	return ok
}

// GetString returns the value associated with the key as a string.
func GetString(w WithParameters, key string) (s string) {
	if val, ok := w.GetParameter(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// GetBool returns the value associated with the key as a boolean.
func GetBool(w WithParameters, key string) (b bool) {
	if val, ok := w.GetParameter(key); ok && val != nil {
		b, _ = val.(bool)
	}
	return
}

// GetInt returns the value associated with the key as an integer.
func GetInt(w WithParameters, key string) (i int) {
	if val, ok := w.GetParameter(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetInt64 returns the value associated with the key as an integer.
func GetInt64(w WithParameters, key string) (i64 int64) {
	if val, ok := w.GetParameter(key); ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

// GetUint returns the value associated with the key as an unsigned integer.
func GetUint(w WithParameters, key string) (ui uint) {
	if val, ok := w.GetParameter(key); ok && val != nil {
		ui, _ = val.(uint)
	}
	return
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func GetUint64(w WithParameters, key string) (ui64 uint64) {
	if val, ok := w.GetParameter(key); ok && val != nil {
		ui64, _ = val.(uint64)
	}
	return
}

// GetFloat64 returns the value associated with the key as a float64.
func GetFloat64(w WithParameters, key string) (f64 float64) {
	if val, ok := w.GetParameter(key); ok && val != nil {
		f64, _ = val.(float64)
	}
	return
}

// GetTime returns the value associated with the key as time.
func GetTime(w WithParameters, key string) (t time.Time) {
	if val, ok := w.GetParameter(key); ok && val != nil {
		t, _ = val.(time.Time)
	}
	return
}

// GetDuration returns the value associated with the key as a duration.
func GetDuration(w WithParameters, key string) (d time.Duration) {
	if val, ok := w.GetParameter(key); ok && val != nil {
		d, _ = val.(time.Duration)
	}
	return
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func GetStringSlice(w WithParameters, key string) (ss []string) {
	if val, ok := w.GetParameter(key); ok && val != nil {
		ss, _ = val.([]string)
	}
	return
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func GetStringMap(w WithParameters, key string) (sm map[string]any) {
	if val, ok := w.GetParameter(key); ok && val != nil {
		sm, _ = val.(map[string]any)
	}
	return
}

// GetStringMapString returns the value associated with the key as a map of strings.
func GetStringMapString(w WithParameters, key string) (sms map[string]string) {
	if val, ok := w.GetParameter(key); ok && val != nil {
		sms, _ = val.(map[string]string)
	}
	return
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func GetStringMapStringSlice(w WithParameters, key string) (smss map[string][]string) {
	if val, ok := w.GetParameter(key); ok && val != nil {
		smss, _ = val.(map[string][]string)
	}
	return
}
