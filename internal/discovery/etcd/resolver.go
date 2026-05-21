package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	etcdresolver "go.etcd.io/etcd/client/v3/naming/resolver"
	gresolver "google.golang.org/grpc/resolver"
)

// ResolverBuilder returns a gRPC resolver.Builder that resolves targets
// via etcd using the "etcd" scheme.
//
// Usage:
//
//	builder := etcd.ResolverBuilder(client)
//	conn, _ := grpc.NewClient("etcd:///kim-gate",
//	    grpc.WithResolvers(builder),
//	    grpc.WithTransportCredentials(insecure.NewCredentials()),
//	)
func ResolverBuilder(cli *clientv3.Client) gresolver.Builder {
	b, err := etcdresolver.NewBuilder(cli)
	if err != nil {
		return nil
	}
	return b
}
