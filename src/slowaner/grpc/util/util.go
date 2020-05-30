package util

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/golang/protobuf/ptypes/wrappers"
	"time"
)

func DecodeTimeP(timestamp *timestamp.Timestamp) *time.Time {
	if timestamp == nil {
		return nil
	}
	tm := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	return &tm
}
func EncodeTimeP(time *time.Time) *timestamp.Timestamp {
	if time == nil {
		return nil
	}
	return &timestamp.Timestamp{Seconds: time.Unix(), Nanos: int32(time.Nanosecond())}
}

func DecodeTime(timestamp *timestamp.Timestamp) time.Time {
	if timestamp == nil {
		return time.Time{}
	}
	tm := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	return tm
}
func EncodeTime(time time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{Seconds: time.Unix(), Nanos: int32(time.Nanosecond())}
}

func UnwrapString(stringValue *wrappers.StringValue) *string {
	if stringValue == nil {
		return nil
	}
	return &stringValue.Value
}
func WrapString(str *string) *wrappers.StringValue {
	if str == nil {
		return nil
	}
	return &wrappers.StringValue{Value: *str}
}

func UnwrapUInt32(i *wrappers.UInt32Value) *uint32 {
	if i == nil {
		return nil
	}
	return &i.Value
}
func WrapUInt32(i *uint32) *wrappers.UInt32Value {
	if i == nil {
		return nil
	}
	return &wrappers.UInt32Value{Value: *i}
}

func UnwrapInt32(i *wrappers.Int32Value) *int32 {
	if i == nil {
		return nil
	}
	return &i.Value
}
func WrapInt32(i *int32) *wrappers.Int32Value {
	if i == nil {
		return nil
	}
	return &wrappers.Int32Value{Value: *i}
}

func UnwrapFloat(i *wrappers.FloatValue) *float32 {
	if i == nil {
		return nil
	}
	return &i.Value
}
func WrapFloat(i *float32) *wrappers.FloatValue {
	if i == nil {
		return nil
	}
	return &wrappers.FloatValue{Value: *i}
}

func UnwrapDouble(i *wrappers.DoubleValue) *float64 {
	if i == nil {
		return nil
	}
	return &i.Value
}
func WrapDouble(i *float64) *wrappers.DoubleValue {
	if i == nil {
		return nil
	}
	return &wrappers.DoubleValue{Value: *i}
}

func UnwrapBool(i *wrappers.BoolValue) *bool {
	if i == nil {
		return nil
	}
	return &i.Value
}
func WrapBool(i *bool) *wrappers.BoolValue {
	if i == nil {
		return nil
	}
	return &wrappers.BoolValue{Value: *i}
}
