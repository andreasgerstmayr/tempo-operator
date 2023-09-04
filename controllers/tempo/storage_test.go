package controllers

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
)

func TestGetS3ParamsInsecure(t *testing.T) {
	storageSecret := &corev1.Secret{
		Data: map[string][]byte{
			"endpoint": []byte("http://minio:9000"),
			"bucket":   []byte("testbucket"),
		},
	}
	s3 := GetS3Params(storageSecret)
	assert.Equal(t, "minio:9000", s3.Endpoint)
	assert.True(t, s3.Insecure)
	assert.Equal(t, "testbucket", s3.Bucket)
}

func TestGetS3ParamsSecure(t *testing.T) {
	storageSecret := &corev1.Secret{
		Data: map[string][]byte{
			"endpoint": []byte("https://minio:9000"),
			"bucket":   []byte("testbucket"),
		},
	}
	s3 := GetS3Params(storageSecret)
	assert.Equal(t, "minio:9000", s3.Endpoint)
	assert.False(t, s3.Insecure)
	assert.Equal(t, "testbucket", s3.Bucket)
}
