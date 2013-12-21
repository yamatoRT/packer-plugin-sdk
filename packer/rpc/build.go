package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// An implementation of packer.Build where the build is actually executed
// over an RPC connection.
type build struct {
	client *rpc.Client
	mux    *MuxConn
}

// BuildServer wraps a packer.Build implementation and makes it exportable
// as part of a Golang RPC server.
type BuildServer struct {
	build packer.Build
	mux   *MuxConn
}

type BuildPrepareResponse struct {
	Warnings []string
	Error    error
}

func (b *build) Name() (result string) {
	b.client.Call("Build.Name", new(interface{}), &result)
	return
}

func (b *build) Prepare(v map[string]string) ([]string, error) {
	var resp BuildPrepareResponse
	if cerr := b.client.Call("Build.Prepare", v, &resp); cerr != nil {
		return nil, cerr
	}

	return resp.Warnings, resp.Error
}

func (b *build) Run(ui packer.Ui, cache packer.Cache) ([]packer.Artifact, error) {
	nextId := b.mux.NextId()
	server := newServerWithMux(b.mux, nextId)
	server.RegisterCache(cache)
	server.RegisterUi(ui)
	go server.Serve()

	var result []uint32
	if err := b.client.Call("Build.Run", nextId, &result); err != nil {
		return nil, err
	}

	artifacts := make([]packer.Artifact, len(result))
	for i, streamId := range result {
		client, err := newClientWithMux(b.mux, streamId)
		if err != nil {
			return nil, err
		}

		artifacts[i] = client.Artifact()
	}

	return artifacts, nil
}

func (b *build) SetDebug(val bool) {
	if err := b.client.Call("Build.SetDebug", val, new(interface{})); err != nil {
		panic(err)
	}
}

func (b *build) SetForce(val bool) {
	if err := b.client.Call("Build.SetForce", val, new(interface{})); err != nil {
		panic(err)
	}
}

func (b *build) Cancel() {
	if err := b.client.Call("Build.Cancel", new(interface{}), new(interface{})); err != nil {
		panic(err)
	}
}

func (b *BuildServer) Name(args *interface{}, reply *string) error {
	*reply = b.build.Name()
	return nil
}

func (b *BuildServer) Prepare(v map[string]string, resp *BuildPrepareResponse) error {
	warnings, err := b.build.Prepare(v)
	*resp = BuildPrepareResponse{
		Warnings: warnings,
		Error:    err,
	}
	return nil
}

func (b *BuildServer) Run(streamId uint32, reply *[]uint32) error {
	client, err := newClientWithMux(b.mux, streamId)
	if err != nil {
		return NewBasicError(err)
	}
	defer client.Close()

	artifacts, err := b.build.Run(client.Ui(), client.Cache())
	if err != nil {
		return NewBasicError(err)
	}

	*reply = make([]uint32, len(artifacts))
	for i, artifact := range artifacts {
		streamId := b.mux.NextId()
		server := newServerWithMux(b.mux, streamId)
		server.RegisterArtifact(artifact)
		go server.Serve()

		(*reply)[i] = streamId
	}

	return nil
}

func (b *BuildServer) SetDebug(val *bool, reply *interface{}) error {
	b.build.SetDebug(*val)
	return nil
}

func (b *BuildServer) SetForce(val *bool, reply *interface{}) error {
	b.build.SetForce(*val)
	return nil
}

func (b *BuildServer) Cancel(args *interface{}, reply *interface{}) error {
	b.build.Cancel()
	return nil
}