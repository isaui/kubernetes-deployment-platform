package utils

import "k8s.io/apimachinery/pkg/util/intstr"

// Helper function to convert int to *int32
func int32Ptr(i int32) *int32 {
    return &i
}

// Helper function to convert int to *int64
func int64Ptr(i int64) *int64 {
    return &i
}

// intToQuantity converts an int to an IntOrString for Kubernetes API
func IntToQuantity(val int) intstr.IntOrString {
	return intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: int32(val),
	}
}