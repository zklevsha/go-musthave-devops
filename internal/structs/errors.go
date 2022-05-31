package structs

import "errors"

var ErrMetricNotFound = errors.New("metric not found")
var ErrMetricBadType = errors.New("metric has unsupported type")
var ErrMetricNullAttr = errors.New("metric has not set attribute")
var ErrMetricBadAttrValue = errors.New("metric has bad attribute value")
