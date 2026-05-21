package etcd

import (
	"testing"
)

func TestResolverBuilderReturnsBuilder(t *testing.T) {
	_ = ResolverBuilder
}

func TestResolverBuilderNilClient(t *testing.T) {
	b := ResolverBuilder(nil)
	_ = b
}
